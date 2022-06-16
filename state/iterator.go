package state

import (
	"fmt"
	"io"
	"os"

	"github.com/dav-m85/hashsnap/core"
)

// Iterator goes through a Node slice with a similar interface to sql.Rows.
type Iterator interface {
	Node() *core.Node
	Close() error
	Error() error
	Next() bool
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

// FileIterator decodes a StateFile Node per Node.
type FileIterator struct {
	file *os.File
	ndec *core.NodeDecoder
	err  error
	n    *core.Node
	pos  int
	path string
}

// Node currently being decoded
func (ni *FileIterator) Node() *core.Node {
	if ni.n == nil {
		panic("Call Next at least once")
	}
	return ni.n
}

// Close all file operation
func (ni *FileIterator) Close() error {
	return ni.file.Close()
}

// Error returned after a call to Next if any
func (ni *FileIterator) Error() error {
	// fmt.Errorf("state %s nodes error: %w", st.Path, err)
	return ni.err
}

// Next yields true if a new Node has been decoded. On end of file or failure,
// it'll return false. Inspect NodeIterator.Error for possible errors.
func (ni *FileIterator) Next() bool {
	n := core.Node{}
	err := ni.ndec.Decode(&n)
	if err == io.EOF {
		return false
	}
	if err != nil {
		ni.err = fmt.Errorf("cannot decode node at pos %d in %s: %s", ni.pos, ni.path, err)
		return false
	}
	ni.n = &n
	ni.pos++
	return true
}
