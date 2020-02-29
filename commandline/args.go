package commandline

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CommandCompletionHandler describes a function that returns all completion options for a given command and entry.
//
// The returned options must include the current user input for the given entry and are filtered by the entered prefix.
type CommandCompletionHandler func(currentCommand []string, entryIndex int) (options []CompletionOption)

// CompletionOption deontes a completion option for a command.
type CompletionOption interface {
	fmt.Stringer
	// Value denotes the full replacement string of the completed command part.
	Replacement() string
	// IsPartial is true, when the replacement is not the final value. This will prevent emitting a whitespace after inserting the command part so the input stays in the current command.
	IsPartial() bool
}

type completionOption struct {
	label       string
	replacement string
	isPartial   bool
}

// NewCompletionOption returns a new completion option.
func NewCompletionOption(replacement string, isPartial bool) CompletionOption {
	return &completionOption{replacement: replacement, isPartial: isPartial}
}

// NewLabelledCompletionOption returns a new completion option with label.
func NewLabelledCompletionOption(label, replacement string, isPartial bool) CompletionOption {
	return &completionOption{label: label, replacement: replacement, isPartial: isPartial}
}

func (c *completionOption) String() string {
	if len(c.label) > 0 {
		return c.label
	}
	return c.replacement
}

func (c *completionOption) Replacement() string {
	return c.replacement
}

func (c *completionOption) IsPartial() bool {
	return c.isPartial
}

//TODO util func to easily map any slice/array/map to options with IsPartial flag

// PrepareCompletionOptions returns a list of completion options with given isPartial flag.
func PrepareCompletionOptions(list []string, isPartial bool) []CompletionOption {
	options := make([]CompletionOption, len(list))
	for i := range list {
		options[i] = &completionOption{replacement: list[i], isPartial: isPartial}
	}
	return options
}

// ArgCompletion denotes an abstract definition for an argument in a completion chain.
type ArgCompletion interface {
	GetCompletionOptions(currentCommand []string, entryIndex int) (options []CompletionOption)
}

// NewFixedArgCompletion returns a completion handler for a fixed set of arguments
//
// The result can directly be used as completion handler for Command definitions.
func NewFixedArgCompletion(args ...ArgCompletion) CommandCompletionHandler {
	return func(currentCommand []string, entryIndex int) (options []CompletionOption) {
		if entryIndex > 0 && entryIndex <= len(args) {
			return args[entryIndex-1].GetCompletionOptions(currentCommand, entryIndex)
		}
		return nil
	}
}

type oneOfArgCompletion struct {
	options []CompletionOption
}

// NewOneOfArgCompletion returns a completion handler for a static list of options.
func NewOneOfArgCompletion(options ...string) ArgCompletion {
	return &oneOfArgCompletion{options: PrepareCompletionOptions(options, false)}
}

func (a *oneOfArgCompletion) GetCompletionOptions(currentCommand []string, entryIndex int) []CompletionOption {
	return a.options
}

type localFileSystemArgCompletion struct {
	withFiles bool
}

// NewLocalFileSystemArgCompletion returns a completion handler to browse the local file system.
func NewLocalFileSystemArgCompletion(withFiles bool) ArgCompletion {
	return &localFileSystemArgCompletion{withFiles}
}

func (a *localFileSystemArgCompletion) GetCompletionOptions(currentCommand []string, entryIndex int) []CompletionOption {
	options, _ := LocalFileSystemCompletion("", currentCommand[entryIndex], a.withFiles)
	return options
}

// LocalFileSystemCompletion returns the completion options for browsing the given directory.
//
// The typical usage is to browse the current working directory using the current command entry:
// LocalFileSystemCompletion("", currentCommand[entryIndex], ...)
func LocalFileSystemCompletion(workingDir, currentCommandEntry string, withFiles bool) ([]CompletionOption, error) {
	if len(workingDir) == 0 {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	//TODO support for . and ..

	var dir string
	if filepath.IsAbs(currentCommandEntry) {
		// absolute paths completely ignore the working dir
		if len(currentCommandEntry) == 1 {
			dir = currentCommandEntry
		} else {
			if strings.HasSuffix(currentCommandEntry, string(filepath.Separator)) {
				// search the given dir
				dir = currentCommandEntry
			} else {
				// search the parent dir of the given path (only a desired prefix has been entered yet)
				dir = filepath.Dir(currentCommandEntry)
			}
		}
	} else {
		if len(currentCommandEntry) == 0 {
			// no path entered yet? search in working dir
			dir = workingDir
		} else {
			if strings.HasSuffix(currentCommandEntry, string(filepath.Separator)) {
				// search the given dir
				dir = filepath.Join(workingDir, currentCommandEntry)
			} else {
				// search the parent dir of the given path (only a desired prefix has been entered yet)
				dir = filepath.Dir(filepath.Join(workingDir, currentCommandEntry))
			}
		}
		// path separator is part of the working directory
		if !strings.HasSuffix(workingDir, string(filepath.Separator)) {
			workingDir += string(filepath.Separator)
		}
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	options := make([]CompletionOption, 0)
	for _, f := range files {
		if withFiles || f.IsDir() {
			var suffix string
			if filepath.IsAbs(currentCommandEntry) {
				// complete path is required
				suffix = filepath.Join(dir, f.Name())
			} else {
				// only the part without working dir is required
				suffix = filepath.Join(dir, f.Name())[len(workingDir):]
			}
			label := f.Name()
			if f.IsDir() {
				// end path with path separator to allow tabbing to child dir
				suffix += string(filepath.Separator)
				label += string(filepath.Separator)
			}
			options = append(options, &completionOption{label: label, replacement: suffix, isPartial: f.IsDir()})
		}
	}
	return options, nil
}
