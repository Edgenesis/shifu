package deviceshifu

import (
	"io"
	"net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func TestStart(t *testing.T) {
	mockds := New("TestStart")

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}
}
func TestDeviceHealthHandler(t *testing.T) {
	mockds := New("TestStartHttpServer")

	go mockds.startHttpServer(wait.NeverStop)

	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://127.0.0.1:8080")
	if err != nil {
		t.Errorf("HTTP GET returns an error %v", err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if string(body) != DEVICEISHEALTHYSTR {
		t.Errorf("%+v", body)
	}
}
