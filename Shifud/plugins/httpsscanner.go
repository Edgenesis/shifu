package main

import (
	"crypto/tls"
	"example.com/Shifud/plugincommon"
	"example.com/Shifud/shifud"
	"net"
	"strconv"
	"time"
)

type HttpsScanner struct{}

func (s *HttpsScanner) Scan() ([]shifud.DeviceConfig, error) {
	startPort := 80
	endPort := 10000
	timeout := 5 * time.Second
	handler := httpsHandler
	devices := plugincommon.WebScanner(startPort, endPort, timeout, handler)
	return devices, nil
}

func httpsHandler(ip string, port int, timeout time.Duration) bool {
	address := ip + ":" + strconv.Itoa(port)
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{InsecureSkipVerify: true})
	if err == nil {
		defer conn.Close()
		return true
	}
	return false
}

var ScannerPlugin HttpsScanner
