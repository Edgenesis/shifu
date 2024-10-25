package deviceshifulwm2m

import (
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
)

func TestCreateLwM2MInstructions(t *testing.T) {
	tests := []struct {
		InstructionName string
		input           *deviceshifubase.DeviceShifuInstructions
		expected        *LwM2MInstruction
		expectingPanic  bool
	}{
		{
			InstructionName: "Valid instructions and enableObserveStr",
			input: &deviceshifubase.DeviceShifuInstructions{
				Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{
					"instruction": {
						DeviceShifuProtocolProperties: map[string]string{
							objectIdStr:      "1",
							enableObserveStr: "true",
						},
					},
				},
			},
			expected: &LwM2MInstruction{
				Instructions: map[string]*LwM2MProtocolProperty{
					"instruction": {
						ObjectId:      "1",
						EnableObserve: true,
					},
				},
			},
			expectingPanic: false,
		},
		{
			InstructionName: "Valid instructions and disableObserveStr",
			input: &deviceshifubase.DeviceShifuInstructions{
				Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{
					"instruction": {
						DeviceShifuProtocolProperties: map[string]string{
							objectIdStr:      "1",
							enableObserveStr: "false",
						},
					},
				},
			},
			expected: &LwM2MInstruction{
				Instructions: map[string]*LwM2MProtocolProperty{
					"instruction": {
						ObjectId:      "1",
						EnableObserve: false,
					},
				},
			},
			expectingPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.InstructionName, func(t *testing.T) {
			// CreateLwM2MInstructions's logger.Fatalf will end program with no panic
			res := CreateLwM2MInstructions(tt.input)
			if !tt.expectingPanic && !compareLwM2MInstructions(res, tt.expected) {
				t.Errorf("Test case %s failed: expected %v, got %v", tt.InstructionName, tt.expected, res)
			}
		})
	}
}

func compareLwM2MInstructions(a, b *LwM2MInstruction) bool {
	if len(a.Instructions) != len(b.Instructions) {
		return false
	}
	for key, valA := range a.Instructions {
		valB, ok := b.Instructions[key]
		if !ok || valA.ObjectId != valB.ObjectId || valA.EnableObserve != valB.EnableObserve {
			return false
		}
	}
	return true
}
