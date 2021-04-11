package core

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	bar "github.com/schollz/progressbar/v3"
)

type exclusions []string

// func makeExclusions(in []string) exclusions {
// 	var e exclusions
// 	copy(e, in)
// 	sort.Strings(e)
// 	return e
// }

func (e exclusions) Has(name string) bool {
	// i := sort.SearchStrings(e, name)
	// return i < len(e) && e[i] == name
	for _, x := range e {
		if x == name {
			return true
		}
	}
	return false
}

func Create(target, outfile string, progress bool) error {
	excludes := exclusions{".git", ".DS_Store"}

	var pbar *bar.ProgressBar
	if progress {
		pbar = bar.DefaultBytes(
			-1,
			"Hashing",
		)
	}

	nodes := make(chan *Node)
	hashNodes := make(chan *Node)

	go walker(nodes, target, excludes)

	wg := &sync.WaitGroup{}
	for w := 0; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go hasher(nodes, hashNodes, wg, pbar)
	}

	done := make(chan uint64)
	local := MakeHsnap(outfile)
	go local.ChannelWrite(hashNodes, done)

	wg.Wait()
	close(hashNodes)
	var count uint64 = <-done
	log.Printf("Created snapshot %s with %d files\n", outfile, count)

	return nil
}

// walker walks a filetree in a breadth first manner
// It generates a stream of *Nodes to be used. Once a Node has been sent, it is
// not accessed anymore making it safe for user from another thread.
// Once the walker has explored all files, it closes the emitting channel.
// Each node receives a unique increment id, starting at 1 (0 being null)
func walker(out chan<- *Node, path string, excludes exclusions) {
	var q []*Node
	var id uint64 = 1

	if !filepath.IsAbs(path) {
		log.Fatalf("Path should be absolute: %s", path)
	}

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
			names, err := readDirNames(node.path)
			if err != nil {
				log.Printf("Listing directory %s failed: %s", node.path, err)
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

	close(out)
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

func hasher(in <-chan *Node, out chan<- *Node, wg *sync.WaitGroup, pbar *bar.ProgressBar) {
	defer wg.Done()
	for node := range in {
		if !node.Mode.IsDir() {
			err := node.ComputeHash(pbar)
			if err != nil {
				log.Printf("Cannot hash %s: %s", node, err)
				continue
			}
		}
		out <- node
	}
}
