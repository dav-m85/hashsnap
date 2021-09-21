package core

import (
	"context"
	"time"
)

type Info struct {
	RootPath  string
	CreatedAt time.Time
	Version   int
}

// TODO Rename to Noder
type Hsnap interface {
	Info() Info
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
