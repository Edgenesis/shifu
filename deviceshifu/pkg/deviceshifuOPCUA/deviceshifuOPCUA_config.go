package deviceshifuOPCUA

import (
	"errors"
	"gopkg.in/yaml.v3"
	"knative.dev/pkg/configmap"
	"log"
)

type OPCUAInstructions struct {
	Instructions map[string]*OPCUAInstruction
}

type OPCUAInstruction struct {
	OPCUAInstructionProperties *OPCUAInstructionProperty `yaml:"instructionProperties,omitempty"`
}

type OPCUAInstructionProperty struct {
	OPCUANodeID string `yaml:"OPCUANodeID"`
}

const (
	OPCUA_INSTRUCTIONS_STR = "opcuaInstructions"
)

// Read the configuration under the path directory and return configuration
func NewOPCUAInstructions(path string) (*OPCUAInstructions, error) {
	if path == "" {
		return nil, errors.New("DeviceShifuConfig path can't be empty")
	}

	cfg, err := configmap.Load(path)
	if err != nil {
		return nil, err
	}

	dsc := &OPCUAInstructions{}
	// TODO: add validation to types and readwrite mode
	if instructions, ok := cfg[OPCUA_INSTRUCTIONS_STR]; ok {
		err := yaml.Unmarshal([]byte(instructions), &dsc.Instructions)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", OPCUA_INSTRUCTIONS_STR, err)
			return nil, err
		}
	}

	return dsc, nil
}
