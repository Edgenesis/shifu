package main

import (
	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifu"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	deviceName := "edgedevice-sample"
	namespace := "crd-system"
	// kubeconfigPath := "/root/.kube/config"
	// config_folder := "etc/edgedevice/config"

	ds := deviceshifu.New(
		deviceName,
		deviceshifu.DEVICE_CONFIGMAP_FOLDER_STR,
		deviceshifu.KUBERNETES_CONFIG_DEFAULT,
		namespace,
	)

	if err := ds.Start(wait.NeverStop); err != nil {
		panic(err.Error())
	}

	select {}
}
