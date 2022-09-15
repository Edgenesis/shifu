package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCustomizedDataProcessing(t *testing.T) {
	rawData, _ := os.ReadFile("testdata/raw_data")
	expectedProcessed, _ := os.ReadFile("testdata/expected_data")
	processed := ProcessInstruction("customized_handlers", "humidity", string(rawData), "testdata/pythoncustomizedhandlersfortest")
	assert.Equal(t, string(expectedProcessed), processed)
}
