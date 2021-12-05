package cmd

import (
	"context"
	"fmt"

	"github.com/dav-m85/hashsnap/core"
)

func Info(local core.Noder) {
	var size uint64
	var count uint64

	stream, err := local.Read(context.Background())
	if err != nil {
		panic(err)
	}
	for n := range stream {
		// if n.RootPath != "" {
		// 	fmt.Printf("Snapshot captured in %s\n", n.RootPath)
		// }
		if n.Mode.IsDir() {
			continue
		}
		size = size + n.Size
		count++
	}
	fmt.Printf("Totalling %s and %d files\n", core.ByteSize(size), count)
}
