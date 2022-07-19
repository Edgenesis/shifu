package deviceshifuOPCUA

import (
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func TestNew(t *testing.T) {
	deviceShifuMetadata := &DeviceShifuMetaData{
		"TestStartHttpServer",
		"etc/edgedevice/config",
		DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
		"TestStartHttpServerNamespace",
	}

	_, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceShifu")
	}
}

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

	mockds.Start(wait.NeverStop)

	if err := mockds.Stop(); err != nil {
		log.Printf("Error stopping mock deviceShifu, error: %v", err.Error())
	}
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
	if err != nil {
		log.Printf("cannot read HTTP body, error: %v", err.Error())
	}

	if string(body) != DEVICE_IS_HEALTHY_STR {
		t.Errorf("%+v", body)
	}

	if err := mockds.Stop(); err != nil {
		log.Printf("Error stopping mock deviceShifu, error: %v", err.Error())
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
