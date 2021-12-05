package cmd

import (
	"github.com/dav-m85/hashsnap/core"
)

func Dedup(local core.Noder) {
	matches := make(core.HashGroup)
	matches.Load(local)
	matches.Dedup()
}
