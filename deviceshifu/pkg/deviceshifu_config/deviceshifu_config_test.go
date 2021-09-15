package deviceshifuconfig

import (
	"fmt"
	"testing"
)

const (
	CM_NAME                                = "etc/edgedevice/config"
	SKU_STR                                = "Edgenesis Mock Device"
	DRIVER_STR                             = "edgenesis/mockdevice-0.0.1"
	INSTRUCTION_SET_READING_NAME           = "set_reading"
	INSTRUCTION_SET_READING_VALUETYPE      = "Int32"
	INSTRUCTION_SET_READING_READWRITE      = "W"
	TELEMETRY_DEVICE_HEALTH_NAME           = "device_health"
	TELEMETRY_DEVICE_HEALTH_INSTRUCTION    = "get_status"
	TELEMETRY_DEVICE_HEALTH_INITIALDELAYMS = 1000
	TELEMETRY_DEVICE_HEALTH_INTERVALMS     = 1000
)

var functions = []string{"get_reading", "get_status", "set_reading", "start", "stop"}
var telemetries = []string{"device_health", "device_random"}

func TestStart(t *testing.T) {
	mockdsc := New(CM_NAME)
	fmt.Println(mockdsc.DriverImage, mockdsc.DriverSKU)
	if mockdsc.DriverSKU != SKU_STR {
		t.Errorf("%+v", mockdsc.DriverSKU)
	}

	if mockdsc.DriverImage != DRIVER_STR {
		t.Errorf("%+v", mockdsc.DriverImage)
	}

	if len(mockdsc.Instruction) != len(functions) {
		t.Errorf("instruction length mismatch!")
	}

	if len(mockdsc.Telemetry) != len(telemetries) {
		t.Errorf("instruction length mismatch!")
	}

	for _, v := range mockdsc.Instruction {
		inArray := false
		for _, i := range functions {
			if v.Name == i {
				if v.Name == INSTRUCTION_SET_READING_NAME {
					for _, j := range v.Properties {
						if j.ValueType != INSTRUCTION_SET_READING_VALUETYPE {
							t.Errorf("Instruction %v valuetype incorrect: %v", INSTRUCTION_SET_READING_NAME, j.ValueType)
						}

						if j.ReadWrite != INSTRUCTION_SET_READING_READWRITE {
							t.Errorf("Instruction %v readwrite incorrect: %v", INSTRUCTION_SET_READING_NAME, j.ReadWrite)
						}

						if j.DefaultValue != nil {
							t.Errorf("Instruction %v readwrite incorrect: %v", INSTRUCTION_SET_READING_NAME, j.DefaultValue)
						}
					}
				}
				inArray = true
				break
			}
		}

		if inArray != true {
			t.Errorf("Key %v not in instruction", v.Name)
		}
	}

	for _, v := range mockdsc.Telemetry {
		inArray := false
		for _, i := range telemetries {
			if v.Name == i {
				if v.Name == TELEMETRY_DEVICE_HEALTH_NAME {
					for _, j := range v.Properties {
						if j.Instruction != TELEMETRY_DEVICE_HEALTH_INSTRUCTION {
							t.Errorf("Instruction %v valuetype incorrect: %v", INSTRUCTION_SET_READING_NAME, j.Instruction)
						}

						if j.InitialDelayMs != TELEMETRY_DEVICE_HEALTH_INITIALDELAYMS {
							t.Errorf("Instruction %v readwrite incorrect: %v", INSTRUCTION_SET_READING_NAME, j.InitialDelayMs)
						}

						if j.IntervalMs != TELEMETRY_DEVICE_HEALTH_INTERVALMS {
							t.Errorf("Instruction %v readwrite incorrect: %v", INSTRUCTION_SET_READING_NAME, j.IntervalMs)
						}
					}
				}
				inArray = true
				break
			}
		}

		if inArray != true {
			t.Errorf("Key %v not in telemetris", v.Name)
		}
	}
}
