// +build windows

package input

import (
	"github.com/sbreitf1/go-console"
)

// Text opens a fullscreen text editor in console mode and returns the entered string.
func Text(str string) (string, bool, error) {
	screen, err := newScreen()
	if err != nil {
		return "", false, err
	}
	defer screen.Close()

	editor := newTextEditor(str)

	// currently visible rectangle
	firstLine := 0
	firstPos := 0

	for {
		// render current editor view
		screen.Clear()
		viewportWidth, viewportHeight := screen.Size()
		editorOffsetX := 1
		editorOffsetY := 1
		editorWidth := viewportWidth - 2
		editorHeight := viewportHeight - 3

		//TODO skip draw when inserting large data

		// draw outer box
		for x := editorOffsetX; x < (editorOffsetX + editorWidth); x++ {
			screen.SetCell(x, editorOffsetY-1, '─')
			screen.SetCell(x, editorOffsetY+editorHeight, '─')
		}
		for y := editorOffsetY; y < (editorOffsetY + editorHeight); y++ {
			screen.SetCell(editorOffsetX-1, y, '│')
			screen.SetCell(editorOffsetX+editorWidth, y, '│')
		}
		screen.SetCell(editorOffsetX-1, editorOffsetY-1, '┌')
		screen.SetCell(editorOffsetX+editorWidth, editorOffsetY-1, '┐')
		screen.SetCell(editorOffsetX-1, editorOffsetY+editorHeight, '└')
		screen.SetCell(editorOffsetX+editorWidth, editorOffsetY+editorHeight, '┘')

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
		screen.SetCursor(editorOffsetX+caretPos-firstPos, editorOffsetY+caretLine-firstLine)

		// print text
		for i, line := range editor.LineRange(firstLine, editorHeight) {
			runes := []rune(line)
			for j := firstPos; j < min(len(runes), firstPos+editorWidth); j++ {
				screen.SetCell(editorOffsetX+j-firstPos, editorOffsetY+i, runes[j])
			}
		}

		printCells(screen, "Esc to exit", 1, editorOffsetY+editorHeight+1)
		printCells(screen, "Strg+S to save", editorOffsetX+editorWidth-14, editorOffsetY+editorHeight+1)

		// display
		screen.Flush()

		switch e := screen.PollEvent().(type) {
		case keyEvent:
			switch e.Key {
			case console.KeyEscape:
				return str, false, nil

			case console.KeyCtrlW:
				// for all nano fans :)
				fallthrough
			case console.KeyCtrlS:
				return editor.String(), true, nil

			case console.KeyLeft:
				editor.MoveCaretLeft()
			case console.KeyRight:
				editor.MoveCaretRight()
			case console.KeyUp:
				editor.MoveCaretUp(1)
			case console.KeyDown:
				editor.MoveCaretDown(1)

			case console.KeyPageUp:
				editor.MoveCaretUp(editorHeight)
			case console.KeyPageDown:
				editor.MoveCaretDown(editorHeight)

			case console.KeyHome:
				editor.MoveCaretToLineBegin()
			case console.KeyEnd:
				editor.MoveCaretToLineEnd()

			case console.KeyBackspace:
				editor.RemoveLeftOfCaret()
			case console.KeyDelete:
				editor.RemoveRightOfCaret()

			case console.KeyEnter:
				editor.NewLineAtCaret()
			case console.KeySpace:
				editor.InsertAtCaret(" ")
			case console.KeyTab:
				editor.InsertAtCaret("    ")
			default:
				if e.Rune != '\000' {
					editor.InsertAtCaret(string(e.Rune))
				}
			}

		case resizeEvent:
			// do nothing, just redraw in next iteration

		case errorEvent:
			return "", false, e.Error
		}
	}
}
