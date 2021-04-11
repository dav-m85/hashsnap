package core

import (
	"bufio"
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
	ChannelRead(chan<- *Node)
	ChannelWrite(<-chan *Node, chan uint64)
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

func (h *HsnapMem) ChannelRead(out chan<- *Node) {
	for _, n := range h.Nodes {
		out <- n
	}
	close(out)
}

func (h *HsnapMem) ChannelWrite(<-chan *Node, chan uint64) {
	panic("Not implemented")
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
	nodes := make(chan *Node)
	go snap.ChannelRead(nodes)
	for n := range nodes {
		all = append(all, n)
	}
	return
}

// ChannelRead decodes a hsnap file into a stream of *Node. It closes the receiving
// channel when file has been completely read. Call it within a goroutine.
func (h *HsnapFile) ChannelRead(out chan<- *Node) {
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
func (h *HsnapFile) ChannelWrite(in <-chan *Node, done chan uint64) {
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
