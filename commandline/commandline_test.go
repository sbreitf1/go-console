package commandline

import (
	"strings"
	"testing"

	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/consoletest"

	"github.com/stretchr/testify/assert"
)

func TestParseCompleteCommand(t *testing.T) {
	cmd, isComplete := ParseCommand(`echo foo 'say "hello world"' "white space" console.Key\ sequence "\""`)
	assert.Equal(t, []string{"echo", "foo", "say \"hello world\"", "white space", "console.Key sequence", "\""}, cmd)
	assert.True(t, isComplete)
}

func TestParseIncompleteCommand(t *testing.T) {
	cmd, isComplete := ParseCommand(`echo "foo`)
	assert.Equal(t, []string{"echo", "foo"}, cmd)
	assert.False(t, isComplete)
}

func TestReadCommand(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("foo bat\rr\n")
		cmd, err := ReadCommand("", nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo", "bar"}, cmd)
		input.AssertBufferConsumed(t)
	})
}

func TestReadCommandEscapeSequences(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString(`foo "bar\\test" '\blub' "\n_\ aha\$bar"` + "\n")
		cmd, err := ReadCommand("", nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo", "bar\\test", "\\blub", "\\n_\\ aha$bar"}, cmd)
		input.AssertBufferConsumed(t)
	})
}

func TestReadCommandUTF8(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("ö\rr\n")
		cmd, err := ReadCommand("", nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{"r"}, cmd)
		input.AssertBufferConsumed(t)
	})
}

func TestReadCommandEscape(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("foobar")
		input.PutKeys(console.KeyEscape)
		input.PutString("test\n")
		cmd, err := ReadCommand("", nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{"test"}, cmd)
		input.AssertBufferConsumed(t)
	})
}

func TestReadMultilineCommand(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("foo \"foo\nbar\" foo\\\nbar\n")
		cmd, err := ReadCommand("", nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo", "foo\nbar", "foo\nbar"}, cmd)
		input.AssertBufferConsumed(t)
	})
}

func TestCommandLineEnvironmentHistory(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutKeys(console.KeyUp, console.KeyDown)
		input.PutString("\n")
		input.PutString("p\tf\t\n")
		input.PutKeys(console.KeyUp)
		input.PutString("\nprint 1\nprint 2\n")
		input.PutKeys(console.KeyUp, console.KeyUp)
		input.PutString("\n")
		input.PutKeys(console.KeyDown, console.KeyUp, console.KeyUp, console.KeyDown)
		input.PutString("\n")
		input.PutKeys(console.KeyUp, console.KeyUp, console.KeyUp, console.KeyUp, console.KeyUp, console.KeyUp)
		input.PutString("\nexit\n")

		cle, _, sb := prepareTestCLE()

		assert.NoError(t, cle.Run())
		assert.Equal(t, ">foo<|>foo<|>1<|>2<|>1<|>1<|>foo<|", sb.String())
		input.AssertBufferConsumed(t)
	})
}

func TestCommandLineEnvironmentSimpleCompletion(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("p\tf\t\np\tb\t\nexit\n")

		cle, _, sb := prepareTestCLE()

		assert.NoError(t, cle.Run())
		assert.Equal(t, ">foo<|>bar<|", sb.String())
		input.AssertBufferConsumed(t)
	})
}

func TestCommandLineEnvironmentPartialCompletion(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("p\tf\t_\np\tp\t_\nexit\n")

		cle, _, sb := prepareTestCLE()

		assert.NoError(t, cle.Run())
		assert.Equal(t, ">foo<>_<|>part_<|", sb.String())
		input.AssertBufferConsumed(t)
	})
}

func TestCommandLineEnvironmentEmptyInputCompletion(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("\t\n")

		cle, _, sb := prepareTestCLE()
		cle.UnregisterCommand("print")

		assert.NoError(t, cle.Run())
		assert.Equal(t, "", sb.String())
		input.AssertBufferConsumed(t)
	})
}

func TestCommandLineEnvironmentSingleOptionCompletion(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("p\t\t\nexit\n")

		cle, _, sb := prepareTestCLE()
		cle.RegisterCommand(NewCustomCommand("print",
			func(cmd []string, index int) []CompletionOption {
				return []CompletionOption{
					NewCompletionOption("test", false),
				}
			},
			newPrintHandler(sb)))

		assert.NoError(t, cle.Run())
		assert.Equal(t, ">test<|", sb.String())
		input.AssertBufferConsumed(t)
	})
}

func TestCommandLineEnvironmentLongestPrefixCompletion(t *testing.T) {
	consoletest.WithMocks(func(input *consoletest.MockInput) {
		input.PutString("p\tf\t1\np\tf\t2\t1\nexit\n")

		cle, _, sb := prepareTestCLE()
		cle.RegisterCommand(NewCustomCommand("print",
			func(cmd []string, index int) []CompletionOption {
				return []CompletionOption{
					NewCompletionOption("foobar1", false),
					NewCompletionOption("foobar2", false),
					NewCompletionOption("foobar21", false),
				}
			},
			newPrintHandler(sb)))

		assert.NoError(t, cle.Run())
		assert.Equal(t, ">foobar1<|>foobar21<|", sb.String())
		input.AssertBufferConsumed(t)
	})
}

func prepareTestCLE() (*Environment, *int, *strings.Builder) {
	var sb strings.Builder
	var lastCompletionIndex int

	cle := NewEnvironment()
	cle.RegisterCommand(NewExitCommand("exit"))
	cle.RegisterCommand(NewCustomCommand("print",
		func(cmd []string, index int) []CompletionOption {
			lastCompletionIndex = index
			return []CompletionOption{
				NewLabelledCompletionOption("FOO", "foo", false),
				NewLabelledCompletionOption("FOO", "bar", false),
				NewLabelledCompletionOption("PART", "part", true),
			}
		},
		newPrintHandler(&sb)))

	return cle, &lastCompletionIndex, &sb
}

func newPrintHandler(sb *strings.Builder) func(args []string) error {
	return func(args []string) error {
		for _, a := range args {
			sb.WriteString(">")
			sb.WriteString(a)
			sb.WriteString("<")
		}
		sb.WriteString("|")
		return nil
	}
}
