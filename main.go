package main

import (
	"fmt"
	"os"

	"github.com/dav-m85/hashsnap/cmd"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		help()
	}

	var err error
	switch args[0] {

	case "create":
		err = cmd.Create()

	case "convert":
		err = cmd.Convert()

	case "help":
		help()

	case "info":
		err = cmd.Info(args[1:])

	case "trim":
		err = cmd.Trim()

	default:
		fmt.Printf("hsnap: '%s' is not a hsnap command. See 'hsnap help'.\n", args[0])
		return
	}

	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func help() {
	fmt.Print(`usage: hsnap <command> [<args>]

These are common hsnap commands used in various situations:

create    Make a snapshot for current working directory
info      Detail content of a snapshot
trim      Deduplicate current working directory using snapshots
help      This help message
`)
	os.Exit(0)
}
