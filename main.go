package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dav-m85/hashsnap/core"
	"github.com/integrii/flaggy"
)

func main() {
	flaggy.SetName("hashsnap")
	flaggy.SetDescription("A snapshot manipulator to ease deduplication across filesystems")

	var snapshot, root string
	var progress bool

	// create
	createCmd := flaggy.NewSubcommand("create")
	createCmd.Description = "Create a snapshot file"
	flaggy.AttachSubcommand(createCmd, 1)
	createCmd.AddPositionalValue(&snapshot, "file", 1, true, "Output file")
	createCmd.String(&root, "root", "r", "Root of the filetree to snapshot. Must be absolute. Defaults to current work directory.")
	createCmd.Bool(&progress, "progress", "p", "Progress bar for hashing speed.")

	// list
	listCmd := flaggy.NewSubcommand("list")
	flaggy.AttachSubcommand(listCmd, 1)
	listCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	listCmd.Description = "List content of a snapshot file"

	// dedup
	dedupCmd := flaggy.NewSubcommand("dedup")
	flaggy.AttachSubcommand(dedupCmd, 1)
	dedupCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")

	var withs []string
	dedupCmd.StringSlice(&withs, "w", "with", "Hashsnap to dedup against")

	flaggy.Parse()

	// Main subcommand handling switch
	switch {
	case createCmd.Used:
		if !strings.HasSuffix(snapshot, ".hsnap") {
			log.Fatal("Snapshot file name should end with .hsnap")
		}

		var err error
		if root == "" {
			root, err = os.Getwd()
			if err != nil {
				panic(err)
			}
		}

		if err := core.Create(root, snapshot, progress); err != nil {
			log.Fatal("Cannot create:", err)
		}

	case listCmd.Used:
		core.List(snapshot)

	case dedupCmd.Used:
		core.Dedup(snapshot, withs)
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
