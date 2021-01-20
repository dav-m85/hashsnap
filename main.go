package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dav-m85/hashsnap/core"
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
		var roots []string
		base, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		roots = append(roots, base)

		snap := core.Snapshot{Root: base}
		snap.ComputeHashes()

		err = snap.SaveTo(local)
		if err != nil {
			log.Fatal("Cannot save:", err)
		}

	case dedupCmd.Used:
		snap := core.MustReadSnapshotFrom(local)

		if len(withs) == 0 {
			snap.Group().Dedup()
		} else {
			w := core.MustReadSnapshotFrom(withs[0])
			snap.DedupWith(w.Group())
		}

	default:
		fmt.Println("Use --help")
	}
}
