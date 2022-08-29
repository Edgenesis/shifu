package main

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifuOPCUA"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"os"

	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	deviceName := os.Getenv("EDGEDEVICE_NAME")
	namespace := os.Getenv("EDGEDEVICE_NAMESPACE")

	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           deviceName,
		ConfigFilePath: deviceshifubase.DEVICE_CONFIGMAP_FOLDER_PATH,
		KubeConfigPath: deviceshifubase.KUBERNETES_CONFIG_DEFAULT,
		Namespace:      namespace,
	}

	ds, err := deviceshifuOPCUA.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	ds.Start(wait.NeverStop)

	select {}
}
