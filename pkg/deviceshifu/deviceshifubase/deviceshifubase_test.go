package deviceshifubase

import (
	"errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"reflect"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func TestMain(m *testing.M) {
	err := GenerateConfigMapFromSnippet(MockDeviceCmStr, MockDeviceConfigFolder)
	if err != nil {
		klog.Errorf("error when generateConfigmapFromSnippet, err: %v", err)
		os.Exit(-1)
	}
	m.Run()
	err = os.RemoveAll(MockDeviceConfigPath)
	if err != nil {
		klog.Fatal(err)
	}
}

func TestValidateTelemetryConfig(t *testing.T) {
	testCases := []struct {
		Name        string
		inputDevice *DeviceShifuBase
		expErrStr   string
	}{
		{
			"case 1 no setting",
			&DeviceShifuBase{
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{},
				},
			},
			"",
		},
		{
			"case 2 has pushsetting with negative interval",
			&DeviceShifuBase{
				Name: "test",
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:          unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService:     unitest.ToPointer("test_endpoint-1"),
							DeviceShifuTelemetryUpdateIntervalInMilliseconds: unitest.ToPointer(int64(-1)),
						},
					},
				},
			},
			"error deviceShifuTelemetryInterval mustn't be negative number",
		},
		{
			"case 3 has pushsetting with negative initial delay",
			&DeviceShifuBase{
				Name: "test",
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:        unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService:   unitest.ToPointer("test_endpoint-1"),
							DeviceShifuTelemetryInitialDelayInMilliseconds: unitest.ToPointer(int64(-1)),
						},
					},
				},
			},
			"error deviceShifuTelemetryInitialDelay mustn't be negative number",
		},
		{
			"case 4 has pushsetting with negative timeout",
			&DeviceShifuBase{
				Name: "test",
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer("test_endpoint-1"),
							DeviceShifuTelemetryTimeoutInMilliseconds:    unitest.ToPointer(int64(-1)),
						},
					},
				},
			},
			"error deviceShifuTelemetryTimeout mustn't be negative number",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.inputDevice.ValidateTelemetryConfig()
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
			} else {
				assert.Nil(t, err)
			}

		})
	}

}

func TestTelemetryCollection(t *testing.T) {
	mockds := &DeviceShifuBase{
		Name: "test",
		EdgeDevice: &v1alpha1.EdgeDevice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test_namespace",
			},
		},
		DeviceShifuConfig: &DeviceShifuConfig{
			Telemetries: &DeviceShifuTelemetries{
				DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
					DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(true),
					DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer("test_endpoint-1"),
				},
				DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
					"device_healthy": {
						DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
							PushSettings: &DeviceShifuTelemetryPushSettings{
								DeviceShifuTelemetryPushToServer:      unitest.ToPointer(false),
								DeviceShifuTelemetryCollectionService: unitest.ToPointer("test_endpoint-1"),
							},
							InitialDelayMs: unitest.ToPointer(1),
						},
					},
				},
			},
		},
		RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
	}

	testCases := []struct {
		Name        string
		inputDevice *DeviceShifuBase
		fn          func() (bool, error)
		expErrStr   string
	}{
		{
			"case 1 fn true with nil error",
			mockds,
			func() (bool, error) {
				return true, nil
			},
			"",
		},
		{
			"case 2 fn false with nil error",
			mockds,
			func() (bool, error) {
				return false, nil
			},
			"",
		},
		{
			"case 3 fn false with error",
			mockds,
			func() (bool, error) {
				return false, errors.New("exit")
			},
			"",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.inputDevice.telemetryCollection(c.fn)
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestNew(t *testing.T) {
	testCases := []struct {
		Name      string
		metaData  *DeviceShifuMetaData
		expErrStr string
		initEnv   func()
	}{
		{
			"case 1 have empty name can not new device base",
			&DeviceShifuMetaData{},
			"DeviceShifu's name can't be empty",
			func() {},
		},
		{
			"case 2 have empty configpath meta new device base",
			&DeviceShifuMetaData{
				Name: "test",
			},
			"Error parsing ConfigMap at /etc/edgedevice/config",
			func() {},
		},
		{
			"case 3 have empty KubeConfigPath meta new device base",
			&DeviceShifuMetaData{
				Name:           "test",
				ConfigFilePath: "etc/edgedevice/config",
			},
			"unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined",
			func() {},
		},
		{
			"case 4 KubeConfigPath is NULL",
			&DeviceShifuMetaData{
				Name:           "test",
				ConfigFilePath: "etc/edgedevice/config",
				KubeConfigPath: "NULL",
				Namespace:      "default",
			},
			"",
			func() {
				initEnv()
			},
		},
		{
			"case 5 KubeConfigPath not NULL",
			&DeviceShifuMetaData{
				Name:           "test",
				ConfigFilePath: "etc/edgedevice/config",
				Namespace:      "test_namespace",
			},
			"open /var/run/secrets/kubernetes.io/serviceaccount/token: The system cannot find the path specified.",
			func() {
				initEnv()
			},
		},
		{
			"case 6 KubeConfigPath not NULL AND NewDeviceShifuConfig",
			&DeviceShifuMetaData{
				Name:           "test",
				ConfigFilePath: "etc/edgedevice/config",
				Namespace:      "test_namespace",
				KubeConfigPath: "etc/edgedevice/config",
			},
			"error loading config file \"etc/edgedevice/config\": read etc/edgedevice/config: The handle is invalid.",
			func() {
				err := GenerateConfigMapFromSnippet(MockDeviceCmStr, MockDeviceConfigFolder)
				if err != nil {
					return
				}
				initEnv()
			},
		},
	}
	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			c.initEnv()
			base, mux, err := New(c.metaData)
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
				assert.Nil(t, base)
				assert.Nil(t, mux)
			} else {
				assert.NotNil(t, base)
				assert.NotNil(t, mux)
			}
			unSetEnv()
		})
	}

}

func initEnv() {
	os.Setenv("", "localhost")
	os.Setenv("KUBERNETES_SERVICE_PORT", "1080")
	os.Setenv("KUBERNETES_SERVICE_HOST", "127.0.0.1")
}

func unSetEnv() {
	os.Unsetenv("")
	os.Unsetenv("KUBERNETES_SERVICE_PORT")
	os.Unsetenv("KUBERNETES_SERVICE_HOST")
}

func TestStart(t *testing.T) {
	deviceShifuMetadata := &DeviceShifuMetaData{
		Name:           "TestStart",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: DeviceKubeconfigDoNotLoadStr,
		Namespace:      "TestStartNamespace",
	}

	base, _, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	if err := base.Start(wait.NeverStop, func() (bool, error) {
		return true, nil
	}); err != nil {
		t.Errorf("DeviceShifuHTTP.Start failed due to: %v", err.Error())
	}

	if err := base.Stop(); err != nil {
		t.Errorf("unable to stop mock deviceShifu, error: %+v", err)
	}
}

func TestStartTelemetryCollection(t *testing.T) {
	deviceShifuMetadata := &DeviceShifuMetaData{
		Name:           "TestStart",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: DeviceKubeconfigDoNotLoadStr,
		Namespace:      "TestStartNamespace",
	}
	mock, _, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	testCases := []struct {
		Name             string
		inputDevice      *DeviceShifuBase
		collectTelemetry func() (bool, error)
		SetMock          func()
		expErrStr        string
	}{
		{
			"case 1 fn true with getTelemetryCollectionServiceMap error ",
			mock,
			func() (bool, error) {
				return true, nil
			},
			func() {
				mock.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryDefaultCollectionService = unitest.ToPointer("")
			},
			"error generating TelemetryCollectionServiceMap, error: you need to configure defaultTelemetryCollectionService if setting defaultPushToServer to true",
		},
		{
			"case 2 fn true nil Namespace",
			mock,
			func() (bool, error) {
				return true, nil
			},
			func() {
			},
			"error generating TelemetryCollectionServiceMap, error: you need to configure defaultTelemetryCollectionService if setting defaultPushToServer to true",
		},
		{
			"case 3 fn true DeviceShifuTelemetrySettings nil",
			mock,
			func() (bool, error) {
				return true, nil
			},
			func() {
				mock.EdgeDevice.Namespace = "test_namespace"
				mock.EdgeDevice.Spec.Protocol = unitest.ToPointer(v1alpha1.ProtocolMQTT)
				mock.RestClient = mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/endpoint1\",\"type\": \"HTTP\"}}", t)
				mock.DeviceShifuConfig.Telemetries.DeviceShifuTelemetries["device_healthy"] =
					&DeviceShifuTelemetry{DeviceShifuTelemetryProperties{
						PushSettings: &DeviceShifuTelemetryPushSettings{
							DeviceShifuTelemetryPushToServer:      unitest.ToPointer(true),
							DeviceShifuTelemetryCollectionService: unitest.ToPointer("test_endpoint-1"),
						},
					}}
			},
			"invalid memory address or nil pointer dereference",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			stopCh := make(chan struct{})
			defer func() {
				if err := recover(); err != nil {
					assert.Equal(t, c.expErrStr, reflect.ValueOf(err).String())
				} //内置函数，可以捕捉到函数异常
			}()
			c.SetMock()
			go func() {
				err := c.inputDevice.StartTelemetryCollection(c.collectTelemetry, stopCh)
				if len(c.expErrStr) > 0 {
					assert.Equal(t, c.expErrStr, err.Error())
				} else {
					assert.Nil(t, err)
				}
			}()
			close(stopCh)
			unSetEnv()
		})
	}
}
