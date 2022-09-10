package core

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
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
	// x := a[i].Path()
	// y := a[j].Path()
	return false //x < y
}
func (a ByPath) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (g Group) String(current *Tree) string {
	s := fmt.Sprintf("%d nodes (save %s)\n", len(g.Nodes), g.WastedSize())
	sort.Sort(ByPath(g.Nodes))
	for _, n := range g.Nodes {
		if n.tree == current {
			s = s + fmt.Sprintf("\tC %s [%s]\n", n, n.tree.RelPath(n))
		} else {
			s = s + fmt.Sprintf("\td %s\n", n)
		}
	}
	return s
}

func (g Group) WastedSize() ByteSize {
	return ByteSize(g.Size * int64(len(g.Nodes)-1))
}

// HashGroup helps comparing Hashes pretty quickly
type HashGroup map[[sha1.Size]byte]*Group

// Add a Node slice to HashGroup
func (r HashGroup) Add(n *Node) error {
	if n.Mode.IsDir() {
		return nil
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
	return nil
}

// Intersect adds nodes if their hash is already present (does not create new groups)
func (r HashGroup) Intersect(n *Node) error {
	if n.Mode.IsDir() {
		return nil
	}
	if grp, ok := r[n.Hash]; ok {
		if grp.Size != n.Size {
			return fmt.Errorf("collision, same hash but different size")
		}
		// matching group found; add this file to existing group
		grp.Nodes = append(grp.Nodes, n)
	}
	return nil
}

var output io.ReadWriter = os.Stdout

func (r HashGroup) PrintDetails(verbose bool) {
	var count int64
	var waste int64

	for _, g := range r {
		if len(g.Nodes) < 2 {
			continue
		}
		if verbose {
			fmt.Fprintln(output, g)
		}
		count++
		waste = waste + int64(g.WastedSize())
	}

	fmt.Fprintf(output, "%d duplicated groups, totalling %s wasted space\n", count, ByteSize(waste))
}
