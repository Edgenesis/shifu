package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/edgenesis/shifu/pkg/gateway/gatewaylwm2m"
	"github.com/edgenesis/shifu/pkg/logger"
)

func main() {
	client, err := gatewaylwm2m.New()
	if err != nil {
		logger.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := client.Start(); err != nil {
			logger.Errorf("Error starting client: %v", err)
		}
	}()
	<-sigs
	client.ShutDown()
}
