package console

import (
	"fmt"

	"github.com/eiannone/keyboard"
)

// Key represents a key.
type Key keyboard.Key

const (
	// KeyEscape represents the escape key
	KeyEscape = Key(keyboard.KeyEsc)
	// KeyCtrlC represents the key combination Ctrl+C
	KeyCtrlC = Key(keyboard.KeyCtrlC)
	// KeyCtrlW represents the key combination Ctrl+W
	KeyCtrlW = Key(keyboard.KeyCtrlW)
	// KeyCtrlS represents the key combination Ctrl+S
	KeyCtrlS = Key(keyboard.KeyCtrlS)
	// KeyUp represents the arrow up key
	KeyUp = Key(keyboard.KeyArrowUp)
	// KeyDown represents the arrow down key
	KeyDown = Key(keyboard.KeyArrowDown)
	// KeyLeft represents the arrow left key
	KeyLeft = Key(keyboard.KeyArrowLeft)
	// KeyRight represents the arrow right key
	KeyRight = Key(keyboard.KeyArrowRight)
	// KeyHome represents the home (Pos1) key
	KeyHome = Key(keyboard.KeyHome)
	// KeyEnd represents the end key
	KeyEnd = Key(keyboard.KeyEnd)
	// KeyPageUp represents the page up key
	KeyPageUp = Key(keyboard.KeyPgup)
	// KeyPageDown represents the page down key
	KeyPageDown = Key(keyboard.KeyPgdn)
	// KeyBackspace represents the backspace key
	KeyBackspace = Key(keyboard.KeyBackspace2)
	// KeyDelete represents the delete key
	KeyDelete = Key(keyboard.KeyDelete)
	// KeyEnter represents the enter key
	KeyEnter = Key(keyboard.KeyEnter)
	// KeyTab represents the tabulator key
	KeyTab = Key(keyboard.KeyTab)
	// KeySpace represents the space key
	KeySpace = Key(keyboard.KeySpace)
)

func (k Key) String() string {
	switch k {
	case KeyUp:
		return "ArrowUp"
	case KeyDown:
		return "ArrowDown"
	case KeyLeft:
		return "ArrowLeft"
	case KeyRight:
		return "ArrowRight"
	case KeyEscape:
		return "Escape"
	case KeyTab:
		return "Tab"
	case KeySpace:
		return "Space"
	case KeyEnter:
		return "Enter"
	case KeyBackspace:
		return "Backspace"
	case KeyCtrlC:
		return "CtrlC"

	default:
		return fmt.Sprintf("Key[%d]", k)
	}
}
