package main

import (
	"os"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifusocket"

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

	ds, err := deviceshifusocket.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	if err := ds.Start(wait.NeverStop); err != nil {
		panic(err.Error())
	}

	select {}
}
