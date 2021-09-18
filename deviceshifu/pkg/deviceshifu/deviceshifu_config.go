package deviceshifu

import (
	"context"
	"errors"
	"log"

	v1alpha1 "github.com/Edgenesis/shifu/k8s/crd/api/v1alpha1"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/configmap"
)

type DeviceShifuConfig struct {
	driverProperties DeviceShifuDriverProperties
	Instructions     map[string]*DeviceShifuInstruction
	Telemetries      map[string]*DeviceShifuTelemetry
}

type DeviceShifuDriverProperties struct {
	DriverSku   string `yaml:"driverSku"`
	DriverImage string `yaml:"driverImage"`
}

type DeviceShifuInstruction struct {
	DeviceShifuInstructionProperties []DeviceShifuInstructionProperty `yaml:"properties,omitempty"`
}

type DeviceShifuInstructionProperty struct {
	ValueType    string      `yaml:"valueType"`
	ReadWrite    string      `yaml:"readWrite"`
	DefaultValue interface{} `yaml:"defaultValue"`
}

type DeviceShifuTelemetry struct {
	DeviceShifuTelemetryProperties []DeviceShifuTelemetryProperty `yaml:"properties,omitempty"`
}

type DeviceShifuTelemetryProperty struct {
	DeviceInstructionName *string `yaml:"instruction"`
	InitialDelayMs        *int    `yaml:"initialDelayMs,omitempty"`
	IntervalMs            *int    `yaml:"intervalMs,omitempty"`
}

const (
	CM_DRIVERPROPERTIES_STR = "driverProperties"
	CM_INSTRUCTIONS_STR     = "instructions"
	CM_TELEMETRIES_STR      = "telemetries"
)

func NewDeviceShifuConfig(path string) (*DeviceShifuConfig, error) {
	if path == "" {
		return nil, errors.New("DeviceShifuConfig path can't be empty")
	}

	cfg, err := configmap.Load(path)
	if err != nil {
		return nil, err
	}

	dsc := &DeviceShifuConfig{}
	if driverProperties, ok := cfg[CM_DRIVERPROPERTIES_STR]; ok {
		err := yaml.Unmarshal([]byte(driverProperties), &dsc.driverProperties)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_DRIVERPROPERTIES_STR, err)
			return nil, err
		}
	}

	// TODO: add validation to types and readwrite mode
	if instructions, ok := cfg[CM_INSTRUCTIONS_STR]; ok {
		err := yaml.Unmarshal([]byte(instructions), &dsc.Instructions)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_INSTRUCTIONS_STR, err)
			return nil, err
		}
	}

	if telemetries, ok := cfg[CM_TELEMETRIES_STR]; ok {
		err = yaml.Unmarshal([]byte(telemetries), &dsc.Telemetries)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_TELEMETRIES_STR, err)
			return nil, err
		}
	}
	return dsc, nil
}

func NewEdgeDeviceConfig(nameSpace string, deviceName string, kubeconfigPath string) (*v1alpha1.EdgeDevice, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Fatalf("Error parsing incluster/kubeconfig, error: %v", err.Error())
		return nil, err
	}

	client, err := NewEdgeDeviceRestClient(config)
	if err != nil {
		log.Fatalf("Error creating EdgeDevice custom REST client, error: %v", err.Error())
		return nil, err
	}

	result := &v1alpha1.EdgeDevice{}
	err = client.Get().Namespace(nameSpace).Resource("edgedevices").Name(deviceName).Do(context.TODO()).Into(result)
	if err != nil {
		log.Fatalf("Error GET EdgeDevice resource, error: %v", err.Error())
		return nil, err
	}

	return result, nil
}

func NewEdgeDeviceRestClient(config *rest.Config) (*rest.RESTClient, error) {
	v1alpha1.AddToScheme(scheme.Scheme)
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()
	exampleRestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		return nil, err
	}

	return exampleRestClient, nil
}
