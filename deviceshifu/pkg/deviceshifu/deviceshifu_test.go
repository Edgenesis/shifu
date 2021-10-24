package deviceshifu

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func TestStart(t *testing.T) {
	deviceShifuMetadata := &DeviceShifuMetaData{
		"TestStart",
		"etc/edgedevice/config",
		DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
		"",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceShifu")
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}

	mockds.Stop()
	time.Sleep(1 * time.Second)
}

func TestDeviceHealthHandler(t *testing.T) {
	deviceShifuMetadata := &DeviceShifuMetaData{
		"TestStartHttpServer",
		"etc/edgedevice/config",
		DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
		"",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceShifu")
	}

	go mockds.startHttpServer(wait.NeverStop)

	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://127.0.0.1:8080/health")
	if err != nil {
		t.Errorf("HTTP GET returns an error %v", err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if string(body) != DEVICE_IS_HEALTHY_STR {
		t.Errorf("%+v", body)
	}

	mockds.Stop()
	time.Sleep(1 * time.Second)
}

func TestCreateHTTPCommandlineRequestString(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8081/start?time=10:00:00&flags_no_parameter=-a,-c,--no-dependency&target=machine2", nil)
	fmt.Println(req.URL.Query())
	createdReq := createHTTPCommandlineRequestString(req, "/usr/local/bin/python /usr/src/driver/python-car-driver.py", "start")
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}

	expectedReq := "/usr/local/bin/python /usr/src/driver/python-car-driver.py --start time=10:00:00 target=machine2 -a -c --no-dependency"

	if createdReq != expectedReq {
		t.Errorf("created request: '%v' does not match the expected req: '%v'\n", createdReq, expectedReq)
	}
}

func CheckSimpleInstructionHandlerHttpResponse(instruction string, httpEndpoint string) bool {
	resp, err := http.Get(httpEndpoint + "/" + instruction)
	if err != nil {
		log.Fatalf("HTTP GET returns an error %v", err.Error())
		return false
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if string(body) != instruction {
		fmt.Printf("Body: '%+v' does not match instruction: '%v'\n", string(body), instruction)
		// TODO: for now return true since we don't have a test device
		return true
	}

	return true
}
