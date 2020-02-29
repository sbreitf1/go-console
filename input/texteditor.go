package input

import (
	"strings"
)

type textEditor struct {
	InsertMode bool
	lines      [][]rune
	caretLine  int
	caretPos   int
}

func newTextEditor(str string) *textEditor {
	strLines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(str, "\r\n", "\n"), "\r", "\n"), "\n")
	lines := make([][]rune, len(strLines))
	for i := 0; i < len(strLines); i++ {
		lines[i] = []rune(strLines[i])
	}
	return &textEditor{false, lines, 0, 0}
}

func (e *textEditor) Caret() (int, int) {
	caretLine := boundBy(e.caretLine, 0, len(e.lines)-1)
	caretPos := boundBy(e.caretPos, 0, len(e.lines[caretLine]))
	return caretLine, caretPos
}

func (e *textEditor) String() string {
	var sb strings.Builder
	for i := 0; i < len(e.lines); i++ {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(string(e.lines[i]))
	}
	return sb.String()
}

func (e *textEditor) Lines() []string {
	strLines := make([]string, len(e.lines))
	for i := 0; i < len(e.lines); i++ {
		strLines[i] = string(e.lines[i])
	}
	return strLines
}

func (e *textEditor) LineRange(start, count int) []string {
	if start >= len(e.lines) {
		return []string{}
	}

	count = min(len(e.lines)-start, count)
	strLines := make([]string, count)
	for i := start; i < (start + count); i++ {
		strLines[i-start] = string(e.lines[i])
	}
	return strLines
}

func (e *textEditor) MoveCaretLeft() bool {
	e.caretPos--
	if e.caretPos < 0 {
		if e.caretLine <= 0 {
			e.caretPos = 0
		} else {
			e.caretLine--
			e.caretPos = len(e.lines[e.caretLine])
		}
	}
	return true
}

func (e *textEditor) MoveCaretRight() bool {
	e.caretPos++
	if e.caretPos >= len(e.lines[e.caretLine]) {
		if e.caretLine >= (len(e.lines) - 1) {
			e.caretPos = len(e.lines[e.caretLine])
		} else {
			e.caretLine++
			e.caretPos = 0
		}
	}
	return true
}

func (e *textEditor) MoveCaretUp(delta int) bool {
	e.caretLine -= delta
	if e.caretLine < 0 {
		e.caretLine = 0
		//e.caretPos = 0
	}
	return true
}

func (e *textEditor) MoveCaretDown(delta int) bool {
	e.caretLine += delta
	if e.caretLine >= len(e.lines) {
		e.caretLine = len(e.lines) - 1
		//e.caretPos = len(e.lines[e.caretLine])
	}
	return true
}

func (e *textEditor) MoveCaretToLineBegin() bool {
	e.caretPos = 0
	return true
}

func (e *textEditor) MoveCaretToLineEnd() bool {
	e.caretPos = len(e.lines[e.caretLine])
	return true
}

func (e *textEditor) InsertAtCaret(str string) {
	caretLine, caretPos := e.Caret()

	str = strings.ReplaceAll(strings.ReplaceAll(str, "\r\n", "\n"), "\r", "\n")
	insertLines := strings.Split(str, "\n")
	if len(insertLines) == 1 {
		// no new line inserted, simple case:
		prefix := string(e.lines[caretLine][:caretPos])
		suffix := string(e.lines[caretLine][caretPos:])
		e.lines[caretLine] = []rune(prefix + insertLines[0] + suffix)

		// move caret to end of inserted string
		e.caretPos = caretPos + len([]rune(insertLines[0]))

	} else {
		targetCaretPos := len([]rune(insertLines[len(insertLines)-1]))

		// first line first part of old line (to caret) and first line of inserted string
		insertLines[0] = string(e.lines[caretLine][:caretPos]) + insertLines[0]
		// last line is last line of inserted string and last part of old line (behind caret)
		insertLines[len(insertLines)-1] = insertLines[len(insertLines)-1] + string(e.lines[caretLine][caretPos:])
		// insert new lines to slice
		newLines := make([][]rune, len(e.lines)+len(insertLines)-1)
		for i := 0; i < caretLine; i++ {
			newLines[i] = e.lines[i]
		}
		for i := 0; i < len(insertLines); i++ {
			newLines[caretLine+i] = []rune(insertLines[i])
		}
		for i := (caretLine + 1); i < len(e.lines); i++ {
			newLines[len(insertLines)+i-1] = e.lines[i]
		}
		e.lines = newLines

		// move caret to end of inserted string
		e.caretLine = caretLine + len(insertLines) - 1
		e.caretPos = targetCaretPos
	}
}

func (e *textEditor) NewLineAtCaret() {
	e.InsertAtCaret("\n")
}

func (e *textEditor) RemoveLeftOfCaret() bool {
	caretLine, caretPos := e.Caret()
	// fix caret position to currently visible position
	e.caretPos = caretPos

	if caretPos > 0 {
		// removing only in current line
		prefix := string(e.lines[caretLine][:caretPos-1])
		suffix := string(e.lines[caretLine][caretPos:])
		e.lines[caretLine] = []rune(prefix + suffix)
		e.caretPos = caretPos - 1
		return true
	}
	if caretPos == 0 && caretLine > 0 {
		targetCaretPos := len(e.lines[caretLine-1])

		newLines := make([][]rune, len(e.lines)-1)
		for i := 0; i < caretLine; i++ {
			newLines[i] = e.lines[i]
		}
		newLines[caretLine-1] = append(e.lines[caretLine-1], e.lines[caretLine]...)
		for i := (caretLine + 1); i < len(e.lines); i++ {
			newLines[i-1] = e.lines[i]
		}
		e.lines = newLines

		// move caret to old end of previous line
		e.caretLine = caretLine - 1
		e.caretPos = targetCaretPos
		return true
	}
	return false
}

func (e *textEditor) RemoveRightOfCaret() bool {
	caretLine, caretPos := e.Caret()
	// fix caret position to currently visible position
	e.caretPos = caretPos

	if caretPos < len(e.lines[caretLine]) {
		// removing only in current line
		prefix := string(e.lines[caretLine][:caretPos])
		suffix := string(e.lines[caretLine][caretPos+1:])
		e.lines[caretLine] = []rune(prefix + suffix)
		return true
	}
	if caretPos >= len(e.lines[caretLine]) && caretLine < len(e.lines)-1 {
		newLines := make([][]rune, len(e.lines)-1)
		for i := 0; i <= caretLine; i++ {
			newLines[i] = e.lines[i]
		}
		newLines[caretLine] = append(e.lines[caretLine], e.lines[caretLine+1]...)
		for i := (caretLine + 2); i < len(e.lines); i++ {
			newLines[i-1] = e.lines[i]
		}
		e.lines = newLines
		return true
	}
	return false
}

func max(values ...int) int {
	maxVal := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] > maxVal {
			maxVal = values[i]
		}
	}
	return maxVal
}

func min(values ...int) int {
	minVal := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] < minVal {
			minVal = values[i]
		}
	}
	return minVal
}

func boundBy(val int, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
