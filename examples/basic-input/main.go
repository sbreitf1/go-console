package main

import (
	"github.com/sbreitf1/go-console"
)

func main() {
	console.Print("USER: ")
	user, err := console.ReadLine()
	if err != nil {
		console.Fatallnf("ReadLine failed: %s", err.Error())
	}

	console.Print("PASS: ")
	pass, err := console.ReadPassword()
	if err != nil {
		console.Fatallnf("ReadPassword failed: %s", err.Error())
	}

	console.Println("#######################")
	console.Printlnf("%q -> %q", user, pass)
}
