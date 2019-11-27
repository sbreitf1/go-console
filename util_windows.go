// +build windows

package console

import (
	"os"

	"golang.org/x/sys/windows"
)

const newline = "\r\n"

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
