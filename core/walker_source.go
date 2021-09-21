package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

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
// It generates a stream of *Nodes to be used. Once a Node has been sent, it is
// not accessed anymore making it safe for user from another thread.
// Once the walker has explored all files, it closes the emitting channel.
// Each node receives a unique increment id, starting at 1 (0 being null)
func WalkFS(ctx context.Context, path string, skip Skipper) (<-chan *Node, error) {
	if skip != nil {
		skip = DefaultSkipper
	}

	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("Path should be absolute: %s", path)
	}

	var q []*Node
	var id uint64 = 1
	out := make(chan *Node)

	go func() {
		defer close(out)

		// Appending the root to the processing queue, in order to bootstrap
		// the BFS routine below.
		root := MakeRootNode(path)
		root.ID = id
		q = append(q, root)

		// Actual BFS
		for len(q) > 0 {
			// Shift first node
			node := q[0]
			q = q[1:]

			// Walk deeper
			if node.Mode.IsDir() {
				path, err := node.Path()
				if err != nil {
					log.Fatalf("A node has no path in create: %s", err)
				}
				names, err := readDirNames(path)
				if err != nil {
					log.Printf("Listing directory %s failed: %s", path, err)
					continue
				}
				for _, name := range names {
					child, err := MakeChildNode(node, name)
					if err != nil {
						log.Printf("Node creation failed: %s", err)
					}

					if skip != nil && skip(child) {
						continue
					}

					// This file is legit, let's assign it an ID
					id++
					child.ID = id

					q = append(q, child)
				}
			}

			// Send current node downstream, we won't touch it anymore
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
