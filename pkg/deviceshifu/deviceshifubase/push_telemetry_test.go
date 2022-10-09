package deviceshifubase

import (
	"net/http/httptest"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	utiltesting "k8s.io/client-go/util/testing"
)

func mockTestServer(response string, statusCode int, t *testing.T) *httptest.Server {
	fakeHandler := utiltesting.FakeHandler{
		StatusCode:   statusCode,
		ResponseBody: string(response),
		T:            t,
	}
	testServer := httptest.NewServer(&fakeHandler)
	return testServer
}

func mockRestClientFor(resp string, t *testing.T) *rest.RESTClient {
	testServer := mockTestServer(resp, 200, t)
	c, _ := rest.RESTClientFor(&rest.Config{
		Host: testServer.URL,
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		Username: "user",
		Password: "pass",
	})

	return c
}

func TestGetTelemetryCollectionServiceMap(t *testing.T) {
	//test cases
	// case 1 no setting
	// case 2 has default with false
	// case 3 has default with true and empty endpoint
	// case 4 has default with true and valid endpoint
	// case 5 has default with true and valid endpoint, has telemetry with empty push setting
	// case 6 has default with true and valid endpoint, has telemetry with no push setting
	// case 7 has default with true and valid endpoint, has telemetry with valid push setting
	// case 8 has default with true and valid endpoint, has telemetry with same endpoint
	// case 9 has default with true and valid endpoint, has telemetry with push false setting

	testCases := []struct {
		Name        string
		inputDevice *DeviceShifuBase
		expectedMap map[string]string
		expErrStr   string
	}{
		{
			"case 1 no setting",
			&DeviceShifuBase{
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{},
				},
			},
			map[string]string{},
			"",
		},
		{
			"case 2 has default with false",
			&DeviceShifuBase{
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.BoolPointer(false),
							DeviceShifuTelemetryDefaultCollectionService: unitest.StrPointer(""),
						},
					},
				},
			},
			map[string]string{},
			"",
		},
		{
			"case 3 has default with true and empty endpoint",
			&DeviceShifuBase{
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.BoolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.StrPointer(""),
						},
					},
				},
			},
			nil,
			"you need to configure defaultTelemetryCollectionService if setting defaultPushToServer to true",
		},
		{
			"case 4 has default with true and valid endpoint",
			&DeviceShifuBase{
				Name: "test",
				EdgeDevice: &v1alpha1.EdgeDevice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test_namespace",
					},
				},
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.BoolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.StrPointer("test_endpoint-1"),
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/endpoint1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]string{},
			"",
		},
		{
			"case 5 has default with true and valid endpoint, has telemetry with empty push setting",
			&DeviceShifuBase{
				Name: "test",
				EdgeDevice: &v1alpha1.EdgeDevice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test_namespace",
					},
				},
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.BoolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.StrPointer("test_endpoint-1"),
						},
						DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
							"device_healthy": {
								DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
									PushSettings: &DeviceShifuTelemetryPushSettings{
										DeviceShifuTelemetryPushToServer:      unitest.BoolPointer(true),
										DeviceShifuTelemetryCollectionService: unitest.StrPointer(""),
									},
								},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]string{"device_healthy": "http://192.168.15.48:12345/test_endpoint-1"},
			"",
		},
		{
			"case 6 has default with true and valid endpoint, has telemetry with no push setting",
			&DeviceShifuBase{
				Name: "test",
				EdgeDevice: &v1alpha1.EdgeDevice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test_namespace",
					},
				},
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.BoolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.StrPointer("test_endpoint-1"),
						},
						DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
							"device_healthy": {
								DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]string{"device_healthy": "http://192.168.15.48:12345/test_endpoint-1"},
			"",
		},
		{
			"case 7 has default with true and valid endpoint, has telemetry with valid push setting",
			&DeviceShifuBase{
				Name: "test",
				EdgeDevice: &v1alpha1.EdgeDevice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test_namespace",
					},
				},
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.BoolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.StrPointer("test_endpoint-1"),
						},
						DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
							"device_healthy": {
								DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
									PushSettings: &DeviceShifuTelemetryPushSettings{
										DeviceShifuTelemetryPushToServer:      unitest.BoolPointer(true),
										DeviceShifuTelemetryCollectionService: unitest.StrPointer("test-healthy-endpoint"),
									},
								},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test-healthy-endpoint\",\"type\": \"HTTP\"}}", t),
			},
			map[string]string{"device_healthy": "http://192.168.15.48:12345/test-healthy-endpoint"},
			"",
		},
		{
			"case 8 has default with true and valid endpoint, has telemetry with same endpoint",
			&DeviceShifuBase{
				Name: "test",
				EdgeDevice: &v1alpha1.EdgeDevice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test_namespace",
					},
				},
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.BoolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.StrPointer("test_endpoint-1"),
						},
						DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
							"device_healthy": {
								DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
									PushSettings: &DeviceShifuTelemetryPushSettings{
										DeviceShifuTelemetryPushToServer:      unitest.BoolPointer(true),
										DeviceShifuTelemetryCollectionService: unitest.StrPointer("test_endpoint-1"),
									},
								},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]string{"device_healthy": "http://192.168.15.48:12345/test_endpoint-1"},
			"",
		},
		{
			"case 9 has default with true and valid endpoint, has telemetry with push false setting",
			&DeviceShifuBase{
				Name: "test",
				EdgeDevice: &v1alpha1.EdgeDevice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test_namespace",
					},
				},
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.BoolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.StrPointer("test_endpoint-1"),
						},
						DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
							"device_healthy": {
								DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
									PushSettings: &DeviceShifuTelemetryPushSettings{
										DeviceShifuTelemetryPushToServer:      unitest.BoolPointer(false),
										DeviceShifuTelemetryCollectionService: unitest.StrPointer("test_endpoint-1"),
									},
								},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]string{},
			"",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			result, err := getTelemetryCollectionServiceMap(c.inputDevice)
			assert.Equal(t, c.expectedMap, result)
			if len(c.expErrStr) == 0 {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), c.expErrStr)
			}
		})
	}

}
