package deviceshifuconfig

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

const (
	MOCK_DEVICE_CM_STR               = "configmap_snippet.yaml"
	MOCK_DEVICE_WRITEFILE_PERMISSION = 0644
)

type ConfigMapData struct {
	Data struct {
		DriverProperties string `yaml:"driverProperties"`
		Instructions     string `yaml:"instructions"`
		Telemetries      string `yaml:"telemetries"`
	} `yaml:"data"`
}

var InstructionValueTypeInt32 InstructionValueType = "Int32"
var InstructionReadWriteW InstructionReadWrite = "W"
var TelemetryInstructionNameGetStatus TelemetryInstructionName = "get_status"
var TelemetryInstructionNameGetReading TelemetryInstructionName = "get_reading"
var TelemetryInitialDelayMs1000 TelemetryInitialDelayMs = 1000
var TelemetryIntervalMs1000 TelemetryIntervalMs = 1000
var MOCK_DEVICE_CONFIG_FOLDER = path.Join("etc", "edgedevice", "config")

var mockDeviceDriverProperties = DeviceShifuDriverProperties{
	"Edgenesis Mock Device",
	"edgenesis/mockdevice:0.0.1",
}

var mockDeviceInstructions = map[string]*DeviceShifuInstruction{
	"get_reading": nil,
	"get_status":  nil,
	"set_reading": {
		[]DeviceShifuInstructionProperty{
			{
				ValueType:    &InstructionValueTypeInt32,
				ReadWrite:    &InstructionReadWriteW,
				DefaultValue: nil,
			},
		},
	},
	"start": nil,
	"stop":  nil,
}

var mockDeviceTelemetries = map[string]*DeviceShifuTelemetry{
	"device_health": {
		[]DeviceShifuTelemetryProperty{
			{
				InstructionName: &TelemetryInstructionNameGetStatus,
				InitialDelayMs:  &TelemetryInitialDelayMs1000,
				IntervalMs:      &TelemetryIntervalMs1000,
			},
		},
	},
	"device_random": {
		[]DeviceShifuTelemetryProperty{
			{
				InstructionName: &TelemetryInstructionNameGetReading,
				InitialDelayMs:  &TelemetryInitialDelayMs1000,
				IntervalMs:      &TelemetryIntervalMs1000,
			},
		},
	},
}

func TestNew(t *testing.T) {
	err := GenerateConfigMapFromSnippet(MOCK_DEVICE_CM_STR, MOCK_DEVICE_CONFIG_FOLDER)
	if err != nil {
		t.Errorf(err.Error())
	}

	mockdsc, err := New(MOCK_DEVICE_CONFIG_FOLDER)
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
