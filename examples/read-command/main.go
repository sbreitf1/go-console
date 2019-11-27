package main

import (
	"github.com/sbreitf1/go-console"
)

func main() {
	history := console.NewCommandHistory(3)

	console.Println("type exit to leave")
	for {
		cmd, err := console.ReadCommand("command", history.GetHistoryEntry, nil)
		if err != nil {
			if console.IsErrControlC(err) {
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

			history.Put(cmd)

			if cmd[0] == "exit" {
				break
			}
		} else {
			// empty command
		}
	}
}
