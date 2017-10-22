package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/deckarep/golang-set"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
)

const (
	ENABLE_DEBUG           = true
	OP_SAVE                = "save"
	OP_GET                 = "get"
	OP_DELETE              = "delete"
	OP_CONFIG              = "config"
	OP_PUSH                = "push"
	OP_PULL                = "pull"
	OP_LS                  = "ls"
	BASE_DIR               = "tpl"
	CONFIG_FILE            = "config.json"
	DEFAULT_STORAGE_FOLDER = "store"
)

var (
	validConfigSet = mapset.NewSet("StorePath")
	ignoreFileSet  = mapset.NewSet(".git")
)

func init() {
	log.SetFlags(0)
	initBaseDir()
}

// -1: file not exists
//  0: normal filegit
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

func push(baseDir string) {
	gitDir := "--git-dir=" + filepath.Join(baseDir, ".git")
	workTree := "--work-tree=" + baseDir
	cmd := exec.Command("git", gitDir, workTree, "add", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	cmd = exec.Command("git", gitDir, workTree, "commit", "-m", "\"update\"")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	cmd = exec.Command("git", gitDir, workTree, "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func pull(baseDir string) {
	gitDir := "--git-dir=" + filepath.Join(baseDir, ".git")
	workTree := "--work-tree=" + baseDir
	cmd := exec.Command("git", gitDir, workTree, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func printConfig(config Config) {
	s := reflect.ValueOf(&config).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%s: %s", typeOfT.Field(i).Name, f.Interface())
	}
}

func moveFile(source, dest string, isDir bool) (error) {
	err := os.RemoveAll(dest)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dest, 0644)
	if err != nil {
		return err
	}

	if isDir {
		err := CopyDir(source, dest)
		if err != nil {
			return err
		}
	} else {
		fileName := path.Base(source)
		err := CopyFile(source, path.Join(dest, fileName))
		if err != nil {
			return err
		}
	}
	return nil
}

func convertSetToString(set mapset.Set) string {
	it := set.Iterator()
	items := make([]string, 0, len(it.C))

	for elem := range it.C {
		items = append(items, fmt.Sprintf("%v", elem))
	}
	return fmt.Sprintf("%s", strings.Join(items, ", "))
}

func newError(err error) error {
	if err == nil {
		return nil
	}
	return cli.NewExitError(err.Error(), -1)
}

func newErrorWithText(text string) error {
	return cli.NewExitError(text, -1)
}

func handleArgs(op string, args []string) error {

	configPath, err := configFilePath()
	if err != nil {
		return newError(err)
	}
	config, err := LoadConfig(configPath)
	if err != nil {
		return newError(err)
	}

	switch op {
	case OP_SAVE:
		if len(args) != 2 {
			return newErrorWithText("Valid format: save [key] [template path]")
		}

		key := args[0]
		if FileExists(path.Join(config.StorePath, key)) {
			flag := false
			prompt := &survey.Confirm{
				Message: "Do you want to override the exists key?",
			}
			survey.AskOne(prompt, &flag, nil)
			if !flag {
				return nil
			}

		}
		filePath := args[1]
		res := checkPath(filePath)
		if res == -1 {
			return newErrorWithText("Invalid template path: " + filePath)
		}

		if res == 0 {
			err = moveFile(filePath, path.Join(config.StorePath, key), false)
		} else {
			err = moveFile(filePath, path.Join(config.StorePath, key), true)
		}
		return newError(err)
	case OP_GET:
		if len(args) != 1 {
			return newErrorWithText("Valid format: get [key]")
		}
		key := args[0]
		srcPath := path.Join(config.StorePath, key)
		res := checkPath(srcPath)
		if res != 1 {
			return newErrorWithText("Invalid key: " + key)
		}
		currentPath, err := CurrentRunPath()
		if err != nil {
			return newError(err)
		}
		err = CopyDir(srcPath, currentPath)
		return newError(err)
	case OP_DELETE:
		if len(args) != 1 {
			return newErrorWithText("Valid format: delete [key]")
		}
		key := args[0]
		if FileExists(path.Join(config.StorePath, key)) {
			flag := false
			prompt := &survey.Confirm{
				Message: "Do you want to delete this key?",
			}
			survey.AskOne(prompt, &flag, nil)
			if !flag {
				return nil
			}

		}
		fullPath := filepath.Join(config.StorePath, key)
		err := os.RemoveAll(fullPath)
		if err != nil {
			return newError(err)
		}
		break
	case OP_CONFIG:
		if len(args) == 0 {
			printConfig(config)
			return nil
		}
		if len(args) != 2 {
			return newErrorWithText("Valid format: config [type] [value]")
		}
		key := args[0]
		if validConfigSet.Contains(key) {
			value := filepath.FromSlash(args[1])
			absPath, err := filepath.Abs(value)
			if err != nil {
				return newErrorWithText("Invalid config value: " + value)
			}
			err = os.MkdirAll(absPath, 0644)
			if err != nil {
				return newErrorWithText("Invalid config value: " + value)
			}
			config.StorePath = absPath
			configPath, err := configFilePath()
			if err != nil {
				return newError(err)
			}
			err = SaveConfig(configPath, config)
			return newError(err)
		} else {
			return newErrorWithText("Valid configuration types: " + convertSetToString(validConfigSet))
		}
	case OP_LS:
		if len(args) == 0 {
			args = []string{""}
		}
		if len(args) != 1 {
			return newErrorWithText("Valid format: ls [prefix]");
		}
		prefix := args[0]
		fileList, err := GetFileWithPrefix(config.StorePath, prefix, ignoreFileSet)
		if err != nil {
			return newError(err)
		}
		text := strings.Join(fileList, "\n")
		fmt.Println(text)

	case OP_PUSH:
		push(config.StorePath)
		break
	case OP_PULL:
		pull(config.StorePath)
		break
	}
	return nil
}

func main() {

	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
    {{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}    {{join .Names ","}}{{"\t"}}{{.ArgsUsage}} {{"\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}
	{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}{{end}}
   {{end}}
`

	app := cli.NewApp()
	app.Name = "tpl"
	app.Usage = "A simple tool for easy managing of file or project template"
	app.Version = "1.0.0"

	app.Commands = []cli.Command{
		{
			Name:      "save",
			Aliases:   []string{"s"},
			Usage:     "Save a template associated with the specified key to the library.",
			ArgsUsage: "[key] [template path]",
			Action: func(c *cli.Context) error {
				return handleArgs(OP_SAVE, c.Args())
			},
		},

		{
			Name:      "get",
			Aliases:   []string{"g"},
			Usage:     "Get a template associated with the specified key from the library.",
			ArgsUsage: "[key]",
			Action: func(c *cli.Context) error {
				return handleArgs(OP_GET, c.Args())

			},
		},

		{
			Name:      "delete",
			Aliases:   []string{"d"},
			Usage:     "Delete a template associated with the specified key from the library.",
			ArgsUsage: "[key]",
			Action: func(c *cli.Context) error {
				return handleArgs(OP_DELETE, c.Args())

			},
		},

		{
			Name:      "ls",
			Aliases:   []string{"l"},
			Usage:     "List the keys that begin with prefix.",
			ArgsUsage: "[prefix]",
			Action: func(c *cli.Context) error {
				return handleArgs(OP_LS, c.Args())
			},
		},

		{
			Name:      "config",
			Aliases:   []string{"c"},
			Usage:     "Set configurations of tpl. The valid configuration is \"StorePath\".",
			ArgsUsage: "[type] [value]",
			Action: func(c *cli.Context) error {
				return handleArgs(OP_CONFIG, c.Args())
			},
		},

		{
			Name:  "push",
			Usage: "Call git push command based on the template library directory.",
			Action: func(c *cli.Context) error {
				return handleArgs(OP_PUSH, c.Args())
			},
		},

		{
			Name:  "pull",
			Usage: "Call git pull command based on the template library directory.",
			Action: func(c *cli.Context) error {
				return handleArgs(OP_PULL, c.Args())
			},
		},
	}

	app.Run(os.Args)
}
