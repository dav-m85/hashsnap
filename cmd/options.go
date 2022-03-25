package cmd

import "github.com/dav-m85/hashsnap/state"

// Options for running hsnap, set by flags
type Options struct {
	StateFilePath string
	StateFile     *state.StateFile
	WD            string
}

func NewOptions(wd string) (opt Options) {
	opt.WD = wd

	st, err := state.StateIn(opt.WD)
	if err != nil {
		panic(err)
	}
	opt.StateFile = st

	return
}
