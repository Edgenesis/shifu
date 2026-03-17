package deviceapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseInstructionsFromConfigMap(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "deviceshifu",
		},
		Data: map[string]string{
			"instructions": `instructions:
  read_temp:
    readWrite: R
    safe: true
    description: |
      GET /read_temp
      Returns temperature in Celsius.
  write_config:
    readWrite: W
    safe: false
    description: |
      POST /write_config with JSON body.
  legacy_instruction:
`,
		},
	}

	interactions, err := parseInstructionsFromConfigMap(cm)
	require.NoError(t, err)
	require.Len(t, interactions, 3)

	interactionMap := make(map[string]Interaction)
	for _, intr := range interactions {
		interactionMap[intr.Name] = intr
	}

	readTemp := interactionMap["read_temp"]
	assert.Equal(t, "R", readTemp.ReadWrite)
	assert.NotNil(t, readTemp.Safe)
	assert.True(t, *readTemp.Safe)
	assert.Contains(t, readTemp.Description, "GET /read_temp")

	writeConfig := interactionMap["write_config"]
	assert.Equal(t, "W", writeConfig.ReadWrite)
	assert.NotNil(t, writeConfig.Safe)
	assert.False(t, *writeConfig.Safe)

	// Legacy instruction without extended fields.
	legacy := interactionMap["legacy_instruction"]
	assert.Equal(t, "", legacy.ReadWrite)
	assert.Nil(t, legacy.Safe)
	assert.Equal(t, "", legacy.Description)
}

func TestParseInstructionsFromConfigMapNoInstructionsKey(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "deviceshifu",
		},
		Data: map[string]string{
			"driverProperties": "driverSku: testSku",
		},
	}

	interactions, err := parseInstructionsFromConfigMap(cm)
	require.NoError(t, err)
	assert.Nil(t, interactions)
}

func TestParseInstructionsFromConfigMapEmpty(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "deviceshifu",
		},
		Data: map[string]string{
			"instructions": `instructions:
`,
		},
	}

	interactions, err := parseInstructionsFromConfigMap(cm)
	require.NoError(t, err)
	assert.Empty(t, interactions)
}
