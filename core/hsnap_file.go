package core

import (
	"bufio"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var _ Hsnap = &HsnapFile{}

// HsnapFile is a file holding a filetree snapshot
type HsnapFile struct {
	path string
}

func MakeHsnapFile(path string) *HsnapFile {
	if !strings.HasSuffix(path, ".hsnap") {
		log.Fatal("Snapshot file name should end with .hsnap")
	}
	return &HsnapFile{path}
}

// TODO ChannelRead and ChannelWrite could be detyped

func (h *HsnapFile) Info() Info {
	f, err := os.Open(h.path)
	if err != nil {
		panic(err)
	}

	w := bufio.NewReader(f)
	enc := gob.NewDecoder(w)

	var info *Info = &Info{}
	if err := enc.Decode(info); err != nil {
		panic(err)
	}
	if info.Version > 0 {
		return *info
	}

	// Old archive format support
	f.Seek(0, 0)
	w.Reset(f)
	enc = gob.NewDecoder(w)

	var n *Node = &Node{}

	if err := enc.Decode(n); err != nil {
		panic(err)
	}

	return Info{
		RootPath: n.RootPath,
		Version:  0,
	}
}

// ChannelRead decodes a hsnap file into a stream of *Node.
func (h *HsnapFile) ChannelRead(ctx context.Context) (<-chan *Node, error) {
	f, err := os.Open(h.path)
	if err != nil {
		panic(err)
	}

	out := make(chan *Node)
	go func() {
		defer f.Close()
		defer close(out)

		nodes := make(map[uint64]*Node)

		w := bufio.NewReader(f)
		enc := gob.NewDecoder(w)

		var h *Info = &Info{}
		err := enc.Decode(h)
		if err != nil || h.Version == 0 {
			fmt.Printf("Old archive found: %s\n", err)
			f.Seek(0, 0)
			w.Reset(f)
			enc = gob.NewDecoder(w)
		} else {
			fmt.Printf("%#v\n", h)
		}

		for {
			var n *Node = &Node{}
			err := enc.Decode(n)
			if err != nil {
				if err != io.EOF {
					log.Printf("Decoder encountered an issue: %s\n", err)
				}
				break
			}

			// Lets fill parent
			nodes[n.ID] = n
			if n.ParentID != 0 {
				parent, ok := nodes[n.ParentID]
				if !ok {
					log.Printf("Cannot solve parent")
					continue
				}
				n.parent = parent
			}

			select {
			case <-ctx.Done():
				return
			case out <- n:
			}
		}

	}()
	return out, nil
}

// ChannelWrite encodes a hsnap file given a stream of *Node.
func (h *HsnapFile) ChannelWrite(in <-chan *Node) error {
	if _, err := os.Stat(h.path); err == nil {
		return fmt.Errorf("%s already exists, aborting", h.path)
	}

	f, err := os.Create(h.path)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	defer f.Close()
	defer w.Flush()

	enc := gob.NewEncoder(w)

	enc.Encode(Info{
		Version:  1,
		RootPath: h.path,
	})

	for f := range in {
		enc.Encode(f)
	}

	return nil
}
