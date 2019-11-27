package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sbreitf1/go-console"
)

func main() {
	console.Println("Demo browser")

	cle := console.NewCommandLineEnvironment()
	cle.Prompt = func() string {
		pwd, err := os.Getwd()
		if err != nil {
			return ""
		}
		// display current working directory in nice colors as prompt
		return fmt.Sprintf("\033[1;34m%s\033[0m", pwd)
	}

	cle.RegisterCommand(console.NewExitCommand("exit"))

	cle.RegisterCommand(console.NewCustomCommand("cd",
		func(cmd []string, index int) []console.CompletionCandidate {
			// "" will be replaced with the current working dir. do not allow to cd into files
			candidates, _ := console.BrowseCandidates("", cmd[index], false)
			// ignore errors for completion
			return candidates
		},
		func(args []string) error {
			if len(args) == 0 {
				// no dir to enter specified
				return fmt.Errorf("missing arg")
			}
			return os.Chdir(args[0])
		}))

	cle.RegisterCommand(console.NewParameterlessCommand("ls",
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

	if err := cle.Run(); err != nil {
		console.Println()
		if !console.IsErrControlC(err) {
			panic(err)
		}
	}
}
