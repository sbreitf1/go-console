package main

import (
	"fmt"

	"github.com/sbreitf1/go-console"
)

func main() {
	console.Print("USER: ")
	user, err := console.ReadLine()
	if err != nil {
		panic(err)
	}

	console.Print("PASS: ")
	pass, err := console.ReadPassword()
	if err != nil {
		panic(err)
	}

	console.Println("#######################")
	fmt.Println(user, "->", pass)
}
