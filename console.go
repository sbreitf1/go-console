package console

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	listSpaceLen = 2
)

var (
	// DefaultInput can be used to redirect input sources.
	DefaultInput Input
	// DefaultOutput can be used to redirect output destinations.
	DefaultOutput Output

	newline string
)

// Input defines functionality to handle console input.
type Input interface {
	ReadLine() (string, error)
	ReadPassword() (string, error)
	BeginReadKey() error
	ReadKey() (Key, rune, error)
	EndReadKey() error
}

// Output defines functionality to handle console output and os responses.
type Output interface {
	Print(string) (int, error)
	GetSize() (int, int, error)
	SupportsColors() bool
	Exit(int)
}

type defaultInput struct {
	lastCharWasCR bool
}
type defaultOutput struct {
}

func init() {
	DefaultInput = &defaultInput{}
	DefaultOutput = &defaultOutput{}
	newline = fmt.Sprintln()
}

func (d *defaultOutput) Print(str string) (int, error) {
	return fmt.Print(str)
}

// Print writes a set of objects separated by whitespaces to Stdout.
func Print(a ...interface{}) (int, error) {
	return DefaultOutput.Print(fmt.Sprint(a...))
}

// Printf writes a formatted string to Stdout.
func Printf(format string, a ...interface{}) (int, error) {
	return DefaultOutput.Print(fmt.Sprintf(format, a...))
}

// Println writes a set of objects separated by whitespaces to Stdout and ends the line.
func Println(a ...interface{}) (int, error) {
	return DefaultOutput.Print(fmt.Sprintln(a...))
}

// Printlnf writes a formatted string to Stdout and ends the line.
func Printlnf(format string, a ...interface{}) (int, error) {
	return Println(fmt.Sprintf(format, a...))
}

// Fatal calls Print and exits with code 1.
func Fatal(a ...interface{}) {
	fatalWrapper(Print(a...))
}

// Fatalf calls Printf and exits with code 1.
func Fatalf(format string, a ...interface{}) {
	fatalWrapper(Printf(format, a...))
}

// Fatalln calls Println and exits with code 1.
func Fatalln(a ...interface{}) {
	fatalWrapper(Println(a...))
}

// Fatallnf calls Printlnf and exits with code 1.
func Fatallnf(format string, a ...interface{}) {
	fatalWrapper(Printlnf(format, a...))
}

func (d *defaultOutput) Exit(code int) {
	os.Exit(code)
}

func fatalWrapper(int, error) {
	DefaultOutput.Exit(1)
}

// PrintList prints all array or map values in a regular grid.
func PrintList(obj interface{}) error {
	width, _, err := GetSize()
	if err != nil {
		return err
	}

	list := toList(obj)

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
		sb.WriteString(newline)
	}

	_, err = Print(sb.String())
	return err
}

func toList(obj interface{}) []string {
	if obj == nil {
		return nil
	}

	var list []string

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	toString := func(v reflect.Value) string {
		return fmt.Sprintf("%v", v.Interface())
	}

	switch t.Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		list = make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			list[i] = toString(v.Index(i))
		}

	case reflect.Map:
		list = make([]string, v.Len())
		i := 0
		for it := v.MapRange(); it.Next(); {
			list[i] = toString(it.Value())
			i++
		}
	}

	return list
}

func (d *defaultOutput) GetSize() (int, int, error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

// GetSize returns the current terminal dimensions in characters.
func GetSize() (int, int, error) {
	return DefaultOutput.GetSize()
}

func (d *defaultOutput) SupportsColors() bool {
	return supportsColors()
}

// SupportsColors returns true when the current terminal supports ANSI colors.
func SupportsColors() bool {
	return DefaultOutput.SupportsColors()
}

func (d *defaultInput) ReadLine() (string, error) {
	//TODO configurable encoding
	return d.readLine(d.readRuneUTF8)
}

func (d *defaultInput) readLine(readRune func() (rune, error)) (string, error) {
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

func (*defaultInput) readRuneANSI() (rune, error) {
	var buf = [1]byte{0}
	_, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return rune(buf[0]), nil
}

func (*defaultInput) readRuneUTF8() (rune, error) {
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
	return DefaultInput.ReadLine()
}

func (d *defaultInput) ReadPassword() (string, error) {
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
	return DefaultInput.ReadPassword()
}

func (d *defaultInput) BeginReadKey() error {
	return beginReadKey()
}

func (d *defaultInput) ReadKey() (Key, rune, error) {
	return readKey()
}

func (d *defaultInput) EndReadKey() error {
	return endReadKey()
}

// BeginReadKey opens a raw TTY and allows you to use ReadKey.
func BeginReadKey() error {
	return DefaultInput.BeginReadKey()
}

// ReadKey returns a key and the corresponding rune or an error. BeginReadKey needs to be called first.
func ReadKey() (Key, rune, error) {
	return DefaultInput.ReadKey()
}

// EndReadKey closes the raw TTY opened by BeginReadKey and discards all unprocessed key events.
func EndReadKey() error {
	return DefaultInput.EndReadKey()
}

// WithReadKeyContext executes the given function with surrounding BeginReadKey and EndReadKey calls.
func WithReadKeyContext(f func() error) error {
	if err := DefaultInput.BeginReadKey(); err != nil {
		return err
	}
	defer DefaultInput.EndReadKey()

	return f()
}
