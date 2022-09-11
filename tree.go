package hashsnap

import (
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

func ReadTree(r io.Reader) (*Tree, error) { // options ?
	t := new(Tree)

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
	return matches
	// if err := Each(NewTreeIterator(cur), matches.Add); err != nil {
	// 	return err
	// }
	// if err := Each(NewTreeIterator(x), matches.Intersect); err != nil {
	// 	return err
	// }

}
