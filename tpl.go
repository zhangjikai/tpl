package main

import (
	"log"
	"os"
	"path"

	"flag"
	"fmt"
	"strings"

	"path/filepath"

	"github.com/deckarep/golang-set"
)

const (
	ENABLE_DEBUG           = true
	OP_SAVE                = "save"
	OP_GET                 = "get"
	OP_CONFIG              = "config"
	BASE_DIR               = "tpl"
	CONFIG_FILE            = "config.json"
	DEFAULT_STORAGE_FOLDER = "store"
)

var (
	validConfigSet = mapset.NewSet("StorePath")
)

func init() {
	log.SetFlags(0)
}

// -1: file not exists
//  0: normal file
//  1: directory
func checkPath(path string) int {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return -1
	}

	isDir := fileInfo.IsDir()
	if isDir {
		return 1
	} else {
		return 0
	}
}

func moveFile(source, dest string, isDir bool) {
	err := os.RemoveAll(dest)
	if err != nil {
		log.Fatal("Invalid key: " + dest)
	}
	err = os.MkdirAll(dest, 0644)
	if err != nil {
		log.Fatal("Invalid key: " + dest)
	}

	if isDir {
		err := CopyDir(source, dest)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fileName := path.Base(source)
		err := CopyFile(source, path.Join(dest, fileName))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func convertSetToString(set mapset.Set) string {

	it := set.Iterator()
	items := make([]string, 0, len(it.C))

	for elem := range it.C {
		items = append(items, fmt.Sprintf("%v", elem))
	}
	return fmt.Sprintf("%s", strings.Join(items, ", "))

}

func handleArgs(args []string) {
	if args == nil || len(args) == 0 {
		log.Fatal("No parameters.")
	}

	config := LoadConfig(CONFIG_FILE)

	op := args[0]

	switch op {
	case OP_SAVE:
		if len(args) != 3 {
			log.Fatal("Valid format is: save [key] [file/folder]")
		}

		key := args[1]
		filePath := args[2]
		res := checkPath(filePath)
		if res == -1 {
			log.Fatal("Invalid file path: " + filePath)
		}

		if res == 0 {
			moveFile(filePath, path.Join(config.StorePath, key), false)
		} else {
			moveFile(filePath, path.Join(config.StorePath, key), true)
		}
		break
	case OP_GET:
		if len(args) != 2 {
			log.Fatal("Valid format is: get [key]")
		}
		key := args[1]
		srcPath := path.Join(config.StorePath, key)
		res := checkPath(srcPath)
		if res != 1 {
			log.Fatal("Invalid key: " + key)
		}
		currentPath := CurrentRunPath()
		CopyDir(srcPath, currentPath)
		break
	case OP_CONFIG:
		if len(args) != 3 {
			log.Fatal("Valid format: config [type] [value]")
		}
		key := args[1]
		if validConfigSet.Contains(key) {
			value := args[2]
			absPath, err := filepath.Abs(value)
			if err != nil {
				log.Fatal("Invalid config value: " + value)
			}
			err = os.MkdirAll(absPath, 0644)
			if err != nil {
				log.Fatal("Invalid config value: " + value)
			}
			config.StorePath = absPath
			SaveConfig(configFilePath(), config)
		} else {
			log.Fatal("Valid configuration types are: " + convertSetToString(validConfigSet))
		}
	}
}

func main() {
	flag.Parse()
	handleArgs(flag.Args())
}
