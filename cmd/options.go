package cmd

import "github.com/dav-m85/hashsnap/state"

// Options for running hsnap, set by flags
type Options struct {
	StateFilePath string
	State         *state.StateFile
	WD            string
}

func NewOptions(wd string) (opt Options) {
	opt.WD = wd

	st, err := state.LookupFrom(opt.WD)
	if err != nil {
		panic(err)
	}
	opt.State = st

	return
}
