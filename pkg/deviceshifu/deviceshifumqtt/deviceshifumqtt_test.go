package deviceshifumqtt

import (
	"sync"
	"testing"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestCollectMQTTTelemetry(t *testing.T) {
	// Mock Telemetry Settings
	interval := int64(100) // 100ms
	telemetrySettings := &deviceshifubase.DeviceShifuTelemetrySettings{
		DeviceShifuTelemetryUpdateIntervalInMilliseconds: &interval,
	}

	// Mock Instruction
	instructionName := "cmd_test"
	mqttTopic := "test_topic"

	// Mock Telemetries
	telemetries := map[string]*deviceshifubase.DeviceShifuTelemetry{
		"telemetry1": {
			DeviceShifuTelemetryProperties: deviceshifubase.DeviceShifuTelemetryProperties{
				DeviceInstructionName: &instructionName,
			},
		},
	}

	// Mock DeviceSpec
	protocol := v1alpha1.ProtocolMQTT
	address := "localhost:1883"
	edgeDeviceSpec := v1alpha1.EdgeDeviceSpec{
		Protocol: &protocol,
		Address:  &address,
	}

	// Mock Config
	config := &deviceshifubase.DeviceShifuConfig{
		Telemetries: &deviceshifubase.DeviceShifuTelemetries{
			DeviceShifuTelemetrySettings: telemetrySettings,
			DeviceShifuTelemetries:       telemetries,
		},
	}

	// Mock Instructions
	instructions := &MQTTInstructions{
		Instructions: map[string]*MQTTInstruction{
			instructionName: {
				MQTTProtocolProperty: &MQTTProtocolProperty{
					MQTTTopic: mqttTopic,
				},
			},
		},
	}

	// Initialize DeviceShifu
	ds := &DeviceShifu{
		base: &deviceshifubase.DeviceShifuBase{
			Name:              "test-device",
			EdgeDevice:        &v1alpha1.EdgeDevice{Spec: edgeDeviceSpec},
			DeviceShifuConfig: config,
		},
		mqttInstructions: instructions,
		state: &MQTTState{
			mqttMessageInstructionMap:      make(map[string]string),
			mqttMessageReceiveTimestampMap: make(map[string]time.Time),
			controlMsgs:                    make(map[string]string),
			mu:                             sync.RWMutex{},
		},
	}

	// Case 1: No message received yet -> Should return false
	ok, err := ds.collectMQTTTelemetry()
	assert.Nil(t, err)
	assert.False(t, ok)

	// Case 2: Message received recently -> Should return true
	ds.state.mu.Lock()
	ds.state.mqttMessageReceiveTimestampMap[mqttTopic] = time.Now()
	ds.state.mu.Unlock()

	ok, err = ds.collectMQTTTelemetry()
	assert.Nil(t, err)
	assert.True(t, ok)

	// Case 3: Message received long ago -> Should return false
	ds.state.mu.Lock()
	ds.state.mqttMessageReceiveTimestampMap[mqttTopic] = time.Now().Add(-200 * time.Millisecond)
	ds.state.mu.Unlock()

	ok, err = ds.collectMQTTTelemetry()
	assert.Nil(t, err)
	assert.False(t, ok)
}
