package dedup

import (
	"bytes"
	"testing"

	"github.com/dav-m85/hashsnap/core"
	"github.com/matryer/is"
)

type iterator struct {
	nodes []*core.Node
	err   error
	at    int
}

func (it *iterator) Node() *core.Node { return it.nodes[it.at] }
func (it *iterator) Close() error     { return nil }
func (it *iterator) Error() error     { return it.err }
func (it *iterator) Next() bool       { it.at++; return it.at < len(it.nodes) }

func (it *iterator) Nodes() (state.Iterator, error) { return it, nil }

func TestInfo(t *testing.T) {
	a := &core.Node{Name: "a"}
	b := &core.Node{Name: "b"}
	is := is.New(t)

	output = &bytes.Buffer{}

	is.NoErr(Trim(
		&iterator{nodes: []*core.Node{a, b}},
		false,
		&iterator{nodes: []*core.Node{a, b}},
	))

	t.Log(output)
}
