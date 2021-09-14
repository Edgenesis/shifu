package main

import (
	"fmt"

	"knative.dev/pkg/configmap"
)

const CM_NAME = "etc/edgedevice/config"

func main() {
	cfg, err := configmap.Load(CM_NAME)
	if err != nil {
		panic("Unable to load configmap")
	} else {
		// fmt.Println(cfg)
		for k, v := range cfg {
			fmt.Println(k, "==================")
			fmt.Println(v)

		}
		// body := json.Unmarshal(cfg, &jdata)

		// fmt.Println(jdata)
	}
}
