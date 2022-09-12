package hashsnap

import (
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// const MaxInt = int(^uint(0) >> 1)

// Info header
type Info struct {
	RootPath  string
	CreatedAt time.Time
	Version   int
	Nonce     uuid.UUID
}

func (i *Info) String() string {
	return fmt.Sprintf("v%d %s (%s)", i.Version, i.RootPath, i.Nonce.String()[:8])
}

type Tree struct {
	info     *Info
	nodes    map[int]*Node
	children map[int][]int
}

func NewTree() *Tree {
	return &Tree{
		nodes:    make(map[int]*Node),
		children: make(map[int][]int),
	}
}

func ReadTree(r io.Reader) (*Tree, error) { // options ?
	t := NewTree()

	dec := gob.NewDecoder(r)

	i := new(Info)
	if err := dec.Decode(i); err != nil {
		return t, err
	}

	t.info = i

	for {
		n := new(Node)
		err := dec.Decode(n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return t, err
		}
		t.Add(n)
	}

	return t, nil
}

func (t *Tree) Add(n *Node) {
	if _, ok := t.nodes[n.ID]; ok {
		panic("Already added that node")
	}
	t.nodes[n.ID] = n
	if _, ok := t.children[n.ParentID]; !ok {
		t.children[n.ParentID] = []int{}
	}
	t.children[n.ParentID] = append(t.children[n.ParentID], n.ID)
	n.tree = t
}

func (t *Tree) RelPath(n *Node) (path string) {
	if n.tree != t {
		panic("wrong tree used for resolving path")
	}
	on := n
	for n.ID > 0 {
		path = filepath.Join(n.Name, path)
		var ok bool
		if n, ok = t.nodes[n.ParentID]; !ok {
			panic(fmt.Sprintf("cannot resolve full path for %s, missing parent for %s", on, n))
		}
	}
	return
}

func (t *Tree) AbsPath(n *Node) (path string) {
	return filepath.Join(t.info.RootPath, t.RelPath(n))
}

func (t *Tree) Trim(withs ...*Tree) HashGroup {
	matches := make(HashGroup)
	for _, n := range t.nodes {
		matches.Add(n)
	}
	for _, tx := range withs {
		for _, m := range tx.nodes {
			matches.Intersect(m)
		}
	}

	return matches
}

// HashGroup helps comparing Hashes pretty quickly
type HashGroup map[[sha1.Size]byte][]*Node

// Add a Node slice to HashGroup
func (r HashGroup) Add(n *Node) {
	if n.Mode.IsDir() {
		return
	}
	if grp, ok := r[n.Hash]; ok {
		size := r[n.Hash][0].Size
		if size != n.Size {
			panic("collision, same hash but different size")
		}
		// matching group found; add this file to existing group
		r[n.Hash] = append(grp, n)
	} else {
		// create new group in map
		r[n.Hash] = []*Node{n}
	}
}

// Intersect adds nodes if their hash is already present (does not create new groups)
func (r HashGroup) Intersect(n *Node) {
	if n.Mode.IsDir() {
		return
	}
	if grp, ok := r[n.Hash]; ok {
		size := r[n.Hash][0].Size
		if size != n.Size {
			panic("collision, same hash but different size")
		}
		// matching group found; add this file to existing group
		r[n.Hash] = append(grp, n)
	}
}

// Select all nodes in given tree
func (r HashGroup) Select(t *Tree) (ns []*Node) {
	for _, g := range r {
		for _, n := range g {
			if n.tree == t {
				ns = append(ns, n)
			}
		}
	}
	return
}

type WastedSize []*Node

func (ws WastedSize) String() ByteSize {
	if ws == nil || len(ws) < 1 {
		return 0
	} else {
		return ByteSize(ws[0].Size * int64(len(ws)-1))
	}
}

// func (r HashGroup) PrintDetails(verbose bool) {
// 	var count int64
// 	var waste int64

// 	for _, g := range r {
// 		if len(g.Nodes) < 2 {
// 			continue
// 		}
// 		if verbose {
// 			fmt.Fprintln(output, g)
// 		}
// 		count++
// 		waste = waste + int64(g.WastedSize())
// 	}

// 	fmt.Fprintf(output, "%d duplicated groups, totalling %s wasted space\n", count, ByteSize(waste))
// }
