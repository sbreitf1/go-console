package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditorSimple(t *testing.T) {
	e := newTextEditor("")
	e.InsertAtCaret("foobar")
	assert.Equal(t, "foobar", e.String())
	assert.Equal(t, []string{"foobar"}, e.Lines())
	assertCaret(t, 0, 6, e)
}

func TestEnterNewLine(t *testing.T) {
	e := newTextEditor("")
	e.InsertAtCaret("foo")
	e.NewLineAtCaret()
	e.InsertAtCaret("bar")
	assert.Equal(t, "foo\nbar", e.String())
	assert.Equal(t, []string{"foo", "bar"}, e.Lines())
	assertCaret(t, 1, 3, e)
}

func TestLineRange(t *testing.T) {
	e := newTextEditor("foo\nbar\ntest\nblub")
	assert.Equal(t, []string{}, e.LineRange(4, 2))
	assert.Equal(t, []string{"bar", "test"}, e.LineRange(1, 2))
	assert.Equal(t, []string{"foo", "bar", "test"}, e.LineRange(0, 3))
	assert.Equal(t, []string{"bar", "test", "blub"}, e.LineRange(1, 3))
	assert.Equal(t, []string{"blub"}, e.LineRange(3, 1))
}

func TestMoveCaretLeft(t *testing.T) {
	e := newTextEditor("")
	assertCaret(t, 0, 0, e)
	e.MoveCaretLeft()
	assertCaret(t, 0, 0, e)
	e.InsertAtCaret("foo")
	assertCaret(t, 0, 3, e)
	e.MoveCaretLeft()
	assertCaret(t, 0, 2, e)
	e.InsertAtCaret("_")
	assertCaret(t, 0, 3, e)
	e.NewLineAtCaret()
	assertCaret(t, 1, 0, e)
	e.MoveCaretLeft()
	assertCaret(t, 0, 3, e)
	assert.Equal(t, "fo_\no", e.String())
}

func TestMoveCaretRight(t *testing.T) {
	e := newTextEditor("")
	assertCaret(t, 0, 0, e)
	e.MoveCaretRight()
	assertCaret(t, 0, 0, e)
	e.InsertAtCaret("foo")
	assertCaret(t, 0, 3, e)
	e.MoveCaretRight()
	assertCaret(t, 0, 3, e)
	e.MoveCaretLeft()
	assertCaret(t, 0, 2, e)
	e.MoveCaretRight()
	assertCaret(t, 0, 3, e)
	e.NewLineAtCaret()
	assertCaret(t, 1, 0, e)
	e.MoveCaretLeft()
	assertCaret(t, 0, 3, e)
	e.MoveCaretRight()
	assertCaret(t, 1, 0, e)
	assert.Equal(t, "foo\n", e.String())
}

func TestMoveCaretUp(t *testing.T) {
	e := newTextEditor("")
	assertCaret(t, 0, 0, e)
	e.MoveCaretUp(1)
	assertCaret(t, 0, 0, e)
	assert.Equal(t, "", e.String())
}

func assertCaret(t *testing.T, expectedLine, expectedPos int, e *textEditor) bool {
	caretLine, caretPos := e.Caret()
	if assert.Equal(t, expectedLine, caretLine) {
		return assert.Equal(t, expectedPos, caretPos)
	}
	return false
}
