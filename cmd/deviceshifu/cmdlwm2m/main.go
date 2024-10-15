package main

import (
	"os"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifulwm2m"
	"github.com/edgenesis/shifu/pkg/logger"

	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	deviceName := os.Getenv("EDGEDEVICE_NAME")
	namespace := os.Getenv("EDGEDEVICE_NAMESPACE")

	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           deviceName,
		ConfigFilePath: deviceshifubase.DeviceConfigmapFolderPath,
		KubeConfigPath: deviceshifubase.KubernetesConfigDefault,
		Namespace:      namespace,
	}

	ds, err := deviceshifulwm2m.New(deviceShifuMetadata)
	if err != nil {
		logger.Fatalf("Error creating deviceshifu: %v", err)
	}

	if err := ds.Start(wait.NeverStop); err != nil {
		logger.Fatalf("Error starting deviceshifu: %v", err)
	}

	select {}
}
