package deviceshifulwm2m

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/logger"
)

const (
	objectIdStr      string = "ObjectId"
	enableObserveStr string = "EnableObserve"
)

type LwM2MInstruction struct {
	Instructions map[string]*LwM2MProtocolProperty
}

type LwM2MProtocolProperty struct {
	EnableObserve bool
	ObjectId      string
}

func CreateLwM2MInstructions(dsInstructions *deviceshifubase.DeviceShifuInstructions) *LwM2MInstruction {
	instructions := LwM2MInstruction{
		Instructions: make(map[string]*LwM2MProtocolProperty),
	}
	for key, dsInstruction := range dsInstructions.Instructions {
		if dsInstruction.DeviceShifuProtocolProperties != nil && dsInstruction.DeviceShifuProtocolProperties[objectIdStr] == "" {
			logger.Fatalf("Error when Read ObjectId From DeviceShifuInstructions, error: instruction %v has an empty object id", key)
		}
		if dsInstruction.DeviceShifuProtocolProperties != nil && dsInstruction.DeviceShifuProtocolProperties[enableObserveStr] == "" {
			logger.Fatalf("Error when Read EnableObserve From DeviceShifuInstructions, error: instruction %v has an empty enable observe", key)
		}
		instructions.Instructions[objectIdStr] = &LwM2MProtocolProperty{
			ObjectId:      dsInstruction.DeviceShifuProtocolProperties[objectIdStr],
			EnableObserve: dsInstruction.DeviceShifuProtocolProperties[enableObserveStr] == "true",
		}
	}
	return &instructions
}
