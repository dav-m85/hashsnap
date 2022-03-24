package cmd

import "github.com/dav-m85/hashsnap/state"

// Options for running hsnap, set by flags
type Options struct {
	StateFilePath string
	StateFile     *state.StateFile
	WD            string
}
