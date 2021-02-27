package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	// "github.com/dav-m85/hashsnap/core"
	"github.com/dav-m85/hashsnap/ncore"
	"github.com/integrii/flaggy"
)

func main() {
	flaggy.SetName("hashsnap")
	flaggy.SetDescription("A snapshot manipulator to ease deduplication across filesystems")

	var local string

	// create
	createCmd := flaggy.NewSubcommand("create")
	createCmd.Description = "Create a snapshot file"
	flaggy.AttachSubcommand(createCmd, 1)
	createCmd.AddPositionalValue(&local, "file", 1, true, "Output file")

	// dedup
	dedupCmd := flaggy.NewSubcommand("dedup")
	flaggy.AttachSubcommand(dedupCmd, 1)
	dedupCmd.AddPositionalValue(&local, "file", 1, true, "Input file")

	var withs []string
	dedupCmd.StringSlice(&withs, "w", "with", "Hashsnap to dedup against")

	flaggy.Parse()

	// Main subcommand handling switch
	switch {
	case createCmd.Used:
		base, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		// var roots []string
		// roots = append(roots, base)

		base = filepath.Join(base, "../silo")

		err = ncore.Create(base, "yolo.hsnap")
		if err != nil {
			log.Fatal("Cannot save:", err)
		}

	// case dedupCmd.Used:
	// 	snap := core.MustReadSnapshotFrom(local)

	// 	if len(withs) == 0 {
	// 		snap.Group().Dedup()
	// 	} else {
	// 		w := core.MustReadSnapshotFrom(withs[0])
	// 		snap.DedupWith(w.Group())
	// 	}

	default:
		fmt.Println("Use --help")
	}
}
