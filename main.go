package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dav-m85/hashsnap/cmd"
	"github.com/dav-m85/hashsnap/state"
)

func main() {
	opt := cmd.Options{}
	flag.StringVar(&opt.StateFilePath, "statefile", "", "Different state file")
	flag.StringVar(&opt.WD, "wd", "", "Different working directory")
	flag.Parse()

	if opt.WD == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		opt.WD = wd
	}

	if opt.StateFilePath != "" {
		opt.StateFile = state.NewStateFile(opt.StateFilePath)
	} else {
		st, err := state.StateIn(opt.WD)
		if err != nil {
			panic(err)
		}
		opt.StateFile = st
	}

	args := flag.Args()
	if len(args) == 0 {
		help()
	}

	var err error
	switch args[0] {

	case "create":
		err = cmd.Create(opt, os.Args[2:])

	case "convert":
		err = cmd.Convert()

	case "help":
		help()

	case "info":
		err = cmd.Info(opt, args[1:])

	case "trim":
		err = cmd.Trim(opt)

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
