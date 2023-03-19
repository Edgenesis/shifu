# Shifud Plugin Development Guide

Shifud is a flexible network scanning tool with an extensible plugin system. This guide will help you create and integrate a new plugin into the Shifud system.
## Understanding the Plugin System

Shifud uses Go's `plugin` package to dynamically load plugins at runtime. Plugins are compiled as shared libraries (`.so` files) and loaded from the `plugins` folder. Each plugin must implement the `ScannerPlugin` interface, which is defined in the `shifud.go` file:

```go
type ScannerPlugin interface {
	Scan() ([]DeviceConfig, error)
}
```


## Creating a New Plugin
1. ** folder.**  Name the file based on your plugin's functionality, e.g., `http_scanner.go`.
2. ** interface.**  For example:

```go
package plugins

import (
	"example.com/Shifud/shifud"
	"fmt"
)

type HttpScanner struct{}

func (s *HttpScanner) Scan() ([]shifud.DeviceConfig, error) {
	fmt.Println("Scanning with HTTP scanner...")
	// Add your plugin logic here
	return nil, nil
}

var ScannerPlugin HttpScanner
```


1. ** function.**  This is where you'll add the core functionality of your plugin. The `Scan()` function should return a slice of `DeviceConfig` objects and an error if something goes wrong.
## Compiling and Loading the Plugin
1. **Compile the plugin as a shared library.**  Use the following command to compile your plugin:

```bash
go build -buildmode=plugin -o plugins/*.so plugins/*.go
```



Replace `your_plugin_name` with the appropriate name for your plugin.
1. **file.**  The `LoadPlugins` function in the `shifud.go` file handles the loading of plugins. It is called in the `main.go` file as follows:

```go
err := shifud.LoadPlugins(scanner, "plugins")
if err != nil {
	fmt.Println("Error loading plugins:", err)
	os.Exit(1)
}
```



With these steps, your new plugin will be loaded and used by Shifud at runtime. You can create additional plugins by following the same process, and Shifud will automatically load and use them.
## Testing Your Plugin

After creating and loading your plugin, run the `main.go` file to test your plugin's functionality. If everything is set up correctly, you should see the output of your plugin's `Scan()` function when Shifud runs.