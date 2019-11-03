package console

import (
	"fmt"
	"os"
	"strings"
)

//TODO class CLI for command line interpreting, history and auto-complete

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
	return fmt.Println(fmt.Sprintf(format, a...))
}

// Fatal calls Print and os.Exit(1).
func Fatal(a ...interface{}) error {
	return fatal(fmt.Print(a...))
}

// Fatalf calls Printf and os.Exit(1).
func Fatalf(format string, a ...interface{}) error {
	return fatal(fmt.Printf(format, a...))
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
	fmt.Println()
	return pw, nil
}
