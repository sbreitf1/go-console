// +build windows

package input

import (
	"github.com/gdamore/tcell"
	"github.com/sbreitf1/go-console"
)

type windowsScreen struct {
	screen tcell.Screen
}

func newScreen() (screen, error) {
	screen, err := tcell.NewConsoleScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	return &windowsScreen{screen}, nil
}

func (s *windowsScreen) Clear() {
	s.screen.Clear()
}
func (s *windowsScreen) Size() (int, int) {
	return s.screen.Size()
}
func (s *windowsScreen) SetCell(x, y int, r rune) {
	s.screen.SetContent(x, y, r, nil, tcell.StyleDefault)
}
func (s *windowsScreen) Flush() {
	s.screen.Sync()
}
func (s *windowsScreen) SetCursor(x, y int) {
	s.screen.ShowCursor(x, y)
}
func (s *windowsScreen) PollEvent() event {
	// wait for supported event
	for {
		// translate received event
		switch e := s.screen.PollEvent().(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyEscape:
				return keyEvent{console.KeyEscape, '\000'}

			case tcell.KeyCtrlW:
				return keyEvent{console.KeyCtrlW, '\000'}
			case tcell.KeyCtrlS:
				return keyEvent{console.KeyCtrlS, '\000'}

			case tcell.KeyUp:
				return keyEvent{console.KeyUp, '\000'}
			case tcell.KeyDown:
				return keyEvent{console.KeyDown, '\000'}
			case tcell.KeyLeft:
				return keyEvent{console.KeyLeft, '\000'}
			case tcell.KeyRight:
				return keyEvent{console.KeyRight, '\000'}
			case tcell.KeyHome:
				return keyEvent{console.KeyHome, '\000'}
			case tcell.KeyEnd:
				return keyEvent{console.KeyEnd, '\000'}
			case tcell.KeyPgUp:
				return keyEvent{console.KeyPageUp, '\000'}
			case tcell.KeyPgDn:
				return keyEvent{console.KeyPageDown, '\000'}

			case tcell.KeyBackspace:
				fallthrough
			case tcell.KeyBackspace2:
				return keyEvent{console.KeyBackspace, '\r'}
			case tcell.KeyDelete:
				return keyEvent{console.KeyDelete, '\000'}
			case tcell.KeyEnter:
				return keyEvent{console.KeyEnter, '\n'}
			case tcell.KeyTab:
				return keyEvent{console.KeyTab, '\t'}

			default:
				if e.Rune() == ' ' {
					return keyEvent{console.KeySpace, ' '}
				}
				return keyEvent{0, e.Rune()}
			}

		case *tcell.EventResize:
			return resizeEvent{}

		case *tcell.EventError:
			return errorEvent{e}
		}
	}
}
func (s *windowsScreen) Close() {
	s.screen.Fini()
}
