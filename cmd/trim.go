package cmd

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dav-m85/hashsnap/core"
	"github.com/google/uuid"
)

type TrimFlags struct {
	verbose bool
}

var tf = new(TrimFlags)

func Trim() error {
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fl.BoolVar(&tf.verbose, "verbose", false, "list all groups")
	fl.Parse(os.Args[2:])

	if len(fl.Args()) == 0 {
		return fmt.Errorf("wrong usage")
	}

	withs := fl.Args()

	matches := make(core.HashGroup)
	for _, w := range withs {
		ns, err := readNodes(w)
		if err != nil {
			return err
		}
		err = matches.Add(ns)
		if err != nil {
			return err
		}
	}

	var count int64
	var waste int64

	for _, g := range matches {
		if len(g.Nodes) < 2 {
			continue
		}
		if tf.verbose {
			fmt.Println(g)
		}
		count++
		waste = waste + int64(g.WastedSize())
	}

	fmt.Printf("%d duplicated groups, totalling %s wasted space\n", count, core.ByteSize(waste))
	return nil
}

var withsNonce []uuid.UUID

func readNodes(file string) ([]*core.Node, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %s", err)
	}
	defer f.Close()
	dec := gob.NewDecoder(f)

	// Read the info header
	var h *core.Info = &core.Info{}
	err = dec.Decode(h)
	if err != nil {
		return nil, fmt.Errorf("cannot decode info header: %s", err)
	}
	for _, x := range withsNonce {
		if x == h.Nonce {
			return nil, fmt.Errorf("file has already been imported once")
		}
	}
	withsNonce = append(withsNonce, h.Nonce)

	ndec := core.NewDecoder(dec)

	var r []*core.Node

	for {
		n := core.Node{}
		err := ndec.Decode(&n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("cannot decode node: %s", err)
		}
		r = append(r, &n)
	}

	return r, nil
}
