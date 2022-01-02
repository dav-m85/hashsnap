package cmd

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"github.com/dav-m85/hashsnap/core"
)

func Trim() error {
	withs := []string{".hsnap", ".hsnap"}
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

	for _, g := range matches {
		if len(g.Nodes) < 2 {
			continue
		}
		fmt.Println(g)
	}
	return nil
}

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
