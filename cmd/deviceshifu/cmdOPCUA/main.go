package main

import (
	"log"
	"os"

	deviceshifuopcua "github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifuOPCUA"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"

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

	ds, err := deviceshifuopcua.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	err = ds.Start(wait.NeverStop)
	if err != nil {
		log.Println("deviceshifu start default, error: ", err)
		return
	}
	select {}
}
