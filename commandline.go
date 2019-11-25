package console

import (
	"fmt"
)

var (
	// ErrExit can be returned in command handlers to exit the command line interface.
	ErrExit = fmt.Errorf("exit application")
)

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

// CommandLineEnvironment represents a command line interface environment with history and auto-completion.
type CommandLineEnvironment struct {
	prompt   PromptHandler
	history  *CommandHistory
	commands map[string]Command
}

// PromptHandler defines a function that returns the current command line prompt.
type PromptHandler func() string

// NewCommandLineEnvironment returns a new command line environment.
func NewCommandLineEnvironment(prompt string) *CommandLineEnvironment {
	cle := &CommandLineEnvironment{nil, NewCommandHistory(100), make(map[string]Command)}
	cle.SetStaticPrompt(prompt)
	return cle
}

// SetStaticPrompt sets a constant prompt to display for command input.
func (b *CommandLineEnvironment) SetStaticPrompt(prompt string) {
	b.prompt = func() string { return prompt }
}

// SetPrompt sets the handler for dynamic prompts.
func (b *CommandLineEnvironment) SetPrompt(prompt PromptHandler) {
	b.prompt = prompt
}

// RegisterCommand adds a new command to the command line environment.
func (b *CommandLineEnvironment) RegisterCommand(cmd Command) error {
	//TODO check name and conflicts
	b.commands[cmd.Name()] = cmd
	return nil
}

// ReadCommand reads a command for the configured environment.
func (b *CommandLineEnvironment) ReadCommand() ([]string, error) {
	cmd, err := ReadCommand(b.history.GetHistoryEntry, b.GetCompletionCandidatesForEntry)
	if err != nil {
		return nil, err
	}

	b.history.Put(cmd)
	return cmd, nil
}

// Run reads and processes commands until an error is returned. Use ErrExit to gracefully stop processing.
func (b *CommandLineEnvironment) Run() error {
	for {
		Printf("%s> ", b.prompt())
		cmd, err := b.ReadCommand()
		if err != nil {
			return err
		}

		if len(cmd) > 0 {
			if c, exists := b.commands[cmd[0]]; exists {
				if err := c.Exec(cmd[1:]); err != nil {
					if err == ErrExit {
						return nil
					}
					return err
				}

			} else {
				//TODO handle unknown commands
				Printlnf("Unknown command %q", cmd[0])
			}
		}
	}
}

// GetCompletionCandidatesForEntry returns completion candidates for the given command. This method can be used as callback for ReadCommand.
func (b *CommandLineEnvironment) GetCompletionCandidatesForEntry(currentCommand []string, entryIndex int) []CompletionCandidate {
	if entryIndex == 0 {
		// completion for command
		candidates := make([]CompletionCandidate, 0)
		for name := range b.commands {
			candidates = append(candidates, CompletionCandidate{ReplaceString: name, IsFinal: true})
		}
		return candidates
	}

	cmd, exists := b.commands[currentCommand[0]]
	if !exists {
		//TODO handle unknown commands
		return nil
	}

	return cmd.GetCompletionCandidatesForEntry(currentCommand, entryIndex)
}

// Command denotes a named command with completion and execution handler.
type Command interface {
	Name() string
	GetCompletionCandidatesForEntry(currentCommand []string, entryIndex int) []CompletionCandidate
	Exec(args []string) error
}

// ExecCommandHandler is called when processing a command. Return ErrExit to gracefully stop processing.
type ExecCommandHandler func(args []string) error

type parameterlessCommand struct {
	name    string
	handler ExecCommandHandler
}

// NewParameterlessCommand returns a command that takes no parameters.
func NewParameterlessCommand(name string, handler ExecCommandHandler) Command {
	return &parameterlessCommand{name, handler}
}

func (c *parameterlessCommand) Name() string {
	return c.name
}
func (c *parameterlessCommand) GetCompletionCandidatesForEntry(currentCommand []string, entryIndex int) []CompletionCandidate {
	return nil
}
func (c *parameterlessCommand) Exec(args []string) error {
	if c.handler != nil {
		return c.handler(args)
	}
	return nil
}
