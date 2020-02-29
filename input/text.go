// +build !windows

package input

import (
	"github.com/nsf/termbox-go"
)

// Text opens a fullscreen text editor in console mode and returns the entered string.
func Text(str string) (string, bool, error) {
	if err := termbox.Init(); err != nil {
		return "", false, err
	}
	defer termbox.Close()

	editor := newTextEditor(str)

	// currently visible rectangle
	firstLine := 0
	firstPos := 0

	for {
		// render current editor view
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		viewportWidth, viewportHeight := termbox.Size()
		editorOffsetX := 1
		editorOffsetY := 1
		editorWidth := viewportWidth - 2
		editorHeight := viewportHeight - 3

		//TODO skip draw when inserting large data

		// draw outer box
		for x := editorOffsetX; x < (editorOffsetX + editorWidth); x++ {
			termbox.SetCell(x, editorOffsetY-1, '─', termbox.ColorDefault, termbox.ColorDefault)
			termbox.SetCell(x, editorOffsetY+editorHeight, '─', termbox.ColorDefault, termbox.ColorDefault)
		}
		for y := editorOffsetY; y < (editorOffsetY + editorHeight); y++ {
			termbox.SetCell(editorOffsetX-1, y, '│', termbox.ColorDefault, termbox.ColorDefault)
			termbox.SetCell(editorOffsetX+editorWidth, y, '│', termbox.ColorDefault, termbox.ColorDefault)
		}
		termbox.SetCell(editorOffsetX-1, editorOffsetY-1, '┌', termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(editorOffsetX+editorWidth, editorOffsetY-1, '┐', termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(editorOffsetX-1, editorOffsetY+editorHeight, '└', termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(editorOffsetX+editorWidth, editorOffsetY+editorHeight, '┘', termbox.ColorDefault, termbox.ColorDefault)

		caretLine, caretPos := editor.Caret()
		// ensure caret is visible
		if caretLine < firstLine {
			firstLine = caretLine
		}
		if caretLine >= (firstLine + editorHeight - 1) {
			firstLine = caretLine - editorHeight + 1
		}
		if caretPos < firstPos {
			firstPos = caretPos
		}
		if caretPos >= (firstPos + editorWidth - 1) {
			firstPos = caretPos - editorWidth + 1
		}
		// set relative caret location
		termbox.SetCursor(editorOffsetX+caretPos-firstPos, editorOffsetY+caretLine-firstLine)

		// print text
		for i, line := range editor.LineRange(firstLine, editorHeight) {
			runes := []rune(line)
			for j := firstPos; j < min(len(runes), firstPos+editorWidth); j++ {
				termbox.SetCell(editorOffsetX+j-firstPos, editorOffsetY+i, runes[j], termbox.ColorDefault, termbox.ColorDefault)
			}
		}

		tbPrint("Esc to exit", 1, editorOffsetY+editorHeight+1)
		tbPrint("Strg+S to save", editorOffsetX+editorWidth-14, editorOffsetY+editorHeight+1)

		// display
		termbox.Flush()
		termbox.Sync()

		switch e := termbox.PollEvent(); e.Type {
		case termbox.EventKey:
			switch e.Key {
			case termbox.KeyEsc:
				return str, false, nil

			case termbox.KeyCtrlW:
				// for all nano fans :)
				fallthrough
			case termbox.KeyCtrlS:
				return editor.String(), true, nil

			case termbox.KeyArrowLeft:
				editor.MoveCaretLeft()
			case termbox.KeyArrowRight:
				editor.MoveCaretRight()
			case termbox.KeyArrowUp:
				editor.MoveCaretUp(1)
			case termbox.KeyArrowDown:
				editor.MoveCaretDown(1)

			case termbox.KeyPgup:
				editor.MoveCaretUp(editorHeight)
			case termbox.KeyPgdn:
				editor.MoveCaretDown(editorHeight)

			case termbox.KeyHome:
				editor.MoveCaretToLineBegin()
			case termbox.KeyEnd:
				editor.MoveCaretToLineEnd()

			case termbox.KeyBackspace:
				fallthrough
			case termbox.KeyBackspace2:
				editor.RemoveLeftOfCaret()
			case termbox.KeyDelete:
				editor.RemoveRightOfCaret()

			case termbox.KeyEnter:
				editor.NewLineAtCaret()
			case termbox.KeySpace:
				editor.InsertAtCaret(" ")
			case termbox.KeyTab:
				editor.InsertAtCaret("    ")
			default:
				if e.Ch != '\000' {
					editor.InsertAtCaret(string(e.Ch))
				}
			}

		case termbox.EventResize:

		case termbox.EventError:
			return "", false, e.Err
		}
	}
}

func tbPrint(str string, x, y int) {
	runes := []rune(str)
	for i := range runes {
		termbox.SetCell(x+i, y, runes[i], termbox.ColorDefault, termbox.ColorDefault)
	}
}
