package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/dav-m85/hashsnap"
	bar "github.com/schollz/progressbar/v3"
)

var wd, spath string
var delete bool

// st, err := state.LookupFrom(opt.WD)
// State         *state.StateFile

var output io.Writer = os.Stdout

var color = struct {
	Reset, Red, Green, Yellow, Blue, Purple, Cyan, Gray, White string
}{
	Reset:  "\033[0m",
	Red:    "\033[31m",
	Green:  "\033[32m",
	Yellow: "\033[33m",
	Blue:   "\033[34m",
	Purple: "\033[35m",
	Cyan:   "\033[36m",
	Gray:   "\033[37m",
	White:  "\033[97m",
}

var (
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
		fs.StringVar(&spath, "statefile", "", "Use a different state file")
		fs.StringVar(&wd, "wd", "", "Use a different working directory")
	}
}

var verbose bool

func main() {
	setupCommonFlags()

	if len(os.Args) <= 1 {
		help()
	}

	createCmd.BoolVar(&verbose, "verbose", false, "displays hashing speed")
	infoCmd.BoolVar(&verbose, "verbose", false, "enumerates all files (high-mem)")
	trimCmd.BoolVar(&verbose, "verbose", false, "list all groups")
	trimCmd.BoolVar(&delete, "delete", false, "really deletes stuff")

	cm := subcommands[os.Args[1]]
	if cm == nil {
		log.Fatalf("Unknown subcommand '%s', see help for more details.", os.Args[1])
	}

	cm.Parse(os.Args[2:])

	// Making sure wd and spath are properly set
	if err := cleanwd(); err != nil {
		panic(err)
	}
	if spath == "" {
		var err error
		spath, err = hashsnap.LookupFrom(wd)
		if err != nil {
			panic(err)
		}
	}
	if spath == "" {
		spath = filepath.Join(wd, hashsnap.STATE_NAME)
	}

	// Main command switch
	var err error
	switch cm.Name() {

	case createCmd.Name():
		var pbar io.Writer = io.Discard
		if verbose {
			pbar = bar.DefaultBytes(
				-1,
				"Hashing",
			)
		}
		err = create(pbar)

	case helpCmd.Name():
		help()

	case infoCmd.Name():
		err = info()

	// case dedupCmd.Name():
	// 	err = dedup.Dedup(opt.State, verbose, cm.Args()...)

	case trimCmd.Name():
		if len(cm.Args()) == 0 {
			err = fmt.Errorf("wrong usage")
			break
		}
		err = trim(delete, cm.Args()...)

	default:
		log.Fatalf("Subcommand '%s' is not implemented!", cm.Name())
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// dedup     Good old deduplication, with man guards
func help() {
	fmt.Print(`usage: hsnap <command> [<args>]

These are common hsnap commands used in various situations:

create    Make a snapshot for current working directory
info      Detail content of a snapshot
trim      Remove local files that are already present in provided snapshots
help      This help message
`)
	os.Exit(0)
}

func cleanwd() (err error) {
	if wd == "" {
		wd, err = os.Getwd()
	} else {
		wd, err = filepath.Abs(wd)
	}
	return
}

func readTree(path string) (*hashsnap.Tree, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return hashsnap.ReadTree(f)
}

func create(spy io.Writer) error {
	if path, err := hashsnap.LookupFrom(spath); path != "" || err != nil {
		return fmt.Errorf("already a hsnap directory or child in %s: %s", path, err)
	}

	f, err := os.OpenFile(spath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	c := hashsnap.Snapshot(wd, f, spy)

	fmt.Fprintf(output, "Encoded %d files", c)

	return nil
}

// info opens an hsnap, read its info header and counts how many nodes it has
// it does not check for sanity (like child has a valid parent and so on)
// TODO feature: check nodes exists
func info() error {
	f, err := os.OpenFile(spath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	i := new(hashsnap.Info)
	if err = dec.Decode(i); err != nil {
		return err
	}
	fmt.Fprintf(output, "%s\n", i)

	// Cycle through all nodes
	var size int64
	var count int64

	// fmt.Fprintf(output, "\t%s %s\n", color.Green+t.RelPath(n)+color.Reset, hashsnap.ByteSize(n.Size)) // children is not up to date here

	err = hashsnap.DecodeNodes(dec, func(n *hashsnap.Node) error {
		size = size + n.Size
		count++
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(output, "Totalling %s and %d files\n", hashsnap.ByteSize(size), count)
	return nil
}

func trim(delete bool, withs ...string) error {
	cur, err := readTree(spath)
	if err != nil {
		return err
	}

	var trees []*hashsnap.Tree
	for _, w := range withs {
		x, err := readTree(w)
		if err != nil {
			return err
		}
		trees = append(trees, x)
	}

	matches := cur.Trim(trees...)
	matches.PruneSingleTreeGroups()

	var count int64
	var waste int64

	for _, g := range matches {

		if verbose {

			s := fmt.Sprintf("%d nodes (save %s)\n", len(g.Nodes), g.WastedSize())
			for _, n := range g.Nodes {
				if n.Tree() == cur {
					s = s + fmt.Sprintf(color.Red+"\t- %s [%s]\n"+color.Reset, n, n.Tree().RelPath(n))
				} else {
					s = s + fmt.Sprintf(color.Green+"\t+ %s\n"+color.Reset, n)
				}
			}

			fmt.Fprintln(output, s)
		}
		if delete {
			for _, n := range g.Nodes {
				if n.Tree() == cur {
					err := os.Remove(n.Tree().AbsPath(n))
					if err != nil {
						return err
					}
				}
			}
		}
		count++
		waste = waste + int64(WastedSize(g))
	}

	fmt.Fprintf(output, "%d duplicated groups, totalling %s wasted space\n", count, hashsnap.ByteSize(waste))
	return nil
}
