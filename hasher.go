package hashsnap

import (
	"context"
	"crypto/sha1"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
)

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
	fd, err := os.Open(n.Path)
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
