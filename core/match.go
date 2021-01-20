package core

import (
	"crypto/sha1"
	"fmt"
)

type Group struct {
	Files []*File
	Tsize int64
}

func (sn *Snapshot) Dedup() {
	// check for matching hash
	matches := make(map[[sha1.Size]byte]*Group)

	for _, f := range sn.Files {
		match, ok := matches[f.Hash]
		if ok {
			if match.Tsize != f.Size {
				fmt.Printf("Collision, same hash but different size")
			}
			// matching group found; add this file to existing group
			match.Files = append(match.Files, f)
		} else {
			// create new group in map
			matches[f.Hash] = &Group{[]*File{f}, f.Size}
		}
	}

	for _, group := range matches {
		if len(group.Files) > 1 {
			fmt.Println("Duplicates\n", group.Files)
		}
	}
}
