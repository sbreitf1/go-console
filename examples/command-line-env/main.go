package main

import (
	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/commandline"
)

func main() {
	console.Println("Demo command line interface")

	cle := commandline.NewEnvironment()

	cle.RegisterCommand(commandline.NewExitCommand("exit"))

	cle.ExecUnknownCommand = func(cmd string, args []string) error {
		console.Printlnf("Unknown command %q", cmd)
		for _, arg := range args {
			console.Printlnf("-> %q", arg)
		}
		return nil
	}

	cle.RegisterCommand(commandline.NewCustomCommand("duck",
		func(cmd []string, index int) []commandline.CompletionOption {
			options := make([]commandline.CompletionOption, 0)
			for name := range ducks {
				options = append(options, commandline.NewCompletionOption(name, false))
			}
			return options
		},
		func(args []string) error {
			if len(args) > 0 {
				if duck, exists := ducks[args[0]]; exists {
					console.Printlnf("-> %s duck: %s", args[0], duck)
				} else {
					console.Printlnf("-> unknown duck %q", args[0])
				}
			} else {
				console.Println("-> missing duck name!")
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

var (
	ducks = make(map[string]string)
)

func init() {
	ducks["pintail"] = "ancestor"
	ducks["humperdink"] = "grandpa"
	ducks["elvira"] = "grandma"
	ducks["quackmore"] = "father of donald"
	ducks["hortense"] = "mother of donald"
	ducks["daphne"] = "aunt from donald"
	ducks["eider"] = "uncle from donald"
	ducks["lulubelle"] = "wife from eider"
	ducks["dan"] = "sheriff"
	ducks["donald"] = "famous!"
	ducks["della"] = "mother from huey, dewey and louie"
	ducks["huey"] = "famous!"
	ducks["dewey"] = "famous!"
	ducks["louie"] = "famous!"
	ducks["fethry"] = "cousin"
	ducks["whitewater"] = "from log jockey"
	ducks["dudly d."] = "architect"
	ducks["dimwitty"] = "assistant"
	ducks["moby"] = "moby dick?"
	ducks["dugan"] = "very young"
}
