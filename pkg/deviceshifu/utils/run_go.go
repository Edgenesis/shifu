package utils

import (
	"os/exec"

	"k8s.io/klog/v2"
)

const (
	SCRIPTDIR = "../go_custom_handler"
)

func Run(funcName string, rawData string) string {
	cmd := exec.Command(funcName, rawData)
	cmd.Dir = SCRIPTDIR

	processed, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("Failed process command %v, %s", cmd.Path, cmd.Args)
	}
	return string(processed)
}
