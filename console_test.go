package console

import (
	"testing"

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
		input.PutString(str)
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

type readKeyResult struct {
	Key   Key
	Rune  rune
	Error error
}

type mockInput struct {
	buffer    []readKeyResult
	bufferPos int
}

func newMockInput() *mockInput {
	return &mockInput{make([]readKeyResult, 0), 0}
}

func (m *mockInput) PutString(buffer string) {
	for _, r := range buffer {
		switch r {
		case '\r':
			m.buffer = append(m.buffer, readKeyResult{KeyBackspace, 0, nil})
		case '\n':
			m.buffer = append(m.buffer, readKeyResult{KeyEnter, 0, nil})
		case ' ':
			m.buffer = append(m.buffer, readKeyResult{KeySpace, 0, nil})
		case '\t':
			m.buffer = append(m.buffer, readKeyResult{KeyTab, 0, nil})

		default:
			m.buffer = append(m.buffer, readKeyResult{0, r, nil})
		}
	}
}

func (m *mockInput) PutKeys(keys ...Key) {
	for _, k := range keys {
		m.buffer = append(m.buffer, readKeyResult{k, 0, nil})
	}
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

	result := m.buffer[m.bufferPos]
	m.bufferPos++
	return result.Key, result.Rune, result.Error
}

func (m *mockInput) EndReadKey() error {
	return nil
}