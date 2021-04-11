package core

import (
	"crypto/sha1"
	"fmt"
	"log"
)

// Group aka duplicates
type Group struct {
	Nodes []*Node // Localnodes
	Size  uint64
}

func (g Group) String() string {
	var s string
	for _, n := range g.Nodes {
		s = fmt.Sprintf("%s\t%s\n", s, n)
	}
	return s
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

func (hg *HashGroup) Contains(n *Node) (*Group, bool) {
	g, ok := (*hg)[n.Hash]
	return g, ok
}

// Load ignores Dirs
func (hg *HashGroup) Load(snap Hsnap) uint64 {
	nodes := make(chan *Node)
	var i uint64
	go snap.ChannelRead(nodes)
	for n := range nodes {
		if n.Mode.IsDir() {
			continue
		}

		match, ok := (*hg)[n.Hash]
		if ok {
			if match.Size != n.Size {
				log.Fatalln("Collision, same hash but different size")
			}
			// matching group found; add this file to existing group
			match.Nodes = append(match.Nodes, n)
		} else {
			// create new group in map
			(*hg)[n.Hash] = &Group{[]*Node{n}, n.Size}
		}

		i++
	}

	return i
}
