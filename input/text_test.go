package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditorSimple(t *testing.T) {
	e := newTextEditor("")
	e.InsertAtCaret("foobar")
	e.MoveCaretLeft()
	e.MoveCaretLeft()
	e.MoveCaretLeft()
	e.InsertAtCaret(" ")
	assert.Equal(t, "foo bar", e.String())
}
