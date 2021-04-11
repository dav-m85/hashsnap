package core

import (
	"fmt"
	"log"
)

func List(target string) {
	local := MakeHsnap(target)

	nodes := make(chan *Node)
	go local.ChannelRead(nodes)

	var count uint64 = 0
	for n := range nodes {
		if n.RootPath != "" {
			fmt.Printf("Snapshot captured in %s\n", n.RootPath)
		}
		fmt.Printf("%s\n", n)
		count++
	}
	log.Printf("Listed snapshot %s with %d files\n", target, count)
}
