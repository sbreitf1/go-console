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
	oldInput := DefaultInput
	defer func() {
		DefaultInput = oldInput
	}()
	DefaultInput = &keyFakeInput{KeyEnter, '\n', nil, false}

	BeginReadKey()
	defer EndReadKey()

	key, r, err := ReadKey()
	assert.Equal(t, KeyEnter, key)
	assert.Equal(t, '\n', r)
	assert.NoError(t, err)
}

type keyFakeInput struct {
	Key       Key
	Rune      rune
	Error     error
	isReading bool
}

func (i *keyFakeInput) ReadLine() (string, error) {
	panic("ReadLine not implemented on keyFakeInput")
}
func (i *keyFakeInput) ReadPassword() (string, error) {
	panic("ReadLine not implemented on keyFakeInput")
}
func (i *keyFakeInput) BeginReadKey() error {
	if i.isReading {
		panic("BeginReadKey after BeginReadKey")
	}
	i.isReading = true
	return nil
}
func (i *keyFakeInput) ReadKey() (Key, rune, error) {
	if !i.isReading {
		panic("ReadKey before BeginReadKey")
	}
	return i.Key, i.Rune, i.Error
}
func (i *keyFakeInput) EndReadKey() error {
	if !i.isReading {
		panic("EndReadKey before BeginReadKey")
	}
	i.isReading = false
	return nil
}
