package core

import (
	"context"
	"crypto/sha1"
	"io"
	"log"
	"os"
	"runtime"
	"sync"

	bar "github.com/schollz/progressbar/v3"
)

func Hasher(ctx context.Context, pbar *bar.ProgressBar, in <-chan *Node) <-chan *Node {
	out := make(chan *Node)
	go func() {
		defer close(out)

		wg := &sync.WaitGroup{}
		for w := 0; w < runtime.NumCPU(); w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for node := range in {
					if !node.Mode.IsDir() {
						err := computeHash(node, pbar)
						if err != nil {
							log.Printf("Cannot hash %s: %s", node, err)
							continue
						}
					}
					select {
					case out <- node:
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
func computeHash(n *Node, pbar *bar.ProgressBar) error {
	fd, err := os.Open(n.Path())
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

	copy(n.Hash[:], h.Sum(nil)) // [sha1.Size]byte()

	return nil
}
