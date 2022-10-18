package main

import "github.com/edgenesis/shifu/pkg/telemetryservice/mqtt"

func main() {
	stop := make(chan struct{})
	mqtt.New(stop)
}
