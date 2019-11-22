package main

import (
	"github.com/sbreitf1/go-console"
)

func main() {
	console.Println("Press any key:")
	key, char, err := console.ReadKey()
	if err != nil {
		panic(err)
	}
	console.Printlnf("%s -> %q", key, string(char))
}
