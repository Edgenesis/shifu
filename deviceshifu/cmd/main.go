package main

import (
	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifu"
)

func main() {
	ds := &deviceshifu.DeviceShifu{
		Name: "mock-device",
	}
	ds.Start()
}
