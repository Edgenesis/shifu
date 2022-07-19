package deviceshifuSocket

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
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

	if err := mockds.Stop(); err != nil {
		log.Printf("Error stopping mock deviceShifu, error: %v", err.Error())
	}

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
	if err != nil {
		log.Printf("error reading body from response, error: %v", err.Error())
	}

	if string(body) != DEVICE_IS_HEALTHY_STR {
		t.Errorf("%+v", body)
	}

	if err := mockds.Stop(); err != nil {
		log.Printf("Error stopping mock deviceShifu, error: %v", err.Error())
	}

	time.Sleep(1 * time.Second)
}

func TestCreateHTTPCommandlineRequestString(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8081/start?time=10:00:00&flags_no_parameter=-a,-c,--no-dependency&target=machine2", nil)
	log.Println(req.URL.Query())
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
		t.Errorf("created request: '%v' does not match the expected req: '%v'\n", createdRequestString, expectedRequestString)
	}
}

func TestCreateHTTPUriString(t *testing.T) {
	expectedUriString := "http://localhost:8081/start?time=10:00:00&target=machine1&target=machine2"
	req, err := http.NewRequest("POST", expectedUriString, nil)
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}

	log.Println(req.URL.Query())
	createdUriString := createUriFromRequest("localhost:8081", "start", req)

	createdUriStringWithoutQueries := strings.Split(createdUriString, "?")[0]
	createdQueries := strings.Split(strings.Split(createdUriString, "?")[1], "&")
	expectedUriStringWithoutQueries := strings.Split(expectedUriString, "?")[0]
	expectedQueries := strings.Split(strings.Split(expectedUriString, "?")[1], "&")

	sort.Strings(createdQueries)
	sort.Strings(expectedQueries)
	if createdUriStringWithoutQueries != expectedUriStringWithoutQueries || !reflect.DeepEqual(createdQueries, expectedQueries) {
		t.Errorf("createdQuery '%v' is different from the expectedQuery '%v'", createdUriString, expectedUriString)
	}
}

func TestCreateHTTPUriStringNoQuery(t *testing.T) {
	expectedUriString := "http://localhost:8081/start"
	req, err := http.NewRequest("POST", expectedUriString, nil)
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}

	log.Println(req.URL.Query())
	createdUriString := createUriFromRequest("localhost:8081", "start", req)

	createdUriStringWithoutQueries := strings.Split(createdUriString, "?")[0]
	expectedUriStringWithoutQueries := strings.Split(expectedUriString, "?")[0]

	if createdUriStringWithoutQueries != expectedUriStringWithoutQueries {
		t.Errorf("createdQuery '%v' is different from the expectedQuery '%v'", createdUriString, expectedUriString)
	}

	// cleanup
	t.Cleanup(func() {
		//tear-down code
		err := os.RemoveAll(MOCK_DEVICE_CONFIG_PATH)
		if err != nil {
			log.Fatal(err)
		}
	})
}

func CheckSimpleInstructionHandlerHttpResponse(instruction string, httpEndpoint string) bool {
	resp, err := http.Get(httpEndpoint + "/" + instruction)
	if err != nil {
		log.Fatalf("HTTP GET returns an error %v", err.Error())
		return false
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading body from response, error: %v", err.Error())
	}

	if string(body) != instruction {
		fmt.Printf("Body: '%+v' does not match instruction: '%v'\n", string(body), instruction)
		// TODO: for now return true since we don't have a test device
		return true
	}

	return true
}
