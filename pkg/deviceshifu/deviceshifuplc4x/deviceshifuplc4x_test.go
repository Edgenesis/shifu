package deviceshifuplc4x

import (
	plc4go "github.com/apache/plc4x/plc4go/pkg/api"
	"github.com/apache/plc4x/plc4go/pkg/api/drivers"
	"github.com/apache/plc4x/plc4go/pkg/api/transports"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"testing"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
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

func TestMain(m *testing.M) {
	err := GenerateConfigMapFromSnippet(MockDeviceCmStr, MockDeviceConfigFolder)
	if err != nil {
		klog.Errorf("error when generateConfigmapFromSnippet,err: %v", err)
		os.Exit(-1)
	}
	m.Run()
	err = os.RemoveAll(MockDeviceConfigPath)
	if err != nil {
		klog.Fatal(err)
	}
}

func TestStart(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestStart",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}

	if err := mockds.Stop(); err != nil {
		t.Errorf("unable to stop mock deviceShifu, error: %+v", err)
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
		path.Join(MockDeviceConfigFolder, deviceshifubase.ConfigmapDriverPropertiesStr): cmData.Data.DriverProperties,
		path.Join(MockDeviceConfigFolder, deviceshifubase.ConfigmapInstructionsStr):     cmData.Data.Instructions,
		path.Join(MockDeviceConfigFolder, deviceshifubase.ConfigmapTelemetriesStr):      cmData.Data.Telemetries,
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

func TestCollectPLC4XTelemetry(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestCollectPLC4XTelemetry",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}
	mockds.base.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds = unitest.ToPointer(int64(1))

	// Create a new instance of the PlcDriverManager
	driverManager := plc4go.NewPlcDriverManager()

	transports.RegisterTcpTransport(driverManager)

	drivers.RegisterModbusTcpDriver(driverManager)
	connectionRequestChanel := driverManager.GetConnection("modbus-tcp://192.168.23.30?unit-identifier=1")
	connectionResult := <-connectionRequestChanel
	mockds.conn = &connectionResult

	testCases := []struct {
		Name        string
		inputDevice *DeviceShifu
		expected    bool
		expErrStr   string
	}{
		{
			"case 1 Error sending the request",
			mockds,
			false,
			"Error sending the request: error sending request: error writing to transport. No writer available",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := c.inputDevice.collectPLC4XTelemetry()
			assert.Equal(t, c.expected, got)
			if got {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, c.expErrStr, err.Error())
			}
		})
	}
}
