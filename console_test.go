package console

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestNilToList(t *testing.T) {
	assert.Nil(t, toList(nil))
}

func TestSliceToList(t *testing.T) {
	v := []int{1, 2, 3, 4}
	list := toList(v)
	assert.Equal(t, []string{"1", "2", "3", "4"}, list)
}

func TestArrayToList(t *testing.T) {
	v := [4]int{1, 2, 3, 4}
	list := toList(v)
	assert.Equal(t, []string{"1", "2", "3", "4"}, list)
}

func TestMapToList(t *testing.T) {
	v := make(map[string]int)
	v["foo"] = 4
	v["bar"] = 2
	list := toList(v)
	assert.Len(t, list, 2)
	assert.Contains(t, list, "4")
	assert.Contains(t, list, "2")
}

type readKeyResult struct {
	Key   Key
	Rune  rune
	Error error
}

func TestReadKey(t *testing.T) {
	str := "aü ?\n"
	expected := []readKeyResult{
		readKeyResult{0, 'a', nil},
		readKeyResult{0, 'ü', nil},
		readKeyResult{KeySpace, 0, nil},
		readKeyResult{0, '?', nil},
		readKeyResult{KeyEnter, 0, nil},
	}

	withMocks(func(input *mockInput) {
		input.PutBuffer(str)
		for _, e := range expected {
			k, r, err := ReadKey()
			if !assert.Equal(t, e.Key, k, "Expected Key %s", e.Key) {
				break
			}
			if !assert.Equal(t, e.Rune, r, "Expected Rune %s", string(e.Rune)) {
				break
			}
			if !assert.Equal(t, e.Error, err) {
				break
			}
		}
		input.AssertBufferConsumed(t)
	})
}

func withMocks(f func(input *mockInput)) {
	oldInput := DefaultInput
	oldOutput := DefaultOutput

	defer func() {
		DefaultInput = oldInput
		DefaultOutput = oldOutput
	}()

	input := newMockInput()
	DefaultInput = input

	f(input)
}

type mockInput struct {
	buffer    []byte
	bufferPos int
}

func newMockInput() *mockInput {
	return &mockInput{}
}

func (m *mockInput) PutBuffer(buffer string) {
	m.buffer = []byte(buffer)
	m.bufferPos = 0
}

func (m *mockInput) BufferConsumed() bool {
	return m.bufferPos >= len(m.buffer)
}

func (m *mockInput) AssertBufferConsumed(t *testing.T) bool {
	return assert.True(t, m.BufferConsumed(), "Not all input buffer chars have been consumed")
}

func (m *mockInput) ReadLine() (string, error) {
	panic("ReadLine not available for mock")
}

func (m *mockInput) ReadPassword() (string, error) {
	panic("ReadPassword not available for mock")
}

func (m *mockInput) BeginReadKey() error {
	return nil
}

func (m *mockInput) ReadKey() (Key, rune, error) {
	if m.BufferConsumed() {
		panic("too many ReadKey calls detected")
	}

	r, size := utf8.DecodeRune(m.buffer[m.bufferPos:])
	m.bufferPos += size

	switch r {
	case '\r':
		return KeyBackspace, 0, nil
	case '\n':
		return KeyEnter, 0, nil
	case ' ':
		return KeySpace, 0, nil
	case '\t':
		return KeyTab, 0, nil

	default:
		return 0, r, nil
	}
}

func (m *mockInput) EndReadKey() error {
	return nil
}
