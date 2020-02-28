// +build !windows

package console

import (
	"os"
	"time"
	"unicode/utf8"

	"github.com/mattn/go-tty"
	"golang.org/x/sys/unix"
)

const ioctlReadTermios = unix.TCGETS
const ioctlWriteTermios = unix.TCSETS

func withoutEcho(f func() error) error {
	fd := int(os.Stdin.Fd())

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

func supportsColors() bool {
	return true
}

var (
	currentTTY *tty.TTY
	disableRaw func() error
	ttyCh      chan ttyEvent
)

type ttyEvent struct {
	Byte  byte
	Error error
}

func beginReadKey() error {
	obj, err := tty.Open()
	if err != nil {
		return err
	}
	currentTTY = obj
	disableRaw, err = currentTTY.Raw()
	ttyCh = make(chan ttyEvent)
	go func() {
		// resources can be closed elsewhere
		defer recover()

		buf := make([]byte, 1)
		for {
			_, err := obj.Input().Read(buf)
			if err != nil {
				ttyCh <- ttyEvent{0, err}
				return
			}
			ttyCh <- ttyEvent{buf[0], nil}
		}
	}()
	return err
}

func readKey() (Key, rune, error) {
	buf := make([]byte, 4)

	e := <-ttyCh
	if e.Error != nil {
		return 0, 0, e.Error
	}
	buf[0] = e.Byte
	if buf[0] == 27 {
		select {
		case e2 := <-ttyCh:
			if e2.Error != nil {
				return 0, 0, e2.Error
			}

			if e2.Byte == 91 {
				e3 := <-ttyCh
				if e3.Error != nil {
					return 0, 0, e3.Error
				}

				switch e3.Byte {
				case 65:
					return KeyUp, 0, nil
				case 66:
					return KeyDown, 0, nil
				case 68:
					return KeyLeft, 0, nil
				case 67:
					return KeyRight, 0, nil

				default:
					return Key(e3.Byte), 0, nil
				}
			}
			return KeyEscape, 0, nil

		case <-time.After(100 * time.Microsecond):
			// Escape key is denoted by [27] and all other escape sequences by [27 91 ...]
			// can only differentiate between Escape key and sequences, if some data is sent after 27
			// need to read from input channel, but also cancel when no data is received
			return KeyEscape, 0, nil
		}
	}

	// handle some special chars
	switch buf[0] {
	case '\r':
		return KeyEnter, '\n', nil
	case '\u007f':
		return KeyBackspace, '\r', nil
	case '\t':
		return KeyTab, '\t', nil
	case ' ':
		return KeySpace, ' ', nil
	case '\x03':
		return KeyCtrlC, 0, nil
	}

	// assemble utf8-rune
	for i := 1; i < 4; i++ {
		if utf8.FullRune(buf[:i]) {
			break
		}

		e := <-ttyCh
		if e.Error != nil {
			return 0, 0, e.Error
		}
		buf[i] = e.Byte
	}

	r, _ := utf8.DecodeRune(buf)
	return 0, r, nil
}

func endReadKey() error {
	disableRaw()
	return currentTTY.Close()
}

/*
var (
	ttyIn, ttyOut *os.File
)

func beginReadKey() error {
	out, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	in, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
	if err != nil {
		out.Close()
		return err
	}
	ttyIn = in
	ttyOut = out

	return nil
}

func readKey() (Key, rune, error) {
	buffer := make([]byte, 1024)
	l, err := ttyIn.Read(buffer)
	fmt.Println(err, l)

	return 0, 0, fmt.Errorf("blub")
}

func endReadKey() error {
	ttyIn.Close()
	ttyOut.Close()
	return nil
}
*/
