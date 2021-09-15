package deviceshifuconfig

import (
	"fmt"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
	"knative.dev/pkg/configmap"
)

type DeviceShifuConfig struct {
	DriverImage string
	DriverSKU   string
	Instruction []DeviceShifuInstruction
	Telemetry   []DeviceShifuTelemetry
}

type DeviceShifuInstruction struct {
	Name       string `yaml:"name"`
	Properties []struct {
		ValueType    string      `yaml:"valueType"`
		ReadWrite    string      `yaml:"readWrite"`
		DefaultValue interface{} `yaml:"defaultValue"`
	} `yaml:"properties,omitempty"`
}

type DeviceShifuTelemetry struct {
	Name       string `yaml:"name"`
	Properties []struct {
		Instruction    string `yaml:"instruction"`
		InitialDelayMs int    `yaml:"initialDelayMs"`
		IntervalMs     int    `yaml:"intervalMs"`
	} `yaml:"properties"`
}

const (
	CM_DRIVERIMAGE_NAME = "driverProperties.driverImage"
	CM_DRIVERSKU_NAME   = "driverProperties.driverSku"
	CM_INSTRUCTION_NAME = "instructions"
	CM_TELEMETRY_NAME   = "telemetries"
)

func New(path string) *DeviceShifuConfig {
	if path == "" {
		fmt.Println("DeviceShifuConfig path can't be empty")
		return nil
	}

	cfg, err := configmap.Load(path)
	fmt.Println("Path is:", path)

	if err != nil {
		log.Fatalf("Unable to load configmap, error: %v", err)
		return nil
	} else {
		dsc := DeviceShifuConfig{}
		dsc.DriverImage = strings.TrimSpace(cfg[CM_DRIVERIMAGE_NAME])
		dsc.DriverSKU = strings.TrimSpace(cfg[CM_DRIVERSKU_NAME])

		err := yaml.Unmarshal([]byte(cfg[CM_INSTRUCTION_NAME]), &dsc.Instruction)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_INSTRUCTION_NAME, err)
		}

		err = yaml.Unmarshal([]byte(cfg[CM_TELEMETRY_NAME]), &dsc.Telemetry)
		if err != nil {
			log.Fatalf("Error parsing %v from ConfigMap, error: %v", CM_TELEMETRY_NAME, err)
		}

		return &dsc
	}
}
