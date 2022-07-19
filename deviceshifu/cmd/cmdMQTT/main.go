package main

import (
	"os"

	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifuMQTT"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	deviceName := os.Getenv("EDGEDEVICE_NAME")
	namespace := os.Getenv("EDGEDEVICE_NAMESPACE")

	deviceShifuMetadata := &deviceshifuMQTT.DeviceShifuMetaData{
		Name:           deviceName,
		ConfigFilePath: deviceshifuMQTT.DEVICE_CONFIGMAP_FOLDER_PATH,
		KubeConfigPath: deviceshifuMQTT.KUBERNETES_CONFIG_DEFAULT,
		Namespace:      namespace,
	}

	ds, err := deviceshifuMQTT.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	ds.Start(wait.NeverStop)

	select {}
}
