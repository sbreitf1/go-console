// +build aix darwin dragonfly freebsd linux,!appengine netbsd openbsd

package console

import (
	"golang.org/x/sys/unix"
)

func withoutEcho(f func() error) error {
	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return err
	}

	newState := *termios
	newState.Lflag &^= unix.ECHO
	newState.Lflag |= unix.ICANON | unix.ISIG
	newState.Iflag |= unix.ICRNL
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, &newState); err != nil {
		return err
	}

	defer unix.IoctlSetTermios(fd, ioctlWriteTermios, termios)

	return f()
}
