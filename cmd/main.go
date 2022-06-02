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
		deviceName,
		deviceshifu.DEVICE_CONFIGMAP_FOLDER_PATH,
		deviceshifu.KUBERNETES_CONFIG_DEFAULT,
		namespace,
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
