package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCompleteCommand(t *testing.T) {
	cmd, isComplete := ParseCommand(`echo foo 'say "hello world"' "white space" escape\ sequence "\""`)
	assert.Equal(t, []string{"echo", "foo", "say \"hello world\"", "white space", "escape sequence", "\""}, cmd)
	assert.True(t, isComplete)
}

func TestParseIncompleteCommand(t *testing.T) {
	cmd, isComplete := ParseCommand(`echo "foo`)
	assert.Equal(t, []string{"echo", "foo"}, cmd)
	assert.False(t, isComplete)
}

func TestReadCommand(t *testing.T) {
	withMocks(func(input *mockInput) {
		input.PutBuffer("foo bat\rr\n")
		cmd, err := ReadCommand("", nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo", "bar"}, cmd)
		input.AssertBufferConsumed(t)
	})
}
