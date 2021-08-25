package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Exclusions []string

// func makeExclusions(in []string) exclusions {
// 	var e exclusions
// 	copy(e, in)
// 	sort.Strings(e)
// 	return e
// }

func (e Exclusions) Has(name string) bool {
	// i := sort.SearchStrings(e, name)
	// return i < len(e) && e[i] == name
	for _, x := range e {
		if x == name {
			return true
		}
	}
	return false
}

// WalkFileTree walks a filetree in a breadth first manner
// It generates a stream of *Nodes to be used. Once a Node has been sent, it is
// not accessed anymore making it safe for user from another thread.
// Once the walker has explored all files, it closes the emitting channel.
// Each node receives a unique increment id, starting at 1 (0 being null)
// func () Sourcer {
func WalkFS(ctx context.Context, path string, excludes Exclusions) (<-chan *Node, error) {
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
					if excludes.Has(name) {
						continue
					}

					child, err := MakeNode(node, name)
					if err != nil {
						log.Printf("Node creation failed: %s", err)
					}
					// Ignore symlinks and zero-size files
					if !child.Mode.IsDir() && (!child.Mode.IsRegular() || child.Size == 0) {
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
