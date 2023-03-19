package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanner(t *testing.T) {
	scanner := &Scanner{}
	scanner.AddPlugin(&LocalNetworkScannerMock{})

	devices, err := scanner.Scan()
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	assert.Len(t, devices, 2)

	expectedDevices := []DeviceConfig{
		{
			IP:     "192.168.1.2",
			Port:   8080,
			Option: "default",
		},
		{
			IP:     "192.168.1.3",
			Port:   8080,
			Option: "default",
		},
	}

	assert.ElementsMatch(t, expectedDevices, devices)
}

type LocalNetworkScannerMock struct{}

func (s *LocalNetworkScannerMock) Scan() ([]DeviceConfig, error) {
	devices := []DeviceConfig{
		{
			IP:     "192.168.1.2",
			Port:   8080,
			Option: "default",
		},
		{
			IP:     "192.168.1.3",
			Port:   8080,
			Option: "default",
		},
	}

	return devices, nil
}

func TestLocalNetworkScanner(t *testing.T) {
	scanner := &LocalNetworkScanner{}
	devices, err := scanner.Scan()

	assert.NoError(t, err)
	assert.NotNil(t, devices)
	assert.Len(t, devices, 2)

	expectedDevices := []DeviceConfig{
		{
			IP:     "192.168.1.2",
			Port:   8080,
			Option: "default",
		},
		{
			IP:     "192.168.1.3",
			Port:   8080,
			Option: "default",
		},
	}

	assert.ElementsMatch(t, expectedDevices, devices)
}
