package deviceshifuconfig

import (
	"errors"
	"log"

	"gopkg.in/yaml.v2"
	"knative.dev/pkg/configmap"
)

type DeviceShifuConfig struct {
	driverProperties DeviceShifuDriverProperties
	Instructions     map[string]*DeviceShifuInstruction
	Telemetries      map[string]*DeviceShifuTelemetry
}

type DeviceShifuDriverProperties struct {
	DriverSku   string `yaml:"driverSku"`
	DriverImage string `yaml:"driverImage"`
}

type DeviceShifuInstruction struct {
	DeviceShifuInstructionProperties []DeviceShifuInstructionProperty `yaml:"properties,omitempty"`
}

type DeviceShifuInstructionProperty struct {
	ValueType    string      `yaml:"valueType"`
	ReadWrite    string      `yaml:"readWrite"`
	DefaultValue interface{} `yaml:"defaultValue"`
}

type DeviceShifuTelemetry struct {
	DeviceShifuTelemetryProperties []DeviceShifuTelemetryProperty `yaml:"properties,omitempty"`
}

type DeviceShifuTelemetryProperty struct {
	DeviceInstructionName *string `yaml:"instruction"`
	InitialDelayMs        *int    `yaml:"initialDelayMs,omitempty"`
	IntervalMs            *int    `yaml:"intervalMs,omitempty"`
}

const (
	CM_DRIVERPROPERTIES_STR = "driverProperties"
	CM_INSTRUCTIONS_STR     = "instructions"
	CM_TELEMETRIES_STR      = "telemetries"
)

func New(path string) (*DeviceShifuConfig, error) {
	if path == "" {
		return nil, errors.New("DeviceShifuConfig path can't be empty")
	}

	cfg, err := configmap.Load(path)
	if err != nil {
		return nil, err
	}

	dsc := &DeviceShifuConfig{}
	if driverProperties, ok := cfg[CM_DRIVERPROPERTIES_STR]; ok {
		err := yaml.Unmarshal([]byte(driverProperties), &dsc.driverProperties)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_DRIVERPROPERTIES_STR, err)
			return nil, err
		}
	}

	// TODO: add validation to types and readwrite mode
	if instructions, ok := cfg[CM_INSTRUCTIONS_STR]; ok {
		err := yaml.Unmarshal([]byte(instructions), &dsc.Instructions)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_INSTRUCTIONS_STR, err)
			return nil, err
		}
	}

	if telemetries, ok := cfg[CM_TELEMETRIES_STR]; ok {
		err = yaml.Unmarshal([]byte(telemetries), &dsc.Telemetries)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_TELEMETRIES_STR, err)
			return nil, err
		}
	}
	return dsc, nil
}
