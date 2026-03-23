package deviceapi

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
)

// parseInstructionsFromConfigMap parses the instructions key from a ConfigMap.
func parseInstructionsFromConfigMap(cm *corev1.ConfigMap) ([]Interaction, error) {
	instructionsYAML, ok := cm.Data["instructions"]
	if !ok {
		return nil, nil
	}

	var parsed deviceshifubase.DeviceShifuInstructions
	if err := yaml.Unmarshal([]byte(instructionsYAML), &parsed); err != nil {
		return nil, fmt.Errorf("parsing instructions YAML: %w", err)
	}

	var interactions []Interaction
	for name, instr := range parsed.Instructions {
		interaction := Interaction{Name: name}
		if instr != nil {
			interaction.Description = strings.TrimSpace(instr.Description)
			interaction.ReadWrite = instr.ReadWrite
			interaction.Safe = instr.Safe
		}
		interactions = append(interactions, interaction)
	}

	return interactions, nil
}
