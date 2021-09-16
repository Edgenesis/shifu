package deviceshifuconfig

import (
	"errors"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
	"knative.dev/pkg/configmap"
)

type DeviceShifuConfig struct {
	driverImage  string
	driverSKU    string
	Instructions map[string]*DeviceShifuInstruction
	Telemetries  map[string]*DeviceShifuTelemetry
}

type DeviceShifuInstruction struct {
	Properties []DeviceShifuInstructionProperty `yaml:"properties,omitempty"`
}

type DeviceShifuInstructionProperty struct {
	ValueType    string      `yaml:"valueType"`
	ReadWrite    string      `yaml:"readWrite"`
	DefaultValue interface{} `yaml:"defaultValue"`
}

type DeviceShifuTelemetry struct {
	Properties []DeviceShifuTelemetryProperty `yaml:"properties"`
}

type DeviceShifuTelemetryProperty struct {
	Instruction    string `yaml:"instruction"`
	InitialDelayMs int    `yaml:"initialDelayMs"`
	IntervalMs     int    `yaml:"intervalMs"`
}

const (
	CM_DRIVERIMAGE_STR  = "driverProperties.driverImage"
	CM_DRIVERSKU_STR    = "driverProperties.driverSku"
	CM_INSTRUCTIONS_STR = "instructions"
	CM_TELEMETRIES_STR  = "telemetries"
)

func New(path string) (*DeviceShifuConfig, error) {
	if path == "" {
		return nil, errors.New("DeviceShifuConfig path can't be empty")
	}

	cfg, err := configmap.Load(path)
	if err != nil {
		return nil, err
	} else {
		dsc := &DeviceShifuConfig{}
		if val, ok := cfg[CM_DRIVERIMAGE_STR]; ok {
			dsc.driverImage = strings.TrimSpace(val)
		}

		if val, ok := cfg[CM_DRIVERSKU_STR]; ok {
			dsc.driverSKU = strings.TrimSpace(val)
		}

		if val, ok := cfg[CM_INSTRUCTIONS_STR]; ok {
			err := yaml.Unmarshal([]byte(val), &dsc.Instructions)
			if err != nil {
				log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_INSTRUCTIONS_STR, err)
				return nil, err
			}
		}

		if val, ok := cfg[CM_TELEMETRIES_STR]; ok {
			err = yaml.Unmarshal([]byte(val), &dsc.Telemetries)
			if err != nil {
				log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_TELEMETRIES_STR, err)
				return nil, err
			}
		}

		return dsc, nil
	}
}
