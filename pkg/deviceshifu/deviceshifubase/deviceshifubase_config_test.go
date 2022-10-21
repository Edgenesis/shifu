package deviceshifubase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"

	"gopkg.in/yaml.v3"
)

// Str and default value
const (
	MockDeviceCmStr              = "configmap_snippet.yaml"
	MockDeviceWritFilePermission = 0644
	MockDeviceConfigPath         = "etc"
	MockConfigFile               = "mockconfig"
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
			//"stat kubepath: no such file or directory",
			"CreateFile kubepath: The system cannot find the file specified.",
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

func TestNewEdgeDevice(t *testing.T) {

	ms := mockHttpServer(t)
	defer ms.Close()

	writeMockConfigFile(t, ms.URL)

	testCases := []struct {
		Name      string
		config    *EdgeDeviceConfig
		expErrStr string
	}{
		{
			"case 1 have mock config can get mock edge device",
			&EdgeDeviceConfig{
				NameSpace:      "test",
				DeviceName:     "httpdevice",
				KubeconfigPath: MockConfigFile,
			},
			"",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			ds, _, err := NewEdgeDevice(c.config)
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
				assert.Nil(t, ds)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, "test_namespace", ds.Namespace)
				assert.Equal(t, "test_name", ds.Name)
			}
		})
	}
}

type MockResponse struct {
	v1alpha1.EdgeDevice
}

func mockHttpServer(t *testing.T) *httptest.Server {
	mockrs := MockResponse{
		EdgeDevice: v1alpha1.EdgeDevice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test_name",
				Namespace: "test_namespace",
			},
			Status: v1alpha1.EdgeDeviceStatus{
				EdgeDevicePhase: (*v1alpha1.EdgeDevicePhase)(unitest.ToPointer("Success")),
			},
		},
	}

	dsByte, _ := json.Marshal(mockrs)

	// Implements the http.Handler interface to be passed to httptest.NewServer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch path {
		case "/apis/shifu.edgenesis.io/v1alpha1/namespaces/test/edgedevices/httpdevice":
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(dsByte)
			if err != nil {
				t.Errorf("failed to write response")
			}
		default:
			t.Errorf("Not expected to request: %s", r.URL.Path)
		}
	}))
	return server
}

func writeMockConfigFile(t *testing.T, serverURL string) {
	fakeConfig := clientcmdapi.NewConfig()
	fakeConfig.APIVersion = "v1"
	fakeConfig.CurrentContext = "alpha"

	fakeConfig.Clusters["alpha"] = &clientcmdapi.Cluster{
		Server:                serverURL,
		InsecureSkipTLSVerify: true,
	}

	fakeConfig.Contexts["alpha"] = &clientcmdapi.Context{
		Cluster: "alpha",
	}

	err := clientcmd.WriteToFile(*fakeConfig, MockConfigFile)
	if err != nil {
		t.Errorf("write mock file failed")
	}

}
