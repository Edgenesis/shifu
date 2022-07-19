package main

import (
	"os"

	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifu"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	deviceName := os.Getenv("EDGEDEVICE_NAME")
	namespace := os.Getenv("EDGEDEVICE_NAMESPACE")

	deviceShifuMetadata := &deviceshifu.DeviceShifuMetaData{
		Name:           deviceName,
		ConfigFilePath: deviceshifu.DEVICE_CONFIGMAP_FOLDER_PATH,
		KubeConfigPath: deviceshifu.KUBERNETES_CONFIG_DEFAULT,
		Namespace:      namespace,
	}

	ds, err := deviceshifu.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	ds.Start(wait.NeverStop)
	select {}
}
