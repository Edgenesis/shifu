package deviceshifulwm2m

import (
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/stretchr/testify/assert"
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

	// Call the function under test
	result := CreateLwM2MInstructions(dsInstructions)

	// Assert that the result is not nil
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Instructions))

	// Check if each instruction's properties are correctly mapped
	instruction1 := result.Instructions["instruction1"]
	assert.NotNil(t, instruction1)
	assert.Equal(t, "123", instruction1.ObjectId)
	assert.True(t, instruction1.EnableObserve)

	instruction2 := result.Instructions["instruction2"]
	assert.NotNil(t, instruction2)
	assert.Equal(t, "456", instruction2.ObjectId)
	assert.False(t, instruction2.EnableObserve)
}

func TestCreateLwM2MInstructions_EmptyInstructions(t *testing.T) {
	// Test the case of an empty instruction map
	dsInstructions := &deviceshifubase.DeviceShifuInstructions{
		Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{},
	}

	// Call the function under test
	result := CreateLwM2MInstructions(dsInstructions)

	// Assert that the result is not nil and the instruction map is empty
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Instructions))
}
