package core

import (
	"crypto/sha1"
	"fmt"
	"io/fs"
	"path/filepath"
)

// Node is an entry in the filetree. Either file or directory. A .hsnap file is
// actually made of a stream of Nodes.
// A Node is referenced by its ID. Each Node has a parent.
// FileInfo
type Node struct {
	Name string
	Mode fs.FileMode // Dir ? Link ? etc...
	Size uint64

	Hash [sha1.Size]byte // hash.Hash // sha1.New()

	parent   *Node
	children []*Node
	path     string // full absolute path with Name
}

// NewNode creates a Node given its FileInfo
func NewNode(info fs.FileInfo) *Node {
	return &Node{
		Mode: info.Mode(),
		Name: info.Name(),
		Size: uint64(info.Size()),
	}
}

func (n Node) String() string {
	return fmt.Sprintf("%d (in %d), %s %d %s", n.Name, n.Size, n.path)
}

// Path retrieve the absolute path of current Node by walking the parent tree
func (n *Node) Path() string {
	if n.path != "" {
		return n.path
	}
	if n.parent == nil {
		panic("Node has no parent set") // Root Node has always a path set
	}
	return filepath.Join(n.parent.Path(), n.Name)
}

// Attach nodes as children
func (n *Node) Attach(children ...*Node) {
	for _, c := range children {
		if c.parent != nil {
			panic("parent already declared")
		}
		c.parent = n
		n.children = append(n.children, c)
	}
}
