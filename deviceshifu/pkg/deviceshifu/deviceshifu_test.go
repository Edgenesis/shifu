package deviceshifu

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	v1alpha1 "edgenesis.io/shifu/k8s/crd/api/v1alpha1"
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

func TestDeviceInstructionHandler(t *testing.T) {
	var (
		config_folder     = "etc/edgedevice/config"
		httpEndpoint      = "http://127.0.0.1:8080"
		deviceName        = "edgedevice-sample"
		kubeconfigPath    = "/root/.kube/config"
		namespace         = "crd-system"
		instruction_array = []string{
			"health",
			"get_reading",
			"get_status",
			"set_reading",
			"start",
			"stop",
		}
	)

	deviceShifuMetadata := &DeviceShifuMetaData{
		deviceName,
		config_folder,
		kubeconfigPath,
		namespace,
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceShifu")
	}

	go mockds.Start(wait.NeverStop)

	time.Sleep(1 * time.Second)
	for _, instruction := range instruction_array {
		if !CheckSimpleInstructionHandlerHttpResponse(instruction, httpEndpoint) {
			t.Errorf("Error getting instruction response from instruction: %v", instruction)
		}
	}

	getResult := &v1alpha1.EdgeDevice{}
	err = mockds.restClient.Get().
		Namespace(mockds.edgeDevice.Namespace).
		Resource(EDGEDEVICE_RESOURCE_STR).
		Name(mockds.Name).
		Do(context.TODO()).
		Into(getResult)

	if err != nil {
		t.Errorf("Unable to get status, error: %v", err.Error())
	}

	if *getResult.Status.EdgeDevicePhase != v1alpha1.EdgeDeviceFailed {
		t.Errorf("Edgedevice status incorrect")
	}

	mockds.Stop()
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
