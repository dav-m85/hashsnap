package core

import (
	"fmt"
	"path/filepath"
)

type Tree struct {
	nodes    map[int]*Node
	children map[int][]int
}

func NewTree() *Tree {
	return &Tree{
		nodes:    make(map[int]*Node),
		children: make(map[int][]int),
	}
}

func (t *Tree) RelPath(n *Node) (path string, err error) {
	on := n
	for n.ID > 0 {
		path = filepath.Join(n.Name, path)
		var ok bool
		if n, ok = t.nodes[n.ParentID]; !ok {
			err = fmt.Errorf("cannot resolve full path for %s, missing parent for %s", on, n)
			return
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
