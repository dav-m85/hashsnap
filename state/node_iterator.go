package state

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"github.com/dav-m85/hashsnap/core"
)

// NodeIterator allows decoding StateFile node per node, with a similar interface
// to sql.Rows.
type NodeIterator struct {
	file *os.File
	ndec *core.NodeDecoder
	err  error
	n    *core.Node
	pos  int
	path string
}

// Node currently being decoded
func (ni *NodeIterator) Node() *core.Node {
	if ni.n == nil {
		panic("Call Next at least once")
	}
	return ni.n
}

// Close all file operation
func (ni *NodeIterator) Close() error {
	return ni.file.Close()
}

// Error returned after a call to Next if any
func (ni *NodeIterator) Error() error {
	// fmt.Errorf("state %s nodes error: %w", st.Path, err)
	return ni.err
}

// Next yields true if a new Node has been decoded. On end of file or failure,
// it'll return false. Inspect NodeIterator.Error for possible errors.
func (ni *NodeIterator) Next() bool {
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

// Nodes spawns a NodeIterator for current StateFile
func (sf *StateFile) Nodes() (*NodeIterator, error) {
	f, err := os.Open(sf.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %s", err)
	}
	dec := gob.NewDecoder(f)

	// Read the info header
	err = dec.Decode(&sf.Info)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("cannot decode info header: %s", err)
	}

	ndec := core.NewDecoder(dec)

	return &NodeIterator{
		file: f,
		ndec: ndec,
		path: sf.Path,
	}, nil
}

// Nodes spawns a NodeIterator for current StateFile
func (sf *StateFile) ReadInfo() error {
	f, err := os.Open(sf.Path)
	if err != nil {
		return fmt.Errorf("cannot open file: %s", err)
	}
	defer f.Close()
	dec := gob.NewDecoder(f)

	// Read the info header
	err = dec.Decode(&sf.Info)
	if err != nil {
		f.Close()
		return fmt.Errorf("cannot decode info header: %s", err)
	}

	return nil
}

func ReadAll(nodes *NodeIterator) []*core.Node {
	var r []*core.Node

	for nodes.Next() {
		r = append(r, nodes.Node())
	}

	return r
}
