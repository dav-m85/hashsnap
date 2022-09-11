package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dav-m85/hashsnap"
	"github.com/google/uuid"
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
main:
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
		var withsNonce []uuid.UUID
		var i *hashsnap.Info
		if i, err = readInfo(spath); err != nil {
			break
		}
		withsNonce = append(withsNonce, i.Nonce)

		var withs []string
		if len(cm.Args()) == 0 {
			err = fmt.Errorf("wrong usage")
			break
		}

		for _, wpath := range cm.Args() {
			if i, err = readInfo(wpath); err != nil {
				break main
			}
			for _, x := range withsNonce {
				if x == i.Nonce {
					err = fmt.Errorf("%s is already in the trim set", wpath)
					break main
				}
			}
			withsNonce = append(withsNonce, i.Nonce)
			withs = append(withs, wpath)
		}

		err = trim(delete, withs...)

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

func create(spy io.Writer) error {
	// errors.New("already a hsnap directory or child")
	// _, err := os.Stat(sf.Path)
	// st := state.New(filepath.Join(wd, state.STATE_NAME))

	f, err := os.OpenFile(spath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)

	// Write info node
	err = enc.Encode(hashsnap.Info{
		Version:   1,
		RootPath:  wd,
		CreatedAt: time.Now(),
		Nonce:     uuid.New(),
	})
	if err != nil {
		return err
	}

	// Pipeline context... cancelling it cancels them all
	ctx, cleanup := context.WithCancel(context.Background())
	defer cleanup()

	skipper := func(n fs.FileInfo) bool {
		return !n.Mode().IsDir() && (!n.Mode().IsRegular() || n.Size() == 0 || n.Name() == hashsnap.STATE_NAME)
	}

	// ⛲️ Source by exploring all nodes
	files, err := hashsnap.WalkFS(ctx, skipper, wd)
	if err != nil {
		return err
	}

	// 🏭 Hash them all and write hashes to statefile
	var c int
	for x := range hashsnap.Hasher(ctx, wd, spy, files) {
		c++
		if err := enc.Encode(x); err != nil {
			return err
		}
	}

	fmt.Fprintf(output, "Encoded %d files", c)

	return nil
}

// info opens an hsnap, read its info header and counts how many nodes it has
// it does not check for sanity (like child has a valid parent and so on)
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

	nodes := &hashsnap.DecoderIterator{
		Decoder: dec,
	}

	// TODO check file exists

	// Cycle through all nodes
	var size int64
	var count int64

	if verbose {
		t := hashsnap.NewTree(i)
		if err := t.ReadIterator(nodes); err != nil {
			return fmt.Errorf("statefile %s nodes error: %w", spath, err)
		}
		ti := hashsnap.NewTreeIterator(t)
		for ti.Next() {
			n := ti.Node()
			if n.Mode.IsDir() {
				continue
			}

			fmt.Fprintf(output, "\t%s %s\n", color.Green+t.RelPath(n)+color.Reset, hashsnap.ByteSize(n.Size)) // children is not up to date here

			size = size + n.Size
			count++
		}
		if err := ti.Error(); err != nil {
			return err
		}
	} else {
		for nodes.Next() {
			n := nodes.Node()
			size = size + n.Size
			count++
		}
		if err := nodes.Error(); err != nil {
			return err
		}
	}

	fmt.Fprintf(output, "Totalling %s and %d files\n", hashsnap.ByteSize(size), count)
	return nil
}

func readTree(path string) (*hashsnap.Tree, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	i := new(hashsnap.Info)
	if err = dec.Decode(i); err != nil {
		return nil, err
	}

	nodes := &hashsnap.DecoderIterator{
		Decoder: dec,
	}

	t := hashsnap.NewTree(i)

	if err := t.ReadIterator(nodes); err != nil {
		return nil, err
	}

	return t, nil
}

func readInfo(path string) (*hashsnap.Info, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	i := new(hashsnap.Info)
	if err = dec.Decode(i); err != nil {
		return nil, err
	}

	return i, nil
}

func trim(delete bool, withs ...string) error {
	matches := make(hashsnap.HashGroup)

	cur, err := readTree(spath)
	if err != nil {
		return err
	}
	if err := hashsnap.Each(hashsnap.NewTreeIterator(cur), matches.Add); err != nil {
		return err
	}

	for _, w := range withs {
		x, err := readTree(w)
		if err != nil {
			return err
		}
		if err := hashsnap.Each(hashsnap.NewTreeIterator(x), matches.Intersect); err != nil {
			return err
		}
	}

	var count int64
	var waste int64

	for _, g := range matches {
		if len(g.Nodes) < 2 {
			continue
		}
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
		waste = waste + int64(g.WastedSize())
	}

	fmt.Fprintf(output, "%d duplicated groups, totalling %s wasted space\n", count, hashsnap.ByteSize(waste))
	return nil
}
