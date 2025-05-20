package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/edgenesis/shifu/pkg/gateway/natsio"
	"github.com/edgenesis/shifu/pkg/logger"
)

func main() {
	client, err := natsio.New()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Client created")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := client.Start(); err != nil {
			logger.Errorf("Error starting client: %v", err)
		}
	}()
	<-sigs
	client.ShutDown()
	logger.Info("Client shutdown")
}
