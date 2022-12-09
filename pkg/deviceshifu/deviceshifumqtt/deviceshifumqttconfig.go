package deviceshifumqtt


import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
)

const (
	mqttTopic = "MQTTTopic"
)

// ReturnBody Body of mqtt's reply
type ReturnBody struct {
	MQTTMessage   string `json:"mqtt_message"`
	MQTTTimestamp string `json:"mqtt_receive_timestamp"`
}

// RequestBody Body of mqtt's request by POST method
type RequestBody string


// MQTTInstructions MQTT Instructions
type MQTTInstructions struct {
	Instructions map[string]*MQTTInstruction
}

// MQTTInstruction MQTT Instruction
type MQTTInstruction struct {
	MQTTInstructionProperty *MQTTInstructionProperty `yaml:"instructionProperties,omitempty"`
}

// MQTTInstructionProperty MQTT Instruction's Property
type MQTTInstructionProperty struct {
	MQTTTopic string `yaml:"MQTTTopic"`
}

// CreateMQTTInstructions Create MQTT Instructions
func CreateMQTTInstructions(dsInstructions *deviceshifubase.DeviceShifuInstructions) *MQTTInstructions {
	instructions := make(map[string]*MQTTInstruction)

	for key, dsInstruction := range dsInstructions.Instructions {
		instruction := &MQTTInstruction{
			&MQTTInstructionProperty{
				MQTTTopic: dsInstruction.DeviceShifuProtocolProperties[mqttTopic],
			},
		}
		instructions[key] = instruction
	}
	return &MQTTInstructions{instructions}
}
