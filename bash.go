package console

// CommandHistory saves a fixed number of the latest commands.
type CommandHistory struct {
	history    [][]string
	count, pos int
}

// NewCommandHistory returns a new command history for maxCount entries.
func NewCommandHistory(maxCount int) *CommandHistory {
	return &CommandHistory{make([][]string, maxCount), 0, 0}
}

// Put saves a new command to the history as latest entry.
func (h *CommandHistory) Put(cmd []string) {
	//TODO command deduplication

	h.history[h.pos] = cmd
	h.pos = (h.pos + 1) % len(h.history)
	if h.count < len(h.history) {
		h.count++
	}
}

// GetHistoryEntry can be used as history callback for ReadCommand.
func (h *CommandHistory) GetHistoryEntry(index int) []string {
	if index >= h.count {
		return nil
	}

	return h.history[(h.pos-1-index+len(h.history))%len(h.history)]
}

// Bash represents a command line interface environment with history and auto-completion.
type Bash struct {
	history *CommandHistory
}

// NewBash returns a new Bash object.
func NewBash(maxHistoryCount int) *Bash {
	return &Bash{NewCommandHistory(maxHistoryCount)}
}

//TODO register commands and auto-completion handlers

// ReadCommand reads a command for the configured environment.
func (b *Bash) ReadCommand() ([]string, error) {
	cmd, err := ReadCommand(b.history.GetHistoryEntry, nil)
	if err != nil {
		return nil, err
	}

	b.history.Put(cmd)
	return cmd, nil
}
