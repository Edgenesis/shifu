package deviceshifumqtt

import (
	"io"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/utils"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

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
		Namespace:      "",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu %v", err.Error())
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
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
		Namespace:      "",
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

	expectedRequestString := "/usr/local/bin/python /usr/src/driver/python-car-driver.py --start time=10:00:00 target=machine2 -a -c --no-dependency"
	expectedRequestArguments := strings.Fields(expectedRequestString)

	sort.Strings(createdRequestArguments)
	sort.Strings(expectedRequestArguments)

	if !reflect.DeepEqual(createdRequestArguments, expectedRequestArguments) {
		t.Errorf("created request: '%v' does not match the expected req: '%v'", createdRequestString, expectedRequestString)
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
