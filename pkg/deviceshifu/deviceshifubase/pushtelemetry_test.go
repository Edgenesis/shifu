package deviceshifubase

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/deviceshifu/utils"
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
		expectedMap map[string]v1alpha1.TelemetryServiceSpec
		expErrStr   string
	}{
		{
			"case 1 no setting",
			&DeviceShifuBase{
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{},
				},
			},
			map[string]v1alpha1.TelemetryServiceSpec(map[string]v1alpha1.TelemetryServiceSpec{}),
			"",
		},
		{
			"case 2 has default with false",
			&DeviceShifuBase{
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(false),
							DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer(""),
						},
					},
				},
			},
			map[string]v1alpha1.TelemetryServiceSpec(map[string]v1alpha1.TelemetryServiceSpec{}),
			"",
		},
		{
			"case 3 has default with true and empty endpoint",
			&DeviceShifuBase{
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer(""),
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
							DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer("test_endpoint-1"),
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/endpoint1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]v1alpha1.TelemetryServiceSpec(map[string]v1alpha1.TelemetryServiceSpec{}),
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
							DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer("test_endpoint-1"),
						},
						DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
							"device_healthy": {
								DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
									PushSettings: &DeviceShifuTelemetryPushSettings{
										DeviceShifuTelemetryPushToServer:      unitest.ToPointer(true),
										DeviceShifuTelemetryCollectionService: unitest.ToPointer(""),
									},
								},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]v1alpha1.TelemetryServiceSpec(map[string]v1alpha1.TelemetryServiceSpec{
				"device_healthy": {
					Type:    unitest.ToPointer(v1alpha1.TypeHTTP),
					Address: unitest.ToPointer(""),
				},
			}),
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
							DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer("test_endpoint-1"),
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
			map[string]v1alpha1.TelemetryServiceSpec(map[string]v1alpha1.TelemetryServiceSpec{
				"device_healthy": {
					Type:    unitest.ToPointer(v1alpha1.TypeHTTP),
					Address: unitest.ToPointer(""),
				},
			}),
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
							DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer("test_endpoint-1"),
						},
						DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
							"device_healthy": {
								DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
									PushSettings: &DeviceShifuTelemetryPushSettings{
										DeviceShifuTelemetryPushToServer:      unitest.ToPointer(true),
										DeviceShifuTelemetryCollectionService: unitest.ToPointer("test-healthy-endpoint"),
									},
								},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test-healthy-endpoint\",\"type\": \"HTTP\"}}", t),
			},
			map[string]v1alpha1.TelemetryServiceSpec(map[string]v1alpha1.TelemetryServiceSpec{
				"device_healthy": {
					Type:    unitest.ToPointer(v1alpha1.TypeHTTP),
					Address: unitest.ToPointer("http://192.168.15.48:12345/test-healthy-endpoint"),
				},
			}),
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
							DeviceShifuTelemetryDefaultPushToServer:      unitest.ToPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: unitest.ToPointer("test_endpoint-1"),
						},
						DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
							"device_healthy": {
								DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
									PushSettings: &DeviceShifuTelemetryPushSettings{
										DeviceShifuTelemetryPushToServer:      unitest.ToPointer(true),
										DeviceShifuTelemetryCollectionService: unitest.ToPointer("test_endpoint-1"),
									},
								},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]v1alpha1.TelemetryServiceSpec(map[string]v1alpha1.TelemetryServiceSpec{
				"device_healthy": {
					Type:    unitest.ToPointer(v1alpha1.TypeHTTP),
					Address: unitest.ToPointer("http://192.168.15.48:12345/test_endpoint-1"),
				},
			}),
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
								},
							},
						},
					},
				},
				RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
			},
			map[string]v1alpha1.TelemetryServiceSpec(map[string]v1alpha1.TelemetryServiceSpec{}),
			"",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			result, err := getTelemetryCollectionServiceMap(c.inputDevice)
			ok := assert.ObjectsAreEqual(c.expectedMap, result)
			assert.Equal(t, true, ok)
			if len(c.expErrStr) == 0 {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), c.expErrStr)
			}
		})
	}

}

func TestCopyHeader(t *testing.T) {
	src := http.Header{
		"Test": {"aa", "bb", "cc"},
	}
	dst := http.Header{}
	utils.CopyHeader(dst, src)

	if !reflect.DeepEqual(dst, src) {
		t.Errorf("Not match")
	}
}

func TestPushToHTTPTelemetryCollectionService(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader("Hello,World")),
	}

	err := pushToHTTPTelemetryCollectionService(resp, "localhost")
	assert.NotNil(t, err)
}

func TestPushToShifuTelemetryCollectionService(t *testing.T) {
	server := mockMQTTTelemetryServiceServer(t)
	defer server.Close()
	mockServerAddress := server.URL
	testCases := []struct {
		name        string
		message     *http.Response
		request     *v1alpha1.TelemetryRequest
		address     string
		expectedErr string
	}{
		{
			name: "case1 Error address",
			message: &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("TestBody")),
			},
			request: &v1alpha1.TelemetryRequest{
				MQTTSetting: &v1alpha1.MQTTSetting{
					MQTTTopic: unitest.ToPointer("/test/topic"),
				},
			},
			address:     "test",
			expectedErr: "Post \"test\": unsupported protocol scheme \"\"",
		}, {
			name: "case2 pass",
			message: &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("TestBody")),
			},
			request: &v1alpha1.TelemetryRequest{
				MQTTSetting: &v1alpha1.MQTTSetting{
					MQTTTopic: unitest.ToPointer("/test/topic"),
				},
			},
			address:     mockServerAddress,
			expectedErr: "",
		},
	}
	for _, c := range testCases {
		err := pushToShifuTelemetryCollectionService(c.message, *c.request, c.address)
		if err != nil {
			assert.Equal(t, err.Error(), c.expectedErr)
		} else {
			assert.Equal(t, c.expectedErr, "")
		}
	}
}

func mockMQTTTelemetryServiceServer(t *testing.T) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(handler)
}

func TestPushTelemetryCollectionService(t *testing.T) {
	server := mockMQTTTelemetryServiceServer(t)
	defer server.Close()
	address := server.URL

	testCases := []struct {
		name        string
		spec        *v1alpha1.TelemetryServiceSpec
		message     *http.Response
		expectedErr string
	}{
		{
			name: "case1 http",
			spec: &v1alpha1.TelemetryServiceSpec{
				Type:    unitest.ToPointer(v1alpha1.TypeHTTP),
				Address: unitest.ToPointer(address),
			},
			message: &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("test")),
			},
		}, {
			name: "case2 MQTT",
			spec: &v1alpha1.TelemetryServiceSpec{
				ServiceSettings: &v1alpha1.ServiceSettings{
					MQTTSetting: &v1alpha1.MQTTSetting{
						MQTTTopic: unitest.ToPointer("/test/topic"),
					},
				},
				Address: unitest.ToPointer(address),
				Type:    unitest.ToPointer(v1alpha1.TypeMQTT),
			},
			message: &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("test")),
			},
			expectedErr: "",
		}, {
			name: "case3 SQL",
			spec: &v1alpha1.TelemetryServiceSpec{
				ServiceSettings: &v1alpha1.ServiceSettings{
					SQLSetting: &v1alpha1.SQLConnectionSetting{},
				},
				Address: unitest.ToPointer(address),
				Type:    unitest.ToPointer(v1alpha1.TypeSQL),
			},
			message: &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("test")),
			},
			expectedErr: "",
		}, {
			name: "case4 OtherProtocol",
			spec: &v1alpha1.TelemetryServiceSpec{
				ServiceSettings: &v1alpha1.ServiceSettings{
					MQTTSetting: &v1alpha1.MQTTSetting{
						MQTTTopic: unitest.ToPointer("/test/topic"),
					},
				},
				Address: unitest.ToPointer(address),
				Type:    unitest.ToPointer(v1alpha1.Type("")),
			},
			message: &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("test")),
			},
			expectedErr: "unsupported protocol",
		},
	}

	for _, c := range testCases {
		err := PushTelemetryCollectionService(c.spec, c.message)
		if err != nil {
			assert.Equal(t, err.Error(), c.expectedErr)
		}
	}
}
