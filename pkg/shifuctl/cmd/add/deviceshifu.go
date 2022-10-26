package add

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	DEVICESHIFU_TEMPLATE = "deviceshifutemplate"
)

// addcmd creates the new deviceShifu cmd source file from cmd/deviceshifu/cmdtemplate
func addCmd(ds, shifuRootDir string) {
	cmdDeviceShifuDir := filepath.Join(shifuRootDir, "cmd", "deviceshifu")
	cmdDeviceShifuTemplate := filepath.Join(cmdDeviceShifuDir, "cmdtemplate", "main.go")
	templateByteSlice, err := os.ReadFile(cmdDeviceShifuTemplate)
	if err != nil {
		panic(err)
	}

	templateStr := string(templateByteSlice)
	templateStr = strings.ReplaceAll(templateStr, DEVICESHIFU_TEMPLATE, ds)
	templateByteSlice = []byte(templateStr)

	cmdDeviceShifuGeneratedDir := filepath.Join(cmdDeviceShifuDir, ds)
	_, err = exec.Command("mkdir", cmdDeviceShifuGeneratedDir).Output()
	if err != nil {
		panic(err)
	}

	cmdDeviceShifuGeneratedSource := filepath.Join(cmdDeviceShifuGeneratedDir, "main.go")
	err = ioutil.WriteFile(cmdDeviceShifuGeneratedSource, templateByteSlice, 0644)
	if err != nil {
		panic(err)
	}
}

func addPkg(ds, shifuRootDir string) {
	pkgDeviceShifuDir := filepath.Join(shifuRootDir, "pkg", "deviceshifu")
	pkgDeviceShifuTemplateDir := filepath.Join(pkgDeviceShifuDir, DEVICESHIFU_TEMPLATE)
	pkgDeviceShifuGeneratedDir := filepath.Join(pkgDeviceShifuDir, ds)
	_, err := exec.Command("mkdir", pkgDeviceShifuGeneratedDir).Output()
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(pkgDeviceShifuTemplateDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		templateByteSlice, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		templateStr := string(templateByteSlice)
		templateStr = strings.ReplaceAll(templateStr, DEVICESHIFU_TEMPLATE, ds)
		templateByteSlice = []byte(templateStr)

		inputFileName := info.Name()
		outputFileName := strings.ReplaceAll(inputFileName, DEVICESHIFU_TEMPLATE, ds)
		outputFilePath := filepath.Join(pkgDeviceShifuGeneratedDir, outputFileName)

		err = ioutil.WriteFile(outputFilePath, templateByteSlice, 0644)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}

func addDeviceShifu(ds string) {
	shifuRootDir := os.Getenv("SHIFU_ROOT_DIR")
	if shifuRootDir == "" {
		errStr := "Please set SHIFU_ROOT_DIR environment to Shifu source code root directory." +
			"For example: export SHIFU_ROOT_DIR=~/go/src/github.com/edgenesis/shifu/"
		fmt.Println(errStr)
	}

	addCmd(ds, shifuRootDir)
	addPkg(ds, shifuRootDir)
}
