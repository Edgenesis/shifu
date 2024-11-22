package deviceshifulwm2m

import (
	"reflect"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
)

func TestCreateLwM2MInstructions(t *testing.T) {
	testCases := []struct {
		desc           string
		dsInstructions *deviceshifubase.DeviceShifuInstructions
		expected       *LwM2MInstruction
	}{
		{
			desc: "multiple instructions",
			dsInstructions: &deviceshifubase.DeviceShifuInstructions{
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
			},
			expected: &LwM2MInstruction{
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
			},
		},
		{
			desc: "empty instructions",
			dsInstructions: &deviceshifubase.DeviceShifuInstructions{
				Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{},
			},
			expected: &LwM2MInstruction{
				Instructions: map[string]*LwM2MProtocolProperty{},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := CreateLwM2MInstructions(tC.dsInstructions)
			if !reflect.DeepEqual(tC.expected, result) {
				t.Errorf("Case %s: Unexpected result.\nExpected: %+v\nGot: %+v",
					tC.desc, tC.expected, result)
			}
		})
	}
}
