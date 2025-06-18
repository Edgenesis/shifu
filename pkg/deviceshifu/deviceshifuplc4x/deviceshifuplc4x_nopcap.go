//go:build nopcap

package deviceshifuplc4x

import (
	"fmt"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
)

// DeviceShifu deviceshifu stub for environments without pcap
type DeviceShifu struct {
	base *deviceshifubase.DeviceShifuBase
}

// New creates a new PLC4X deviceshifu (stub implementation)
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	return nil, fmt.Errorf("PLC4X support is not available in this build (missing pcap dependencies)")
}

// Start starts the deviceshifu (stub implementation)
func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	return fmt.Errorf("PLC4X support is not available in this build")
}

// Stop stops the deviceshifu (stub implementation)
func (ds *DeviceShifu) Stop() error {
	return fmt.Errorf("PLC4X support is not available in this build")
}