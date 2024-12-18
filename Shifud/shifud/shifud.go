package shifud

import (
	"fmt"
	"io/ioutil"
	"os"
	"plugin"
	"sync"
	"time"

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

type Scanner struct {
	plugins []ScannerPlugin
}

func (s *Scanner) AddPlugin(p ScannerPlugin) {
	s.plugins = append(s.plugins, p)
}

func (s *Scanner) Scan() ([]DeviceConfig, error) {
	var devices []DeviceConfig
	var mu sync.Mutex
	var wg sync.WaitGroup

	fmt.Println("Hold on to your hat, we're starting the scanning party! 🎉")

	// Scan with each plugin in parallel
	for i, p := range s.plugins {
		wg.Add(1)
		go func(plugin ScannerPlugin, index int) {
			defer wg.Done()

			start := time.Now()
			ds, err := plugin.Scan()
			duration := time.Since(start)

			if err != nil {
				fmt.Printf("Oops! Plugin %d stumbled a bit: %v\n", index, err)
			} else {
				fmt.Printf("Plugin %d finished scanning in %v! 🚀\n", index, duration)
			}

			output, err := yaml.Marshal(ds)
			if err != nil {
				fmt.Printf("Plugin %d had trouble packing its bags: %v\n", index, err)
			} else {
				filename := fmt.Sprintf("devices_plugin_%d.yml", index)
				err = os.WriteFile(filename, output, 0644)
				if err != nil {
					fmt.Printf("Plugin %d's suitcase got lost: %v\n", index, err)
				} else {
					fmt.Printf("Plugin %d's results are ready for you in %s! 📦\n", index, filename)
				}
			}

			mu.Lock()
			devices = append(devices, ds...)
			mu.Unlock()
		}(p, i)
	}

	wg.Wait()

	fmt.Println("The scanning party is over. Thanks for joining! 🥳")

	return devices, nil
}

func LoadPlugins(scanner *Scanner, pluginDir string) error {
	files, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	fmt.Println("Knocking on the plugins' doors... 🚪")

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

			fmt.Printf("Plugin %s is ready to join the fun! 🎊\n", file.Name())
		}
	}

	fmt.Println("All the plugins are on board! Let's get this party started! 🕺")

	return nil
}
