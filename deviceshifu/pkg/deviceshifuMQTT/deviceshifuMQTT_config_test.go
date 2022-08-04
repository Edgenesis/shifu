package deviceshifuMQTT

import (
	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifubase"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	MOCK_DEVICE_CM_STR               = "configmap_snippet.yaml"
	MOCK_DEVICE_WRITEFILE_PERMISSION = 0644
	MOCK_DEVICE_CONFIG_PATH          = "etc"
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
		InstructionValueTypeInt32       = "Int32"
		InstructionReadWriteW           = "W"
		TelemetrySettingInterval  int64 = 1000
	)

	var mockDeviceDriverProperties = deviceshifubase.DeviceShifuDriverProperties{
		"Edgenesis Mock Device",
		"edgenesis/mockdevice:v0.0.1",
		"python mock_driver.py",
	}

	var mockDeviceInstructions = map[string]*deviceshifubase.DeviceShifuInstruction{
		"get_reading": nil,
		"get_status":  nil,
		"set_reading": {
			[]deviceshifubase.DeviceShifuInstructionProperty{
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

	var mockDeviceTelemetries = &deviceshifubase.DeviceShifuTelemetries{
		DeviceShifuTelemetrySettings: &deviceshifubase.DeviceShifuTelemetrySettings{
			DeviceShifuTelemetryUpdateIntervalInMilliseconds: &TelemetrySettingInterval,
		},
	}

	err := GenerateConfigMapFromSnippet(MOCK_DEVICE_CM_STR, MOCK_DEVICE_CONFIG_FOLDER)
	if err != nil {
		t.Errorf(err.Error())
	}

	mockdsc, err := deviceshifubase.NewDeviceShifuConfig(MOCK_DEVICE_CONFIG_FOLDER)
	if err != nil {
		t.Errorf(err.Error())
	}

	eq := reflect.DeepEqual(mockDeviceDriverProperties, mockdsc.DriverProperties)
	if !eq {
		t.Errorf("DriverProperties mismatch")
	}

	eq = reflect.DeepEqual(mockDeviceInstructions, mockdsc.Instructions.Instructions)
	if !eq {
		t.Errorf("Instruction mismatch")
	}

	eq = reflect.DeepEqual(mockDeviceTelemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds,
		mockdsc.Telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds)
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
		path.Join(MOCK_DEVICE_CONFIG_FOLDER, deviceshifubase.CM_DRIVERPROPERTIES_STR): cmData.Data.DriverProperties,
		path.Join(MOCK_DEVICE_CONFIG_FOLDER, deviceshifubase.CM_INSTRUCTIONS_STR):     cmData.Data.Instructions,
		path.Join(MOCK_DEVICE_CONFIG_FOLDER, deviceshifubase.CM_TELEMETRIES_STR):      cmData.Data.Telemetries,
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
