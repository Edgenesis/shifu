package main

import (
	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifubase"
	"os"

	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifu"
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

	ds, err := deviceshifu.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	if err := ds.Start(wait.NeverStop); err != nil {
		panic(err.Error())
	}

	select {}
}
