package core

import (
	"crypto/sha1"
	"fmt"
)

// Group aka duplicates
type Group struct {
	Files []*File
	Size  int64
}

// HashGroup helps comparing Hashes pretty quickly
type HashGroup map[[sha1.Size]byte]*Group

// Dedup reports duplicates within an HashGroup
func (h *HashGroup) Dedup() {
	for _, group := range *h {
		if len(group.Files) > 1 {
			fmt.Println("Duplicates\n", group.Files)
		}
	}
}

// DedupWith reports duplicates belonging both to a Snapshot and a given HashGroup
func (sn *Snapshot) DedupWith(hb *HashGroup) {
	for _, f := range sn.Files {
		match, ok := (*hb)[f.Hash]
		if ok {
			if match.Size != f.Size {
				panic("Collision, same hash but different size")
			}
			// matching group found; add this file to existing group
			fmt.Printf("Duplicates:\n\t%s\n\t%s\n", f, match.Files[0])
		}
	}
}

// Group computes the Snapshot's HashGroup
func (sn *Snapshot) Group() *HashGroup {
	// check for matching hash
	matches := make(HashGroup)

	for _, f := range sn.Files {
		match, ok := matches[f.Hash]
		if ok {
			if match.Size != f.Size {
				panic("Collision, same hash but different size")
			}
			// matching group found; add this file to existing group
			match.Files = append(match.Files, f)
		} else {
			// create new group in map
			matches[f.Hash] = &Group{[]*File{f}, f.Size}
		}
	}

	return &matches
}
