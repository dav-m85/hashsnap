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

type StateFile struct {
	Path string
	Info core.Info
}

const STATE_NAME = ".hsnap"

// NewStateFileIn a directory
func NewStateFileIn(dir string) *StateFile {
	return NewStateFile(filepath.Join(dir, STATE_NAME))
}

// NewStateFile at specified statepath
func NewStateFile(statepath string) *StateFile {
	return &StateFile{
		Path: statepath,
	}
}

// StateIn traverses dir ancestors looking for a StateFile.
func StateIn(dir string) (*StateFile, error) {
	if !path.IsAbs(dir) {
		return nil, fmt.Errorf("statepath %s is not absolute", dir)
	}
	for {
		fp := filepath.Join(dir, STATE_NAME)
		_, err := os.Stat(fp)
		if err == nil {
			// found
			return NewStateFile(fp), nil
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
