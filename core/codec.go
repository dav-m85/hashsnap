package core

import (
	"encoding/gob"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const MaxInt = int(^uint(0) >> 1)

// Info header
type Info struct {
	RootPath  string
	CreatedAt time.Time
	Version   int
	Nonce     uuid.UUID
}

type NodeIDMap map[int]*Node

// NodeDecoder rebuilds Node.parent and Node.children thanks to IDs
type NodeDecoder struct {
	dec *gob.Decoder
	nrm NodeIDMap
}

func NewDecoder(dec *gob.Decoder) *NodeDecoder {
	ne := &NodeDecoder{
		dec: dec,
		nrm: make(NodeIDMap),
	}
	return ne
}

func (nd *NodeDecoder) Decode(n *Node) error {
	err := nd.dec.Decode(&n)
	if err != nil {
		return err
	}

	nd.nrm[n.ID] = n
	// Does not solve for root
	if n.ID == 1 || n.ID == 0 {
		return nil
	}

	var parent *Node
	var ok bool
	if parent, ok = nd.nrm[n.ParentID]; !ok {
		return fmt.Errorf("parent of %s has not been decoded yet", n)
	}
	parent.Attach(n)

	return nil
}
