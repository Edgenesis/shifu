package plugins

import (
	"example.com/Shifud/shifud"
	"fmt"
	"sync"
	"time"
)

type ProtocolHandler func(ip string, port int, timeout time.Duration) bool

func WebScanner(startPort, endPort int, timeout time.Duration, handler ProtocolHandler) []shifud.DeviceConfig {
	var devices []shifud.DeviceConfig
	var wg sync.WaitGroup

	for i := 1; i <= 254; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i)
		wg.Add(1)

		go func(ip string) {
			defer wg.Done()

			for port := startPort; port <= endPort; port++ {
				if handler(ip, port, timeout) {
					devices = append(devices, shifud.DeviceConfig{
						IP:   ip,
						Port: port,
					})
				}
			}
		}(ip)
	}

	wg.Wait()
	return devices
}
