package main

import (
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/dav-m85/hashsnap"
	bar "github.com/schollz/progressbar/v3"
)

var wd, spath string
var delete, quiet bool

var version string = "dev"

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
	createCmd  = flag.NewFlagSet("create", flag.ExitOnError)
	infoCmd    = flag.NewFlagSet("info", flag.ExitOnError)
	nodeCmd    = flag.NewFlagSet("node", flag.ExitOnError)
	helpCmd    = flag.NewFlagSet("help", flag.ExitOnError)
	trimCmd    = flag.NewFlagSet("trim", flag.ExitOnError)
	listCmd    = flag.NewFlagSet("ls", flag.ExitOnError)
	diffCmd    = flag.NewFlagSet("diff", flag.ExitOnError)
	checkCmd   = flag.NewFlagSet("check", flag.ExitOnError)
	versionCmd = flag.NewFlagSet("version", flag.ExitOnError)
)

var subcommands = map[string]*flag.FlagSet{
	createCmd.Name():  createCmd,
	helpCmd.Name():    helpCmd,
	infoCmd.Name():    infoCmd,
	nodeCmd.Name():    nodeCmd,
	trimCmd.Name():    trimCmd,
	listCmd.Name():    listCmd,
	diffCmd.Name():    diffCmd,
	checkCmd.Name():   checkCmd,
	versionCmd.Name(): versionCmd,
}

func setupCommonFlags() {
	for _, fs := range subcommands {
		fs.StringVar(&spath, "hsnap", "", "Use a different .hsnap file")
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
	trimCmd.BoolVar(&delete, "delete", false, "really deletes stuff")
	trimCmd.BoolVar(&quiet, "quiet", false, "do not list stuff")

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
		err = info(cm.Args()...)

	case diffCmd.Name():
		err = diff(cm.Args()...)

	case checkCmd.Name():
		err = check()

	case nodeCmd.Name():
		err = node(cm.Args()...)

	case listCmd.Name():
		if len(cm.Args()) == 0 {
			err = list("")
		} else {
			err = list(cm.Args()...)
		}

	case trimCmd.Name():
		if len(cm.Args()) == 0 {
			err = fmt.Errorf("wrong usage")
			break
		}
		err = trim(delete, cm.Args()...)

	case versionCmd.Name():
		fmt.Fprintln(output, version)

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
info      Basic information about current snapshot
check     Existence of files in current snapshot
trim      Remove local files that are present in provided snapshots
list
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

	start := time.Now()

	c := hashsnap.Snapshot(wd, f, spy)

	fmt.Fprintf(output, "Encoded %d files in %s\n", c, time.Since(start))

	return nil
}

func check() error {
	cur, err := readTree(spath)
	if err != nil {
		return err
	}
	missing := cur.Check(wd)
	if len(missing) == 0 {
		fmt.Fprint(output, "Snapshot is complete\n")
		return nil
	}
	for _, n := range missing {
		fmt.Fprintf(output, "%s\n", n.Path())
	}
	fmt.Fprintf(output, "\nMissing %d files\n", len(missing))
	return nil
}

func info(paths ...string) error {
	if paths == nil {
		paths = []string{spath}
	}
	for _, x := range paths {
		fmt.Fprintf(output, color.Blue+"%s\n"+color.Reset, x)
		err := infoSingle(x)
		if err != nil {
			return err
		}
	}
	return nil
}

func diff(paths ...string) error {
	if paths == nil {
		panic("give me some paths")
	}

	matches := make(hashsnap.HashGroup)

	var trees []*hashsnap.Tree

	for _, p := range paths {
		t, err := readTree(p)
		t.Name = p
		trees = append(trees, t)
		if err != nil {
			return err
		}
		matches.AddTree(t)
	}
	gr := matches.Groups()
	gr = gr.Filter(func(n hashsnap.Nodes) bool {
		t := n.Trees()
		return len(t) == 1 && t[0] == trees[0]
	})

	tots := len(matches)
	dels := matches.PruneSingleTreeGroups()

	fmt.Fprintf(output, "%d file groups\n", tots)
	for t, v := range dels {
		fmt.Fprintf(output, "%s had %d specific files not found elsewhere\n", t.Name, v)
	}

	for _, gg := range gr {
		for _, x := range gg {
			fmt.Fprintf(output, "\t%s\t%s\n", hashsnap.ByteSize(x.Size), trees[0].RelPath(x))
		}
	}

	return nil
}

// info opens an hsnap, read its info header and counts how many nodes it has
// it does not check for sanity (like child has a valid parent and so on)
func infoSingle(path string) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
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

	err = hashsnap.DecodeNodes(dec, func(n *hashsnap.Node) error {
		if !n.Mode.IsDir() {
			size = size + n.Size
			count++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(output, "Totalling %s and %d files\n", hashsnap.ByteSize(size), count)
	return nil
}

func list(paths ...string) error {
	cur, err := readTree(spath)
	if err != nil {
		return err
	}
	path := filepath.Join(paths...)
	var at *hashsnap.Node
	if path == "" {
		at = cur.Root()
	} else {
		at = cur.Search(path)
	}
	if at == nil {
		fmt.Fprintf(output, "Not found\n")
		return nil
	}

	w := new(tabwriter.Writer)

	w.Init(output, 5, 4, 1, '\t', tabwriter.AlignRight)
	children := cur.ChildrenOf(at)

	// Sort per size, then per name
	// Folders on top

	for _, x := range children {
		d := " "
		if x.Mode.IsDir() {
			d = "d"
		}
		fmt.Fprintf(w, "%s%d(%d)\t%s\t%s\n", d, x.ID, x.ParentID, hashsnap.ByteSize(x.Size), x.Name)
	}

	w.Flush()

	return nil
}

func trim(delete bool, withs ...string) error {
	cur, err := readTree(spath)
	if err != nil {
		return err
	}
	cur.Name = "a"
	fmt.Fprintf(output, color.Red+"%s %s (%s)\n"+color.Reset, cur.Name, cur.Info, spath)

	cur.Info.RootPath = wd

	var trees []*hashsnap.Tree
	for k, w := range withs {
		x, err := readTree(w)
		if err != nil {
			return err
		}
		trees = append(trees, x)
		x.Name = string("bcdefghijkl"[k])
		fmt.Fprintf(output, color.Green+"%s %s (%s)\n"+color.Reset, x.Name, x.Info, w)
	}

	matches := cur.Trim(trees...)
	tots := len(matches)
	dels := matches.PruneSingleTreeGroups()
	fmt.Fprintf(output, "%d file groups\n", tots)
	for t, v := range dels {
		fmt.Fprintf(output, "%s had %d specific files not found elsewhere\n", t.Name, v)
	}

	var count, errc int
	var groups int
	var waste int64

	if delete {
		for _, ma := range matches {
			in, _ := hashsnap.SplitNodes(cur, ma)

			groups++

			for _, n := range in {
				p := n.Tree().AbsPath(n)
				if err := os.Remove(p); err != nil {
					fmt.Fprintf(output, "Cannot remove %s: %s\n", p, err)
					errc++
				} else {
					fmt.Fprintf(output, "Removed %s\n", p)
					count++
					waste = waste + int64(hashsnap.Nodes(in).ByteSize())
				}
			}
		}
		fmt.Fprintf(output, "%d duplicated groups, removed %d files totalling %s wasted space, %d errors\n", groups, count, hashsnap.ByteSize(waste), errc)
	} else {
		for _, ma := range matches {
			var str strings.Builder
			in, out := hashsnap.SplitNodes(cur, ma)

			count = count + len(in)
			bs := hashsnap.Nodes(in).ByteSize()
			waste = waste + int64(bs)
			groups++
			if quiet {
				continue
			}
			str.WriteString(fmt.Sprintf("%d files (wasting %s)\n", len(in), bs))

			for _, n := range in {
				str.WriteString(fmt.Sprintf(color.Red+"\t-%s %s\n"+color.Reset, n.Tree().Name, n.Path()))
			}
			for _, n := range out {
				str.WriteString(fmt.Sprintf(color.Green+"\t+%s %s\n"+color.Reset, n.Tree().Name, n.Path()))
			}

			fmt.Fprintln(output, str.String())
		}
		fmt.Fprintf(output, "%d duplicated groups, totalling %s wasted space in %d files\n", groups, hashsnap.ByteSize(waste), count)
	}

	if errc != 0 {
		return errors.New("Delete got some errors while processing")
	}

	return nil
}

func node(ids ...string) error {
	cur, err := readTree(spath)
	if err != nil {
		return err
	}

	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			return err
		}
		n := cur.Node(i)
		if n == nil {
			fmt.Fprintf(output, "%s not found\n", id)
		} else {
			fmt.Fprintf(output, "%s %s\n", n, cur.RelPath(n))
		}
	}
	return nil
}
