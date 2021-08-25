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
	ChannelRead(context.Context) (<-chan *Node, error)
	ChannelWrite(<-chan *Node) error
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

func (h *HsnapMem) ChannelRead(context.Context) (<-chan *Node, error) {
	out := make(chan *Node)
	go func() {
		defer close(out)
		for _, n := range h.Nodes {
			out <- n
		}
	}()
	return out, nil
}

func (h *HsnapMem) ChannelWrite(<-chan *Node) error {
	return nil
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
	out, _ := snap.ChannelRead(ctx)
	for n := range out {
		all = append(all, n)
	}
	return
}

// TODO ChannelRead and ChannelWrite could be detyped

// ChannelRead decodes a hsnap file into a stream of *Node. It closes the receiving
// channel when file has been completely read. Call it within a goroutine.
func (h *HsnapFile) ChannelRead(ctx context.Context) (<-chan *Node, error) {
	f, err := os.Open(h.path)
	if err != nil {
		panic(err)
	}

	out := make(chan *Node)
	go func() {
		defer f.Close()
		defer close(out)

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

			select {
			case <-ctx.Done():
				return
			case out <- n:
			}

		}

	}()
	return out, nil
}

// ChannelWrite encodes a hsnap file given a stream of *Node. Signal end of processing
// by sending on the done channel the number of written Node.
// done chan uint64
func (h *HsnapFile) ChannelWrite(in <-chan *Node) error {
	f, err := os.Create(h.path)
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)

	defer f.Close()
	defer w.Flush()

	enc := gob.NewEncoder(w)

	for f := range in {
		enc.Encode(f)
	}

	return nil
}
