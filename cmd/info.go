package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dav-m85/hashsnap/core"
)

type InfoFlags struct {
	prefix string
}

var ifl = new(InfoFlags)

// Info opens an hsnap, read its info header and counts how many nodes it has
// it does not check for sanity (like child has a valid parent and so on)
func Info(opt Options, args []string) error {
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// fl.StringVar(&ifl.input, "input", ".hsnap", "help message for flagname")
	fl.Parse(args)
	rargs := fl.Args()

	st := opt.StateFile
	if st == nil {
		return errors.New("no state here")
	}

	nodes, err := st.Nodes()
	if err != nil {
		return err
	}
	defer nodes.Close()

	// Cycle through all nodes
	var size int64
	var count int64

	var prefix string
	if len(rargs) > 0 {
		prefix = rargs[0]
	}

	for nodes.Next() {
		n := nodes.Node()
		if n.Mode.IsDir() {
			continue
		}
		if prefix != "" && !strings.HasPrefix(n.Path(), prefix) {
			continue
		}
		fmt.Fprintf(Output, "\t%s\n", color.Green+n.Path()+color.Reset) // children is not up to date here
		size = size + n.Size
		count++
	}
	if err := nodes.Error(); err != nil {
		return err
	}

	// Write some report on stdout
	fmt.Fprintf(Output, "Totalling %s and %d files\n", core.ByteSize(size), count)
	return nil
}
