package cmd

import (
	"context"
	"log"

	"github.com/dav-m85/hashsnap/core"
	bar "github.com/schollz/progressbar/v3"
)

func Create(target string, outfile core.Hsnap, progress bool) error {
	excludes := core.Exclusions{".git", ".DS_Store"}

	var pbar *bar.ProgressBar
	if progress {
		pbar = bar.DefaultBytes(
			-1,
			"Hashing",
		)
	}

	// Pipeline context... cancelling it cancels them all
	ctx, cleanup := context.WithCancel(context.Background())

	var errcList []<-chan error

	// ⛲️ Source by exploring all files
	nodes, err := core.WalkFS(ctx, target, excludes)
	if err != nil {
		panic(err)
	}

	// 🏭 Hash them all
	nodes2, err := core.Hasher(ctx, pbar, nodes)
	if err != nil {
		panic(err)
	}

	// 🛁 Write hashes to hashfile
	err = outfile.ChannelWrite(nodes2)
	if err != nil {
		panic(err)
	}

	log.Printf("Pipeline started, processing...")
	err = core.WaitForPipeline(errcList...)
	if err != nil {
		panic(err)
	}
	cleanup()

	log.Printf("Pipeline done!")

	// log.Printf("Created snapshot with %d files\n", count)

	return nil
}
