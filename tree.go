package hashsnap

import (
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"io/fs"
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
	Hostname  string
}

func (i *Info) String() string {
	return fmt.Sprintf("%s@%s (v%d %s)", i.Hostname, i.RootPath, i.Version, i.Nonce.String()[:8])
}

// Tree structure that holds a filesystem
type Tree struct {
	Info     *Info
	Name     string
	nodes    map[int]*Node
	children map[int][]int
}

func NewTree() *Tree {
	return &Tree{
		nodes:    make(map[int]*Node),
		children: make(map[int][]int),
	}
}

// ReadTree into a Tree, usually from a fs.Open
func ReadTree(r io.Reader) (*Tree, error) {
	t := NewTree()

	dec := gob.NewDecoder(r)

	i := new(Info)
	if err := dec.Decode(i); err != nil {
		return t, err
	}

	if i.Version != VERSION {
		panic("Not version 1")
	}

	t.Info = i

	err := DecodeNodes(dec, func(n *Node) error {
		t.Add(n)
		return nil
	})

	return t, err
}

func DecodeNodes(dec *gob.Decoder, hf func(*Node) error) error {
	for {
		n := new(Node)
		err := dec.Decode(n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = hf(n)
		if err != nil {
			return err
		}
	}
	return nil
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

func (t *Tree) Root() *Node {
	if n, ok := t.nodes[0]; ok {
		return n
	}
	if n, ok := t.nodes[1]; ok {
		return n
	}
	panic("No root node in tree")
}

func (t *Tree) Search(path string) *Node {
	for _, x := range t.nodes {
		rel := t.RelPath(x)
		if rel == path {
			return x
		}
	}
	return nil
}

func (t *Tree) ChildrenOf(n *Node) (ns Nodes) {
	if n == nil {
		panic("oops")
	}
	for _, x := range t.nodes {
		if x.ParentID == n.ID {
			ns = append(ns, x)
		}
	}
	return
}

func (t *Tree) RelPath(n *Node) (path string) {
	if n.tree != t {
		panic("wrong tree used for resolving path")
	}
	on := n
	for n.ID > 0 /*&& n.ParentID != 0*/ {
		path = filepath.Join(n.Name, path)
		var ok bool
		if n, ok = t.nodes[n.ParentID]; !ok {
			panic(fmt.Sprintf("cannot resolve full path for %s, missing parent for %s in %s", on, n, t.Info))
		}
	}
	return
}

func (t *Tree) AbsPath(n *Node) (path string) {
	return filepath.Join(t.Info.RootPath, t.RelPath(n))
}

func (t *Tree) Trim(withs ...*Tree) HashGroup {
	matches := make(HashGroup)
	for _, n := range t.nodes {
		matches.Add(n)
	}
	for _, tx := range withs {
		if t.Info.Nonce == tx.Info.Nonce {
			panic("cannot trim with self")
		}
		for _, m := range tx.nodes {
			matches.Intersect(m)
		}
	}

	return matches
}

func (t *Tree) Check() (missing Nodes) {
	for _, n := range t.nodes {
		lstat := FS.(fs.StatFS).Stat
		_, err := lstat(n.Path())
		if err != nil {
			missing = append(missing, n)
		}
	}
	return
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

func (r HashGroup) PruneSingleTreeGroups() {
	for hash, g := range r {
		if len(g) <= 1 {
			delete(r, hash)
			continue
		}
		ts := make(map[uuid.UUID]struct{})
		for _, n := range g {
			ts[n.tree.Info.Nonce] = struct{}{}
		}
		if len(ts) <= 1 {
			delete(r, hash)
		}
	}
}

// Select all nodes in given tree, when
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

type Nodes []*Node

func (ns Nodes) ByteSize() ByteSize {
	if ns == nil || len(ns) < 1 {
		return 0
	} else {
		return ByteSize(ns[0].Size * int64(len(ns)))
	}
}

// SplitNodes by tree appartenance
// If owning tree is t, then node is in, else is out
func SplitNodes(t *Tree, ns []*Node) (in, out []*Node) {
	for _, n := range ns {
		if n.tree == t {
			in = append(in, n)
		} else {
			out = append(out, n)
		}
	}
	return
}
