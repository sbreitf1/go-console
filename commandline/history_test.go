package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandHistory(t *testing.T) {
	hist := NewCommandHistory(2)
	requireHistEntryNil(t, hist, 0)

	hist.Put([]string{"foo", "bar"})
	requireHistEntry(t, hist, 0, []string{"foo", "bar"})
	requireHistEntryNil(t, hist, 1)

	hist.Put([]string{"42"})
	requireHistEntry(t, hist, 0, []string{"42"})
	requireHistEntry(t, hist, 1, []string{"foo", "bar"})
	requireHistEntryNil(t, hist, 2)

	hist.Put([]string{})
	requireHistEntry(t, hist, 0, []string{})
	requireHistEntry(t, hist, 1, []string{"42"})
	requireHistEntryNil(t, hist, 2)

	hist.Put([]string{"new", "stuff"})
	requireHistEntry(t, hist, 0, []string{"new", "stuff"})
	requireHistEntry(t, hist, 1, []string{})
	requireHistEntryNil(t, hist, 2)
}

func TestCommandHistoryDeduplication(t *testing.T) {
	hist := NewCommandHistory(4)
	requireHistEntryNil(t, hist, 0)

	hist.Put([]string{"the", "very", "first", "entry"})
	requireHistEntry(t, hist, 0, []string{"the", "very", "first", "entry"})
	requireHistEntryNil(t, hist, 1)

	hist.Put([]string{"the", "very", "first", "entry"})
	requireHistEntry(t, hist, 0, []string{"the", "very", "first", "entry"})
	requireHistEntryNil(t, hist, 1)

	hist.Put([]string{"2"})
	requireHistEntry(t, hist, 0, []string{"2"})
	requireHistEntry(t, hist, 1, []string{"the", "very", "first", "entry"})
	requireHistEntryNil(t, hist, 2)

	hist.Put([]string{"the", "very", "first", "entry"})
	requireHistEntry(t, hist, 0, []string{"the", "very", "first", "entry"})
	requireHistEntry(t, hist, 1, []string{"2"})
	requireHistEntryNil(t, hist, 2)

	hist.Put([]string{"foo", "bar"})
	requireHistEntry(t, hist, 0, []string{"foo", "bar"})
	requireHistEntry(t, hist, 1, []string{"the", "very", "first", "entry"})
	requireHistEntry(t, hist, 2, []string{"2"})
	requireHistEntryNil(t, hist, 3)

	hist.Put([]string{"42"})
	requireHistEntry(t, hist, 0, []string{"42"})
	requireHistEntry(t, hist, 1, []string{"foo", "bar"})
	requireHistEntry(t, hist, 2, []string{"the", "very", "first", "entry"})
	requireHistEntry(t, hist, 3, []string{"2"})
	requireHistEntryNil(t, hist, 4)

	hist.Put([]string{"2"})
	requireHistEntry(t, hist, 0, []string{"2"})
	requireHistEntry(t, hist, 1, []string{"42"})
	requireHistEntry(t, hist, 2, []string{"foo", "bar"})
	requireHistEntry(t, hist, 3, []string{"the", "very", "first", "entry"})
	requireHistEntryNil(t, hist, 4)

	hist.Put([]string{"foo", "bar"})
	requireHistEntry(t, hist, 0, []string{"foo", "bar"})
	requireHistEntry(t, hist, 1, []string{"2"})
	requireHistEntry(t, hist, 2, []string{"42"})
	requireHistEntry(t, hist, 3, []string{"the", "very", "first", "entry"})
	requireHistEntryNil(t, hist, 4)

	hist.Put([]string{"new", "stuff"})
	requireHistEntry(t, hist, 0, []string{"new", "stuff"})
	requireHistEntry(t, hist, 1, []string{"foo", "bar"})
	requireHistEntry(t, hist, 2, []string{"2"})
	requireHistEntry(t, hist, 3, []string{"42"})
	requireHistEntryNil(t, hist, 4)

	hist.Put([]string{"42"})
	requireHistEntry(t, hist, 0, []string{"42"})
	requireHistEntry(t, hist, 1, []string{"new", "stuff"})
	requireHistEntry(t, hist, 2, []string{"foo", "bar"})
	requireHistEntry(t, hist, 3, []string{"2"})
	requireHistEntryNil(t, hist, 4)
}

func requireHistEntry(t *testing.T, hist CommandHistory, i int, expected []string) {
	cmd, ok := hist.GetHistoryEntry(i)
	require.True(t, ok)
	require.Equal(t, expected, cmd)
}

func requireHistEntryNil(t *testing.T, hist CommandHistory, i int) {
	cmd, ok := hist.GetHistoryEntry(i)
	require.False(t, ok)
	require.Nil(t, cmd)
}
