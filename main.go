package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dav-m85/hashsnap/cmd"
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

	// info
	infoCmd := flaggy.NewSubcommand("info")
	flaggy.AttachSubcommand(infoCmd, 1)
	infoCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	infoCmd.Description = "Information about a snapshot file"

	// dedup
	dedupCmd := flaggy.NewSubcommand("dedup")
	flaggy.AttachSubcommand(dedupCmd, 1)
	dedupCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")

	var withs []string
	dedupCmd.StringSlice(&withs, "w", "with", "Hashsnap to dedup against")

	flaggy.Parse()

	// Main subcommand handling switch
	local := core.MakeHsnapFile(snapshot)

	switch {
	case createCmd.Used:
		var err error
		if root == "" {
			root, err = os.Getwd()
			if err != nil {
				panic(err)
			}
		}

		if err := cmd.Create(root, local, progress); err != nil {
			log.Fatal("Cannot create:", err)
		}

	case listCmd.Used:
		cmd.List(local)

	case infoCmd.Used:
		cmd.Info(local)

	case dedupCmd.Used:
		var withSnaps []core.Hsnap
		for _, w := range withs {
			withSnaps = append(withSnaps, core.MakeHsnapFile(w))
		}
		cmd.DedupWith(local, withSnaps)
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
