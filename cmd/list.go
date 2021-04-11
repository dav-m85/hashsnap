package cmd

import (
	"fmt"
	"log"

	"github.com/dav-m85/hashsnap/core"
)

func List(local core.Hsnap) {
	nodes := make(chan *core.Node)
	go local.ChannelRead(nodes)

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
