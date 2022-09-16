package deviceshifubase

import (
	"context"
	"errors"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"knative.dev/pkg/configmap"
)

// DeviceShifuConfig data under Configmap, Settings of deviceShifu
type DeviceShifuConfig struct {
	DriverProperties DeviceShifuDriverProperties
	Instructions     DeviceShifuInstructions
	Telemetries      *DeviceShifuTelemetries
}

// DeviceShifuDriverProperties properties of deviceshifuDriver
type DeviceShifuDriverProperties struct {
	DriverSku       string `yaml:"driverSku"`
	DriverImage     string `yaml:"driverImage"`
	DriverExecution string `yaml:"driverExecution"`
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
			klog.Fatalf("parsing %v from ConfigMap, error: %v", ConfigmapDriverPropertiesStr, err)
			return nil, err
		}
	}

	// TODO: add validation to types and readwrite mode
	if instructions, ok := cfg[ConfigmapInstructionsStr]; ok {
		err := yaml.Unmarshal([]byte(instructions), &dsc.Instructions)
		if err != nil {
			klog.Fatalf("parsing %v from ConfigMap, error: %v", ConfigmapInstructionsStr, err)
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
	return dsc, nil
}

// NewEdgeDevice new edgeDevice
func NewEdgeDevice(edgeDeviceConfig *EdgeDeviceConfig) (*v1alpha1.EdgeDevice, *rest.RESTClient, error) {
	var config *rest.Config
	var err error

	if edgeDeviceConfig.KubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", edgeDeviceConfig.KubeconfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		klog.Fatalf("Error parsing incluster/kubeconfig, error: %v", err.Error())
		return nil, nil, err
	}

	client, err := NewEdgeDeviceRestClient(config)
	if err != nil {
		klog.Fatalf("Error creating EdgeDevice custom REST client, error: %v", err.Error())
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
		klog.Fatalf("Error GET EdgeDevice resource, error: %v", err.Error())
		return nil, nil, err
	}
	return ed, client, nil
}

// NewEdgeDeviceRestClient new edgeDevice rest Client
func NewEdgeDeviceRestClient(config *rest.Config) (*rest.RESTClient, error) {
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
