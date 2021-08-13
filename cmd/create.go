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
	nodes, errc, err := core.WalkFileTree(target, excludes)(ctx)
	if err != nil {
		panic(err)
	}
	errcList = append(errcList, errc)

	// 🏭 Hash them all
	nodes2, errc, err := core.Hasher(pbar)(ctx, nodes)
	if err != nil {
		panic(err)
	}
	errcList = append(errcList, errc)

	// 🛁 Write hashes to hashfile
	errc, err = outfile.ChannelWrite()(ctx, nodes2)
	if err != nil {
		panic(err)
	}
	errcList = append(errcList, errc)

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
