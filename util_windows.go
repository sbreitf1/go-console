//go:build windows

package console

import (
	"os"

	"github.com/eiannone/keyboard"
	"golang.org/x/sys/windows"
)

func withoutEcho(f func() error) error {
	fd := os.Stdin.Fd()

	var st uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &st); err != nil {
		return err
	}
	old := st

	st &^= (windows.ENABLE_ECHO_INPUT)
	st |= (windows.ENABLE_PROCESSED_INPUT | windows.ENABLE_LINE_INPUT | windows.ENABLE_PROCESSED_OUTPUT)
	if err := windows.SetConsoleMode(windows.Handle(fd), st); err != nil {
		return err
	}

	defer windows.SetConsoleMode(windows.Handle(fd), old)

	var h windows.Handle
	p, _ := windows.GetCurrentProcess()
	if err := windows.DuplicateHandle(p, windows.Handle(fd), p, &h, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return err
	}

	file := os.NewFile(uintptr(h), "stdin")
	defer file.Close()

	return f()
}

func supportsColors() bool {
	//TODO check for ANSI color support
	return false
}

func beginReadKey() error {
	//return keyboard.Open()
	return nil
}

func readKey() (Key, rune, error) {
	// does not work when inserting text
	//char, key, err := keyboard.GetKey()
	char, key, err := keyboard.GetSingleKey()
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

func endReadKey() error {
	//keyboard.Close()
	return nil
}
