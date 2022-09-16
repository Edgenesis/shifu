package main

import (
	"os"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifuopcua"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
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
		klog.Errorf("deviceshifu start default, error: %v", err)
		return
	}
	select {}
}
