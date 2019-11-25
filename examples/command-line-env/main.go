package main

import (
	"github.com/sbreitf1/go-console"
)

func main() {
	console.Println("Demo command line interface")

	cle := console.NewCommandLineEnvironment("cli")

	cle.RegisterCommand(console.NewParameterlessCommand("exit", func(args []string) error { return console.ErrExit }))

	if err := cle.Run(); err != nil {
		console.Println()
		if err != console.ErrControlC {
			panic(err)
		}
	}

	/*for {
		console.Print("cli> ")
		cmd, err := cle.ReadCommand()
		if err != nil {
			if err == console.ErrControlC {
				console.Println()
				break
			}

			panic(err)
		}

		if len(cmd) > 0 {
			console.Printlnf("# %q", cmd[0])
			for i := 1; i < len(cmd); i++ {
				console.Printlnf("-> %q", cmd[i])
			}

			if cmd[0] == "exit" {
				break
			}
		} else {
			// empty command
		}
	}*/
}
