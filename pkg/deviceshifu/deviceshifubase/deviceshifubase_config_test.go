package deviceshifubase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewShifuConfig(t *testing.T) {

	ds, err := NewDeviceShifuConfig("")
	assert.Nil(t, ds)
	assert.Equal(t, "DeviceShifuConfig path can't be empty", err.Error())

	ds, err2 := NewDeviceShifuConfig("etc/edgedevice/config")

	assert.Nil(t, ds)
	assert.Equal(t, "lstat etc/edgedevice/config: no such file or directory", err2.Error())

}
