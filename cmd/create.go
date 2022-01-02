package cmd

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dav-m85/hashsnap/core"
	bar "github.com/schollz/progressbar/v3"
)

type CreateFlags struct {
	progress bool
}

var cf = new(CreateFlags)

func Create() error {
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fl.BoolVar(&cf.progress, "progress", false, "help message for flagname")
	fl.Parse(os.Args[2:])

	// target string, outfile core.Noder, progress bool

	target, err := os.Getwd()
	if err != nil {
		return err
	}

	// excludes := core.Exclusions{".git", ".DS_Store"}

	var pbar *bar.ProgressBar
	if cf.progress {
		pbar = bar.DefaultBytes(
			-1,
			"Hashing",
		)
	}

	filename := ".hsnap"

	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("%s already exists, aborting", filename)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)

	// Write info node
	enc.Encode(core.Info{
		Version:   1,
		RootPath:  target,
		CreatedAt: time.Now(),
	})

	// Pipeline context... cancelling it cancels them all
	ctx, cleanup := context.WithCancel(context.Background())
	defer cleanup()

	// ⛲️ Source by exploring all files
	nodes, err := core.WalkFS(ctx, target, nil)
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

	return nil
}
