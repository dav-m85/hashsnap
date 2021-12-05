package core

import (
	"context"
	"time"
)

// Info header
type Info struct {
	RootPath  string
	CreatedAt time.Time
	Version   int
}

type Noder interface {
	Info() Info
	Read(context.Context) (<-chan *Node, error)
	Write(<-chan *Node) error
}

func ReadAll(n Noder) (all []*Node) {
	ctx := context.Background()
	out, _ := n.Read(ctx)
	for n := range out {
		all = append(all, n)
	}
	return
}
