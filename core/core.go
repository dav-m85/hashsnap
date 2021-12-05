package core

import (
	"context"
	"time"
)

// Info header for a .hsnap file
type Info struct {
	RootPath  string
	CreatedAt time.Time
	Version   int
}

// TODO Rename to Noder
type Noder interface {
	Info() Info
	Read(context.Context) (<-chan *Node, error)
	Write(<-chan *Node) error
}

func Read(snap Noder) (all []*Node) {
	ctx := context.Background()
	out, _ := snap.Read(ctx)
	for n := range out {
		all = append(all, n)
	}
	return
}
