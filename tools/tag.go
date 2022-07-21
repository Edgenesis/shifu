package main

/*
This Tool is to modify all deployment.yaml's image's oldTag to newTag in the workDirectory,
arg: [1] workDirectory [2] oldTag [3] newTag
*/
import (
	"bufio"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		log.Fatalf("few parameters")
	}

	path, oldTag, newTag := args[1], args[2], args[3]
	files, err := traverseDir(path)
	if err != nil {
		log.Fatalf("err cannot get all deployment.yaml file! %v", err)
	}

	err = updateTag(files, oldTag, newTag)
	if err != nil {
		log.Fatalf("err make an error during update Tag! %v", err)
	}
}

func traverseDir(workPath string) ([]string, error) {
	var files []string
	var yamls []string
	err := filepath.Walk(workPath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if path.Ext(file) == ".yaml" {
			filename := path.Base(file)
			if strings.Contains(filename, "deployment") && strings.Contains(filename, "deviceshifu") {
				yamls = append(yamls, file)
			}
		}
	}

	return yamls, nil
}

func updateTag(files []string, oldTag string, newTag string) error {
	for _, filePath := range files {
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.Printf("err cannot open file %s %v", file, err)
			return err
		}
		err = replaceTag(file, oldTag, newTag)
		if err != nil {
			log.Printf("err file %s cannot modify %v", file.Name(), err)
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
		file.Truncate(0)
		file.Seek(0, 0)
		out.WriteString(context)
		out.Flush()
		file.Close()
	}()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}

			log.Printf("cannot read file %s %v", file.Name(), err)
			return err
		}

		if strings.Contains(line, "image") && strings.Contains(line, "deviceshifu") && strings.Contains(line, oldTag) {
			line = strings.Replace(line, oldTag, newTag, 1)
		}

		context += line
	}
	return nil
}
