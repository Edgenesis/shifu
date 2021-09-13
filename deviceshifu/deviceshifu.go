package deviceshifu

import (
	"fmt"
)

type DeviceShifu struct {
	Name string
}

func (ds *DeviceShifu) Start() {
	fmt.Println("%s", ds.Name)
}