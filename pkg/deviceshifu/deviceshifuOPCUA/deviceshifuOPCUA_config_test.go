package deviceshifuOPCUA

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
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
		InstructionNameGetValue               = "get_value"
		InstructionNameGetTime                = "get_time"
		InstructionNameGetServerVersion       = "get_server"
		InstructionNodeIDValue                = "ns=2;i=2"
		InstructionNodeIDTime                 = "i=2258"
		InstructionNodeIDServerVersion        = "i=2261"
		TelemetryMs1000                       = 1000
		TelemetryMs1000Int64            int64 = 1000
	)

	var mockDeviceDriverProperties = deviceshifubase.DeviceShifuDriverProperties{
		DriverSku:       "Edgenesis Mock Device",
		DriverImage:     "edgenesis/mockdevice:v0.0.1",
		DriverExecution: "python mock_driver.py",
	}

	var mockDeviceInstructions = map[string]*OPCUAInstruction{
		InstructionNameGetValue: {
			&OPCUAInstructionProperty{
				OPCUANodeID: InstructionNodeIDValue,
			},
		},
		InstructionNameGetTime: {
			&OPCUAInstructionProperty{
				OPCUANodeID: InstructionNodeIDTime,
			},
		},
		InstructionNameGetServerVersion: {
			&OPCUAInstructionProperty{
				OPCUANodeID: InstructionNodeIDServerVersion,
			},
		},
	}

	var mockDeviceTelemetries = &deviceshifubase.DeviceShifuTelemetries{
		DeviceShifuTelemetrySettings: &deviceshifubase.DeviceShifuTelemetrySettings{
			DeviceShifuTelemetryUpdateIntervalInMilliseconds: &TelemetryMs1000Int64,
		},
		DeviceShifuTelemetries: map[string]*deviceshifubase.DeviceShifuTelemetry{
			"device_health": {
				deviceshifubase.DeviceShifuTelemetryProperties{
					DeviceInstructionName: &InstructionNameGetServerVersion,
					InitialDelayMs:        &TelemetryMs1000,
					IntervalMs:            &TelemetryMs1000,
				},
			},
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

	mockintructions := CreateOPCUAInstructions(&mockdsc.Instructions)

	eq := reflect.DeepEqual(mockDeviceDriverProperties, mockdsc.DriverProperties)
	if !eq {
		t.Errorf("DriverProperties mismatch")
	}

	eq = reflect.DeepEqual(mockDeviceInstructions, mockintructions.Instructions)
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
