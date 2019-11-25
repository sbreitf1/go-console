package console

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	// ErrExit can be returned in command handlers to exit the command line interface.
	ErrExit = fmt.Errorf("exit application")

	// remember the last time Tab was pressed to detect double-tab.
	lastTabPress  = time.Unix(0, 0)
	doubleTabSpan = 250 * time.Millisecond
)

// ReadCommand reads a command from console input and offers history, aswell as completion functionality.
func ReadCommand(prompt string, getHistoryEntry CommandHistoryEntry, getCompletionCandidates CompletionCandidatesForEntry) ([]string, error) {
	var sb strings.Builder

	if err := withReadKeyContext(func() error {
		for {
			line, err := readCommandLine(prompt, sb.String(), getHistoryEntry, getCompletionCandidates)
			if err != nil {
				return err
			}

			sb.WriteString(line)

			if _, isComplete := ParseCommand(sb.String()); isComplete {
				return nil
			}

			// line break is part of command -> append to command because it has been omitted by the line reader
			sb.WriteRune('\n')
			prompt = ""
		}
	}); err != nil {
		return nil, err
	}

	cmd, _ := ParseCommand(sb.String())
	return cmd, nil
}

func readCommandLine(prompt, currentCommand string, getHistoryEntry CommandHistoryEntry, getCompletionCandidates CompletionCandidatesForEntry) (string, error) {
	Printf("%s> ", prompt)

	var sb strings.Builder

	lineLen := 0

	putRune := func(r rune) {
		sb.WriteRune(r)
		Print(string(r))
		lineLen++
	}

	putString := func(str string) {
		sb.WriteString(str)
		Print(str)
		lineLen += len(str)
	}

	clearLine := func() {
		sb.Reset()
		str1 := strings.Repeat("\b", lineLen)
		str2 := strings.Repeat(" ", lineLen)
		Printf("%s%s%s", str1, str2, str1)
		lineLen = 0
	}

	replaceLine := func(newLine string) {
		clearLine()
		putString(newLine)
	}

	reprintLine := func() {
		Printf("%s> %s", prompt, sb.String())
	}

	removeLastChar := func() {
		if lineLen > 0 {
			str := sb.String()
			sb.Reset()
			if len(str) > 0 {
				sb.WriteString(str[:len(str)-1])
			}

			Print("\b \b")
			lineLen--
		}
	}

	historyIndex := -1

	for {
		key, r, err := readKey()
		if err != nil {
			return "", err
		}

		switch key {
		case KeyCtrlC:
			return "", ErrControlC

		case KeyEscape:
			clearLine()

		case KeyUp:
			if getHistoryEntry != nil {
				newCmd := getHistoryEntry(historyIndex + 1)
				if newCmd != nil {
					historyIndex++
					replaceLine(GetCommandString(newCmd))
				}
			}
		case KeyDown:
			if getHistoryEntry != nil {
				if historyIndex >= 0 {
					historyIndex--

					if historyIndex >= 0 {
						newCmd := getHistoryEntry(historyIndex)
						if newCmd != nil {
							replaceLine(GetCommandString(newCmd))
						} else {
							// something seems to have changed -> return to initial state
							historyIndex = -1
							clearLine()
						}
					} else {
						clearLine()
					}
				}
			}

		case KeyTab:
			if getCompletionCandidates != nil {
				str := sb.String()
				cmd, _ := ParseCommand(fmt.Sprintf("%s%s", currentCommand, str))

				if len(cmd) == 0 {
					// append virtual entry to complete commands
					cmd = []string{""}
				} else {
					if str[len(str)-1] == ' ' {
						// new command part already started by whitespace, but not recognized as part of command
						// -> append empty command part for processing
						cmd = append(cmd, "")
					}
				}

				prefix := cmd[len(cmd)-1]
				candidates := filterCandidates(getCompletionCandidates(cmd, len(cmd)-1), prefix)
				if candidates != nil && len(candidates) > 0 {
					if time.Since(lastTabPress) < doubleTabSpan {
						// double-tab detected -> print candidates
						Println()
						//TODO ask for large lists
						list := make([]string, len(candidates))
						for i := range candidates {
							list[i] = Quote(candidates[i].ReplaceString)
						}
						sort.Strings(list)
						PrintList(list)
						reprintLine()

					} else {
						if len(candidates) == 1 {
							suffix := Escape(candidates[0].ReplaceString[len(prefix):])
							putString(suffix)

							if candidates[0].IsFinal {
								putRune(' ')
							}

						} else {
							longestCommonPrefix := findLongestCommonPrefix(candidates)
							suffix := Escape(longestCommonPrefix[len(prefix):])
							putString(suffix)
						}
					}
				}
			}
			lastTabPress = time.Now()

		case KeyEnter:
			Println()
			return sb.String(), nil

		case KeyBackspace:
			removeLastChar()

		case KeySpace:
			putRune(' ')

		case 0:
			putRune(r)

		default:
			// ignore unknown special keys
		}
	}
}

func filterCandidates(candidates []CompletionCandidate, prefix string) []CompletionCandidate {
	if candidates == nil {
		return nil
	}

	filtered := make([]CompletionCandidate, 0)
	for _, c := range candidates {
		if strings.HasPrefix(c.ReplaceString, prefix) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func findLongestCommonPrefix(candidates []CompletionCandidate) string {
	if len(candidates) == 0 {
		return ""
	}

	longestCommonPrefix := ""
	for i := 1; ; i++ {
		if len(candidates[0].ReplaceString) < i {
			// prefix cannot be any longer
			return longestCommonPrefix
		}

		prefix := candidates[0].ReplaceString[:i]
		for _, c := range candidates {
			if !strings.HasPrefix(c.ReplaceString, prefix) {
				// the next prefix would not be valid for all candidates
				return longestCommonPrefix
			}
		}

		longestCommonPrefix = prefix
	}
}

// ParseCommand parses a command input with escape sequences, single quotes and double quotes. The return parameter isComplete is false when a quote or escape sequence is not closed.
func ParseCommand(str string) (parts []string, isComplete bool) {
	cmd := make([]string, 0)

	var sb strings.Builder
	escape := false
	doubleQuote := false
	singleQuote := false

	for _, r := range str {
		if singleQuote {
			if r == '\'' {
				singleQuote = false
			} else {
				sb.WriteRune(r)
			}

		} else if doubleQuote {
			if escape {
				sb.WriteRune(r)
				escape = false

			} else {
				if r == '"' {
					doubleQuote = false
				} else if r == '\\' {
					escape = true
				} else {
					sb.WriteRune(r)
				}
			}
		} else if escape {
			sb.WriteRune(r)
			escape = false

		} else {
			if r == '\\' {
				escape = true
			} else if r == '\'' {
				singleQuote = true
			} else if r == '"' {
				doubleQuote = true
			} else if r == ' ' {
				if sb.Len() > 0 {
					cmd = append(cmd, sb.String())
					sb.Reset()
				}
			} else {
				sb.WriteRune(r)
			}
		}
	}

	if len(sb.String()) > 0 {
		cmd = append(cmd, sb.String())
	}

	return cmd, (!escape && !singleQuote && !doubleQuote)
}

// GetCommandString is the inverse function of Parse() and outputs a single string equal to the given command.
func GetCommandString(cmd []string) string {
	var sb strings.Builder
	for i, str := range cmd {
		if i > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteString(Quote(str))
	}
	return sb.String()
}

// Quote returns a quoted string if it contains special chars.
func Quote(str string) string {
	if NeedQuote(str) {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(strings.ReplaceAll(str, "\\", "\\\\"), "\"", "\\\""))
	}
	return str
}

// NeedQuote returns true when the string contains characters that need to be quoted or escaped.
func NeedQuote(str string) bool {
	if strings.Contains(str, " ") {
		return true
	}
	return false
}

// Escape returns a string that escapes all special chars.
func Escape(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\"", "\\\"")
	str = strings.ReplaceAll(str, "'", "\\'")
	str = strings.ReplaceAll(str, " ", "\\ ")
	str = strings.ReplaceAll(str, "\n", "\\\n")
	str = strings.ReplaceAll(str, "\r", "\\\r")
	return str
}

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
	prompt             PromptHandler
	execUnknownHandler ExecUnknownCommandHandler
	history            *CommandHistory
	commands           map[string]Command
}

// PromptHandler defines a function that returns the current command line prompt.
type PromptHandler func() string

// ExecUnknownCommandHandler is called when processing an unknown command.
type ExecUnknownCommandHandler func(cmd string, args []string) error

// NewCommandLineEnvironment returns a new command line environment.
func NewCommandLineEnvironment(prompt string) *CommandLineEnvironment {
	return &CommandLineEnvironment{
		prompt: func() string { return prompt },
		execUnknownHandler: func(cmd string, _ []string) error {
			_, err := Printlnf("Unknown function %q", cmd)
			return err
		},
		history:  NewCommandHistory(100),
		commands: make(map[string]Command),
	}
}

// SetStaticPrompt sets a constant prompt to display for command input.
func (b *CommandLineEnvironment) SetStaticPrompt(prompt string) {
	b.prompt = func() string { return prompt }
}

// SetPrompt sets the handler for dynamic prompts.
func (b *CommandLineEnvironment) SetPrompt(prompt PromptHandler) {
	b.prompt = prompt
}

// SetExecUnknownCommandHandler sets callback function to handle unknown commands.
func (b *CommandLineEnvironment) SetExecUnknownCommandHandler(handler ExecUnknownCommandHandler) {
	b.execUnknownHandler = handler
}

// RegisterCommand adds a new command to the command line environment.
func (b *CommandLineEnvironment) RegisterCommand(cmd Command) error {
	//TODO check name and conflicts
	b.commands[cmd.Name()] = cmd
	return nil
}

// ReadCommand reads a command for the configured environment.
func (b *CommandLineEnvironment) ReadCommand() ([]string, error) {
	cmd, err := ReadCommand(b.prompt(), b.history.GetHistoryEntry, b.GetCompletionCandidatesForEntry)
	if err != nil {
		return nil, err
	}

	b.history.Put(cmd)
	return cmd, nil
}

// Run reads and processes commands until an error is returned. Use ErrExit to gracefully stop processing.
func (b *CommandLineEnvironment) Run() error {
	for {
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
				if err := b.execUnknownHandler(cmd[0], cmd[1:]); err != nil {
					if err == ErrExit {
						return nil
					}
					return err
				}
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
	// Name returns the name of the command as used in the command line.
	Name() string
	// GetCompletionCandidatesForEntry denotes a custom completion handler as used for ReadCommand.
	GetCompletionCandidatesForEntry(currentCommand []string, entryIndex int) []CompletionCandidate
	// Exec is called to execute the command with a set of arguments.
	Exec(args []string) error
}

// ExecCommandHandler is called when processing a command. Return ErrExit to gracefully stop processing.
type ExecCommandHandler func(args []string) error

type customCommand struct {
	name              string
	completionHandler CompletionCandidatesForEntry
	execHandler       ExecCommandHandler
}

func (c *customCommand) Name() string {
	return c.name
}
func (c *customCommand) GetCompletionCandidatesForEntry(currentCommand []string, entryIndex int) []CompletionCandidate {
	if c.completionHandler != nil {
		return c.completionHandler(currentCommand, entryIndex)
	}
	return nil
}
func (c *customCommand) Exec(args []string) error {
	if c.execHandler != nil {
		return c.execHandler(args)
	}
	return nil
}

// NewExitCommand returns a named command to stop command line processing.
func NewExitCommand(name string) Command {
	return &customCommand{name, func([]string, int) []CompletionCandidate { return nil }, func([]string) error { return ErrExit }}
}

// NewParameterlessCommand returns a named command that takes no parameters.
func NewParameterlessCommand(name string, handler ExecCommandHandler) Command {
	return &customCommand{name, func([]string, int) []CompletionCandidate { return nil }, handler}
}

// NewCustomCommand returns a named command with completion and execution handler.
func NewCustomCommand(name string, completionHandler CompletionCandidatesForEntry, execHandler ExecCommandHandler) Command {
	return &customCommand{name, completionHandler, execHandler}
}

// CommandHistoryEntry describes a function that returns a command from history at the given index.
//
// Index 0 denotes the latest command. nil is returned when the number of entries in history is exceeded. The index will never be negative.
type CommandHistoryEntry func(index int) []string

// CompletionCandidatesForEntry describes a function that returns all completion candidates for a given command and entry.
//
// The returned candidates must include the current user input for the given entry and are filtered by the entered prefix.
type CompletionCandidatesForEntry func(currentCommand []string, entryIndex int) (candidates []CompletionCandidate)

// CompletionCandidate denotes a completion entity for a command.
type CompletionCandidate struct {
	// ReplaceString denotes the full replacement string of the completed command part.
	ReplaceString string
	// IsFinal is true, when the replacement is the final value. This will also emit a whitespace after inserting the command part.
	IsFinal bool
}

//TODO convenience method to better prepare completion candidates from arrays and maps
