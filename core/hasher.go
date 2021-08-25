package core

import (
	"context"
	"log"
	"runtime"
	"sync"

	bar "github.com/schollz/progressbar/v3"
)

func Hasher(ctx context.Context, pbar *bar.ProgressBar, in <-chan *Node) (<-chan *Node, error) {
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
						err := node.ComputeHash(pbar)
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
	return out, nil
}
