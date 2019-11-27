package console

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CompletionCandidatesForEntry describes a function that returns all completion candidates for a given command and entry.
//
// The returned candidates must include the current user input for the given entry and are filtered by the entered prefix.
type CompletionCandidatesForEntry func(currentCommand []string, entryIndex int) (candidates []CompletionCandidate)

// CompletionCandidate denotes a completion entity for a command.
type CompletionCandidate struct {
	// Label denotes the label that is visible for completion lists. ReplaceString is shown if Label is empty.
	Label string
	// ReplaceString denotes the full replacement string of the completed command part.
	ReplaceString string
	// IsFinal is true, when the replacement is the final value. This will also emit a whitespace after inserting the command part.
	IsFinal bool
}

// PrepareCandidates returns a list of completion candidates with isFinal flag.
func PrepareCandidates(list []string, isFinal bool) []CompletionCandidate {
	candidates := make([]CompletionCandidate, len(list))
	for i := range list {
		candidates[i] = CompletionCandidate{ReplaceString: list[i], IsFinal: isFinal}
	}
	return candidates
}

// ArgCompletion denotes an abstract definition for an argument in a completion chain.
type ArgCompletion interface {
	GetCompletionCandidates(currentCommand []string, entryIndex int) (candidates []CompletionCandidate)
}

// NewFixedArgCompletion returns a completion handler for a fixed set of arguments
//
// The result can directly be used as completion handler for Command definitions.
func NewFixedArgCompletion(args ...ArgCompletion) CompletionCandidatesForEntry {
	return func(currentCommand []string, entryIndex int) (candidates []CompletionCandidate) {
		if entryIndex > 0 && entryIndex <= len(args) {
			return args[entryIndex-1].GetCompletionCandidates(currentCommand, entryIndex)
		}
		return nil
	}
}

type oneOfArgCompletion struct {
	candidates []CompletionCandidate
}

// NewOneOfArgCompletion returns a completion handler for a static list of options.
func NewOneOfArgCompletion(options ...string) ArgCompletion {
	return &oneOfArgCompletion{candidates: PrepareCandidates(options, true)}
}

func (a *oneOfArgCompletion) GetCompletionCandidates(currentCommand []string, entryIndex int) []CompletionCandidate {
	return a.candidates
}

type localFileSystemArgCompletion struct {
	withFiles bool
}

// NewLocalFileSystemArgCompletion returns a completion handler to browse the local file system.
func NewLocalFileSystemArgCompletion(withFiles bool) ArgCompletion {
	return &localFileSystemArgCompletion{withFiles}
}

func (a *localFileSystemArgCompletion) GetCompletionCandidates(currentCommand []string, entryIndex int) []CompletionCandidate {
	candidates, _ := BrowseCandidates("", currentCommand[entryIndex], a.withFiles)
	return candidates
}

// BrowseCandidates returns the completion candidates for browsing the given directory.
//
// The typical usage is to browse the current working directory using the current command entry:
// BrowseCandidates("", currentCommand[entryIndex], ...)
func BrowseCandidates(workingDir, currentCommandEntry string, withFiles bool) ([]CompletionCandidate, error) {
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

	candidates := make([]CompletionCandidate, 0)
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
			candidates = append(candidates, CompletionCandidate{Label: label, ReplaceString: suffix, IsFinal: !f.IsDir()})
		}
	}
	return candidates, nil
}