package main

/*
This Tool is to modify all deployment.yaml's image's oldTag to newTag in the workDirectory,
arg: [1] workDirectory [2] oldTag [3] newTag
*/
import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args
	if len(args) < 4 {
		panic("few parameters")
	}

	workDir, oldTag, newTag := args[1], args[2], args[3]
	files, err := traverseDir(workDir)
	if err != nil {
		fmt.Printf("err cannot get all deployment.yaml and .go files! %v", err)
		panic(err)
	}

	err = updateTag(files, oldTag, newTag)
	if err != nil {
		fmt.Printf("err make an error during update Tag! %v", err)
		panic(err)
	}
}

func traverseDir(workPath string) ([]string, error) {
	var files []string
	var targetFiles []string
	err := filepath.Walk(workPath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if path.Ext(file) == ".yaml" || path.Ext(file) == ".go" {
			filename := path.Base(file)
			if strings.Contains(filename, "deployment") ||
				strings.Contains(filename, "install") ||
				filename == "telemetryservice_controller.go" {
				targetFiles = append(targetFiles, file)
			}
		}
	}

	return targetFiles, nil
}

func updateTag(files []string, oldTag string, newTag string) error {
	for _, filePath := range files {
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			fmt.Printf("err cannot open file %s %v", file.Name(), err)
			return err
		}

		err = replaceTag(file, oldTag, newTag)
		if err != nil {
			fmt.Printf("err file %s cannot modify %v", file.Name(), err)
			return err
		}
	}

	return nil
}

func replaceTag(file *os.File, oldTag string, newTag string) error {
	var context string
	reader := bufio.NewReader(file)
	out := bufio.NewWriter(file)
	defer func() {
		_ = file.Truncate(0)
		_, _ = file.Seek(0, 0)
		_, _ = out.WriteString(context)
		_ = out.Flush()
		_ = file.Close()
	}()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}

			fmt.Printf("cannot read file %s %v", file.Name(), err)
			return err
		}

		// Handle image updates in both deployment files and controller constants
		if (strings.Contains(line, "image") &&
			(strings.Contains(line, "deviceshifu") ||
				strings.Contains(line, "mockdevice") ||
				strings.Contains(line, "telemetryservice") ||
				strings.Contains(line, "mockserver") ||
				strings.Contains(line, "humidity") ||
				strings.Contains(line, "gateway"))) ||
			(strings.Contains(line, "const IMAGE = ") &&
				strings.Contains(line, "edgehub/telemetryservice")) {
			line = strings.Replace(line, oldTag, newTag, 1)
		}

		context += line
	}
}
