package state

import (
	"fmt"
	"io"

	"github.com/dav-m85/hashsnap/core"
)

// Iterator goes through a Node slice with a similar interface to sql.Rows.
type Iterator interface {
	Node() *core.Node
	Error() error
	Next() bool
}

type Decoder interface {
	Decode(e interface{}) error
}

// ReadAll Iterator's nodes and return them as a slice. Inspect Iterator's Error
// method to check if everything went well.
func ReadAll(nodes Iterator) []*core.Node {
	var r []*core.Node

	for nodes.Next() {
		r = append(r, nodes.Node())
	}

	return r
}

// DecoderIterator decodes a StateFile Node per Node.
type DecoderIterator struct {
	Decoder Decoder
	err     error
	n       *core.Node
	pos     int
}

// Node currently being decoded
func (ni *DecoderIterator) Node() *core.Node {
	if ni.n == nil {
		panic("Call Next at least once")
	}
	return ni.n
}

// Error returned after a call to Next if any
func (ni *DecoderIterator) Error() error {
	return ni.err
}

// Next yields true if a new Node has been decoded. On end of file or failure,
// it'll return false. Inspect DecoderIterator.Error for possible errors.
func (ni *DecoderIterator) Next() bool {
	n := core.Node{}
	err := ni.Decoder.Decode(&n)
	if err == io.EOF {
		return false
	}
	if err != nil {
		ni.err = fmt.Errorf("cannot decode node at pos %d: %s", ni.pos, err)
		return false
	}
	ni.n = &n
	ni.pos++
	return true
}
