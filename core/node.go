package core

import (
	"crypto/sha1"
	"fmt"
	"io/fs"
	"sync"
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
	tree         *Tree
}

func (n *Node) Tree() *Tree {
	return n.tree
}

type NodeP struct {
	Node *Node
	Path string
}

var incrementID = struct {
	value int
	sync.Mutex
}{}

func Allocate() int {
	incrementID.Lock()
	defer incrementID.Unlock()
	incrementID.value++
	return incrementID.value
}

func Reset() int {
	incrementID.Lock()
	defer incrementID.Unlock()
	incrementID.value = 0
	return incrementID.value
}

func (n Node) String() string {
	return fmt.Sprintf("%d(%d) %s", n.ID, n.ParentID, n.Name)
}
