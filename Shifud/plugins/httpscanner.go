package main

import (
	"example.com/Shifud/plugincommon"
	"example.com/Shifud/shifud"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

type HttpScanner struct{}

func (s *HttpScanner) Scan() ([]shifud.DeviceConfig, error) {
	startPort := 80
	endPort := 10000
	timeout := 5 * time.Second
	handler := httpHandler
	log.Println("Scanning for HTTP devices in the wild... Let's see what we find!")
	devices := plugincommon.WebScanner(startPort, endPort, timeout, handler)
	log.Printf("Finished scanning HTTP devices! Found %d devices. Time to party! 🥳\n", len(devices))
	return devices, nil
}

func httpHandler(ip string, port int, timeout time.Duration) bool {
	address := ip + ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err == nil {
		defer conn.Close()

		// Check if it's an HTTP server
		_, err := http.Get("http://" + address)
		if err == nil {
			log.Printf("HTTP server spotted at %s! 🎯\n", address)
			return true
		}
	}
	return false
}

var ScannerPlugin HttpScanner
