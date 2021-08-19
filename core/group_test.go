package core

import (
	"crypto/sha1"
	"io"
	"io/fs"
	"testing"

	"github.com/matryer/is"
)

func flatten(nodes [][]*Node) []*Node {
	var res []*Node
	for _, n := range nodes {
		res = append(res, n...)
	}
	return res
}

var ni uint64 = 1

func n(name, content string, children ...[]*Node) []*Node {
	n := &Node{
		ID:       ni,
		ParentID: 0,
		Name:     name,
	}
	ni++

	if content != "" {
		if len(children) > 0 {
			panic("cannot have content and children")
		}
		h := sha1.New()
		io.WriteString(h, content)
		copy(n.Hash[:], h.Sum(nil))
		// n.Mode: default file
		n.Size = uint64(len(content))
	} else {
		if len(children) == 0 {
			panic("cannot have no children and no content")
		}
		n.Mode = fs.ModeDir
	}

	for _, nc := range children {
		nc[0].ParentID = n.ID
	}

	flat := flatten(children)

	return append([]*Node{n}, flat...)
}

func TestAbs(t *testing.T) {
	is := is.New(t)

	a := &HsnapMem{
		Nodes: n(
			"/", "", // dir
			n("a", "", // dir
				n("c", "aze"),
				n("d", "foo"),
			),
			n("b", "bar"),
		),
	}

	t.Logf("%s\n", a)

	h := make(HashGroup)
	h.Load(a)

	// This one is their
	_, ok := h[n("z", "aze")[0].Hash]
	is.True(ok)

	// This one not
	_, ok = h[n("z", "foe")[0].Hash]
	is.True(!ok)

	// This one is a dir
	_, ok = h[n("/", "", n("z", "foe"))[0].Hash]
	is.True(!ok)
}
