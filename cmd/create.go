package cmd

import (
	"context"
	"errors"

	"github.com/dav-m85/hashsnap/core"
	"github.com/dav-m85/hashsnap/state"
	bar "github.com/schollz/progressbar/v3"
)

func Create(opt Options, pbar *bar.ProgressBar) error {
	if opt.State != nil {
		return errors.New("already a hsnap directory or child")
	}

	st := state.NewIn(opt.WD)
	enc, close, err := st.Create()
	if err != nil {
		return err
	}
	defer close()

	// Pipeline context... cancelling it cancels them all
	ctx, cleanup := context.WithCancel(context.Background())
	defer cleanup()

	// ⛲️ Source by exploring all dirs
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
