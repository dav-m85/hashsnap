package core

import (
	"bufio"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var _ Hsnap = &HsnapFile{}
var _ Hsnap = &HsnapMem{}

type Hsnap interface {
	ChannelRead() Sourcer
	ChannelWrite() Sinker
}

// HsnapMem in memory filetree snapshot, useful for testing
type HsnapMem struct {
	Nodes []*Node
}

func (hs HsnapMem) String() string {
	var s string
	for _, n := range hs.Nodes {
		s = fmt.Sprintf("%s\t%s\n", s, n)
	}
	return s
}

func (h *HsnapMem) ChannelRead() Sourcer {
	return func(_ context.Context) (<-chan *Node, <-chan error, error) {
		out := make(chan *Node)
		go func() {
			defer close(out)
			for _, n := range h.Nodes {
				out <- n
			}
		}()
		return out, nil, nil
	}
}

func (h *HsnapMem) ChannelWrite() Sinker {
	return func(_ context.Context, _ <-chan *Node) (<-chan error, error) {
		panic("Not implemented")
	}
}

// HsnapFile is a file holding a filetree snapshot
type HsnapFile struct {
	path string
}

func MakeHsnapFile(path string) *HsnapFile {
	if !strings.HasSuffix(path, ".hsnap") {
		log.Fatal("Snapshot file name should end with .hsnap")
	}
	return &HsnapFile{path}
}

func Read(snap Hsnap) (all []*Node) {
	ctx := context.Background()
	out, _, _ := snap.ChannelRead()(ctx)
	for n := range out {
		all = append(all, n)
	}
	return
}

// ChannelRead decodes a hsnap file into a stream of *Node. It closes the receiving
// channel when file has been completely read. Call it within a goroutine.
func (h *HsnapFile) ChannelRead() Sourcer {
	return func(_ context.Context) (<-chan *Node, <-chan error, error) {

		f, err := os.Open(h.path)
		if err != nil {
			panic(err)
		}

		out := make(chan *Node)
		errc := make(chan error, 1)
		go func() {
			defer f.Close()
			defer close(out)
			defer close(errc)

			nodes := make(map[uint64]*Node)

			w := bufio.NewReader(f)
			enc := gob.NewDecoder(w)
			for {
				var n *Node = &Node{}
				err := enc.Decode(n)
				if err != nil {
					if err != io.EOF {
						log.Printf("Decoder encountered an issue: %s\n", err)
					}
					break
				}

				// Lets fill parent
				nodes[n.ID] = n
				if n.ParentID != 0 {
					parent, ok := nodes[n.ParentID]
					if !ok {
						log.Printf("Cannot solve parent")
						continue
					}
					n.parent = parent
				}

				// TODO context
				out <- n
			}

		}()
		return out, errc, nil
	}
}

// ChannelWrite encodes a hsnap file given a stream of *Node. Signal end of processing
// by sending on the done channel the number of written Node.
// done chan uint64
func (h *HsnapFile) ChannelWrite() Sinker {
	return func(ctx context.Context, in <-chan *Node) (<-chan error, error) {

		f, err := os.Create(h.path)
		if err != nil {
			panic(err)
		}

		w := bufio.NewWriter(f)

		errc := make(chan error, 1)

		go func() {
			defer f.Close()
			defer w.Flush()
			defer close(errc)

			enc := gob.NewEncoder(w)
			var count uint64 = 0

			for f := range in {
				enc.Encode(f)
				count++
			}
		}()
		return errc, nil
	}
}
