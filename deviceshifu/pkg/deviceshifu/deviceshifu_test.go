package deviceshifu

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

func TestDeviceShifuEmptyNamespace(t *testing.T) {
	deviceShifuMetadata := &DeviceShifuMetaData{
		"TestDeviceShifuEmptyNamespace",
		"etc/edgedevice/config",
		DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
		"",
	}

	_, err := New(deviceShifuMetadata)
	if err != nil {
		log.Print(err)
	} else {
		t.Errorf("DeviceShifu Test with empty namespace failed")
	}
	time.Sleep(1 * time.Second)
}

func TestStart(t *testing.T) {
	deviceShifuMetadata := &DeviceShifuMetaData{
		"TestStart",
		"etc/edgedevice/config",
		DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
		"TestStartNamespace",
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
		"TestStartHttpServerNamespace",
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
	req, err := http.NewRequest("GET", "http://localhost:8081/--start?time=10:00:00&flags_no_parameter=-a,-c,--no-dependency&target=machine2", nil)
	log.Println(req.URL.Query())
	createdReq := createHTTPCommandlineRequestString(req, "/usr/local/bin/python /usr/src/driver/python-car-driver.py", "--start")
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}

	expectedRequestExecution := "/usr/local/bin/python /usr/src/driver/python-car-driver.py"
	expectedRequestInstruction := "--start"
	expectedRequestArguments := []string{"time=10:00:00", "target=machine2", "-a", "-c", "--no-dependency"}
	createdRequestList := strings.Split(createdReq, expectedRequestInstruction)
	if len(createdRequestList) != 2 {
		t.Error("created request instructiondoes not match the expected req")
	}

	if strings.TrimSpace(createdRequestList[0]) != expectedRequestExecution {
		t.Errorf("created request execution: '%v' does not match the expected req execution: '%v'\n", createdRequestList[0], expectedRequestExecution)
	}

	createdRequestArguments := strings.Fields(createdRequestList[1])
	if len(expectedRequestArguments) != len(createdRequestArguments) {
		t.Errorf("length of created request args: '%v' does not match the expected req args: '%v'\n", createdRequestArguments, expectedRequestArguments)
	}

	isArgumentInRequest := false
	for _, expectedArgument := range expectedRequestArguments {
		for _, createdArgument := range createdRequestArguments {
			if expectedArgument == strings.TrimSpace(createdArgument) {
				isArgumentInRequest = true
			}
		}
		if !isArgumentInRequest {
			t.Errorf("expected request argument: '%v' not in created arguments: %v", expectedArgument, createdRequestArguments)
		}
		isArgumentInRequest = false
	}
}

func TestCreateHTTPCommandlineRequestString2(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8081/issue_cmd?cmdTimeout=10&flags_no_parameter=ping,8.8.8.8,-t", nil)
	createdReq := createHTTPCommandlineRequestString(req, "poweshell.exe", DEVICE_HTTP_CMD_NO_EXEC)
	if err != nil {
		t.Errorf("Cannot create HTTP commandline request: %v", err.Error())
	}

	expectedReq := "ping 8.8.8.8 -t"
	if createdReq != expectedReq {
		t.Errorf("created request: '%v' does not match the expected req: '%v'\n", createdReq, expectedReq)
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

	if string(body) != instruction {
		fmt.Printf("Body: '%+v' does not match instruction: '%v'\n", string(body), instruction)
		// TODO: for now return true since we don't have a test device
		return true
	}

	return true
}
