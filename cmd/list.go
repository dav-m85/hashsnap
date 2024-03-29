package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/dav-m85/hashsnap/core"
)

func List(local core.Hsnap) {
	nodes, _ := local.ChannelRead(context.Background())

	var count uint64 = 0
	for n := range nodes {
		if n.RootPath != "" {
			fmt.Printf("Snapshot captured in %s\n", n.RootPath)
		}
		fmt.Printf("%s\n", n)
		count++
	}
	log.Printf("Listed %d files\n", count)
}
