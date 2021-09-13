package deviceshifu

import (
	"fmt"
)

type DeviceShifu struct {
	Name string
}

func (ds *DeviceShifu) Start() error {
	fmt.Println(ds.Name)

	return nil
}
