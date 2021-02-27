package ncore

import (
	"bufio"
	"crypto/sha1"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	bar "github.com/schollz/progressbar/v3"
)

var (
	ErrNoParent = errors.New("parent is required to get correct path")
)

// Node a hsnap file contains a stream of nodes discovered by the walker. Each node
// is either file or dir.
type Node struct {
	ID       uint64
	ParentID uint64

	// Directory specific. Complete absolute path for this directory.
	Path string

	// File specific
	Name string
	Size uint64
	Hash [sha1.Size]byte // hash.Hash // sha1.New()

	parent *Node // Unexported yup
}

func (n Node) String() string {
	return fmt.Sprintf("%d (in %d), %s %s", n.ID, n.ParentID, n.Path, n.Name)
}

func (f *Node) LStat() (fs.FileInfo, error) {
	var path string
	if f.Name != "" {
		if f.parent == nil {
			return nil, ErrNoParent
		}
		path = filepath.Join(f.parent.Path, f.Name)
	} else {
		path = f.parent.Path
	}

	return os.Lstat(path)
}

// ComputeHash reads the file and computes the sha1 of it
func (f *Node) ComputeHash(pbar *bar.ProgressBar) error {
	if f.parent == nil {
		return ErrNoParent
	}
	filename := filepath.Join(f.parent.Path, f.Name)
	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	h := sha1.New()
	defer fd.Close()

	var writeTo io.Writer = h
	if pbar != nil {
		writeTo = io.MultiWriter(h, pbar)
	}
	if _, err = io.Copy(writeTo, fd); err != nil {
		return err
	}

	copy(f.Hash[:], h.Sum(nil)) // [sha1.Size]byte()

	return nil
}

func Create(target, outfile string) error {

	hasher := func(in <-chan *Node, out chan<- *Node, wg *sync.WaitGroup, pbar *bar.ProgressBar) {
		defer wg.Done()
		for f := range in {
			if f.Path != "" {
				out <- f // Dir
			} else {
				err := f.ComputeHash(pbar)
				if err != nil {
					fmt.Printf("Cannot hash %s: %s", f, err)
				} else {
					out <- f
				}
			}
		}
	}

	writer := func(nodeChan <-chan *Node, path string) {
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		enc := gob.NewEncoder(w)

		for f := range nodeChan {
			// fmt.Printf("%s\n", f)
			enc.Encode(f)
		}
		w.Flush()
	}

	var pbar *bar.ProgressBar = bar.DefaultBytes(
		-1,
		"Hashing",
	)

	nodeChan := make(chan *Node)
	hashedNodeChan := make(chan *Node)

	go walker(nodeChan, target)

	wg := &sync.WaitGroup{}
	for w := 0; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go hasher(nodeChan, hashedNodeChan, wg, pbar)
	}

	go writer(hashedNodeChan, "yolo.hsnap")

	wg.Wait()
	close(hashedNodeChan)

	return nil
}
