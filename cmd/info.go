package cmd

import (
	"fmt"

	"github.com/dav-m85/hashsnap/core"
)

func Info(local core.Hsnap) {
	var size uint64
	var count uint64

	for _, n := range core.Read(local) {
		if n.RootPath != "" {
			fmt.Printf("Snapshot captured in %s\n", n.RootPath)
		}
		if n.Mode.IsDir() {
			continue
		}
		size = size + n.Size
		count++
	}
	fmt.Printf("Totalling %s and %d files\n", core.ByteSize(size), count)
}
