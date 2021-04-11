package cmd

import (
	"fmt"

	"github.com/dav-m85/hashsnap/core"
)

func DedupWith(local core.Hsnap, withs []core.Hsnap) {
	matches := make(core.HashGroup)
	for _, w := range withs {
		matches.Load(w)
	}

	var size uint64
	var count uint64

	for _, n := range core.Read(local) {
		if g, ok := matches.Contains(n); ok {
			fmt.Printf("DUP %s:\n%s\n", n, g)
			size = size + n.Size
			count++
		}
	}

	fmt.Printf("Duplication totalling %s and %d files\n", core.ByteSize(size), count)
}

func Dedup(local core.Hsnap) {
	matches := make(core.HashGroup)
	matches.Load(local)
	matches.Dedup()
}