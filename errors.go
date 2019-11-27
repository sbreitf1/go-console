package console

import (
	"fmt"
)

/* ################################################ */
/* ###                   exit                   ### */
/* ################################################ */

type errExit struct{}

func (e errExit) Error() string {
	return "exit application"
}

// ErrExit returns an error that indicates a graceful CLE shutdown.
func ErrExit() error {
	return errExit{}
}

// IsErrExit returns true when the error indicates a graceful CLE shutdown.
func IsErrExit(err error) bool {
	_, ok := err.(errExit)
	return ok
}

/* ################################################ */
/* ###                 ctrl + c                 ### */
/* ################################################ */

type errCtrlC struct{}

func (e errCtrlC) Error() string {
	return "Ctrl+C"
}

// ErrCtrlC returns a new error that indicates user input Ctrl+C.
func ErrCtrlC() error {
	return errCtrlC{}
}

// IsErrCtrlC returns true when the error indicates user input Ctrl+C.
func IsErrCtrlC(err error) bool {
	_, ok := err.(errCtrlC)
	return ok
}

/* ################################################ */
/* ###             unknown command              ### */
/* ################################################ */

type errUnknownCommand struct {
	commandName string
}

func (e errUnknownCommand) Error() string {
	return fmt.Sprintf("unknown command %q", e.commandName)
}

// ErrUnknownCommand returns a new error that indicates an unknown command.
func ErrUnknownCommand(commandName string) error {
	return errUnknownCommand{commandName}
}

// IsErrUnknownCommand returns true when the error indicates an unknown command.
func IsErrUnknownCommand(err error) bool {
	_, ok := err.(errUnknownCommand)
	return ok
}

/* ################################################ */
/* ###             command panicked             ### */
/* ################################################ */

type errCommandPanicked struct {
	recovered interface{}
}

func (e errCommandPanicked) Error() string {
	return fmt.Sprintf("%v", e.recovered)
}

// ErrCommandPanicked returns a new error that indicates a panicked command.
func ErrCommandPanicked(recovered interface{}) error {
	return errCommandPanicked{recovered}
}

// IsErrCommandPanicked returns true when the error indicates a panicked command.
func IsErrCommandPanicked(err error) bool {
	_, ok := err.(errCommandPanicked)
	return ok
}
