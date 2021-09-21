package cmd

import (
	"context"

	"github.com/dav-m85/hashsnap/core"
	bar "github.com/schollz/progressbar/v3"
)

func Create(target string, outfile core.Hsnap, progress bool) error {
	// excludes := core.Exclusions{".git", ".DS_Store"}

	var pbar *bar.ProgressBar
	if progress {
		pbar = bar.DefaultBytes(
			-1,
			"Hashing",
		)
	}

	// Pipeline context... cancelling it cancels them all
	ctx, cleanup := context.WithCancel(context.Background())

	// ⛲️ Source by exploring all files
	nodes, err := core.WalkFS(ctx, target, nil)
	if err != nil {
		panic(err)
	}

	// 🏭 Hash them all
	nodes2 := core.Hasher(ctx, pbar, nodes)

	// 🛁 Write hashes to hashfile
	err = outfile.ChannelWrite(nodes2)
	if err != nil {
		panic(err)
	}

	cleanup()

	return nil
}
