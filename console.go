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
