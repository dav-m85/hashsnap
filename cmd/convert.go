package cmd

import (
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/dav-m85/hashsnap/core"
)

type legacyNode struct {
	ID   uint64
	Mode fs.FileMode // Dir ? Link ? etc...
	Name string

	Size uint64
	Hash [sha1.Size]byte // hash.Hash // sha1.New()

	// Parent node
	ParentID uint64

	// Root only has this
	RootPath string
}

// Convert old files...
func Convert() error {
	if len(os.Args) != 4 {
		return fmt.Errorf("wrong usage")
	}
	input := os.Args[2]
	output := os.Args[3]

	f, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("cannot open file: %s", err)
	}
	defer f.Close()

	o, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("cannot create file: %s", err)
	}
	defer o.Close()

	dec := gob.NewDecoder(f)
	enc := gob.NewEncoder(o)

	var headerWritten bool
	for {
		n := legacyNode{}
		err := dec.Decode(&n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot decode node: %s", err)
		}

		// Transform to new node
		nn := core.Node{
			Name: n.Name,
			Mode: n.Mode,
			Size: int64(n.Size),

			Hash: n.Hash,

			ID:       int(n.ID),
			ParentID: int(n.ParentID),
		}

		// Root node ?
		if n.RootPath != "" {
			if headerWritten {
				return fmt.Errorf("Rootpath has already been written")
			}
			headerWritten = true

			if err := enc.Encode(core.Info{
				Version:   1,
				RootPath:  n.RootPath,
				CreatedAt: time.Now(),
			}); err != nil {
				return err
			}
		}

		if err := enc.Encode(nn); err != nil {
			return err
		}
	}

	return nil
}
