// +build !windows

package input

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

type unixScreen struct{}

func newScreen() (screen, error) {
	return nil, fmt.Errorf("newScreen not implemented on Unix")
}

func (s *unixScreen) Clear() {
	termbox.Clear()
}
func (s *unixScreen) Size() (int, int) {
	return termbox.Size()
}
func (s *unixScreen) SetCell(x, y int, r rune) {
	termbox.SetCell(x, y, r, termbox.ColorDefault, termbox.ColorDefault)
}
func (s *unixScreen) Flush() {
	termbox.Flush()
}
func (s *unixScreen) SetCursor(x, y int) {
	termbox.SetCursor(x, y)
}
func (s *unixScreen) PollEvent() event {
	return errorEvent{fmt.Errorf("unixScreen.PollEvent not implemented")}
}
func (s *unixScreen) Close() {
	termbox.Close()
}
