package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/eiannone/keyboard"
)

var (
	// ErrControlC is returned when a read command has been aborted by Ctrl+C user input.
	ErrControlC = fmt.Errorf("Ctrl+C")
)

// Print writes a set of objects separated by whitespaces to Stdout.
func Print(a ...interface{}) (int, error) {
	return fmt.Print(a...)
}

// Printf writes a formatted string to Stdout.
func Printf(format string, a ...interface{}) (int, error) {
	return fmt.Printf(format, a...)
}

// Println writes a set of objects separated by whitespaces to Stdout and ends the line.
func Println(a ...interface{}) (int, error) {
	return fmt.Println(a...)
}

// Printlnf writes a formatted string to Stdout and ends the line.
func Printlnf(format string, a ...interface{}) (int, error) {
	return Println(fmt.Sprintf(format, a...))
}

// Fatal calls Print and os.Exit(1).
func Fatal(a ...interface{}) error {
	return fatal(Print(a...))
}

// Fatalf calls Printf and os.Exit(1).
func Fatalf(format string, a ...interface{}) error {
	return fatal(Printf(format, a...))
}

// Fatalln calls Println and os.Exit(1).
func Fatalln(a ...interface{}) error {
	return fatal(Println(a...))
}

// Fatallnf calls Printlnf and os.Exit(1).
func Fatallnf(format string, a ...interface{}) error {
	return fatal(Printlnf(format, a...))
}

func fatal(_ int, err error) error {
	if err != nil {
		return err
	}
	os.Exit(1)
	return nil
}

var lastCharWasCR bool

// ReadLine reads a line from Stdin.
//
// This method should not be used in conjunction with Stdin read from other packages as it might leave an orphaned '\n' in the input buffer for '\r\n' line breaks.
func ReadLine() (string, error) {
	return readLineANSI()
}

func readLineANSI() (string, error) {
	var sb strings.Builder

	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return sb.String(), err
		}

		if n > 0 {
			if buf[0] == '\r' {
				lastCharWasCR = true
				return sb.String(), nil
			} else if buf[0] == '\n' {
				if lastCharWasCR {
					// just ignore that char to be compatible with windows \r\n
					lastCharWasCR = false
				} else {
					lastCharWasCR = false
					return sb.String(), nil
				}
			} else {
				lastCharWasCR = false
				sb.Write(buf)
			}
		}
	}
}

// ReadPassword reads a line from Stdin while hiding the user input.
//
// This method should not be used in conjunction with Stdin read from other packages as it might leave an orphaned '\n' in the input buffer for '\r\n' line breaks.
func ReadPassword() (string, error) {
	var pw string
	if err := withoutEcho(func() error {
		line, err := ReadLine()
		pw = line
		return err
	}); err != nil {
		return "", err
	}
	// print suppressed line-feed
	Println()
	return pw, nil
}

// ReadKey reads a single key from terminal input and returns it along with the corresponding rune.
func ReadKey() (Key, rune, error) {
	var key Key
	var r rune
	var err error
	withReadKeyContext(func() error {
		key, r, err = readKey()
		return nil
	})
	return key, r, err
}

func readKey() (Key, rune, error) {
	char, key, err := keyboard.GetKey()
	if err != nil {
		return 0, 0, err
	}

	// re-map keys
	switch key {
	case keyboard.KeyBackspace:
		key = keyboard.KeyBackspace2
	}

	return Key(key), char, nil
}

func withReadKeyContext(f func() error) error {
	if err := keyboard.Open(); err != nil {
		return err
	}
	defer keyboard.Close()

	return f()
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

// ReadCommand reads a command from console input and offers history, aswell as completion functionality.
func ReadCommand(getHistoryEntry CommandHistoryEntry, getCompletionCandidates CompletionCandidatesForEntry) ([]string, error) {
	var sb strings.Builder

	if err := withReadKeyContext(func() error {
		for {
			line, err := readCommandLine(sb.String(), getHistoryEntry, getCompletionCandidates)
			if err != nil {
				return err
			}

			sb.WriteString(line)

			if _, isComplete := ParseCommand(sb.String()); isComplete {
				return nil
			}

			// line break is part of command -> append to command because it has been omitted by the line reader
			sb.WriteRune('\n')
			Print("> ")
		}
	}); err != nil {
		return nil, err
	}

	cmd, _ := ParseCommand(sb.String())
	return cmd, nil
}

func readCommandLine(currentCommand string, getHistoryEntry CommandHistoryEntry, getCompletionCandidates CompletionCandidatesForEntry) (string, error) {
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
		sb.WriteString(newLine)
		Print(newLine)
		lineLen = len(newLine)
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
			//TODO print on double-tab
			if getCompletionCandidates != nil {
				str := sb.String()
				cmd, _ := ParseCommand(fmt.Sprintf("%s%s", currentCommand, str))

				//TODO use next command part when space is typed
				if len(cmd) == 0 {
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
				if candidates != nil {
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

		case KeyEnter:
			Println()
			return sb.String(), nil

		case KeyBackspace:
			if lineLen > 0 {
				str := sb.String()
				sb.Reset()
				if len(str) > 0 {
					sb.WriteString(str[:len(str)-1])
				}

				Print("\b \b")
				lineLen--
			}

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
	//TODO quote string
	return str
}

// Escape returns a string that escapes all special chars.
func Escape(str string) string {
	return str
}
