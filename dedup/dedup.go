package dedup

import (
	"github.com/dav-m85/hashsnap/core"
	"github.com/dav-m85/hashsnap/state"
)

type State interface {
	Nodes() (state.Iterator, error)
}

// var output io.ReadWriter = os.Stdout

// TODO interactive version ?
func Dedup(st State, verbose bool, keeps ...string) error {
	nodes, err := st.Nodes()
	if err != nil {
		return err
	}
	defer nodes.Close()

	matches := make(core.HashGroup)

	// Process
	// If have two nodes, select one using keeps
	// if keeps is undeciseful (keeps all nodes), then ask user prompt

	if err := matches.Add(state.ReadAll(nodes)); err != nil {
		return err
	}
	if err := nodes.Error(); err != nil {
		return err
	}

	matches.PrintDetails(verbose)
	return nil
}
