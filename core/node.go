package core

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	bar "github.com/schollz/progressbar/v3"
)

// lstat is a proxy for os.Lstat
var lstat = os.Lstat

// Node is an entry in the filetree. Either file or directory. A .hsnap file is
// actually made of a stream of Nodes.
// A Node is referenced by its ID. Each Node has a parent.
// FileInfo
type Node struct {
	ID   uint64
	Mode fs.FileMode // Dir ? Link ? etc...
	Name string

	Size uint64
	Hash [sha1.Size]byte // hash.Hash // sha1.New()

	// Parent node
	ParentID uint64
	parent   *Node  // Unexported, we do not want this in the .hsnap file
	path     string // full absolute path with Name
	depth    uint64

	// Root only has this
	RootPath string
}

func (n Node) String() string {
	return fmt.Sprintf("%d (in %d), %s %d %s", n.ID, n.ParentID, n.Name, n.Size, n.path)
}

// Path retrieve the absolute path of current Node by walking the parent tree
func (n Node) Path() (string, error) {
	if n.RootPath != "" {
		return n.RootPath, nil
	}
	if n.parent == nil {
		return "", fmt.Errorf("Node %d has no parent set", n.ID)
	}
	res, err := n.parent.Path()
	return filepath.Join(res, n.Name), err
}

func MakeNode(parent *Node, name string) (*Node, error) {
	if parent.ID == 0 {
		panic("parent.ID has to be set")
	}
	path := filepath.Join(parent.path, name)
	info, err := lstat(path)
	if err != nil {
		return nil, fmt.Errorf("lstat %s failed: %s", path, err)
	}

	return &Node{
		Mode:     info.Mode(),
		Name:     name,
		Size:     uint64(info.Size()),
		parent:   parent,
		ParentID: parent.ID,
		path:     path,
		depth:    parent.depth + 1,
	}, nil
}

func MakeRootNode(path string) *Node {
	info, err := lstat(path)
	if err != nil {
		log.Fatalf("Cannot create root node: %s", err)
	}
	return &Node{
		Mode:     info.Mode(),
		Name:     info.Name(),
		Size:     uint64(info.Size()),
		parent:   nil,
		ParentID: 0, // Trivial, root Node has no parent
		path:     path,
		depth:    0,
		RootPath: path,
	}
}

// ComputeHash reads the file and computes the sha1 of it
func (n *Node) ComputeHash(pbar *bar.ProgressBar) error {
	fd, err := os.Open(n.path)
	if err != nil {
		return err
	}
	h := sha1.New()
	defer fd.Close()

	var writeTo io.Writer = h
	if pbar != nil {
		writeTo = io.MultiWriter(h, pbar)
	}
	if _, err = io.Copy(writeTo, fd); err != nil {
		return err
	}

	copy(n.Hash[:], h.Sum(nil)) // [sha1.Size]byte()

	return nil
}
