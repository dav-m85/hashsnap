package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dav-m85/hashsnap/file"
	bar "github.com/schollz/progressbar/v3"
)

// 2min pour 6Go SSD avec mon quad core

var mutex = &sync.Mutex{}

type Group struct {
	files []*File
	fsize int64
}

type File struct {
	path string
	size int64
	hash [sha1.Size]byte // hash.Hash // sha1.New()
}

func (f *File) ComputeHash(pbar *bar.ProgressBar) error {
	fd, err := os.Open(f.path)
	if err != nil {
		return err
	}
	h := sha1.New()
	defer fd.Close()

	if _, err = io.Copy(io.MultiWriter(h, pbar), fd); err != nil {
		return err
	}

	copy(f.hash[:], h.Sum(nil)) // [sha1.Size]byte()

	return nil
}

type Snapshot struct {
	files []*File
	tsize int64
}

func (sn Snapshot) String() string {
	return fmt.Sprintf("Snapshot holds %d files, totalling %d bytes", len(sn.files), sn.tsize)
}

// Walker collects files found by file.Walk into the Snapshot
func (sn *Snapshot) Walker() file.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		// fmt.Printf("%s\n", path)
		if err != nil {
			log.Println(err)
			return nil
		}

		// hidden file or folder; skip file, don't recurse directory
		if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// don't store directory entry
		if info.IsDir() {
			return nil
		}

		// ignore symlinks
		if !info.Mode().IsRegular() {
			return nil
		}

		// excludes zero-size files
		if info.Size() == 0 {
			return nil
		}

		sn.files = append(sn.files, &File{path: path, size: info.Size()})
		sn.tsize = sn.tsize + info.Size()
		return nil
	}
}

func main() {
	start := time.Now()

	// var help = flag.Bool("h", false, "Display this message")
	flag.Parse()
	// if *help {
	// 	fmt.Println("\nduplicates is a command line tool to find duplicate files in a folder\n")
	// 	fmt.Println("usage: duplicates [options...] path\n")
	// 	flag.PrintDefaults()
	// 	os.Exit(0)
	// }
	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "You have to specify at least a directory to explore ...\n")
		os.Exit(-1)
	}
	root := flag.Arg(0)

	var wg sync.WaitGroup

	hasher := func(fileCh chan *File, pbar *bar.ProgressBar) {
		defer wg.Done()
		for f := range fileCh {
			f.ComputeHash(pbar)
		}
	}

	// Très rapide !
	snap := Snapshot{}
	file.Walk(root, snap.Walker())
	fmt.Printf("%s\n", snap)

	pbar := bar.DefaultBytes(
		snap.tsize,
		"Hashing",
	)

	fileChan := make(chan *File)

	for w := 0; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go hasher(fileChan, pbar)
	}

	go func() {
		for _, f := range snap.files {
			fileChan <- f
		}
		close(fileChan)
	}()

	wg.Wait()

	t := time.Now()
	elapsed := t.Sub(start)

	fmt.Printf("Hashing finished after %s", elapsed)

	// We got a file here !

	// files is full here
	// fmt.Printf("%v\n", files)
	// fmt.Printf("%v\n", matches)

	// check for matching hash
	matches := make(map[[sha1.Size]byte]*Group)

	for _, f := range snap.files {
		match, ok := matches[f.hash]
		if ok {
			// matching group found; add this file to existing group
			match.files = append(match.files, f)
		} else {
			// create new group in map
			matches[f.hash] = &Group{[]*File{f}, f.size}
		}
	}

	// for _, group := range matches {
	// 	if len(group.files) > 1 {
	// 		fmt.Println("Duplicates\n", group.files)
	// 	}
	// }
}

func (f File) String() string {
	return fmt.Sprintf("%s(%d)[sha1:%x]", f.path, f.size, f.hash)
}
