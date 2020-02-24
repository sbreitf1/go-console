package main

import (
	"github.com/sbreitf1/go-console"
)

func main() {
	history := console.NewLineHistory(5)

	for {
		console.Print("enter> ")
		line, err := console.ReadLineWithHistory(history)
		if err != nil {
			console.Fatallnf("ReadLineWithHistory failed: %s", err.Error())
		}

		history.Put(line)
		console.Printlnf("-> %q", line)
	}
}
