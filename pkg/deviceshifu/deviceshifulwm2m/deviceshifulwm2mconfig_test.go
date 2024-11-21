package deviceshifulwm2m

import (
	"reflect"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
)

func TestCreateLwM2MInstructions(t *testing.T) {
	// Initialize test data
	dsInstructions := &deviceshifubase.DeviceShifuInstructions{
		Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{
			"instruction1": {
				DeviceShifuProtocolProperties: map[string]string{
					objectIdStr:      "123",
					enableObserveStr: "true",
				},
			},
			"instruction2": {
				DeviceShifuProtocolProperties: map[string]string{
					objectIdStr:      "456",
					enableObserveStr: "false",
				},
			},
		},
	}

	// Expected result
	expectedResult := &LwM2MInstruction{
		Instructions: map[string]*LwM2MProtocolProperty{
			"instruction1": {
				ObjectId:      "123",
				EnableObserve: true,
			},
			"instruction2": {
				ObjectId:      "456",
				EnableObserve: false,
			},
		},
	}

	// Call the function under test
	result := CreateLwM2MInstructions(dsInstructions)

	// Assert that the result matches the expected result using reflect.DeepEqual
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Unexpected result. Expected: %+v, Got: %+v", expectedResult, result)
	}
}

func TestCreateLwM2MInstructions_EmptyInstructions(t *testing.T) {
	// Test the case of an empty instruction map
	dsInstructions := &deviceshifubase.DeviceShifuInstructions{
		Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{},
	}

	// Expected result
	expectedResult := &LwM2MInstruction{
		Instructions: map[string]*LwM2MProtocolProperty{},
	}

	// Call the function under test
	result := CreateLwM2MInstructions(dsInstructions)

	// Assert that the result matches the expected result using reflect.DeepEqual
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Unexpected result. Expected: %+v, Got: %+v", expectedResult, result)
	}
}
