package state

import (
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dav-m85/hashsnap/core"
	"github.com/google/uuid"
)

// package state retrieves current hsnap file, in current
// or parent dirs

// StateFile is a file holding a bunch of nodes. It is decoded using
// an iterator
type StateFile struct {
	Path string
	Info core.Info
}

const STATE_NAME = ".hsnap"

// NewIn a directory
func NewIn(dir string) *StateFile {
	return New(filepath.Join(dir, STATE_NAME))
}

// New at specified statepath
func New(statepath string) *StateFile {
	return &StateFile{
		Path: statepath,
	}
}

// LookupFrom traverses dir ancestors looking for a StateFile.
func LookupFrom(dir string) (*StateFile, error) {
	if !path.IsAbs(dir) {
		return nil, fmt.Errorf("statepath %s is not absolute", dir)
	}
	for {
		fp := filepath.Join(dir, STATE_NAME)
		_, err := os.Stat(fp)
		if err == nil {
			// found
			return New(fp), nil
		}
		if os.IsNotExist(err) {
			nd := path.Dir(dir)
			if nd == dir {
				return nil, nil
			}
			dir = nd
			continue
		}
		return nil, err
	}
}

// Create writes a new StateFile on disk, or fails if it already exists.
func (sf *StateFile) Create() (*gob.Encoder, func() error, error) {
	if _, err := os.Stat(sf.Path); err == nil {
		return nil, nil, fmt.Errorf("%s already exists, aborting", sf.Path)
	}

	f, err := os.Create(sf.Path)
	if err != nil {
		return nil, nil, err
	}

	enc := gob.NewEncoder(f)

	// Write info node
	err = enc.Encode(core.Info{
		Version:   1,
		RootPath:  path.Dir(sf.Path),
		CreatedAt: time.Now(),
		Nonce:     uuid.New(),
	})
	if err != nil {
		return nil, nil, err
	}

	return enc, f.Close, err
}

// Nodes spawns a NodeIterator for current StateFile
func (sf *StateFile) Nodes() (Iterator, error) {
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

	return &FileIterator{
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
