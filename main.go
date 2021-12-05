package main

import (
	"fmt"
	"os"

	"github.com/dav-m85/hashsnap/cmd"
	"github.com/dav-m85/hashsnap/core"
	"github.com/integrii/flaggy"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		help()
	}

	switch args[0] {
	case "help":
		help()

	case "create":
		cmd.Create()

	default:
		fmt.Printf("hsnap: '%s' is not a hsnap command. See 'hsnap help'.\n", args[1])
		return
	}
	return

	flaggy.SetName("hashsnap")
	flaggy.SetDescription("A snapshot manipulator to ease deduplication across filesystems")

	var snapshot, root, output string
	var progress bool

	// convert
	convertCmd := flaggy.NewSubcommand("convert")
	convertCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	convertCmd.AddPositionalValue(&output, "file", 2, true, "Output file")
	flaggy.AttachSubcommand(convertCmd, 1)

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

	// info
	infoCmd := flaggy.NewSubcommand("info")
	infoCmd.Description = "Information about a snapshot file"
	infoCmd.AddPositionalValue(&snapshot, "file", 1, true, "Input file")
	flaggy.AttachSubcommand(infoCmd, 1)

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
	case convertCmd.Used:
		local2 := core.MakeHsnapFile(output)
		cmd.Convert(local, local2)

	// case createCmd.Used:
	// 	var err error
	// 	if root == "" {
	// 		root, err = os.Getwd()
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	}

	// 	if err := cmd.Create(root, local, progress); err != nil {
	// 		log.Fatal("Cannot create:", err)
	// 	}

	// case listCmd.Used:
	// 	cmd.List(local)

	case infoCmd.Used:
		cmd.Info(local)

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

func help() {
	fmt.Println(`usage: hsnap [--version] [--help] [-C <path>] [-c <name>=<value>]
	[--exec-path[=<path>]] [--html-path] [--man-path] [--info-path]
	[-p | --paginate | -P | --no-pager] [--no-replace-objects] [--bare]
	[--git-dir=<path>] [--work-tree=<path>] [--namespace=<name>]
	<command> [<args>]

These are common Hsnap commands used in various situations:

start a working area (see also: git help tutorial)
clone     Clone a repository into a new directory
init      Create an empty Git repository or reinitialize an existing one

work on the current change (see also: git help everyday)
add       Add file contents to the index
mv        Move or rename a file, a directory, or a symlink
restore   Restore working tree files
rm        Remove files from the working tree and from the index
`)
	os.Exit(0)
}
