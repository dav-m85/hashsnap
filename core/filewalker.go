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

// WalkFS walks a filetree in a breadth first manner
// It generates a stream of *Nodes to be used.
// Once the walker has explored all files, it closes the emitting channel.
// Each node receives a unique increment id, starting at 1.
func WalkFS(ctx context.Context, skip Skipper, root string) (<-chan NodeP, error) {
	out := make(chan NodeP)

	if skip == nil {
		skip = DefaultSkipper
	}

	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("wd should be absolute: %s", root)
	}

	go func() {
		defer close(out)

		info, err := lstat(root)
		if err != nil {
			log.Printf("Root node creation failed on %s: %s", root, err)
			return
		}

		rootNode := &Node{
			ID:   Reset(),
			Mode: info.Mode(),
			Name: info.Name(),
			Size: info.Size(),
		}

		q := []NodeP{{
			rootNode, root,
		}}
		var np NodeP

		// Actual BFS
		for len(q) > 0 {
			// Shift first node
			np, q = q[0], q[1:]

			// Walk deeper in directory
			if np.Node.Mode.IsDir() {
				names, err := readDirNames(np.Path)
				if err != nil {
					log.Printf("Listing directory %s failed: %s", np.Path, err)
					continue
				}
				for _, name := range names {
					cpath := filepath.Join(np.Path, name)
					info, err := lstat(cpath)
					if err != nil {
						log.Printf("Node creation failed: %s", err)
					}
					if skip != nil && skip(info) {
						continue
					}
					child := &Node{
						ID:       Allocate(),
						ParentID: np.Node.ID,
						Mode:     info.Mode(),
						Name:     info.Name(),
						Size:     info.Size(),
					}

					q = append(q, NodeP{child, cpath})
				}
			}

			out <- np
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
