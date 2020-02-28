// +build !windows

package console

import (
	"os"
	"syscall"
	"unicode/utf8"
	"unsafe"

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
	ttyIn         *os.File
	ttyOut        *os.File
	ttyOldTermios syscall.Termios
	ttyBuffer     []byte
)

func beginReadKey() error {
	in, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	ttyIn = in
	out, err := os.OpenFile("/dev/tty", syscall.O_WRONLY, 0)
	if err != nil {
		ttyIn.Close()
		return err
	}
	ttyOut = out

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(in.Fd()), ioctlReadTermios, uintptr(unsafe.Pointer(&ttyOldTermios))); err != 0 {
		ttyIn.Close()
		ttyOut.Close()
		return err
	}
	newTermios := ttyOldTermios
	newTermios.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXOFF
	newTermios.Lflag &^= syscall.ECHO | syscall.ICANON
	//TODO catch Ctrl+C
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(in.Fd()), ioctlWriteTermios, uintptr(unsafe.Pointer(&newTermios))); err != 0 {
		ttyIn.Close()
		ttyOut.Close()
		return err
	}

	ttyBuffer = []byte{}
	return nil
}

func readKey() (Key, rune, error) {
	buf := make([]byte, 4)
	var bufLen int
	if len(ttyBuffer) > 0 {
		// use buffered data from previous readKey call
		bufLen = len(ttyBuffer)
		for i := 0; i < len(ttyBuffer); i++ {
			buf[i] = ttyBuffer[i]
		}
		ttyBuffer = []byte{}

	} else {
		// read new buffer
		len, err := ttyIn.Read(buf)
		if err != nil {
			return 0, 0, err
		}

		// huge amounts of data leading to multiple keys at once in the buffer will probably be caused by inserting text
		// thus, explicit actions keys are assumed to be received in a single read call

		// handle escape key sequences
		if buf[0] == 27 {
			if len == 1 {
				return KeyEscape, '^', nil
			}

			if len == 2 {
				// malformed escape sequence
				return KeyEscape, rune(buf[1]), nil
			}

			switch buf[2] {
			case 65:
				return KeyUp, 0, nil
			case 66:
				return KeyDown, 0, nil
			case 68:
				return KeyLeft, 0, nil
			case 67:
				return KeyRight, 0, nil

			default:
				// unknown escape sequence
				return KeyEscape, rune(buf[2]), nil
			}
		}

		bufLen = len
	}

	// handle some special chars
	switch buf[0] {
	case '\r':
		ttyBuffer = buf[1:bufLen]
		return KeyEnter, '\n', nil
	case '\u007f':
		ttyBuffer = buf[1:bufLen]
		return KeyBackspace, '\r', nil
	case '\t':
		ttyBuffer = buf[1:bufLen]
		return KeyTab, '\t', nil
	case ' ':
		ttyBuffer = buf[1:bufLen]
		return KeySpace, ' ', nil
	case '\x03':
		ttyBuffer = buf[1:bufLen]
		return KeyCtrlC, 0, nil
	}

	// assemble utf8-rune
	for i := 1; i < 4; i++ {
		// return complete buffer rune
		if utf8.FullRune(buf[:i]) {
			// return remainder after rune
			ttyBuffer = buf[i:bufLen]
			break
		}

		if i >= bufLen {
			// need to fill the buffer from tty input to complete the rune
			if _, err := ttyIn.Read(buf[i : i+1]); err != nil {
				return 0, 0, err
			}
			bufLen++
		}
	}

	r, _ := utf8.DecodeRune(buf)
	return 0, r, nil
}

func endReadKey() error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(ttyIn.Fd()), ioctlWriteTermios, uintptr(unsafe.Pointer(&ttyOldTermios))); err != 0 {
		return err
	}
	ttyIn.Close()
	ttyOut.Close()
	return nil
}
