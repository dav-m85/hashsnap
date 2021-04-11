package core

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"os"
)

// Hsnap is a file holding a filetree snapshot
type Hsnap struct {
	path string
}

func MakeHsnap(path string) *Hsnap {
	return &Hsnap{path}
}

// ChannelRead decodes a hsnap file into a stream of *Node. It closes the receiving
// channel when file has been completely read. Call it within a goroutine.
func (h *Hsnap) ChannelRead(out chan<- *Node) {
	f, err := os.Open(h.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

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

		out <- n
	}
	close(out)
}

// ChannelWrite encodes a hsnap file given a stream of *Node. Signal end of processing
// by sending on the done channel the number of written Node.
func (h *Hsnap) ChannelWrite(in <-chan *Node, done chan uint64) {
	f, err := os.Create(h.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	enc := gob.NewEncoder(w)
	var count uint64 = 0

	for f := range in {
		enc.Encode(f)
		count++
	}
	w.Flush()

	done <- count
}
