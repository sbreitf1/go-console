package commandline

// CommandHistory defines the interface to a history of commands.
type CommandHistory interface {
	Put([]string)
	GetHistoryEntry(int) ([]string, bool)
}

type memoryCommandHistory struct {
	history    [][]string
	count, pos int
}

// NewCommandHistory returns a new command history for maxCount entries.
func NewCommandHistory(maxCount int) CommandHistory {
	return &memoryCommandHistory{make([][]string, maxCount), 0, 0}
}

// Put saves a new command to the history as latest entry.
func (h *memoryCommandHistory) Put(cmd []string) {
	//TODO command deduplication

	h.history[h.pos] = cmd
	h.pos = (h.pos + 1) % len(h.history)
	if h.count < len(h.history) {
		h.count++
	}
}

// GetHistoryEntry can be used as history callback for ReadCommand.
func (h *memoryCommandHistory) GetHistoryEntry(index int) ([]string, bool) {
	if index >= h.count {
		return nil, false
	}

	return h.history[(h.pos-1-index+len(h.history))%len(h.history)], true
}
