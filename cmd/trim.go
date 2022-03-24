package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/dav-m85/hashsnap/core"
	"github.com/dav-m85/hashsnap/state"
	"github.com/google/uuid"
)

type TrimFlags struct {
	verbose bool
}

var tf = new(TrimFlags)

func Trim(opt Options) error {
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fl.BoolVar(&tf.verbose, "verbose", false, "list all groups")
	fl.Parse(os.Args[2:])

	if len(fl.Args()) == 0 {
		return fmt.Errorf("wrong usage")
	}

	withs := fl.Args()

	// TEST this could be provided by main
	st := opt.StateFile
	if st == nil {
		return errors.New("not an hsnap directory or child")
	}
	nodes, err := st.Nodes()
	if err != nil {
		return err
	}
	defer nodes.Close()
	withsNonce = append(withsNonce, st.Info.Nonce)

	matches := make(core.HashGroup)

	if err := matches.Add(state.ReadAll(nodes)); err != nil {
		return err
	}
	if err := nodes.Error(); err != nil {
		return err
	}

	for _, w := range withs {
		ns := state.NewStateFile(w)
		nodes, err := ns.Nodes()
		if err != nil {
			return err
		}
		// TODO withsNonce = append(withsNonce, st.Info.Nonce)
		// for _, x := range withsNonce {
		// 	if x == st.Info.Nonce {
		// 		return nil, fmt.Errorf("file has already been imported once")
		// 	}
		// }
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
