package main

import (
	"github.com/sbreitf1/go-console"
)

func main() {
	if err := console.BeginReadKey(); err != nil {
		console.Fatallnf("Failed to open Keyboard: %s", err.Error())
	}
	defer console.EndReadKey()

	console.Println("Press Escape to exit")
	for {
		key, char, err := console.ReadKey()
		if err != nil {
			console.Fatallnf("ReadKey failed: %s", err.Error())
		}
		console.Printlnf("%s -> %q", key, string(char))

		if key == console.KeyEscape {
			break
		}
	}
}
