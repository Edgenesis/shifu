package deviceshifuOPCUA

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
)

const OPCUAID = "OPCUANodeID"

type OPCUAInstructions struct {
	Instructions map[string]*OPCUAInstruction
}

type OPCUAInstruction struct {
	OPCUAInstructionProperty *OPCUAInstructionProperty `yaml:"instructionProperties,omitempty"`
}

type OPCUAInstructionProperty struct {
	OPCUANodeID string `yaml:"OPCUANodeID"`
}

func CreateOPCUAInstructions(dsInstructions *deviceshifubase.DeviceShifuInstructions) *OPCUAInstructions {
	instructions := make(map[string]*OPCUAInstruction)

	for key, dsInstruction := range dsInstructions.Instructions {
		instruction := &OPCUAInstruction{
			&OPCUAInstructionProperty{
				OPCUANodeID: dsInstruction.DeviceShifuProtocolProperties[OPCUAID],
			},
		}
		instructions[key] = instruction
	}
	return &OPCUAInstructions{instructions}
}

const (
	OPCUA_INSTRUCTIONS_STR = "opcuaInstructions"
)
