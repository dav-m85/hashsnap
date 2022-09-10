package core

import (
	"fmt"
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

func NewTree(info *Info) *Tree {
	return &Tree{
		nodes:    make(map[int]*Node),
		children: make(map[int][]int),
	}
}

func (t *Tree) ReadIterator(it Iterator) error {
	for it.Next() {
		n := it.Node()
		t.Add(n)
		n.tree = t
	}
	return it.Error()
}

func (t *Tree) RelPath(n *Node) (path string) {
	// on := n
	for n.ID > 0 {
		path = filepath.Join(n.Name, path)
		var ok bool
		if n, ok = t.nodes[n.ParentID]; !ok {
			// err = fmt.Errorf("cannot resolve full path for %s, missing parent for %s", on, n)
			return ""
		}
	}
	return
}

// TODO check it is connected
func (t *Tree) IsOK() (bool, error) {
	if len(t.nodes) != len(t.children) {
		return false, fmt.Errorf("children entries and nodes entries are not the same length")
	}
	return false, fmt.Errorf("Unimplemented")
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
}

var _ Iterator = &TreeIterator{}

// TreeIterator decodes a StateFile Node per Node.
type TreeIterator struct {
	n     int
	nodes []*Node
}

func NewTreeIterator(tree *Tree) *TreeIterator {
	tr := new(TreeIterator)
	for _, n := range tree.nodes {
		tr.nodes = append(tr.nodes, n)
	}
	tr.n = -1
	return tr
}

// Node currently being decoded
func (ni *TreeIterator) Node() *Node {
	if ni.n == -1 {
		panic("Call Next at least once")
	}
	return ni.nodes[ni.n]
}

// Error returned after a call to Next if any
func (ni *TreeIterator) Error() error {
	return nil
}

// Next yields true if a new Node has been decoded. On end of tree or failure,
// it'll return false.
func (ni *TreeIterator) Next() bool {
	ni.n++
	return ni.n < len(ni.nodes)
}
