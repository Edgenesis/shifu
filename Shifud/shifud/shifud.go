package shifud

import (
	"fmt"
	"io/ioutil"
	"plugin"
)

type DeviceConfig struct {
	IP     string `yaml:"ip"`
	Port   int    `yaml:"port"`
	Option string `yaml:"option"`
}

type ScannerPlugin interface {
	Scan() ([]DeviceConfig, error)
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

func LoadPlugins(scanner *Scanner, pluginDir string) error {
	files, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && file.Name()[len(file.Name())-3:] == ".so" {
			p, err := plugin.Open(pluginDir + "/" + file.Name())
			if err != nil {
				return err
			}

			symScannerPlugin, err := p.Lookup("ScannerPlugin")
			if err != nil {
				return err
			}

			scannerPlugin, ok := symScannerPlugin.(ScannerPlugin)
			if !ok {
				return fmt.Errorf("unexpected type from module symbol")
			}

			scanner.AddPlugin(scannerPlugin)
		}
	}

	return nil
}
