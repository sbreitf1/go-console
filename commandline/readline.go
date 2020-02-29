package commandline

import (
	"github.com/sbreitf1/go-console"
)

// LineHistory defines the interface to a history of raw lines.
type LineHistory interface {
	Put(string)
	GetHistoryEntry(int) (string, bool)
}

type memoryLineHistory struct {
	history    []string
	count, pos int
}

// NewLineHistory returns a new line history for maxCount entries.
func NewLineHistory(maxCount int) LineHistory {
	return &memoryLineHistory{make([]string, maxCount), 0, 0}
}

// Put saves a new command to the history as latest entry.
func (h *memoryLineHistory) Put(line string) {
	//TODO command deduplication

	h.history[h.pos] = line
	h.pos = (h.pos + 1) % len(h.history)
	if h.count < len(h.history) {
		h.count++
	}
}

// GetHistoryEntry can be used as history callback for ReadLineWithHistory.
func (h *memoryLineHistory) GetHistoryEntry(index int) (string, bool) {
	if index >= h.count {
		return "", false
	}

	return h.history[(h.pos-1-index+len(h.history))%len(h.history)], true
}

// ReadLineWithHistory reads a line from Stdin and allows to select previous options using the Up and Down keys.
func ReadLineWithHistory(history LineHistory) (string, error) {
	if err := console.BeginReadKey(); err != nil {
		return "", err
	}
	defer console.EndReadKey()

	return readLineWithHistory(history)
}

func readLineWithHistory(history LineHistory) (string, error) {
	opts := ReadCommandOptions{
		GetHistoryEntry: func(index int) ([]string, bool) {
			if line, ok := history.GetHistoryEntry(index); ok {
				return []string{line}, true
			}
			return nil, false
		},
	}

	return readCommandLine(nil, "", false, &opts)
}
