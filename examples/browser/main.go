package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/commandline"
	"github.com/sbreitf1/go-console/input"
)

func main() {
	console.Println("Demo browser")

	cle := commandline.NewEnvironment()
	cle.Prompt = func() string {
		pwd, err := os.Getwd()
		if err != nil {
			return ""
		}
		if console.SupportsColors() {
			// display current working directory in nice colors as prompt
			return fmt.Sprintf("\033[1;34m%s\033[0m", pwd)
		}
		// fallback: no colors supported
		return pwd
	}

	cle.RegisterCommand(commandline.NewExitCommand("exit"))

	cle.RegisterCommand(commandline.NewCustomCommand("cd",
		commandline.NewFixedArgCompletion(commandline.NewLocalFileSystemArgCompletion(false)),
		func(args []string) error {
			if len(args) == 0 {
				// no dir to enter specified
				return fmt.Errorf("missing arg")
			}
			return os.Chdir(args[0])
		}))

	cle.RegisterCommand(commandline.NewParameterlessCommand("ls",
		func(args []string) error {
			// print current working dir content
			files, err := ioutil.ReadDir("./")
			if err != nil {
				return err
			}
			for _, f := range files {
				if f.IsDir() {
					console.Print("  D  ")
				} else {
					console.Print("  F  ")
				}
				console.Println(f.Name())
			}
			return nil
		}))

	cle.RegisterCommand(commandline.NewCustomCommand("edit",
		commandline.NewFixedArgCompletion(commandline.NewLocalFileSystemArgCompletion(true)),
		func(args []string) error {
			if len(args) == 0 {
				// no dir to enter specified
				return fmt.Errorf("missing arg")
			}

			var content string
			fi, err := os.Stat(args[0])
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				if fi.IsDir() {
					return fmt.Errorf("path is a directory")
				}

				data, err := ioutil.ReadFile(args[0])
				if err != nil {
					return err
				}

				content = string(data)
			}

			newContent, ok, err := input.Text(content)
			if err != nil {
				return err
			}

			if ok {
				if err := ioutil.WriteFile(args[0], []byte(newContent), os.ModePerm); err != nil {
					return err
				}
			}

			return nil
		}))

	if err := cle.Run(); err != nil {
		console.Println()
		if !commandline.IsErrCtrlC(err) {
			console.Fatallnf("Run failed: %s", err.Error())
		}
	}
}
