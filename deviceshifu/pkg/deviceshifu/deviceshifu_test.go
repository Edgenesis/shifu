package deviceshifu

import (
	"testing"
)

func TestStart(t *testing.T) {
	mockds := &DeviceShifu{
		Name: "test",
	}

	if err := mockds.Start(); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}
}
