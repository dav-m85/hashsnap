package cmd

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dav-m85/hashsnap/core"
)

type InfoFlags struct {
	input string
}

var ifl = new(InfoFlags)

// Info opens an hsnap, read its info header and counts how many nodes it has
// it does not check for sanity (like child has a valid parent and so on)
func Info() error {
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fl.StringVar(&ifl.input, "input", ".hsnap", "help message for flagname")
	fl.Parse(os.Args[2:])

	f, err := os.Open(ifl.input)
	if err != nil {
		return fmt.Errorf("cannot open file: %s", err)
	}
	defer f.Close()
	dec := gob.NewDecoder(f)

	// Read the info header
	var h *core.Info = &core.Info{}
	err = dec.Decode(h)
	if err != nil {
		return fmt.Errorf("cannot decode info header: %s", err)
	}

	ndec := core.NewDecoder(dec)

	// Cycle through all nodes
	var size int64
	var count int64

	for {
		n := core.Node{}
		err := ndec.Decode(&n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot decode node: %s", err)
		}
		if n.Mode.IsDir() {
			continue
		}
		// fmt.Println(n) // children is not up to date here
		size = size + n.Size
		count++
	}

	// Write some report on stdout
	fmt.Printf("Totalling %s and %d files\n", core.ByteSize(size), count)
	return nil
}
