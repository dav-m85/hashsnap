package internal

import (
	"io"
	"testing"

	"github.com/dav-m85/hsnap/memfs"
	"github.com/matryer/is"
)

func readTree(is *is.I, root string) (tr *Tree) {
	r, w := io.Pipe()
	go func() {
		Snapshot(root, w, io.Discard)
		w.Close()
	}()

	tr, err := ReadTree(r)
	is.NoErr(err)
	return
}

func TestBasicTrim(t *testing.T) {
	is := is.New(t)

	rootFS := memfs.New()

	is.NoErr(rootFS.MkdirAll("d1/d2", 0777))
	is.NoErr(rootFS.WriteFile("d1/d2/f1.txt", []byte("abc"), 0755))
	is.NoErr(rootFS.WriteFile("d1/d2/f1dup.txt", []byte("abc"), 0755)) // == f1

	is.NoErr(rootFS.MkdirAll("d2/d3", 0777))
	is.NoErr(rootFS.WriteFile("d2/d3/f2.txt", []byte("abc"), 0755)) // == f1

	FS = rootFS

	t1 := readTree(is, "d1")
	t2 := readTree(is, "d2")

	hg := t1.Trim(t2)
	nodes := hg.Select(t1)

	N(nodes).Equal(is, "f1.txt", "f1dup.txt")
	is.True(!N(nodes).Contains("f2.txt"))
}

// TestTrimWithDuplicate makes sure that a duplicated file in t1 won't get trimmed
// if not present in t2 (this is the job of dup)
func TestTrimWithDuplicate(t *testing.T) {
	is := is.New(t)

	rootFS := memfs.New()

	is.NoErr(rootFS.MkdirAll("d1/d2", 0777))
	is.NoErr(rootFS.WriteFile("d1/d2/f1.txt", []byte("abc"), 0755))
	is.NoErr(rootFS.WriteFile("d1/d2/f1dup.txt", []byte("abc"), 0755)) // == f1

	is.NoErr(rootFS.MkdirAll("d2/d3", 0777))
	is.NoErr(rootFS.WriteFile("d2/d3/f2.txt", []byte("def"), 0755)) // != f1

	FS = rootFS

	t1 := readTree(is, "d1")
	t2 := readTree(is, "d2")

	hg := t1.Trim(t2)
	hg.PruneSingleTreeGroups()
	nodes := hg.Select(t1)

	is.Equal(len(nodes), 0)
}

func TestSelfTrim(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	t1 := NewTree()
	t1.Info = new(Info)

	t1.Trim(t1)

}

type N []*Node

func (ns N) Contains(name string) bool {
	for _, n := range ns {
		if n.Name == name {
			return true
		}
	}
	return false
}

func (ns N) Equal(is *is.I, name ...string) {
	is.Equal(len(ns), len(name))
	for _, n := range name {
		is.True(ns.Contains(n))
	}
}

// func logTree(t *testing.T, tree *Tree) {
// 	for id, n := range tree.nodes {
// 		t.Logf("%d %s\n", id, n)
// 	}
// }
