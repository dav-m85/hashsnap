package cmd

import (
	"context"

	"github.com/dav-m85/hashsnap/core"
)

func Convert(in core.Noder, out core.Noder) {
	inInfo := in.Info()
	if inInfo.Version != 0 {
		panic("Only Version=0 files can be converted")
	}

	nodes, err := in.Read(context.Background())
	if err != nil {
		panic(err)
	}

	// TODO convert should receives paths to infer files...

	err = out.Write(nodes)
	if err != nil {
		panic(err)
	}
}
