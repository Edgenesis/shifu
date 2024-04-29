package main

import (
	"github.com/edgenesis/shifu/pkg/gateway/gatewaylwm2m"
	"github.com/edgenesis/shifu/pkg/logger"
)

func main() {
	client, err := gatewaylwm2m.New()
	if err != nil {
		logger.Fatal(err)
	}

	err = client.LoadCfg()
	if err != nil {
		logger.Fatal(err)
	}

	panic(client.Start())
}
