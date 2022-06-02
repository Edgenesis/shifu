package main

import (
	"os"

	"github.com/edgenesis/shifu/pkg/deviceshifuSocket"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	deviceName := os.Getenv("EDGEDEVICE_NAME")
	namespace := os.Getenv("EDGEDEVICE_NAMESPACE")

	deviceShifuMetadata := &deviceshifuSocket.DeviceShifuMetaData{
		deviceName,
		deviceshifuSocket.DEVICE_CONFIGMAP_FOLDER_PATH,
		deviceshifuSocket.KUBERNETES_CONFIG_DEFAULT,
		namespace,
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
