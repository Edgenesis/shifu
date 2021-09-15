package main

import (
	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifu"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	ds := &deviceshifu.DeviceShifu{
		Name: "mock-device",
	}

	if err := ds.Start(wait.NeverStop); err != nil {
		panic(err.Error())
	}

	select {}
}
