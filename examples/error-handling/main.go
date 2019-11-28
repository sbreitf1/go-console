package main

import (
	"fmt"
	"strings"

	"github.com/sbreitf1/go-console"
)

func main() {
	cle := console.NewCommandLineEnvironment()
	cle.RegisterCommand(console.NewExitCommand("exit"))

	errHandler := func(cmd string, args []string, err error) error {
		if console.IsErrCommandPanicked(err) {
			console.Printlnf("PANIC: %s", err.Error())
		} else {
			console.Printlnf("ERROR: %s", err.Error())
		}
		return nil
	}
	cle.ErrorHandler = errHandler

	cle.RegisterCommand(console.NewCustomCommand("toggle",
		console.NewFixedArgCompletion(
			console.NewOneOfArgCompletion("error", "panic"),
		),
		func(args []string) error {
			if len(args) > 0 {
				switch args[0] {
				case "error":
					if cle.ErrorHandler == nil {
						cle.ErrorHandler = errHandler
						console.Printlnf("Handle Errors")
					} else {
						cle.ErrorHandler = nil
						console.Printlnf("Escalate Errors")
					}

				case "panic":
					cle.RecoverPanickedCommands = !cle.RecoverPanickedCommands
					console.Printlnf("Recover Panics = %v", cle.RecoverPanickedCommands)

				default:
					console.Printlnf("invalid arg %q", args[0])
				}

			} else {
				console.Println("missing arg")
			}
			return nil
		}))

	cle.RegisterCommand(console.NewParameterlessCommand("error", func(args []string) error {
		return fmt.Errorf(strings.Join(args, " "))
	}))

	cle.RegisterCommand(console.NewParameterlessCommand("panic", func(args []string) error {
		panic(strings.Join(args, " "))
	}))

	if err := cle.Run(); err != nil {
		if console.IsErrCommandPanicked(err) {
			console.Printlnf("FATAL PANIC: %s", err.Error())
		} else {
			console.Printlnf("FATAL ERROR: %s", err.Error())
		}
	}
}
