package console

import (
	"fmt"

	"github.com/eiannone/keyboard"
)

// Key represents a key.
type Key keyboard.Key

const (
	// KeyUp represents the arrow up key
	KeyUp = Key(keyboard.KeyArrowUp)
	// KeyDown represents the arrow down key
	KeyDown = Key(keyboard.KeyArrowDown)
	// KeyLeft represents the arrow left key
	KeyLeft = Key(keyboard.KeyArrowLeft)
	// KeyRight represents the arrow right key
	KeyRight = Key(keyboard.KeyArrowRight)
	// KeyEscape represents the escape key
	KeyEscape = Key(keyboard.KeyEsc)
	// KeyTab represents the tabulator key
	KeyTab = Key(keyboard.KeyTab)
	// KeySpace represents the space key
	KeySpace = Key(keyboard.KeySpace)
	// KeyEnter represents the enter key
	KeyEnter = Key(keyboard.KeyEnter)
	// KeyBackspace represents the backspace key
	KeyBackspace = Key(keyboard.KeyBackspace2)
	// KeyCtrlC represents the key combination Ctrl+C
	KeyCtrlC = Key(keyboard.KeyCtrlC)
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
