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

	var snapshot, output, root string
	var progress bool

	// // convert
	// convertCmd := flaggy.NewSubcommand("convert")
	// convertCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	// convertCmd.AddPositionalValue(&output, "file", 2, true, "Output file")
	// flaggy.AttachSubcommand(convertCmd, 1)

	// create
	createCmd := flaggy.NewSubcommand("create")
	createCmd.Description = "Create a snapshot file"
	createCmd.AddPositionalValue(&snapshot, "file", 1, true, "Output file")
	createCmd.String(&root, "root", "r", "Root of the filetree to snapshot. Must be absolute. Defaults to current work directory.")
	createCmd.Bool(&progress, "progress", "p", "Progress bar for hashing speed.")
	flaggy.AttachSubcommand(createCmd, 1)

	// // list
	// listCmd := flaggy.NewSubcommand("list")
	// listCmd.Description = "List content of a snapshot file"
	// listCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	// flaggy.AttachSubcommand(listCmd, 1)

	// // info
	// infoCmd := flaggy.NewSubcommand("info")
	// infoCmd.Description = "Information about a snapshot file"
	// infoCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	// flaggy.AttachSubcommand(infoCmd, 1)

	// // trim
	// trimCmd := flaggy.NewSubcommand("trim")
	// trimCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	// var withs []string
	// trimCmd.StringSlice(&withs, "w", "with", "Hashsnap to dedup against")
	// flaggy.AttachSubcommand(trimCmd, 1)

	// // dedup
	// dedupCmd := flaggy.NewSubcommand("dedup")
	// dedupCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	// flaggy.AttachSubcommand(dedupCmd, 1)

	flaggy.Parse()

	// Main subcommand handling switch
	local := core.MakeHsnapFile(snapshot)

	switch {
	// case convertCmd.Used:
	// 	local2 := core.MakeHsnapFile(snapshot)
	// 	cmd.Convert(local, local2)

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

	// case listCmd.Used:
	// 	cmd.List(local)

	// case infoCmd.Used:
	// 	cmd.Info(local)

	// case dedupCmd.Used:
	// 	cmd.Dedup(local)

	// case trimCmd.Used:
	// 	var withSnaps []core.Hsnap
	// 	for _, w := range withs {
	// 		withSnaps = append(withSnaps, core.MakeHsnapFile(w))
	// 	}
	// 	cmd.Trim(local, withSnaps)

	default:
		fmt.Println("Use --help")
	}
}
