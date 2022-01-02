package cmd

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"github.com/dav-m85/hashsnap/core"
)

// Info opens an hsnap, read its info header and counts how many nodes it has
// it does not check for sanity (like child has a valid parent and so on)
func Info() error {
	f, err := os.Open(".hsnap")
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
	var size uint64
	var count uint64

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
