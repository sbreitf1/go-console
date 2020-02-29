package commandline

import (
	"testing"

	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/consoletest"

	"github.com/stretchr/testify/assert"
)

func TestReadLineWithHistory(t *testing.T) {
	history := NewLineHistory(2)
	history.Put("test")
	history.Put("foo bar")

	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("asdf\n")
		l, err := ReadLineWithHistory(history)
		assert.NoError(t, err)
		assert.Equal(t, "asdf", l)

		input.PutKeys(console.KeyUp, console.KeyEnter)
		l, err = ReadLineWithHistory(history)
		assert.NoError(t, err)
		assert.Equal(t, "foo bar", l)

		input.PutString("asdf")
		input.PutKeys(console.KeyUp, console.KeyUp, console.KeyEnter)
		l, err = ReadLineWithHistory(history)
		assert.NoError(t, err)
		assert.Equal(t, "test", l)

		input.PutKeys(console.KeyUp, console.KeyUp, console.KeyUp, console.KeyUp, console.KeyDown, console.KeyEnter)
		l, err = ReadLineWithHistory(history)
		assert.NoError(t, err)
		assert.Equal(t, "foo bar", l)

		input.AssertBufferConsumed(t)
	})
}
