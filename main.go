package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dav-m85/hashsnap/cmd"
	"github.com/dav-m85/hashsnap/dedup"
	"github.com/dav-m85/hashsnap/state"
	"github.com/dav-m85/hashsnap/trim"
	"github.com/google/uuid"
	bar "github.com/schollz/progressbar/v3"
)

var (
	opt cmd.Options

	createCmd = flag.NewFlagSet("create", flag.ExitOnError)
	infoCmd   = flag.NewFlagSet("info", flag.ExitOnError)
	helpCmd   = flag.NewFlagSet("help", flag.ExitOnError)
	trimCmd   = flag.NewFlagSet("trim", flag.ExitOnError)
	dedupCmd  = flag.NewFlagSet("dedup", flag.ExitOnError)
)

var subcommands = map[string]*flag.FlagSet{
	createCmd.Name(): createCmd,
	helpCmd.Name():   helpCmd,
	infoCmd.Name():   infoCmd,
	trimCmd.Name():   trimCmd,
	dedupCmd.Name():  dedupCmd,
}

func setupCommonFlags() {
	for _, fs := range subcommands {
		fs.StringVar(&opt.StateFilePath, "statefile", "", "Use a different state file")
		fs.StringVar(&opt.WD, "wd", "", "Use a different working directory")
	}
}

var verbose bool

func main() {
	setupCommonFlags()

	if len(os.Args) <= 1 {
		help()
	}

	createCmd.BoolVar(&verbose, "verbose", false, "list all groups")
	trimCmd.BoolVar(&verbose, "verbose", false, "list all groups")

	cm := subcommands[os.Args[1]]
	if cm == nil {
		log.Fatalf("Unknown subcommand '%s', see help for more details.", os.Args[1])
	}

	cm.Parse(os.Args[2:])

	if opt.WD == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		opt.WD = wd
	} else {
		var err error
		opt.WD, err = filepath.Abs(opt.WD)
		if err != nil {
			panic(err)
		}
	}

	if opt.StateFilePath != "" {
		opt.State = state.New(opt.StateFilePath)
	} else {
		st, err := state.LookupFrom(opt.WD)
		if err != nil {
			panic(err)
		}
		opt.State = st
	}

	var err error
main:
	switch cm.Name() {

	case createCmd.Name():
		var pbar *bar.ProgressBar
		if verbose {
			pbar = bar.DefaultBytes(
				-1,
				"Hashing",
			)
		}
		err = cmd.Create(opt, pbar)

	case helpCmd.Name():
		help()

	case infoCmd.Name():
		err = cmd.Info(opt)

	case dedupCmd.Name():
		err = dedup.Dedup(opt.State, verbose, cm.Args()...)

	case trimCmd.Name():
		var withsNonce []uuid.UUID
		err = opt.State.ReadInfo()
		if err != nil {
			break
		}
		withsNonce = append(withsNonce, opt.State.Info.Nonce)

		var withs []trim.State
		if len(cm.Args()) == 0 {
			err = fmt.Errorf("wrong usage")
			break
		}

		for _, wpath := range cm.Args() {
			w := state.New(wpath)
			err = w.ReadInfo()
			if err != nil {
				break main
			}
			for _, x := range withsNonce {
				if x == w.Info.Nonce {
					err = fmt.Errorf("file has already been imported once")
					break main
				}
			}
			withsNonce = append(withsNonce, w.Info.Nonce)
			var ts trim.State = w
			withs = append(withs, ts)
		}

		err = trim.Trim(opt.State, verbose, withs...)

	// case "convert":
	// 	err = cmd.Convert()

	default:
		log.Fatalf("Subcommand '%s' is not implemented!", cm.Name())
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func help() {
	fmt.Print(`usage: hsnap <command> [<args>]

These are common hsnap commands used in various situations:

create    Make a snapshot for current working directory
info      Detail content of a snapshot
dedup     Good old deduplication, with man guards
trim      Remove local files that are already present in provided snapshots
help      This help message
`)
	os.Exit(0)
}
