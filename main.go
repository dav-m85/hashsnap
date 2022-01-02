package main

import (
	"fmt"
	"os"

	"github.com/dav-m85/hashsnap/cmd"
)

// A snapshot manipulator to ease deduplication across filesystems
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		help()
	}

	var err error
	switch args[0] {
	case "help":
		help()

	// trim, dedup, verify, list

	case "create":
		err = cmd.Create()

	case "info":
		err = cmd.Info()

	default:
		fmt.Printf("hsnap: '%s' is not a hsnap command. See 'hsnap help'.\n", args[1])
		return
	}

	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	return
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
