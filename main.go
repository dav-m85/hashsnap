package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dav-m85/hashsnap/file"
)

// 2min pour 6Go SSD avec mon quad core

var mutex = &sync.Mutex{}

type Group struct {
	files []*File
	fsize int64
}

func main() {
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

	files := make([]*File, 0)
	matches := make(map[[sha1.Size]byte]*Group)
	fileChan := make(chan *File)
	ok := make(chan int)

	hasher := func(fileCh chan *File, done chan int) {
		for f := range fileCh {
			f.ComputeHash()
			mutex.Lock()
			files = append(files, f)

			// check for matching hash
			match, ok := matches[f.hash]
			if ok {
				// matching group found; add this file to existing group
				match.files = append(match.files, f)
			} else {
				// create new group in map
				matches[f.hash] = &Group{[]*File{f}, f.size}
			}

			mutex.Unlock()
		}
		done <- 1
	}

	/*
			if singleThread {
		    go worker(1, jobs, results, walkProgress)
		  } else {
		    for w := 1; w <= runtime.NumCPU(); w++ {
		      go worker(w, jobs, results, walkProgress)
		    }
		  }
	*/

	go hasher(fileChan, ok)

	file.Walk(root, visitor(fileChan))
	close(fileChan)

	<-ok

	// We got a file here !

	// files is full here
	// fmt.Printf("%v\n", files)
	// fmt.Printf("%v\n", matches)

	for _, group := range matches {
		if len(group.files) > 1 {
			fmt.Println("Duplicate for %s\n", group.files)
		}
	}
}

// gd file struct
type File struct {
	path string
	size int64
	hash [sha1.Size]byte // hash.Hash // sha1.New()
}

func (f File) String() string {
	return fmt.Sprintf("%s(%d)[sha1:%x]", f.path, f.size, f.hash)
}

func (f *File) ComputeHash() error {
	fd, err := os.Open(f.path)
	if err != nil {
		return err
	}
	h := sha1.New()
	defer fd.Close()
	if _, err := io.Copy(h, fd); err != nil {
		return err
	}

	copy(f.hash[:], h.Sum(nil)) // [sha1.Size]byte()

	return nil
}

// visitor collects files found by file.Walk()
func visitor(files chan *File) file.WalkFunc {
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

		files <- &File{path: path, size: info.Size()}
		return nil
	}
}
