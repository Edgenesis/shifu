package main

import (
	"os"

	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifuOPCUA"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	deviceName := os.Getenv("EDGEDEVICE_NAME")
	namespace := os.Getenv("EDGEDEVICE_NAMESPACE")

	deviceShifuMetadata := &deviceshifuOPCUA.DeviceShifuMetaData{
		Name:           deviceName,
		ConfigFilePath: deviceshifuOPCUA.DEVICE_CONFIGMAP_FOLDER_PATH,
		KubeConfigPath: deviceshifuOPCUA.KUBERNETES_CONFIG_DEFAULT,
		Namespace:      namespace,
	}

	ds, err := deviceshifuOPCUA.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	ds.Start(wait.NeverStop)

	select {}
}
