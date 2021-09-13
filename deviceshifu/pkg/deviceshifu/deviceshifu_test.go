package deviceshifu

import (
	"net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func TestStart(t *testing.T) {
	mockds := &DeviceShifu{
		Name: "test",
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}
}
func TestStartHttpServer(t *testing.T) {
	mockds := &DeviceShifu{
		Name: "test",
	}

	go mockds.startHttpServer(wait.NeverStop)

	time.Sleep(1 * time.Second)

	_, err := http.Get("http://127.0.0.1:8000")
	if err != nil {
		t.Errorf("getInfo returns an error %v", err.Error())
	}
}
