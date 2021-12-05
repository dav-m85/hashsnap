package core

import (
	"context"
	"fmt"
)

var _ Hsnap = &HsnapMem{}

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

func (h *HsnapMem) ChannelWrite(in <-chan *Node) error {
	for n := range in {
		h.Nodes = append(h.Nodes, n)
	}
	return nil
}

func (h *HsnapMem) Info() Info {
	return Info{}
}
