package cmd

import (
	"context"

	"github.com/dav-m85/hashsnap/core"
)

func Convert(in core.Hsnap, out core.Hsnap) {
	nodes, err := in.ChannelRead(context.Background())
	if err != nil {
		panic(err)
	}

	err = out.ChannelWrite(nodes)
	if err != nil {
		panic(err)
	}
}
