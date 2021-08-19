package core

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"sort"
	"strings"
)

// Group aka duplicates
type Group struct {
	Nodes []*Node // Localnodes
	Size  uint64
}

type ByPath []*Node

func (a ByPath) Len() int { return len(a) }
func (a ByPath) Less(i, j int) bool {
	x, err := a[i].Path()
	if err != nil {
		panic(err)
	}
	y, err := a[j].Path()
	if err != nil {
		panic(err)
	}
	return x < y
}
func (a ByPath) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (g Group) String() string {
	var s string
	for _, n := range g.Nodes {
		s = fmt.Sprintf("%s\t%s\n", s, n)
	}
	return s
}

type ByEmbedPath struct {
	Slice [][sha1.Size]byte
	HG    HashGroup
}

// HashGroup helps comparing Hashes pretty quickly
type HashGroup map[[sha1.Size]byte]*Group

// Dedup reports duplicates within an HashGroup
func (h HashGroup) Dedup() {
	// Keysort groups!
	keys := make([][sha1.Size]byte, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}

	var dupGroup uint64 = 0
	var dupSize uint64 = 0
	for _, group := range h {
		if len(group.Nodes) > 1 {
			cnt := len(group.Nodes)
			names := []string{}
			for _, g := range group.Nodes {
				p, _ := g.Path()
				names = append(names, p)
			}
			fmt.Printf("Duplicates %s (%s), %d times\n\t%s\n\n", group.Nodes[0].Name, ByteSize(group.Size), len(group.Nodes), strings.Join(names, "\n\t"))
			dupGroup++
			dupSize = dupSize + group.Size*uint64(cnt-1)
		}
	}
	fmt.Printf("Found %d duplicated groups, totalling %s", dupGroup, ByteSize(dupSize))
}

// Load ignores Dirs
func (hg HashGroup) Load(snap Hsnap) error {
	// var i uint64
	nodes, errc, err := snap.ChannelRead()(context.Background())
	if err != nil {
		return err
	}
	for {
		select {
		case n, ok := <-nodes:
			if !ok {
				// End of processing... let's sort them all
				for _, v := range hg {
					sort.Sort(ByPath(v.Nodes))
				}
				return nil
			}
			if n.Mode.IsDir() {
				continue
			}

			match, ok := hg[n.Hash]
			if ok {
				if match.Size != n.Size {
					log.Fatalln("Collision, same hash but different size")
				}
				// matching group found; add this file to existing group
				match.Nodes = append(match.Nodes, n)
			} else {
				// create new group in map
				hg[n.Hash] = &Group{[]*Node{n}, n.Size}
			}

			// i++
		case err := <-errc:
			return err
		}
	}
}
