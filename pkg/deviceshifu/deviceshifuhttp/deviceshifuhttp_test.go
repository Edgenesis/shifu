package deviceshifuhttp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/utils"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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

	err = GenerateConfigMapFromSnippet(MockDeviceCmStr2, MockDeviceConfigFolder)
	if err != nil {
		klog.Errorf("error when generateConfigmapFromSnippet2, err: %v", err)
		os.Exit(-1)
	}
	m.Run()
	err = os.RemoveAll(MockDeviceConfigPath)
	if err != nil {
		klog.Fatal(err)
	}
}

func TestDeviceShifuEmptyNamespace(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestDeviceShifuEmptyNamespace",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
	}

	_, err := New(deviceShifuMetadata)
	if err != nil {
		klog.Errorf("%v", err)
	} else {
		t.Errorf("DeviceShifuHTTP Test with empty namespace failed")
	}
}

func TestStart(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestStart",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
		Namespace:      "TestStartNamespace",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifuHTTP.Start failed due to: %v", err.Error())
	}

	if err := mockds.Stop(); err != nil {
		t.Errorf("unable to stop mock deviceShifu, error: %+v", err)
	}
}

func TestDeviceHealthHandler(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TeststartHTTPServer",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
		Namespace:      "TeststartHTTPServerNamespace",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}

	resp, err := utils.RetryAndGetHTTP("http://localhost:8080/health", 3)
	if err != nil {
		t.Errorf("HTTP GET returns an error %v", err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unable to read response body, error: %v", err.Error())
	}

	if string(body) != deviceshifubase.DeviceIsHealthyStr {
		t.Errorf("%+v", body)
	}

	if err := mockds.Stop(); err != nil {
		t.Errorf("unable to stop mock deviceShifu, error: %+v", err)
	}
}

func TestCreateHTTPCommandlineRequestString(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8081/start?time=10:00:00&flags_no_parameter=-a,-c,--no-dependency&target=machine2", nil)
	klog.Infof("%v", req.URL.Query())
	createdRequestString := createHTTPCommandlineRequestString(req, "/usr/local/bin/python /usr/src/driver/python-car-driver.py", "start")
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}

	createdRequestArguments := strings.Fields(createdRequestString)

	expectedRequestString := "/usr/local/bin/python /usr/src/driver/python-car-driver.py start time=10:00:00 target=machine2 -a -c --no-dependency"
	expectedRequestArguments := strings.Fields(expectedRequestString)

	sort.Strings(createdRequestArguments)
	sort.Strings(expectedRequestArguments)

	if !reflect.DeepEqual(createdRequestArguments, expectedRequestArguments) {
		t.Errorf("created request: '%v' does not match the expected req: '%v'", createdRequestString, expectedRequestString)
	}
}

func TestCreateHTTPCommandlineRequestString2(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8081/issue_cmd?cmdTimeout=10&flags_no_parameter=ping,8.8.8.8,-t", nil)
	createdReq := createHTTPCommandlineRequestString(req, "poweshell.exe", deviceshifubase.DeviceDefaultCMDDoNotExec)
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}
	expectedReq := "ping 8.8.8.8 -t"
	if createdReq != expectedReq {
		t.Errorf("created request: '%v' does not match the expected req: '%v'\n", createdReq, expectedReq)
	}
}

func TestCreatehttpURIString(t *testing.T) {
	expectedURIString := "http://localhost:8081/start?time=10:00:00&target=machine1&target=machine2"
	req, err := http.NewRequest("POST", expectedURIString, nil)
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}

	klog.Infof("%v", req.URL.Query())
	createdURIString := createURIFromRequest("localhost:8081", "start", req)

	createdURIStringWithoutQueries := strings.Split(createdURIString, "?")[0]
	createdQueries := strings.Split(strings.Split(createdURIString, "?")[1], "&")
	expectedURIStringWithoutQueries := strings.Split(expectedURIString, "?")[0]
	expectedQueries := strings.Split(strings.Split(expectedURIString, "?")[1], "&")

	sort.Strings(createdQueries)
	sort.Strings(expectedQueries)
	if createdURIStringWithoutQueries != expectedURIStringWithoutQueries || !reflect.DeepEqual(createdQueries, expectedQueries) {
		t.Errorf("createdQuery '%v' is different from the expectedQuery '%v'", createdURIString, expectedURIString)
	}
}

func TestCreatehttpURIStringNoQuery(t *testing.T) {
	expectedURIString := "http://localhost:8081/start"
	req, err := http.NewRequest("POST", expectedURIString, nil)
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}

	klog.Infof("%v", req.URL.Query())
	createdURIString := createURIFromRequest("localhost:8081", "start", req)

	createdURIStringWithoutQueries := strings.Split(createdURIString, "?")[0]
	expectedURIStringWithoutQueries := strings.Split(expectedURIString, "?")[0]

	if createdURIStringWithoutQueries != expectedURIStringWithoutQueries {
		t.Errorf("createdQuery '%v' is different from the expectedQuery '%v'", createdURIString, expectedURIString)
	}
}

func Test_commandHandleHTTPFunc(t *testing.T) {

	hs := mockHandlerServer()
	defer hs.Close()
	addr := strings.Split(hs.URL, "//")[1]

	hc, err := mockRestClient(addr, "")
	if err != nil {
		t.Errorf("create handler client error: %s", err.Error())
	}
	mockHandlerHTTP := &DeviceCommandHandlerHTTP{
		client: hc,
		HandlerMetaData: &HandlerMetaData{
			edgeDeviceSpec: v1alpha1.EdgeDeviceSpec{
				Address: &addr,
			},
			instruction: "test_instruction",
			properties:  mockDeviceShifuInstruction(),
		},
	}

	ds := mockDeviceServer(mockHandlerHTTP)
	defer ds.Close()
	dc, err := mockRestClient(ds.URL, "testing")
	if err != nil {
		t.Errorf("create device client error: %s", err.Error())
	}

	// start device client testing
	r := dc.Get().Param("timeout", "1").Do(context.TODO())
	assert.Nil(t, r.Error())

	r = dc.Get().Param("timeout", "aa").Do(context.TODO())
	assert.Equal(t, "the server rejected our request for an unknown reason", r.Error().Error())

	r = dc.Get().Do(context.TODO())
	assert.Nil(t, r.Error())

	r = dc.Post().Do(context.TODO())
	assert.Nil(t, r.Error())

	r = dc.Put().Do(context.TODO())
	assert.Nil(t, r.Error())
}

func Test_commandHandleFuncHTTPCommandLine(t *testing.T) {
	hs := mockHandlerServer()
	defer hs.Close()
	addr := strings.Split(hs.URL, "//")[1]

	hc, err := mockRestClient(addr, "")
	if err != nil {
		t.Errorf("create handler client error: %s", err.Error())
	}
	mockHandlerHTTPCli := &DeviceCommandHandlerHTTPCommandline{
		client: hc,
		CommandlineHandlerMetadata: &CommandlineHandlerMetadata{
			edgeDeviceSpec: v1alpha1.EdgeDeviceSpec{
				Address: &addr,
			},
			instruction: "test_instruction",
			properties:  mockDeviceShifuInstruction(),
		},
	}

	ds := mockDeviceServer(mockHandlerHTTPCli)
	defer ds.Close()
	dc, err := mockRestClient(ds.URL, "testing")
	if err != nil {
		t.Errorf("create device client error: %s", err.Error())
	}

	// start device client testing
	r := dc.Post().Param("timeout", "1").Param("stub_toleration", "1").Do(context.TODO())
	assert.Nil(t, r.Error())

	r = dc.Post().Param("timeout", "aa").Param("stub_toleration", "1").Do(context.TODO())
	assert.Equal(t, "the server rejected our request for an unknown reason", r.Error().Error())

	r = dc.Post().Param("timeout", "-1").Param("stub_toleration", "aa").Do(context.TODO())
	assert.Equal(t, "the server rejected our request for an unknown reason", r.Error().Error())

}

func mockRestClient(url string, path string) (*rest.RESTClient, error) {
	return rest.RESTClientFor(
		&rest.Config{
			Host:    url,
			APIPath: path,
			ContentConfig: rest.ContentConfig{
				GroupVersion:         &v1.SchemeGroupVersion,
				NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
			},
			Username: "user",
			Password: "pass",
		},
	)
}

type MockCommandHandler interface {
	commandHandleFunc() http.HandlerFunc
}

func mockDeviceServer(h MockCommandHandler) *httptest.Server {
	// catch device http request and response properly with specific paths
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		if method == "GET" || method == "POST" {
			path := r.URL.Path
			switch path {
			case "/testing/apps/v1":
				println("ds get testing call, calling the handler server")
				f := h.commandHandleFunc()
				f.ServeHTTP(w, r)
			default:
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				println("ds default request, path:", path)
			}
		} else {
			println("invalid method")
		}
	}))
	return server
}

func mockHandlerServer() *httptest.Server {
	// catch handler http request and response properly with specific paths
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch path {
		case "/test_instruction":
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			println("handler get the instruction and executed.")
		default:
			w.WriteHeader(http.StatusOK)
			println("hs get default request, path:", path)
		}

	}))
	return server
}

func mockDeviceShifuInstruction() *deviceshifubase.DeviceShifuInstruction {
	return &deviceshifubase.DeviceShifuInstruction{
		DeviceShifuInstructionProperties: []deviceshifubase.DeviceShifuInstructionProperty{
			{
				ValueType:    "testing",
				ReadWrite:    "rw",
				DefaultValue: "0",
			},
		},
		DeviceShifuProtocolProperties: map[string]string{
			"test_key": "test_value",
		},
	}
}

func Test_collectHTTPTelemtries(t *testing.T) {
	ts := mockTelemetryServer()
	addr := strings.Split(ts.URL, "//")[1]
	mockDevice := &DeviceShifuHTTP{
		base: &deviceshifubase.DeviceShifuBase{
			Name: "test",
			EdgeDevice: &v1alpha1.EdgeDevice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test_namespace",
				},
				Spec: v1alpha1.EdgeDeviceSpec{
					Address:  &addr,
					Protocol: (*v1alpha1.Protocol)(strPointer(string(v1alpha1.ProtocolHTTP))),
				},
			},
			DeviceShifuConfig: &deviceshifubase.DeviceShifuConfig{
				Telemetries: &deviceshifubase.DeviceShifuTelemetries{
					DeviceShifuTelemetrySettings: &deviceshifubase.DeviceShifuTelemetrySettings{
						DeviceShifuTelemetryTimeoutInMilliseconds:    int64Pointer(10),
						DeviceShifuTelemetryDefaultPushToServer:      boolPointer(true),
						DeviceShifuTelemetryDefaultCollectionService: strPointer("test_endpoint-1"),
					},
					DeviceShifuTelemetries: map[string]*deviceshifubase.DeviceShifuTelemetry{
						"device_healthy": {
							DeviceShifuTelemetryProperties: deviceshifubase.DeviceShifuTelemetryProperties{
								DeviceInstructionName: strPointer("mock_testing"),
								PushSettings: &deviceshifubase.DeviceShifuTelemetryPushSettings{
									DeviceShifuTelemetryPushToServer:      boolPointer(false),
									DeviceShifuTelemetryCollectionService: strPointer("test_endpoint-1"),
								},
								InitialDelayMs: intPointer(1),
							},
						},
					},
				},
			},
			RestClient: mockRestClientFor(addr, "{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
		},
	}

	res, err := mockDevice.collectHTTPTelemtries()
	assert.Equal(t, true, res)
	assert.Nil(t, err)

}

func boolPointer(b bool) *bool {
	return &b
}

func strPointer(s string) *string {
	return &s
}

func intPointer(i int) *int {
	return &i
}

func int64Pointer(i int64) *int64 {
	return &i
}

func mockTelemetryServer() *httptest.Server {
	// catch handler http request and response properly with specific paths
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch path {
		case "/telemetry":
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			println("handler get the instruction and executed.")
		default:
			w.WriteHeader(http.StatusOK)
			println("hs get default request, path:", path)
		}

	}))
	return server
}

func mockRestClientFor(host string, resp string, t *testing.T) *rest.RESTClient {
	c, _ := rest.RESTClientFor(&rest.Config{
		Host: host,
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		Username: "user",
		Password: "pass",
	})

	return c
}
