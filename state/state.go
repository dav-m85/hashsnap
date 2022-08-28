package state

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// package state retrieves current hsnap file, in current
// or parent dirs

const STATE_NAME = ".hsnap"

// LookupFrom traverses dir ancestors looking for a StateFile.
func LookupFrom(dir string) (string, error) {
	if !path.IsAbs(dir) {
		return "", fmt.Errorf("statepath %s is not absolute", dir)
	}
	for {
		fp := filepath.Join(dir, STATE_NAME)
		_, err := os.Stat(fp)
		if err == nil {
			// found
			return fp, nil
		}
		if os.IsNotExist(err) {
			nd := path.Dir(dir)
			if nd == dir {
				return "", nil
			}
			dir = nd
			continue
		}
		return "", err
	}
}
