package hashsnap

import (
	"fmt"
	"io"
)

// Iterator goes through a Node slice with a similar interface to sql.Rows.
type Iterator interface {
	Node() *Node
	Error() error
	Next() bool
}

// Decoder decodes into e or sends an error. Used here to abstract gob.Decoder
type Decoder interface {
	Decode(e interface{}) error
}

// ReadAll Iterator's nodes and return them as a slice. Inspect Iterator's Error
// method to check if everything went well.
func ReadAll(it Iterator) []*Node {
	var r []*Node

	for it.Next() {
		r = append(r, it.Node())
	}

	return r
}

func Each(it Iterator, nf func(n *Node) error) error {
	for it.Next() {
		if err := nf(it.Node()); err != nil {
			return err
		}
	}
	return nil
}

var _ Iterator = &DecoderIterator{}

// DecoderIterator decodes a StateFile Node per Node.
type DecoderIterator struct {
	Decoder Decoder
	err     error
	n       *Node
	pos     int
}

// Node currently being decoded
func (ni *DecoderIterator) Node() *Node {
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
	n := Node{}
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
