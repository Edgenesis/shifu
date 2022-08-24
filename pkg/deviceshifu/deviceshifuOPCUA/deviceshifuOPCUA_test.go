package deviceshifuOPCUA

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func TestNew(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestStartHttpServer",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
		Namespace:      "TestStartHttpServerNamespace",
	}

	_, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}
}

func TestDeviceShifuEmptyNamespace(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestDeviceShifuEmptyNamespace",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
	}

	_, err := New(deviceShifuMetadata)
	if err != nil {
		log.Print(err)
	} else {
		t.Errorf("DeviceShifu Test with empty namespace failed")
	}
}

func TestStart(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestStart",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
		Namespace:      "TestStartNamespace",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	mockds.Start(wait.NeverStop)

	mockds.Stop()
}

func TestDeviceHealthHandler(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestStartHttpServer",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR,
		Namespace:      "TestStartHttpServerNamespace",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	go mockds.startHttpServer(wait.NeverStop)

	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://127.0.0.1:8080/health")
	if err != nil {
		t.Errorf("HTTP GET returns an error %v", err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if string(body) != deviceshifubase.DEVICE_IS_HEALTHY_STR {
		t.Errorf("%+v", body)
	}

	mockds.Stop()

	// cleanup
	t.Cleanup(func() {
		//tear-down code
		err := os.RemoveAll(MOCK_DEVICE_CONFIG_PATH)
		if err != nil {
			log.Fatal(err)
		}
	})
}
