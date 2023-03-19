package main

import (
	"example.com/Shifud/shifud"
	"fmt"
	_ "net"
	"os"
	"sigs.k8s.io/yaml"
)

func main() {
	fmt.Println("Shifud online...")

	// Create a scanner
	scanner := &shifud.Scanner{}

	// Load plugins from the plugins folder
	err := shifud.LoadPlugins(scanner, "plugins")
	if err != nil {
		fmt.Println("Error loading plugins:", err)
		os.Exit(1)
	}

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
