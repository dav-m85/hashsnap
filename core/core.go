package core

import (
	"context"
)

// TODO Rename to Noder
type Hsnap interface {
	ChannelRead(context.Context) (<-chan *Node, error)
	ChannelWrite(<-chan *Node) error
}

func Read(snap Hsnap) (all []*Node) {
	ctx := context.Background()
	out, _ := snap.ChannelRead(ctx)
	for n := range out {
		all = append(all, n)
	}
	return
}
