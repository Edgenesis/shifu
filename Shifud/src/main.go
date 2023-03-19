package main

import (
	"fmt"
	"net"
	"os"

	"sigs.k8s.io/yaml"
)

type DeviceConfig struct {
	IP     string `yaml:"ip"`
	Port   int    `yaml:"port"`
	Option string `yaml:"option"`
}

type ScannerPlugin interface {
	Scan() ([]DeviceConfig, error)
}

type LocalNetworkScanner struct{}

func (s *LocalNetworkScanner) Scan() ([]DeviceConfig, error) {
	var devices []DeviceConfig

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ip, ok := addr.(*net.IPNet)
		if ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			config := DeviceConfig{
				IP:     ip.IP.String(),
				Port:   8080,
				Option: "default",
			}
			devices = append(devices, config)
		}
	}

	return devices, nil
}

type Scanner struct {
	plugins []ScannerPlugin
}

func (s *Scanner) AddPlugin(p ScannerPlugin) {
	s.plugins = append(s.plugins, p)
}

func (s *Scanner) Scan() ([]DeviceConfig, error) {
	var devices []DeviceConfig

	// Scan with each plugin
	for _, p := range s.plugins {
		ds, err := p.Scan()
		if err != nil {
			return nil, err
		}
		devices = append(devices, ds...)
	}

	return devices, nil
}

func main() {
	fmt.Println("Shifud online...")
	// Create a scanner with the default plugin
	scanner := &Scanner{}
	scanner.AddPlugin(&LocalNetworkScanner{})

	// Add plugins for other scanning locations, e.g.:
	// scanner.AddPlugin(&HttpNetworkScanner{})
	// scanner.AddPlugin(&UsbScanner{})

	// Scan devices with all plugins
	devices, err := scanner.Scan()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Export the configuration settings to a YAML file
	output, err := yaml.Marshal(devices)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	err = os.WriteFile("devices.yml", output, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
