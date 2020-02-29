package input

import (
	"github.com/sbreitf1/go-console"
)

type screen interface {
	Clear()
	Size() (int, int)
	SetCell(x, y int, r rune)
	Flush()
	SetCursor(x, y int)
	PollEvent() event
	Close()
}

type event interface{}

type errorEvent struct {
	Error error
}

type keyEvent struct {
	Key  console.Key
	Rune rune
}

type resizeEvent struct{}

func printCells(screen screen, str string, x, y int) {
	//TODO support for multiline string
	runes := []rune(str)
	for i := range runes {
		screen.SetCell(x+i, y, runes[i])
	}
}
