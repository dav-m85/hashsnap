package hashsnap

// var F func(t *Tree, path string, content string)
// var expect func(t *Tree, path string)
// var deleted func(t *Tree, path string)

// func TestTrim() {
// 	x := NewTree(&Info{})
// 	F(x, "a", "a")
// 	F(x, "a2", "a")
// 	F(x, "b", "b")

// 	y := NewTree(&Info{})
// 	F(y, "b", "b")

// 	// trim x with y
// 	// Make sure a duplicate file in x won't get deleted by trimming it without that file
// 	// here y{b} deletes x{b} but not x{a},x{a2} which are duplicates...
// 	expect(x, "a")
// 	expect(x, "a2")
// 	deleted(x, "b")
// }

import (
	"io"
	"testing"

	"github.com/dav-m85/hashsnap/memfs"
	"github.com/matryer/is"
)

func TestSnapshot(t *testing.T) {
	is := is.New(t)

	rootFS := memfs.New()

	err := rootFS.MkdirAll("dir1/dir2", 0777)
	if err != nil {
		panic(err)
	}

	err = rootFS.WriteFile("dir1/dir2/f1.txt", []byte("incinerating-unsubstantial"), 0755)
	if err != nil {
		panic(err)
	}

	FS = rootFS

	r, w := io.Pipe()
	go func() {
		Snapshot("/.", w, io.Discard)
		w.Close()
	}()

	dut, err := ReadTree(r)
	is.NoErr(err)

	t.Logf("%#v", dut)
}
