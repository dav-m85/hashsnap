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
	Size int64

	Hash [sha1.Size]byte // hash.Hash // sha1.New()

	ID, ParentID int

	parent   *Node
	children []*Node
}

func (n *Node) Parent() *Node {
	return n.parent
}

// Children of Node. Beware, while decoding, this field could be not entirely
// filled until whole file has been decoded.
func (n *Node) Children() []*Node {
	return n.children
}

var IncrementID = 0

// NewNode creates a Node given its FileInfo
func NewNode(info fs.FileInfo) *Node {
	IncrementID++
	return &Node{
		ID:   IncrementID,
		Mode: info.Mode(),
		Name: info.Name(),
		Size: info.Size(),
	}
}

func (n Node) String() string {
	return fmt.Sprintf("%d %s (in %d)", n.ID, n.Path(), n.ParentID)
}

// Path relative to the root Node
// TODO cache path result !
func (n *Node) Path() string {
	// if n.path != "" {
	// 	return n.path
	// }
	if n.parent == nil {
		return ""
	}
	pp := n.parent.Path()
	if pp == "" {
		return n.Name
	}
	return filepath.Join(n.parent.Path(), n.Name)
}

// Attach nodes as children
func (n *Node) Attach(children ...*Node) {
	n.children = append(n.children, children...)
	for _, c := range children {
		if c.parent != nil {
			panic("parent already declared")
		}
		c.parent = n
		c.ParentID = n.ID // Useful for decoding
	}
}
