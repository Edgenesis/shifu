package deviceshifuconfig

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	MOCK_DEVICE_CM_STR     = "/workspaces/shifu/deviceshifu/examples/mockdevice/mockdeviceshifu_config/etc/edgedevice/config"
	MOCK_DEVICE_SKU_STR    = "Edgenesis Mock Device"
	MOCK_DEVICE_DRIVER_STR = "edgenesis/mockdevice-0.0.1"
)

var mockDeviceInstructions = map[string]*DeviceShifuInstruction{
	"get_reading": nil,
	"get_status":  nil,
	"set_reading": {
		[]DeviceShifuInstructionProperty{
			{
				"Int32",
				"W",
				nil,
			},
		},
	},
	"start": nil,
	"stop":  nil,
}

var mockDeviceTelemetries = map[string]*DeviceShifuTelemetry{
	"device_health": {
		[]DeviceShifuTelemetryProperty{
			{
				"get_status",
				1000,
				1000,
			},
		},
	},
	"device_random": {
		[]DeviceShifuTelemetryProperty{
			{
				"get_reading",
				1000,
				1000,
			},
		},
	},
}

func TestStart(t *testing.T) {
	mockdsc, err := New(MOCK_DEVICE_CM_STR)
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Printf("Drive image is '%v', SKU is `%v`\n", mockdsc.driverImage, mockdsc.driverSKU)
	if mockdsc.driverSKU != MOCK_DEVICE_SKU_STR {
		t.Errorf("%+v", mockdsc.driverSKU)
	}

	if mockdsc.driverImage != MOCK_DEVICE_DRIVER_STR {
		t.Errorf("%+v", mockdsc.driverImage)
	}

	if len(mockdsc.Instructions) != len(mockDeviceInstructions) {
		t.Errorf("instruction length mismatch!")
	}

	if len(mockdsc.Telemetries) != len(mockDeviceTelemetries) {
		t.Errorf("telemetry length mismatch!")
	}

	eq := reflect.DeepEqual(mockDeviceInstructions, mockdsc.Instructions)
	if !eq {
		t.Errorf("Instruction mismatch")
	}

	eq = reflect.DeepEqual(mockDeviceTelemetries, mockdsc.Telemetries)
	if !eq {
		t.Errorf("Telemetries mismatch")
	}
}
