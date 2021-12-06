package cmd

import (
	"context"
	"flag"
	"os"

	"github.com/dav-m85/hashsnap/core"
	bar "github.com/schollz/progressbar/v3"
)

type CreateFlags struct {
	progress bool
}

var cf = new(CreateFlags)

func Create() {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	f.BoolVar(&cf.progress, "progress", false, "help message for flagname")
	f.Parse(os.Args[2:])

	// target string, outfile core.Noder, progress bool

	target, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	outfile := core.MakeHsnapFile(".hsnap")

	// excludes := core.Exclusions{".git", ".DS_Store"}

	var pbar *bar.ProgressBar
	if cf.progress {
		pbar = bar.DefaultBytes(
			-1,
			"Hashing",
		)
	}

	// Pipeline context... cancelling it cancels them all
	ctx, cleanup := context.WithCancel(context.Background())
	defer cleanup()

	// ⛲️ Source by exploring all files
	nodes, err := core.WalkFS(ctx, target, nil)
	if err != nil {
		panic(err)
	}

	// 🏭 Hash them all
	nodes2 := core.Hasher(ctx, pbar, nodes)

	// 🛁 Write hashes to hashfile
	err = outfile.Write(nodes2)
	if err != nil {
		panic(err)
	}
}
