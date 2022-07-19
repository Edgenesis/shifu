package main

import (
	"os"

	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifuSocket"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	deviceName := os.Getenv("EDGEDEVICE_NAME")
	namespace := os.Getenv("EDGEDEVICE_NAMESPACE")

	deviceShifuMetadata := &deviceshifuSocket.DeviceShifuMetaData{
		Name:           deviceName,
		ConfigFilePath: deviceshifuSocket.DEVICE_CONFIGMAP_FOLDER_PATH,
		KubeConfigPath: deviceshifuSocket.KUBERNETES_CONFIG_DEFAULT,
		Namespace:      namespace,
	}

	ds, err := deviceshifuSocket.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	if err := ds.Start(wait.NeverStop); err != nil {
		panic(err.Error())
	}

	select {}
}
