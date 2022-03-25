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

func Create(opt Options, args []string) error {
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fl.BoolVar(&cf.progress, "progress", false, "help message for flagname")
	fl.Parse(args)

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
	dirs, err := core.WalkFS(ctx, core.NoFiles, opt.WD, false, core.NewNodeFromPath(opt.WD))
	if err != nil {
		return err
	}

	dirtree := []*core.Node{}

	// 🛁 Write hashes to hashfile
	for x := range dirs {
		// fmt.Println(x)
		dirtree = append(dirtree, x)
		if err := enc.Encode(x); err != nil {
			return err
		}
	}

	// ⛲️ Source by exploring all files
	files, err := core.WalkFS(ctx, core.NoDirs, opt.WD, true, dirtree...)
	if err != nil {
		return err
	}

	// 🏭 Hash them all
	files2 := core.Hasher(ctx, opt.WD, pbar, files)

	// 🛁 Write hashes to hashfile
	for x := range files2 {
		// fmt.Println(x)
		if err := enc.Encode(x); err != nil {
			return err
		}
	}

	return nil
}
