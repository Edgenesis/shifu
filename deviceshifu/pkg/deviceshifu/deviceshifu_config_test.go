package deviceshifu

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"testing"

	"edgenesis.io/shifu/k8s/crd/api/v1alpha1"
	"gopkg.in/yaml.v2"
)

const (
	MOCK_DEVICE_CM_STR               = "configmap_snippet.yaml"
	MOCK_DEVICE_WRITEFILE_PERMISSION = 0644
)

var MOCK_DEVICE_CONFIG_FOLDER = path.Join("etc", "edgedevice", "config")

type ConfigMapData struct {
	Data struct {
		DriverProperties string `yaml:"driverProperties"`
		Instructions     string `yaml:"instructions"`
		Telemetries      string `yaml:"telemetries"`
	} `yaml:"data"`
}

func TestNewDeviceShifuConfig(t *testing.T) {
	var (
		TelemetryInstructionNameGetStatus  string = "get_status"
		TelemetryInstructionNameGetReading string = "get_reading"
		InstructionValueTypeInt32          string = "Int32"
		InstructionReadWriteW              string = "W"
		TelemetryMs1000                    int    = 1000
	)

	var mockDeviceDriverProperties = DeviceShifuDriverProperties{
		"Edgenesis Mock Device",
		"edgenesis/mockdevice:v0.0.1",
	}

	var mockDeviceInstructions = map[string]*DeviceShifuInstruction{
		"get_reading": nil,
		"get_status":  nil,
		"set_reading": {
			[]DeviceShifuInstructionProperty{
				{
					ValueType:    InstructionValueTypeInt32,
					ReadWrite:    InstructionReadWriteW,
					DefaultValue: nil,
				},
			},
		},
		"start": nil,
		"stop":  nil,
	}

	var mockDeviceTelemetries = map[string]*DeviceShifuTelemetry{
		"device_health": {
			DeviceShifuTelemetryProperties{
				DeviceInstructionName: &TelemetryInstructionNameGetStatus,
				InitialDelayMs:        &TelemetryMs1000,
				IntervalMs:            &TelemetryMs1000,
			},
		},
		"get_reading": {
			DeviceShifuTelemetryProperties{
				DeviceInstructionName: &TelemetryInstructionNameGetReading,
				InitialDelayMs:        &TelemetryMs1000,
				IntervalMs:            &TelemetryMs1000,
			},
		},
	}

	err := GenerateConfigMapFromSnippet(MOCK_DEVICE_CM_STR, MOCK_DEVICE_CONFIG_FOLDER)
	if err != nil {
		t.Errorf(err.Error())
	}

	mockdsc, err := NewDeviceShifuConfig(MOCK_DEVICE_CONFIG_FOLDER)
	if err != nil {
		t.Errorf(err.Error())
	}

	eq := reflect.DeepEqual(mockDeviceDriverProperties, mockdsc.driverProperties)
	if !eq {
		t.Errorf("DriverProperties mismatch")
	}

	eq = reflect.DeepEqual(mockDeviceInstructions, mockdsc.Instructions)
	if !eq {
		t.Errorf("Instruction mismatch")
	}

	eq = reflect.DeepEqual(mockDeviceTelemetries, mockdsc.Telemetries)
	if !eq {
		t.Errorf("Telemetries mismatch")
	}
}

func TestNewEdgeDeviceConfig(t *testing.T) {
	var (
		kubeconfigPath                                          = "/root/.kube/config"
		deviceName                                              = "edgedevice-sample"
		nameSpace                                               = "crd-system"
		EDGEDEVICE_MOCK_AGV_SPEC_SKU                            = "AGV"
		EDGEDEVICE_MOCK_AGV_SPEC_CONNECTION v1alpha1.Connection = "Ethernet"
		EDGEDEVICE_MOCK_AGV_SPEC_ADDRESS                        = "10.0.0.2:80"
		EDGEDEVICE_MOCK_AGV_SPEC_PROTOCOL   v1alpha1.Protocol   = "HTTP"
	)

	edgeDeviceConfig := &EdgeDeviceConfig{
		nameSpace,
		deviceName,
		kubeconfigPath,
	}

	edgeDevice, _, err := NewEdgeDevice(edgeDeviceConfig)
	if err != nil {
		t.Errorf(err.Error())
	}

	if edgeDevice.Spec.Sku != nil && *edgeDevice.Spec.Sku != EDGEDEVICE_MOCK_AGV_SPEC_SKU {
		t.Errorf("Wrong SKU for edgedevice-simple, should be: %v, actual: %v", EDGEDEVICE_MOCK_AGV_SPEC_SKU, *edgeDevice.Spec.Sku)
	}

	if edgeDevice.Spec.Connection != nil && *edgeDevice.Spec.Connection != EDGEDEVICE_MOCK_AGV_SPEC_CONNECTION {
		t.Errorf("Wrong SKU for edgedevice-simple, should be: %v, actual: %v", EDGEDEVICE_MOCK_AGV_SPEC_CONNECTION, *edgeDevice.Spec.Sku)
	}

	if edgeDevice.Spec.Address != nil && *edgeDevice.Spec.Address != EDGEDEVICE_MOCK_AGV_SPEC_ADDRESS {
		t.Errorf("Wrong SKU for edgedevice-simple, should be: %v, actual: %v", EDGEDEVICE_MOCK_AGV_SPEC_ADDRESS, *edgeDevice.Spec.Sku)
	}

	if edgeDevice.Spec.Protocol != nil && *edgeDevice.Spec.Protocol != EDGEDEVICE_MOCK_AGV_SPEC_PROTOCOL {
		t.Errorf("Wrong SKU for edgedevice-simple, should be: %v, actual: %v", EDGEDEVICE_MOCK_AGV_SPEC_PROTOCOL, *edgeDevice.Spec.Sku)
	}
}

func GenerateConfigMapFromSnippet(fileName string, folder string) error {
	snippetFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	var cmData ConfigMapData
	err = yaml.Unmarshal(snippetFile, &cmData)
	if err != nil {
		log.Fatalf("Error parsing ConfigMap %v, error: %v", fileName, err)
		return err
	}

	var MOCK_DEVICE_CONFIG_MAPPING = map[string]string{
		path.Join(MOCK_DEVICE_CONFIG_FOLDER, CM_DRIVERPROPERTIES_STR): cmData.Data.DriverProperties,
		path.Join(MOCK_DEVICE_CONFIG_FOLDER, CM_INSTRUCTIONS_STR):     cmData.Data.Instructions,
		path.Join(MOCK_DEVICE_CONFIG_FOLDER, CM_TELEMETRIES_STR):      cmData.Data.Telemetries,
	}

	err = os.MkdirAll(MOCK_DEVICE_CONFIG_FOLDER, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating path for: %v", MOCK_DEVICE_CONFIG_FOLDER)
		return err
	}

	for outputDir, data := range MOCK_DEVICE_CONFIG_MAPPING {
		err = os.WriteFile(outputDir, []byte(data), MOCK_DEVICE_WRITEFILE_PERMISSION)
		if err != nil {
			log.Fatalf("Error creating configFile for: %v", outputDir)
			return err
		}
	}
	return nil
}
