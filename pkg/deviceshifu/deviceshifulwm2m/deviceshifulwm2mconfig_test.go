package deviceshifulwm2m

import (
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/stretchr/testify/assert"
)

func TestCreateLwM2MInstructions(t *testing.T) {
	tests := []struct {
		name           string
		input          *deviceshifubase.DeviceShifuInstructions
		expected       *LwM2MInstruction
		expectingPanic bool
	}{
		{
			name: "Valid Instructions",
			input: &deviceshifubase.DeviceShifuInstructions{
				Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{
					"instruction1": {
						DeviceShifuProtocolProperties: map[string]string{
							objectIdStr:      "1",
							enableObserveStr: "true",
						},
					},
				},
			},
			expected: &LwM2MInstruction{
				Instructions: map[string]*LwM2MProtocolProperty{
					"instruction1": {
						ObjectId:      "1",
						EnableObserve: true,
					},
				},
			},
			expectingPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectingPanic {
				assert.Panics(t, func() { CreateLwM2MInstructions(tt.input) })
			} else {
				result := CreateLwM2MInstructions(tt.input)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
