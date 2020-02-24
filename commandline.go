package console

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	maxAutoPrintListLen = 100
)

var (
	// remember the last time Tab was pressed to detect double-tab.
	lastTabPress  = time.Unix(0, 0)
	doubleTabSpan = 250 * time.Millisecond
)

// CommandHistoryHandler describes a function that returns a command from history at the given index.
//
// Index 0 denotes the latest command. nil is returned when the number of entries in history is exceeded. The index will never be negative.
type CommandHistoryHandler func(index int) ([]string, bool)

// ReadCommandOptions configures options and callbacks for ReadComman.
type ReadCommandOptions struct {
	// GetHistoryEntry denotes the handler for reading command history.
	GetHistoryEntry CommandHistoryHandler
	// GetCompletionOptions denotes the handler for auto completion.
	GetCompletionOptions CommandCompletionHandler
	// PrintOptionsHandler denotes the handler to print options on double-tab.
	PrintOptionsHandler PrintOptionsHandler
}

// ReadCommand reads a command from console input and offers history, aswell as completion functionality.
func ReadCommand(prompt string, opts *ReadCommandOptions) ([]string, error) {
	if opts == nil {
		opts = &ReadCommandOptions{
			PrintOptionsHandler: DefaultOptionsPrinter(),
		}
	}

	var cmd []string
	err := withReadKeyContext(func() error {
		var err error
		cmd, err = readCommand(prompt, opts)
		return err
	})
	return cmd, err
}

func readCommand(prompt string, opts *ReadCommandOptions) ([]string, error) {
	var sb strings.Builder

	for {
		line, err := readCommandLine(&prompt, sb.String(), true, opts)
		if err != nil {
			return nil, err
		}

		sb.WriteString(line)

		if cmd, isComplete := ParseCommand(sb.String()); isComplete {
			return cmd, nil
		}

		// line break is part of command -> append to command because it has been omitted by the line reader
		sb.WriteRune('\n')
		// show empty prompt for further lines of same command
		prompt = ""
	}
}

func readCommandLine(prompt *string, currentCommand string, escapeHistory bool, opts *ReadCommandOptions) (string, error) {
	if prompt != nil {
		Printf("%s> ", *prompt)
	}

	var cmdToString func([]string) string
	if escapeHistory {
		cmdToString = GetCommandString
	} else {
		cmdToString = func(cmd []string) string { return strings.Join(cmd, " ") }
	}

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
		if prompt == nil {
			Printf("%s", sb.String())
		} else {
			Printf("%s> %s", *prompt, sb.String())
		}
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

		//TODO move caret along line

		switch key {
		case KeyCtrlC:
			return "", ErrCtrlC()

		case KeyEscape:
			clearLine()

		case KeyUp:
			if opts.GetHistoryEntry != nil {
				if newCmd, ok := opts.GetHistoryEntry(historyIndex + 1); ok {
					historyIndex++
					replaceLine(cmdToString(newCmd))
				}
			}
		case KeyDown:
			if opts.GetHistoryEntry != nil {
				if historyIndex >= 0 {
					historyIndex--

					if historyIndex >= 0 {
						if newCmd, ok := opts.GetHistoryEntry(historyIndex); ok {
							replaceLine(cmdToString(newCmd))
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
			if opts.GetCompletionOptions != nil {
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
				options := filterOptions(opts.GetCompletionOptions(cmd, len(cmd)-1), prefix)
				if options != nil && len(options) > 0 {
					if time.Since(lastTabPress) < doubleTabSpan {
						if opts.PrintOptionsHandler != nil {
							// double-tab detected -> print options
							Println()

							sort.Slice(options, func(i, j int) bool {
								return options[i].String() < options[j].String()
							})
							opts.PrintOptionsHandler(options)
							reprintLine()
						}
						// process next tab as single-press
						lastTabPress = time.Unix(0, 0)

					} else {
						if len(options) == 1 {
							if len(options[0].Replacement()) > 0 {
								suffix := Escape(options[0].Replacement()[len(prefix):])
								putString(suffix)

								if !options[0].IsPartial() {
									putRune(' ')
								}
							} else {
								// nothing changed? start double-tab combo
								lastTabPress = time.Now()
							}

						} else {
							longestCommonPrefix := findLongestCommonPrefix(options)
							suffix := Escape(longestCommonPrefix[len(prefix):])
							if len(suffix) > 0 {
								putString(suffix)
							} else {
								// nothing changed? start double-tab combo
								lastTabPress = time.Now()
							}
						}
					}
				} else {
					// nothing changed? start double-tab combo
					lastTabPress = time.Now()
				}
			}

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

func filterOptions(options []CompletionOption, prefix string) []CompletionOption {
	if options == nil {
		return nil
	}

	filtered := make([]CompletionOption, 0)
	for _, c := range options {
		if strings.HasPrefix(c.Replacement(), prefix) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func findLongestCommonPrefix(options []CompletionOption) string {
	if len(options) == 0 {
		return ""
	}

	longestCommonPrefix := ""
	for i := 1; ; i++ {
		if len(options[0].Replacement()) < i {
			// prefix cannot be any longer
			return longestCommonPrefix
		}

		prefix := options[0].Replacement()[:i]
		for _, c := range options {
			if !strings.HasPrefix(c.Replacement(), prefix) {
				// the next prefix would not be valid for all options
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

// CommandLineEnvironment represents a command line interface environment with history and auto-completion.
type CommandLineEnvironment struct {
	// Prompt is called when displaying a command line prompt.
	Prompt PromptHandler
	// PrintOptions is called for double-tab option printing.
	PrintOptions PrintOptionsHandler
	// UnknownCommandHandler is called to handle unknown commands. Can be nil to return an unknown command error instead.
	ExecUnknownCommand ExecUnknownCommandHandler
	// UnknownCommandCompletionHandler is used for completion of unknown commands.
	CompleteUnknownCommand CommandCompletionHandler
	// ErrorHandler is called to handle errors returned from commands. Use RecoverPanickedCommands to also handle panics here.
	ErrorHandler CommandErrorHandler
	// RecoverPanickedCommands sets whether the ErrorHandler is also called for panics.
	RecoverPanickedCommands bool
	// UseCommandNameCompletion denotes whether completion is available for command names.
	UseCommandNameCompletion bool

	history  CommandHistory
	commands map[string]Command
}

// PromptHandler defines a function that returns the current command line prompt.
type PromptHandler func() string

// ExecUnknownCommandHandler is called when processing an unknown command.
type ExecUnknownCommandHandler func(cmd string, args []string) error

// CommandErrorHandler is called when a command has returned an error.
//
// Should return nil when the error has been handled, otherwise the command handler will stop and return the error.
type CommandErrorHandler func(cmd string, args []string, err error) error

// NewCommandLineEnvironment returns a new command line environment.
func NewCommandLineEnvironment() *CommandLineEnvironment {
	return &CommandLineEnvironment{
		Prompt:       func() string { return "cle" },
		PrintOptions: DefaultOptionsPrinter(),
		ExecUnknownCommand: func(cmd string, _ []string) error {
			_, err := Printlnf("Unknown command %q", cmd)
			return err
		},
		CompleteUnknownCommand: nil,
		ErrorHandler: func(_ string, _ []string, err error) error {
			if IsErrCommandPanicked(err) {
				Printlnf("PANIC: %s", err.Error())
			} else {
				Printlnf("ERROR: %s", err.Error())
			}
			return nil
		},
		RecoverPanickedCommands:  true,
		UseCommandNameCompletion: true,
		history:                  NewCommandHistory(100),
		commands:                 make(map[string]Command),
	}
}

// SetStaticPrompt sets a constant prompt to display for command input.
func (b *CommandLineEnvironment) SetStaticPrompt(prompt string) {
	b.Prompt = func() string { return prompt }
}

func (b *CommandLineEnvironment) prompt() string {
	if b.Prompt == nil {
		return ""
	}
	return b.Prompt()
}

// RegisterCommand adds a new command to the command line environment.
func (b *CommandLineEnvironment) RegisterCommand(cmd Command) {
	b.commands[cmd.Name()] = cmd
}

// UnregisterCommand removes a command from the command line environment and returns true if it was existent before.
func (b *CommandLineEnvironment) UnregisterCommand(commandName string) bool {
	_, exists := b.commands[commandName]
	if exists {
		delete(b.commands, commandName)
	}
	return exists
}

// ReadCommand reads a command for the configured environment.
func (b *CommandLineEnvironment) ReadCommand() ([]string, error) {
	return b.readCommand(ReadCommand)
}

func (b *CommandLineEnvironment) readCommand(handler func(prompt string, opts *ReadCommandOptions) ([]string, error)) ([]string, error) {
	opts := &ReadCommandOptions{
		GetHistoryEntry:      b.history.GetHistoryEntry,
		GetCompletionOptions: b.GetCompletionOptions,
		PrintOptionsHandler:  b.PrintOptions,
	}
	cmd, err := handler(b.prompt(), opts)
	if err != nil {
		return nil, err
	}

	if len(cmd) > 0 && len(cmd[0]) > 0 {
		b.history.Put(cmd)
	}
	return cmd, nil
}

// Run reads and processes commands until an error is returned. Use ErrExit to gracefully stop processing.
func (b *CommandLineEnvironment) Run() error {
	return withReadKeyContext(func() error {
		for {
			cmd, err := b.readCommand(readCommand)
			if err != nil {
				return err
			}

			if len(cmd) > 0 {
				if err := b.ExecCommand(cmd[0], cmd[1:]); err != nil {
					if IsErrExit(err) {
						return nil
					}
					if b.ErrorHandler == nil {
						return err
					}
					b.ErrorHandler(cmd[0], cmd[1:], err)
				}
			}
		}
	})
}

// GetCompletionOptions returns completion options for the given command. This method can be used as callback for ReadCommand.
func (b *CommandLineEnvironment) GetCompletionOptions(currentCommand []string, entryIndex int) []CompletionOption {
	if entryIndex == 0 {
		if b.UseCommandNameCompletion {
			// completion for command
			options := make([]CompletionOption, 0)
			for name := range b.commands {
				options = append(options, &completionOption{replacement: name})
			}
			return options
		}
		return nil
	}

	cmd, exists := b.commands[currentCommand[0]]
	if !exists {
		if b.CompleteUnknownCommand != nil {
			return b.CompleteUnknownCommand(currentCommand, entryIndex)
		}
		return nil
	}

	return cmd.GetCompletionOptions(currentCommand, entryIndex)
}

// ExecCommand executes a command as if it has been entered in terminal.
func (b *CommandLineEnvironment) ExecCommand(cmd string, args []string) error {
	var recovered interface{}

	err := func() error {
		defer func() {
			if b.RecoverPanickedCommands {
				// recover from panic and save reason for error handling
				recovered = recover()
			}
		}()

		// execute command
		if c, exists := b.commands[cmd]; exists {
			return c.Exec(args)
		}
		if b.ExecUnknownCommand == nil {
			return ErrUnknownCommand(cmd)
		}
		return b.ExecUnknownCommand(cmd, args)
	}()

	if recovered != nil {
		return ErrCommandPanicked(recovered)
	}
	return err
}

// PrintOptionsHandler specifies a method to print options on double-tab.
type PrintOptionsHandler func([]CompletionOption)

// DefaultOptionsPrinter returns a function that is used to print options on double-tab.
//
// This method will ask the user for large lists to confirm printin.
func DefaultOptionsPrinter() PrintOptionsHandler {
	return func(options []CompletionOption) {
		if len(options) > maxAutoPrintListLen {
			Printlnf("  print all %d options? (y/N)", len(options))
			// assume is only called during command reading here (keyboard needs to be prepared)
			_, r, err := readKey()
			if err != nil {
				return
			}

			if strings.ToLower(string(r)) != "y" {
				return
			}
		}

		PrintList(options)
	}
}

// Command denotes a named command with completion and execution handler.
type Command interface {
	// Name returns the name of the command as used in the command line.
	Name() string
	// GetCompletionOptions denotes a custom completion handler as used for ReadCommand.
	GetCompletionOptions(currentCommand []string, entryIndex int) []CompletionOption
	// Exec is called to execute the command with a set of arguments.
	Exec(args []string) error
}

// ExecCommandHandler is called when processing a command. Return ErrExit to gracefully stop processing.
type ExecCommandHandler func(args []string) error

type customCommand struct {
	name              string
	completionHandler CommandCompletionHandler
	execHandler       ExecCommandHandler
}

func (c *customCommand) Name() string {
	return c.name
}
func (c *customCommand) GetCompletionOptions(currentCommand []string, entryIndex int) []CompletionOption {
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
	return &customCommand{name, func([]string, int) []CompletionOption { return nil }, func([]string) error { return ErrExit() }}
}

// NewParameterlessCommand returns a named command that takes no parameters.
func NewParameterlessCommand(name string, handler ExecCommandHandler) Command {
	return &customCommand{name, func([]string, int) []CompletionOption { return nil }, handler}
}

// NewCustomCommand returns a named command with completion and execution handler.
func NewCustomCommand(name string, completionHandler CommandCompletionHandler, execHandler ExecCommandHandler) Command {
	return &customCommand{name, completionHandler, execHandler}
}
