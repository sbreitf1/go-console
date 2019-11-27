package console

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/eiannone/keyboard"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	listSpaceLen = 2
)

var (
	// DefaultIO can be used to redirect input and output sources.
	DefaultIO IO
)

// IO defines functionality to handle console input and output.
type IO interface {
	Print(string) (int, error)
	ReadLine() (string, error)
	ReadPassword() (string, error)
	ReadKey() (Key, rune, error)
	GetSize() (int, int, error)
}

type defaultIO struct {
	lastCharWasCR bool
}

func init() {
	DefaultIO = &defaultIO{}
}

func (d *defaultIO) Print(str string) (int, error) {
	return fmt.Print(str)
}

// Print writes a set of objects separated by whitespaces to Stdout.
func Print(a ...interface{}) (int, error) {
	return DefaultIO.Print(fmt.Sprint(a...))
}

// Printf writes a formatted string to Stdout.
func Printf(format string, a ...interface{}) (int, error) {
	return DefaultIO.Print(fmt.Sprintf(format, a...))
}

// Println writes a set of objects separated by whitespaces to Stdout and ends the line.
func Println(a ...interface{}) (int, error) {
	return DefaultIO.Print(fmt.Sprintln(a...))
}

// Printlnf writes a formatted string to Stdout and ends the line.
func Printlnf(format string, a ...interface{}) (int, error) {
	return Println(fmt.Sprintf(format, a...))
}

// Fatal calls Print and os.Exit(1).
func Fatal(a ...interface{}) {
	fatal(Print(a...))
}

// Fatalf calls Printf and os.Exit(1).
func Fatalf(format string, a ...interface{}) {
	fatal(Printf(format, a...))
}

// Fatalln calls Println and os.Exit(1).
func Fatalln(a ...interface{}) {
	fatal(Println(a...))
}

// Fatallnf calls Printlnf and os.Exit(1).
func Fatallnf(format string, a ...interface{}) {
	fatal(Printlnf(format, a...))
}

func fatal(int, error) {
	os.Exit(1)
}

// PrintList prints a list of strings in a regular grid.
func PrintList(list []string) error {
	width, _, err := GetSize()
	if err != nil {
		return err
	}

	maxItemLen := 0
	for _, item := range list {
		if len(item) > maxItemLen {
			maxItemLen = len(item)
		}
	}

	var sb strings.Builder
	space := strings.Repeat(" ", listSpaceLen)

	itemsPerLine := (width + listSpaceLen) / (maxItemLen + listSpaceLen)
	lineCount := len(list) / itemsPerLine
	if len(list) > (lineCount * itemsPerLine) {
		lineCount++
	}

	if itemsPerLine == 0 {
		// fallback for very small terminals (or exceedingly large list items)
		itemsPerLine = 1
		lineCount = len(list)
	}

	for l := 0; l < lineCount; l++ {
		for i := 0; i < itemsPerLine; i++ {
			index := l*itemsPerLine + i
			if index >= len(list) {
				break
			}
			if i > 0 {
				sb.WriteString(space)
			}
			sb.WriteString(list[index])
			sb.WriteString(strings.Repeat(" ", maxItemLen-len(list[index])))
		}
		sb.WriteString(fmt.Sprintln())
	}

	_, err = Print(sb.String())
	return err
}

func (d *defaultIO) GetSize() (int, int, error) {
	return terminal.GetSize(0)
}

// GetSize returns the current terminal dimensions in characters.
func GetSize() (int, int, error) {
	return DefaultIO.GetSize()
}

func (d *defaultIO) ReadLine() (string, error) {
	//TODO configurable encoding
	return d.readLine(d.readRuneUTF8)
}

func (d *defaultIO) readLine(readRune func() (rune, error)) (string, error) {
	var sb strings.Builder

	for {
		r, err := readRune()
		if err != nil {
			return sb.String(), err
		}

		if r == '\r' {
			d.lastCharWasCR = true
			return sb.String(), nil
		} else if r == '\n' {
			if d.lastCharWasCR {
				// just ignore that char to be compatible with windows \r\n
				d.lastCharWasCR = false
			} else {
				d.lastCharWasCR = false
				return sb.String(), nil
			}
		} else {
			d.lastCharWasCR = false
			sb.WriteRune(r)
		}
	}
}

func (*defaultIO) readRuneANSI() (rune, error) {
	var buf = [1]byte{0}
	_, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return rune(buf[0]), nil
}

func (*defaultIO) readRuneUTF8() (rune, error) {
	// utf8 runes can take 1 up to 4 bytes
	var buf = [4]byte{0}
	_, err := os.Stdin.Read(buf[0:1])
	if err != nil {
		return 0, err
	}

	// most common case: rune takes exactly one byte
	if !utf8.FullRune(buf[0:1]) {
		// not complete yet? read next byte and check again
		for i := 1; i < 4; i++ {
			// put next byte into buffer
			_, err := os.Stdin.Read(buf[i : i+1])
			if err != nil {
				return 0, err
			}
			if i < 3 {
				// skip check for last rune -> will terminate either way
				if utf8.FullRune(buf[0 : i+1]) {
					break
				}
			}
		}
	}

	r, _ := utf8.DecodeRune(buf[:])
	return r, nil
}

// ReadLine reads a line from Stdin.
//
// This method should not be used in conjunction with Stdin read from other packages as it might leave an orphaned '\n' in the input buffer for '\r\n' line breaks.
func ReadLine() (string, error) {
	return DefaultIO.ReadLine()
}

func (d *defaultIO) ReadPassword() (string, error) {
	var pw string
	if err := withoutEcho(func() error {
		line, err := d.ReadLine()
		pw = line
		return err
	}); err != nil {
		return "", err
	}
	// print suppressed line-feed
	Println()
	return pw, nil
}

// ReadPassword reads a line from Stdin while hiding the user input.
//
// This method should not be used in conjunction with Stdin read from other packages as it might leave an orphaned '\n' in the input buffer for '\r\n' line breaks.
func ReadPassword() (string, error) {
	return DefaultIO.ReadPassword()
}

func (d *defaultIO) ReadKey() (Key, rune, error) {
	var key Key
	var r rune
	var err error
	withReadKeyContext(func() error {
		key, r, err = readKey()
		return nil
	})
	return key, r, err
}

// ReadKey reads a single key from terminal input and returns it along with the corresponding rune.
func ReadKey() (Key, rune, error) {
	return DefaultIO.ReadKey()
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
