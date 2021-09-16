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

var MOCK_DEVICE_CONFIG_FOLDER = path.Join("etc", "edgedevice", "config")

var mockDeviceDriverProperties = DeviceShifuDriverProperties{
	"Edgenesis Mock Device",
	"edgenesis/mockdevice-0.0.1",
}

var mockDeviceInstructions = map[string]*DeviceShifuInstruction{
	"get_reading": nil,
	"get_status":  nil,
	"set_reading": {
		[]DeviceShifuInstructionProperty{
			{
				"Int32",
				"W",
				nil,
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
				"get_status",
				1000,
				1000,
			},
		},
	},
	"device_random": {
		[]DeviceShifuTelemetryProperty{
			{
				"get_reading",
				1000,
				1000,
			},
		},
	},
}

func TestStart(t *testing.T) {
	err := GenerateConfigMapFromSnippet(MOCK_DEVICE_CM_STR, MOCK_DEVICE_CONFIG_FOLDER)
	if err != nil {
		t.Errorf(err.Error())
	}

	mockdsc, err := New(MOCK_DEVICE_CONFIG_FOLDER)
	if err != nil {
		t.Errorf(err.Error())
	}

	if mockdsc.driverProperties.DriverSku != mockDeviceDriverProperties.DriverSku {
		t.Errorf("Driver SKU does not match, config: %+v, testdata: %+v", mockdsc.driverProperties.DriverSku, mockDeviceDriverProperties.DriverSku)
	}

	if mockdsc.driverProperties.DriverImage != mockDeviceDriverProperties.DriverImage {
		t.Errorf("Driver Image does not match, config: %+v, testdata: %+v", mockdsc.driverProperties.DriverImage, mockDeviceDriverProperties.DriverImage)
	}

	if len(mockdsc.Instructions) != len(mockDeviceInstructions) {
		t.Errorf("instruction length mismatch!")
	}

	if len(mockdsc.Telemetries) != len(mockDeviceTelemetries) {
		t.Errorf("telemetry length mismatch!")
	}

	eq := reflect.DeepEqual(mockDeviceInstructions, mockdsc.Instructions)
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
