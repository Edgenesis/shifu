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
		deviceName,
		deviceshifuOPCUA.DEVICE_CONFIGMAP_FOLDER_PATH,
		deviceshifuOPCUA.KUBERNETES_CONFIG_DEFAULT,
		namespace,
	}

	ds, err := deviceshifuOPCUA.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	if err := ds.Start(wait.NeverStop); err != nil {
		panic(err.Error())
	}

	select {}
}
