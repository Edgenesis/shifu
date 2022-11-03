package deviceshifubase

import (
	"context"
	"errors"

	"k8s.io/klog/v2"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/imdario/mergo"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/configmap"
)

// DeviceShifuConfig data under Configmap, Settings of deviceShifu
type DeviceShifuConfig struct {
	DriverProperties         DeviceShifuDriverProperties
	Instructions             DeviceShifuInstructions
	Telemetries              *DeviceShifuTelemetries
	CustomInstructionsPython map[string]string `yaml:"customInstructionsPython"`
}

// DeviceShifuDriverProperties properties of deviceshifuDriver
type DeviceShifuDriverProperties struct {
	DriverSku       string `yaml:"driverSku"`
	DriverImage     string `yaml:"driverImage"`
	DriverExecution string `yaml:"driverExecution,omitempty"`
}

// DeviceShifuInstructions Instructions of devicehsifu
type DeviceShifuInstructions struct {
	Instructions        map[string]*DeviceShifuInstruction `yaml:"instructions"`
	InstructionSettings *DeviceShifuInstructionSettings    `yaml:"instructionSettings,omitempty"`
}

// DeviceShifuInstructionSettings Settings of all instructions
type DeviceShifuInstructionSettings struct {
	DefaultTimeoutSeconds *int `yaml:"defaultTimeoutSeconds,omitempty"`
}

// DeviceShifuInstruction Instruction of deviceshifu
type DeviceShifuInstruction struct {
	DeviceShifuInstructionProperties []DeviceShifuInstructionProperty `yaml:"argumentPropertyList,omitempty"`
	DeviceShifuProtocolProperties    map[string]string                `yaml:"protocolPropertyList,omitempty"`
}

// DeviceShifuInstructionProperty property of instruction
type DeviceShifuInstructionProperty struct {
	ValueType    string      `yaml:"valueType"`
	ReadWrite    string      `yaml:"readWrite"`
	DefaultValue interface{} `yaml:"defaultValue"`
}

// DeviceShifuTelemetryPushSettings settings of push under telemetry
type DeviceShifuTelemetryPushSettings struct {
	DeviceShifuTelemetryCollectionService *string `yaml:"telemetryCollectionService,omitempty"`
	DeviceShifuTelemetryPushToServer      *bool   `yaml:"pushToServer,omitempty"`
}

// DeviceShifuTelemetryProperties properties of Telemetry
type DeviceShifuTelemetryProperties struct {
	DeviceInstructionName *string                           `yaml:"instruction"`
	InitialDelayMs        *int                              `yaml:"initialDelayMs,omitempty"`
	IntervalMs            *int                              `yaml:"intervalMs,omitempty"`
	PushSettings          *DeviceShifuTelemetryPushSettings `yaml:"pushSettings,omitempty"`
}

// DeviceShifuTelemetrySettings settings of Telemetry
type DeviceShifuTelemetrySettings struct {
	DeviceShifuTelemetryUpdateIntervalInMilliseconds *int64  `yaml:"telemetryUpdateIntervalInMilliseconds,omitempty"`
	DeviceShifuTelemetryTimeoutInMilliseconds        *int64  `yaml:"telemetryTimeoutInMilliseconds,omitempty"`
	DeviceShifuTelemetryInitialDelayInMilliseconds   *int64  `yaml:"telemetryInitialDelayInMilliseconds,omitempty"`
	DeviceShifuTelemetryDefaultPushToServer          *bool   `yaml:"defaultPushToServer,omitempty"`
	DeviceShifuTelemetryDefaultCollectionService     *string `yaml:"defaultTelemetryCollectionService,omitempty"`
}

// DeviceShifuTelemetries Telemetries of deviceshifu
type DeviceShifuTelemetries struct {
	DeviceShifuTelemetrySettings *DeviceShifuTelemetrySettings    `yaml:"telemetrySettings,omitempty"`
	DeviceShifuTelemetries       map[string]*DeviceShifuTelemetry `yaml:"telemetries,omitempty"`
}

// DeviceShifuTelemetry properties of telemetry
type DeviceShifuTelemetry struct {
	DeviceShifuTelemetryProperties DeviceShifuTelemetryProperties `yaml:"properties,omitempty"`
}

// EdgeDeviceConfig config of EdgeDevice
type EdgeDeviceConfig struct {
	NameSpace      string
	DeviceName     string
	KubeconfigPath string
}

// NewDeviceShifuConfig Read the configuration under the path directory and return configuration
func NewDeviceShifuConfig(path string) (*DeviceShifuConfig, error) {
	if path == "" {
		return nil, errors.New("DeviceShifuConfig path can't be empty")
	}

	cfg, err := configmap.Load(path)
	if err != nil {
		return nil, err
	}

	dsc := &DeviceShifuConfig{}
	if driverProperties, ok := cfg[ConfigmapDriverPropertiesStr]; ok {
		err := yaml.Unmarshal([]byte(driverProperties), &dsc.DriverProperties)
		if err != nil {
			klog.Fatalf("Error parsing %v from ConfigMap, error: %v", ConfigmapDriverPropertiesStr, err)
			return nil, err
		}
	}

	// TODO: add validation to types and readwrite mode
	if instructions, ok := cfg[ConfigmapInstructionsStr]; ok {
		err := yaml.Unmarshal([]byte(instructions), &dsc.Instructions)
		if err != nil {
			klog.Fatalf("Error parsing %v from ConfigMap, error: %v", ConfigmapInstructionsStr, err)
			return nil, err
		}
	}

	if telemetries, ok := cfg[ConfigmapTelemetriesStr]; ok {
		err = yaml.Unmarshal([]byte(telemetries), &dsc.Telemetries)
		if err != nil {
			klog.Fatalf("Error parsing %v from ConfigMap, error: %v", ConfigmapTelemetriesStr, err)
			return nil, err
		}
	}

	if customInstructionsPython, ok := cfg[ConfigmapCustomizedInstructionsStr]; ok {
		err = yaml.Unmarshal([]byte(customInstructionsPython), &dsc.CustomInstructionsPython)
		if err != nil {
			klog.Fatalf("Error parsing %v from ConfigMap, error: %v", ConfigmapCustomizedInstructionsStr, err)
			return nil, err
		}
	}

	err = dsc.init()
	return dsc, err
}

// NewEdgeDevice new edgeDevice
func NewEdgeDevice(edgeDeviceConfig *EdgeDeviceConfig) (*v1alpha1.EdgeDevice, *rest.RESTClient, error) {
	config, err := getRestConfig(edgeDeviceConfig.KubeconfigPath)
	if err != nil {
		klog.Errorf("Error parsing incluster/kubeconfig, error: %v", err.Error())
		return nil, nil, err
	}

	client, err := newEdgeDeviceRestClient(config)
	if err != nil {
		klog.Errorf("Error creating EdgeDevice custom REST client, error: %v", err.Error())
		return nil, nil, err
	}
	ed := &v1alpha1.EdgeDevice{}
	err = client.Get().
		Namespace(edgeDeviceConfig.NameSpace).
		Resource(EdgedeviceResourceStr).
		Name(edgeDeviceConfig.DeviceName).
		Do(context.TODO()).
		Into(ed)
	if err != nil {
		klog.Errorf("Error GET EdgeDevice resource, error: %v", err.Error())
		return nil, nil, err
	}
	return ed, client, nil
}

func getRestConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	} else {
		return rest.InClusterConfig()
	}
}

// newEdgeDeviceRestClient new edgeDevice rest Client
func newEdgeDeviceRestClient(config *rest.Config) (*rest.RESTClient, error) {
	err := v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Errorf("cannot add to scheme, error: %v", err)
		return nil, err
	}
	crdConfig := config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()
	exampleRestClient, err := rest.UnversionedRESTClientFor(crdConfig)
	if err != nil {
		return nil, err
	}

	return exampleRestClient, nil
}

// init DeviceShifuConfig With default
func (dsc *DeviceShifuConfig) init() error {
	if err := dsc.DriverProperties.init(); err != nil {
		klog.Errorf("Error to init DriverProperties, error %s", err.Error())
		return err
	}

	if dsc.Telemetries == nil {
		dsc.Telemetries = &DeviceShifuTelemetries{}
	}
	if err := dsc.Telemetries.init(); err != nil {
		klog.Errorf("Error to init Telemetries, error %s", err.Error())
		return err
	}

	if err := dsc.Instructions.init(); err != nil {
		klog.Errorf("Error to init Instructions, error %s", err.Error())
		return err
	}

	return nil
}

func (dsdp *DeviceShifuDriverProperties) init() error {
	defaultProperties := &DeviceShifuDriverProperties{
		DriverSku:       "defaultSku",
		DriverImage:     "defaultImage",
		DriverExecution: "defaultExecution",
	}
	return mergo.Merge(dsdp, defaultProperties)
}

func (dsis *DeviceShifuInstructions) init() error {
	if dsis.Instructions == nil {
		dsis.Instructions = map[string]*DeviceShifuInstruction{}
	}

	if dsis.InstructionSettings == nil {
		dsis.InstructionSettings = &DeviceShifuInstructionSettings{}
	}

	return dsis.InstructionSettings.init()
}

func (dsiss *DeviceShifuInstructionSettings) init() error {
	var (
		defaultTimeoutSeconds = DeviceDefaultGlobalTimeoutInSeconds
	)

	defaultDeviceshifuInstructionSettings := &DeviceShifuInstructionSettings{
		DefaultTimeoutSeconds: &defaultTimeoutSeconds,
	}

	return mergo.Merge(dsiss, defaultDeviceshifuInstructionSettings)
}

func (dsts *DeviceShifuTelemetries) init() error {
	if dsts.DeviceShifuTelemetries == nil {
		dsts.DeviceShifuTelemetries = map[string]*DeviceShifuTelemetry{}
	}
	for id := range dsts.DeviceShifuTelemetries {
		if dsts.DeviceShifuTelemetries[id] == nil {
			dsts.DeviceShifuTelemetries[id] = &DeviceShifuTelemetry{}
		}
		err := dsts.DeviceShifuTelemetries[id].init()
		if err != nil {
			klog.Errorf("Error to init telemetry, error %s", err.Error())
			return err
		}
	}

	if dsts.DeviceShifuTelemetrySettings == nil {
		dsts.DeviceShifuTelemetrySettings = &DeviceShifuTelemetrySettings{}
	}
	return dsts.DeviceShifuTelemetrySettings.init()
}

func (dst *DeviceShifuTelemetry) init() error {
	var (
		defaultInitialDelay = DeviceInstructionInitialDelay
	)

	defaultDeviceShifuTelemetry := &DeviceShifuTelemetry{
		DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
			InitialDelayMs: &defaultInitialDelay,
			IntervalMs:     &defaultInitialDelay,
		},
	}

	return mergo.Merge(dst, defaultDeviceShifuTelemetry)
}

func (dsts *DeviceShifuTelemetrySettings) init() error {
	var (
		defaultUpdateInterval = DeviceDefaultTelemetryUpdateIntervalInMS
		defaultTimeout        = DeviceTelemetryTimeoutInMS
		defaultInitialDelay   = DeviceTelemetryInitialDelayInMS
		defaultPushToServer   = false
	)
	defaultDeviceshifuTelemtrySettings := DeviceShifuTelemetrySettings{
		DeviceShifuTelemetryUpdateIntervalInMilliseconds: &defaultUpdateInterval,
		DeviceShifuTelemetryTimeoutInMilliseconds:        &defaultTimeout,
		DeviceShifuTelemetryInitialDelayInMilliseconds:   &defaultInitialDelay,
		DeviceShifuTelemetryDefaultPushToServer:          &defaultPushToServer,
	}
	return mergo.Merge(dsts, defaultDeviceshifuTelemtrySettings)
}
