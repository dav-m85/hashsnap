package core

import (
	"context"
	"log"
	"runtime"
	"sync"

	bar "github.com/schollz/progressbar/v3"
)

func Hasher(pbar *bar.ProgressBar) Transformer {
	return func(ctx context.Context, in <-chan *Node) (<-chan *Node, <-chan error, error) {
		out := make(chan *Node)
		errc := make(chan error, 1)
		go func() {
			defer close(out)
			defer close(errc)

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
		return out, errc, nil
	}
}
