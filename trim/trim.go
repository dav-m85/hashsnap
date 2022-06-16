package trim

import (
	"fmt"
	"io"
	"os"

	"github.com/dav-m85/hashsnap/core"
	"github.com/dav-m85/hashsnap/state"
)

type State interface {
	Nodes() (state.Iterator, error)
}

var output io.ReadWriter = os.Stdout

func Trim(st State, verbose bool, withs ...State) error {
	nodes, err := st.Nodes()
	if err != nil {
		return err
	}
	defer nodes.Close()

	matches := make(core.HashGroup)

	if err := matches.Add(state.ReadAll(nodes)); err != nil {
		return err
	}
	if err := nodes.Error(); err != nil {
		return err
	}

	for _, w := range withs {
		nodes, err := w.Nodes()
		if err != nil {
			return err
		}
		if err := matches.Intersect(state.ReadAll(nodes)); err != nil {
			return err
		}
		if err := nodes.Error(); err != nil {
			return err
		}
	}

	var count int64
	var waste int64

	for _, g := range matches {
		if len(g.Nodes) < 2 {
			continue
		}
		if verbose {
			fmt.Fprintln(output, g)
		}
		count++
		waste = waste + int64(g.WastedSize())
	}

	fmt.Fprintf(output, "%d duplicated groups, totalling %s wasted space\n", count, core.ByteSize(waste))
	return nil
}
