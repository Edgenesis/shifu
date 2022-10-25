package add

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func AddDeviceShifu(ds string) {
	fmt.Println(ds)

	shifuRootDir := os.Getenv("SHIFU_ROOT_DIR")
	if shifuRootDir == "" {
		errStr := "Please set SHIFU_ROOT_DIR environment to Shifu source code root directory." +
			"For example: export SHIFU_ROOT_DIR=~/go/src/github.com/edgenesis/shifu/"
		fmt.Println(errStr)
	}

	CmdDeviceshifuDir := filepath.Join(shifuRootDir, "cmd", "deviceshifu")
	CmdDeviceshifuTemplate := filepath.Join(CmdDeviceshifuDir, "cmdtemplate", "main.go")
	templateByteSlice, err := os.ReadFile(CmdDeviceshifuTemplate)
	if err != nil {
		panic(err)
	}

	templateStr := string(templateByteSlice)
	templateStr = strings.ReplaceAll(templateStr, "deviceshifutemplate", ds)
	templateByteSlice = []byte(templateStr)

	CmdDeviceshifuGeneratedDir := filepath.Join(CmdDeviceshifuDir, ds)
	_, err = exec.Command("mkdir", CmdDeviceshifuGeneratedDir).Output()
	if err != nil {
		panic(err)
	}

	CmdDeviceshifuGeneratedSource := filepath.Join(CmdDeviceshifuGeneratedDir, "main.go")
	err = ioutil.WriteFile(CmdDeviceshifuGeneratedSource, templateByteSlice, 0644)
	if err != nil {
		panic(err)
	}
}
