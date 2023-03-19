package plugins

import (
	"example.com/Shifud/shifud"
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
	devices := WebScanner(startPort, endPort, timeout, handler)
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
			return true
		}
	}
	return false
}

var ScannerPlugin HttpScanner
