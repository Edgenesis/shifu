package deviceshifubase

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"gopkg.in/yaml.v3"
)

// Str and default value
const (
	MockDeviceCmStr              = "configmap_snippet.yaml"
	MockDeviceWritFilePermission = 0644
	MockDeviceConfigPath         = "etc"
)

var MockDeviceConfigFolder = path.Join("etc", "edgedevice", "config")

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

	var DriverProperties = DeviceShifuDriverProperties{
		DriverSku:       "Edgenesis Mock Device",
		DriverImage:     "edgenesis/mockdevice:v0.0.1",
		DriverExecution: "python mock_driver.py",
	}

	var mockDeviceInstructions = map[string]*DeviceShifuInstruction{
		"get_reading": nil,
		"get_status":  nil,
		"set_reading": {
			DeviceShifuInstructionProperties: []DeviceShifuInstructionProperty{
				{
					ValueType:    InstructionValueTypeInt32,
					ReadWrite:    InstructionReadWriteW,
					DefaultValue: nil,
				},
			},
			DeviceShifuProtocolProperties: nil,
		},
		"start": nil,
		"stop":  nil,
	}

	var mockDeviceTelemetries = &DeviceShifuTelemetries{
		DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
			DeviceShifuTelemetryUpdateIntervalInMilliseconds: &TelemetrySettingInterval,
		},
	}

	mockdsc, err := NewDeviceShifuConfig(MockDeviceConfigFolder)
	if err != nil {
		t.Errorf(err.Error())
	}

	eq := reflect.DeepEqual(DriverProperties, mockdsc.DriverProperties)
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
	snippetFile, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	var cmData ConfigMapData
	err = yaml.Unmarshal(snippetFile, &cmData)
	if err != nil {
		klog.Fatalf("Error parsing ConfigMap %v, error: %v", fileName, err)
		return err
	}

	var MockDeviceConfigMapping = map[string]string{
		path.Join(MockDeviceConfigFolder, ConfigmapDriverPropertiesStr): cmData.Data.DriverProperties,
		path.Join(MockDeviceConfigFolder, ConfigmapInstructionsStr):     cmData.Data.Instructions,
		path.Join(MockDeviceConfigFolder, ConfigmapTelemetriesStr):      cmData.Data.Telemetries,
	}

	err = os.MkdirAll(MockDeviceConfigFolder, os.ModePerm)
	if err != nil {
		klog.Fatalf("Error creating path for: %v", MockDeviceConfigFolder)
		return err
	}

	for outputDir, data := range MockDeviceConfigMapping {
		err = os.WriteFile(outputDir, []byte(data), MockDeviceWritFilePermission)
		if err != nil {
			klog.Fatalf("Error creating configFile for: %v", outputDir)
			return err
		}
	}
	return nil
}

func Test_createDevice(t *testing.T) {
	testCases := []struct {
		Name      string
		path      string
		expErrStr string
	}{
		{
			"case 1 have empty kubepath get config failure",
			"",
			"unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined",
		},
		{
			"case 2 use kubepath get config failure",
			"kubepath",
			"stat kubepath: no such file or directory",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			config, err := getRestConfig(c.path)
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
				assert.Nil(t, config)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_getRestConfig(t *testing.T) {
	testCases := []struct {
		Name      string
		path      string
		expErrStr string
	}{
		{
			"case 1 have empty kubepath get config failure",
			"",
			"unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined",
		},
		{
			"case 2 use kubepath get config failure",
			"kubepath",
			"stat kubepath: no such file or directory",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			config, err := getRestConfig(c.path)
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
				assert.Nil(t, config)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_newEdgeDeviceRestClient(t *testing.T) {
	testCases := []struct {
		Name      string
		config    *rest.Config
		expResult string
		expErrStr string
	}{
		{
			"case 1 can generate client with empty config",
			&rest.Config{},
			"v1alpha1",
			"",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			res, err := newEdgeDeviceRestClient(c.config)
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, res.APIVersion().Version, c.expResult)
		})
	}
}
