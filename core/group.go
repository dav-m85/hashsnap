package core

import (
	"crypto/sha1"
	"fmt"
	"sort"
)

// Group of duplicated Nodes sharing the same Hash (and Size)
type Group struct {
	Nodes []*Node
	Hash  [sha1.Size]byte
	Size  int64
}

type ByPath []*Node

func (a ByPath) Len() int { return len(a) }
func (a ByPath) Less(i, j int) bool {
	x := a[i].Path()
	y := a[j].Path()
	return x < y
}
func (a ByPath) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (g Group) String() string {
	s := fmt.Sprintf("%d nodes (save %s)\n", len(g.Nodes), g.WastedSize())
	sort.Sort(ByPath(g.Nodes))
	for _, n := range g.Nodes {
		s = s + fmt.Sprintf("\t%s\n", n)
	}
	return s
}

func (g Group) WastedSize() ByteSize {
	return ByteSize(g.Size * int64(len(g.Nodes)-1))
}

// HashGroup helps comparing Hashes pretty quickly
type HashGroup map[[sha1.Size]byte]*Group

// Add a Node slice to HashGroup
func (r HashGroup) Add(ns []*Node) error {
	for _, n := range ns {
		if n.Mode.IsDir() {
			continue
		}
		if grp, ok := r[n.Hash]; ok {
			if grp.Size != n.Size {
				return fmt.Errorf("collision, same hash but different size")
			}
			// matching group found; add this file to existing group
			grp.Nodes = append(grp.Nodes, n)
		} else {
			// create new group in map
			r[n.Hash] = &Group{[]*Node{n}, n.Hash, n.Size}
		}
	}
	return nil
}

// Intersect adds nodes if their hash is already present (does not create new groups)
func (r HashGroup) Intersect(ns []*Node) error {
	for _, n := range ns {
		if n.Mode.IsDir() {
			continue
		}
		if grp, ok := r[n.Hash]; ok {
			if grp.Size != n.Size {
				return fmt.Errorf("collision, same hash but different size")
			}
			// matching group found; add this file to existing group
			grp.Nodes = append(grp.Nodes, n)
		}
	}
	return nil
}
