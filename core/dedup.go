package core

import (
	"crypto/sha1"
	"fmt"
	"log"
)

// Group aka duplicates
type Group struct {
	Nodes []*Node
	Size  uint64
}

// HashGroup helps comparing Hashes pretty quickly
type HashGroup map[[sha1.Size]byte]*Group

// Dedup reports duplicates within an HashGroup
func (h *HashGroup) Dedup() {
	var dupGroup uint64 = 0
	var dupSize uint64 = 0
	for _, group := range *h {
		if len(group.Nodes) > 1 {
			cnt := len(group.Nodes)
			fmt.Printf("Duplicates %s, %d times\n", group.Nodes[0].Name, len(group.Nodes))
			dupGroup++
			dupSize = dupSize + group.Size*uint64(cnt-1)

			fmt.Println(group.Nodes[0].Path())
		}
	}
	fmt.Printf("Found %d duplicated groups, totalling %d bytes", dupGroup, dupSize)
}

// DedupWith reports duplicates belonging both to a Snapshot and a given HashGroup
// func (sn *Snapshot) DedupWith(hb *HashGroup) {
// 	for _, f := range sn.Files {
// 		match, ok := (*hb)[f.Hash]
// 		if ok {
// 			if match.Size != f.Size {
// 				panic("Collision, same hash but different size")
// 			}
// 			// matching group found; add this file to existing group
// 			fmt.Printf("Duplicates:\n\t%s\n\t%s\n", f, match.Files[0])
// 		}
// 	}
// }

func Dedup(target string) {
	nodes := make(chan *Node)

	go reader(target, nodes)
	// var count uint64 = 0

	matches := make(HashGroup)

	for n := range nodes {
		if n.RootPath != "" {
			fmt.Printf("Snapshot captured in %s\n", n.RootPath)
		}

		if n.Mode.IsDir() {
			continue
		}

		match, ok := matches[n.Hash]
		if ok {
			if match.Size != n.Size {
				log.Println("Collision, same hash but different size")
				continue
			}
			// matching group found; add this file to existing group
			match.Nodes = append(match.Nodes, n)
		} else {
			// create new group in map
			matches[n.Hash] = &Group{[]*Node{n}, n.Size}
		}
	}

	matches.Dedup()
}
