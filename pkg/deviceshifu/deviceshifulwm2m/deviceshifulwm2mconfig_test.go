package deviceshifulwm2m

import (
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
)

func TestCreateLwM2MInstructions(t *testing.T) {
	tests := []struct {
		name          string
		input         *deviceshifubase.DeviceShifuInstructions
		expectedError bool
		expected      *LwM2MInstruction
	}{
		{
			name: "Valid instructions",
			input: &deviceshifubase.DeviceShifuInstructions{
				Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{
					"instruction1": {
						DeviceShifuProtocolProperties: map[string]string{
							objectIdStr:      "123",
							enableObserveStr: "true",
						},
					},
				},
			},
			expectedError: false,
			expected: &LwM2MInstruction{
				Instructions: map[string]*LwM2MProtocolProperty{
					"instruction1": {
						ObjectId:      "123",
						EnableObserve: true,
					},
				},
			},
		},
		{
			name: "Missing ObjectId",
			input: &deviceshifubase.DeviceShifuInstructions{
				Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{
					"instruction1": {
						DeviceShifuProtocolProperties: map[string]string{
							enableObserveStr: "true",
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name: "Missing EnableObserve",
			input: &deviceshifubase.DeviceShifuInstructions{
				Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{
					"instruction1": {
						DeviceShifuProtocolProperties: map[string]string{
							objectIdStr: "123",
						},
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectedError {
						t.Errorf("unexpected error: %v", r)
					}
				}
			}()

			result := CreateLwM2MInstructions(tt.input)
			if !tt.expectedError && result != nil {
				for key, prop := range result.Instructions {
					expectedProp, exists := tt.expected.Instructions[key]
					if !exists {
						t.Errorf("unexpected instruction key: %v", key)
					}
					if prop.ObjectId != expectedProp.ObjectId || prop.EnableObserve != expectedProp.EnableObserve {
						t.Errorf("unexpected instruction properties for key %v: got %v, want %v", key, prop, expectedProp)
					}
				}
			}
		})
	}
}
