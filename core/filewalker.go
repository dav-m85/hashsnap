package core

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

var lstat = os.Lstat

// Skipper indicate a Node should be skipped by returning true
type Skipper func(fs.FileInfo) bool

// DefaultSkipper
var DefaultSkipper = func(fs.FileInfo) bool {
	return false
}

var NoDirs = func(n fs.FileInfo) bool {
	return (!n.IsDir() && (!n.Mode().IsRegular() || n.Size() == 0 || n.Name() == ".hsnap" /*state.STATE_NAME*/))
}

var NoFiles = func(n fs.FileInfo) bool {
	return !n.IsDir()
}

// WalkFS walks a filetree in a breadth first manner
// It generates a stream of *Nodes to be used.
// Once the walker has explored all files, it closes the emitting channel.
// Each node receives a unique increment id, starting at 1.
func WalkFS(ctx context.Context, skip Skipper, wd string, skipNodes bool, q ...*Node) (<-chan *Node, error) {
	out := make(chan *Node)

	skipUntil := len(q)

	if skip == nil {
		skip = DefaultSkipper
	}

	if !filepath.IsAbs(wd) {
		return nil, fmt.Errorf("wd should be absolute: %s", wd)
	}

	go func() {
		defer close(out)

		// Actual BFS
		for len(q) > 0 {
			// Shift first node
			node := q[0]
			q = q[1:]

			// Walk deeper in directory
			if node.Mode.IsDir() {
				dpath := filepath.Join(wd, node.Path())
				names, err := readDirNames(dpath)
				if err != nil {
					log.Printf("Listing directory %s failed: %s", dpath, err)
					continue
				}
				for _, name := range names {
					cpath := filepath.Join(dpath, name)
					info, err := lstat(cpath)
					if err != nil {
						log.Printf("Node creation failed: %s", err)
					}
					if skip != nil && skip(info) {
						continue
					}
					child := NewNode(info)

					node.Attach(child)

					q = append(q, child)
				}
			}

			if skipUntil <= 0 || !skipNodes {
				out <- node
			} else {
				skipUntil--
			}
		}
	}()

	return out, nil
}

// readDirNames reads the directory named by dirname and returns
// a list of directory entries.
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return names, nil
}
