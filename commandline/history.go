package commandline

// CommandHistory defines the interface to a history of commands.
type CommandHistory interface {
	Put([]string)
	GetHistoryEntry(int) ([]string, bool)
}

type memoryCommandHistory struct {
	maxCount int
	history  [][]string
}

// NewCommandHistory returns a new command history for maxCount entries.
func NewCommandHistory(maxCount int) CommandHistory {
	return &memoryCommandHistory{maxCount, make([][]string, 0)}
}

// Put saves a new command to the history as latest entry.
func (h *memoryCommandHistory) Put(cmd []string) {
	if oldPos := h.find(cmd); oldPos >= 0 {
		// remove old entry from list
		h.history = append(h.history[:oldPos], h.history[oldPos+1:]...)
	}

	h.history = append([][]string{cmd}, h.history...)
	if len(h.history) > h.maxCount {
		h.history = h.history[:h.maxCount]
	}
}

func (h *memoryCommandHistory) find(cmd []string) int {
HistLoop:
	for i := range h.history {
		if len(h.history[i]) == len(cmd) {
			for j := range h.history[i] {
				if cmd[j] != h.history[i][j] {
					continue HistLoop
				}
			}
			return i
		}
	}
	return -1
}

// GetHistoryEntry can be used as history callback for ReadCommand.
func (h *memoryCommandHistory) GetHistoryEntry(index int) ([]string, bool) {
	if index >= len(h.history) {
		return nil, false
	}

	return h.history[index], true
}
