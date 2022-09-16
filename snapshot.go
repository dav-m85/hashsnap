package hashsnap

import (
	"context"
	"crypto/sha1"
	"encoding/gob"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

const STATE_NAME = ".hsnap"
const VERSION = 1

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
func WalkFS(ctx context.Context, skip Skipper, root string) <-chan NodeP {
	lstat := FS.(fs.StatFS).Stat

	out := make(chan NodeP)

	if skip == nil {
		skip = DefaultSkipper
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
					cpath := filepath.Join(np.Path, name.Name())
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

	return out
}

// readDirNames reads the directory named by dirname and returns
// a list of directory entries.
func readDirNames(dirname string) ([]fs.DirEntry, error) {
	readdir := FS.(fs.ReadDirFS).ReadDir

	names, err := readdir(dirname)
	if err != nil {
		return nil, err
	}
	return names, nil
}

// Hasher... spy allows to follow hashing speed by having every hashed byte copied to it
func Hasher(ctx context.Context, wd string, spy io.Writer, in <-chan NodeP) <-chan *Node {
	out := make(chan *Node)
	go func() {
		defer close(out)

		wg := &sync.WaitGroup{}
		for w := 0; w < runtime.NumCPU(); w++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for np := range in {
					if !np.Node.Mode.IsDir() {
						err := computeHash(np, spy)
						if err != nil {
							log.Printf("Cannot hash %s: %s", np.Path, err)
							continue
						}
					}
					select {
					case out <- np.Node:
					case <-ctx.Done():
						return
					}
				}
			}()
		}
		wg.Wait()
	}()
	return out
}

// computeHash reads the file and computes the sha1 of it
func computeHash(n NodeP, spy io.Writer) error {
	fd, err := FS.Open(n.Path)
	if err != nil {
		return err
	}
	h := sha1.New()
	defer fd.Close()

	if _, err = io.Copy(io.MultiWriter(h, spy), fd); err != nil {
		return err
	}

	copy(n.Node.Hash[:], h.Sum(nil)) // [sha1.Size]byte()

	return nil
}

func Snapshot(root string, out, spy io.Writer) (c int) {
	enc := gob.NewEncoder(out)

	hs, err := os.Hostname()
	if err != nil {
		hs = "localhost"
		log.Printf("Cannot get hostname: %s", err)
	}
	// Write info node
	err = enc.Encode(Info{
		Version:   VERSION,
		RootPath:  root,
		CreatedAt: time.Now(),
		Nonce:     uuid.New(),
		Hostname:  hs,
	})
	if err != nil {
		panic(err)
	}

	// Context for the pipelines, cancel the workers
	ctx, cleanup := context.WithCancel(context.Background())
	defer cleanup()

	skipper := func(n fs.FileInfo) bool {
		return !n.Mode().IsDir() && (!n.Mode().IsRegular() || n.Size() == 0 || n.Name() == STATE_NAME)
	}

	// Source by exploring all nodes and hash them
	for x := range Hasher(ctx, root, spy, WalkFS(ctx, skipper, root)) {
		c++
		if err := enc.Encode(x); err != nil {
			panic(err)
		}
	}

	return
}
