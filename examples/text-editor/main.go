package main

import (
	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/input"
)

func main() {
	//again:
	str, ok, err := input.Text("default string")
	if err != nil {
		console.Fatallnf("FATAL: %s", err.Error())
	}

	if ok {
		console.Printlnf("You entered %q", str)
		//goto again
	} else {
		console.Println("Abort")
	}
}
