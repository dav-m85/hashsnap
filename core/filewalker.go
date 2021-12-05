package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var lstat = os.Lstat

// Skipper indicate a Node should be skipped by returning true
type Skipper func(*Node) bool

// DefaultSkipper ignores symlinks and zero-size files
var DefaultSkipper = func(n *Node) bool {
	return !n.Mode.IsDir() && (!n.Mode.IsRegular() || n.Size == 0)
}

// MakeNameSkipper extends DefaultSkipper to ignore some names
func MakeNameSkipper(names []string) Skipper {
	return func(n *Node) bool {
		for _, x := range names {
			if x == n.Name {
				return true
			}
		}
		return DefaultSkipper(n)
	}
}

// WalkFS walks a filetree in a breadth first manner
// It generates a stream of *Nodes to be used.
// Once the walker has explored all files, it closes the emitting channel.
// Each node receives a unique increment id, starting at 1.
func WalkFS(ctx context.Context, path string, skip Skipper) (<-chan *Node, error) {
	if skip == nil {
		skip = DefaultSkipper
	}

	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("Path should be absolute: %s", path)
	}

	out := make(chan *Node)

	go func() {
		var q []*Node

		defer close(out)

		// Appending the root to the processing queue, in order to bootstrap
		// the BFS routine below.
		info, err := lstat(path)
		if err != nil {
			panic("Node creation failed: " + err.Error())
		}
		root := NewNode(info)
		q = append(q, root)

		// Actual BFS
		for len(q) > 0 {
			// Shift first node
			node := q[0]
			q = q[1:]

			// Walk deeper
			if node.Mode.IsDir() {
				names, err := readDirNames(node.Path())
				if err != nil {
					log.Printf("Listing directory %s failed: %s", path, err)
					continue
				}
				for _, name := range names {
					cpath := filepath.Join(path, name)
					info, err := lstat(cpath)
					if err != nil {
						log.Printf("Node creation failed: %s", err)
					}
					child := NewNode(info)
					child.path = cpath

					if skip != nil && skip(child) {
						continue
					}

					node.Attach(child)

					q = append(q, child)
				}
			}

			out <- node
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
