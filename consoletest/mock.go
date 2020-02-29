package consoletest

import (
	"fmt"
	"testing"

	"github.com/sbreitf1/go-console"

	"github.com/stretchr/testify/assert"
)

type ReadKeyResult struct {
	Key   console.Key
	Rune  rune
	Error error
}

type MockInput struct {
	buffer          []ReadKeyResult
	bufferPos       int
	isReadKeyActive bool
}

func NewMockInput() *MockInput {
	return &MockInput{make([]ReadKeyResult, 0), 0, false}
}

func (m *MockInput) PutString(buffer string) {
	for _, r := range buffer {
		switch r {
		case '\r':
			m.buffer = append(m.buffer, ReadKeyResult{console.KeyBackspace, 0, nil})
		case '\n':
			m.buffer = append(m.buffer, ReadKeyResult{console.KeyEnter, 0, nil})
		case ' ':
			m.buffer = append(m.buffer, ReadKeyResult{console.KeySpace, 0, nil})
		case '\t':
			m.buffer = append(m.buffer, ReadKeyResult{console.KeyTab, 0, nil})

		default:
			m.buffer = append(m.buffer, ReadKeyResult{0, r, nil})
		}
	}
}

func (m *MockInput) PutKeys(keys ...console.Key) {
	for _, k := range keys {
		m.buffer = append(m.buffer, ReadKeyResult{k, 0, nil})
	}
}

func (m *MockInput) BufferConsumed() bool {
	return m.bufferPos >= len(m.buffer)
}

func (m *MockInput) AssertBufferConsumed(t *testing.T) bool {
	return assert.True(t, m.BufferConsumed(), "Not all input buffer chars have been consumed")
}

func (m *MockInput) ReadLine() (string, error) {
	panic("ReadLine not available for mock")
}

func (m *MockInput) ReadPassword() (string, error) {
	panic("ReadPassword not available for mock")
}

func (m *MockInput) BeginReadKey() error {
	if m.isReadKeyActive {
		return fmt.Errorf("double BeginReadKey call")
	}
	m.isReadKeyActive = true
	return nil
}

func (m *MockInput) ReadKey() (console.Key, rune, error) {
	if m.BufferConsumed() {
		panic("too many ReadKey calls detected")
	}

	if !m.isReadKeyActive {
		return 0, 0, fmt.Errorf("call to ReadKey before BeginReadKey")
	}

	result := m.buffer[m.bufferPos]
	m.bufferPos++
	return result.Key, result.Rune, result.Error
}

func (m *MockInput) EndReadKey() error {
	if !m.isReadKeyActive {
		return fmt.Errorf("call to EndReadKey before BeginReadKey")
	}
	m.isReadKeyActive = false
	return nil
}

func WithMocks(f func(input *MockInput)) {
	oldInput := console.DefaultInput
	oldOutput := console.DefaultOutput

	defer func() {
		console.DefaultInput = oldInput
		console.DefaultOutput = oldOutput
	}()

	input := NewMockInput()
	console.DefaultInput = input

	f(input)
}
