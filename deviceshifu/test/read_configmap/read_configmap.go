package main

import (
	"fmt"

	"knative.dev/pkg/configmap"
	"sigs.k8s.io/yaml"
)

const CM_NAME = "test-mockdevice-configmap.yaml"

func main() {
	// ds := &deviceshifu.DeviceShifu{
	// 	Name: "mock-device",
	// }

	cfg, err := configmap.Load(CM_NAME)
	if err != nil {
		panic("Unable to load configmap")
	} else {
		body, _ := yaml.Marshal(cfg)
		fmt.Println(cfg)
		fmt.Println(body)
	}

	// if err := ds.Start(wait.NeverStop); err != nil {
	// 	panic(err.Error())
	// }

	// select {}

}
