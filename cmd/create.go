package cmd

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/dav-m85/hashsnap/core"
	"github.com/dav-m85/hashsnap/state"
	bar "github.com/schollz/progressbar/v3"
)

type CreateFlags struct {
	progress bool
}

var cf = new(CreateFlags)

func Create(opt Options) error {
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fl.BoolVar(&cf.progress, "progress", false, "help message for flagname")
	fl.Parse(os.Args[2:])

	if opt.StateFile != nil {
		return errors.New("already a hsnap directory or child")
	}

	// excludes := core.Exclusions{".git", ".DS_Store"}

	var pbar *bar.ProgressBar
	if cf.progress {
		pbar = bar.DefaultBytes(
			-1,
			"Hashing",
		)
	}

	st := state.NewStateFileIn(opt.WD)
	enc, close, err := st.Create()
	if err != nil {
		return err
	}
	defer close()

	// Pipeline context... cancelling it cancels them all
	ctx, cleanup := context.WithCancel(context.Background())
	defer cleanup()

	// ⛲️ Source by exploring all files
	nodes, err := core.WalkFS(ctx, opt.WD, nil)
	if err != nil {
		return err
	}

	// 🏭 Hash them all
	nodes2 := core.Hasher(ctx, pbar, nodes)

	// 🛁 Write hashes to hashfile
	for x := range nodes2 {
		// fmt.Println(x)
		if err := enc.Encode(x); err != nil {
			return err
		}
	}

	// TODO explains what happned !

	return nil
}
