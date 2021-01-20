package core

import (
	"bufio"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	bar "github.com/schollz/progressbar/v3"
)

type File struct {
	Path string
	Size int64
	Hash [sha1.Size]byte // hash.Hash // sha1.New()
}

func (f File) String() string {
	return fmt.Sprintf("%s(%d)[sha1:%x]", f.Path, f.Size, f.Hash)
}

func (f *File) ComputeHash(pbar *bar.ProgressBar) error {
	fd, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	h := sha1.New()
	defer fd.Close()

	if _, err = io.Copy(io.MultiWriter(h, pbar), fd); err != nil {
		return err
	}

	copy(f.Hash[:], h.Sum(nil)) // [sha1.Size]byte()

	return nil
}

type Snapshot struct {
	Files []*File
	Tsize int64
	Root  string
}

func (sn Snapshot) String() string {
	return fmt.Sprintf("Snapshot holds %d files, totalling %d bytes", len(sn.Files), sn.Tsize)
}

// Walker collects files found by file.Walk into the Snapshot
func (sn *Snapshot) Walker() WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		// fmt.Printf("%s\n", path)
		if err != nil {
			log.Println(err)
			return nil
		}

		// hidden file or folder; skip file, don't recurse directory
		if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// don't store directory entry
		if info.IsDir() {
			return nil
		}

		// ignore symlinks
		if !info.Mode().IsRegular() {
			return nil
		}

		// excludes zero-size files
		if info.Size() == 0 {
			return nil
		}

		sn.Files = append(sn.Files, &File{Path: path, Size: info.Size()})
		sn.Tsize = sn.Tsize + info.Size()
		return nil
	}
}

func (sn *Snapshot) ComputeHashes() {
	var wg sync.WaitGroup

	hasher := func(fileCh chan *File, pbar *bar.ProgressBar) {
		defer wg.Done()
		for f := range fileCh {
			f.ComputeHash(pbar)
		}
	}

	// Très rapide !
	Walk(sn.Root, sn.Walker())
	fmt.Printf("%s\n", sn)

	pbar := bar.DefaultBytes(
		sn.Tsize,
		"Hashing",
	)

	fileChan := make(chan *File)

	for w := 0; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go hasher(fileChan, pbar)
	}

	go func() {
		for _, f := range sn.Files {
			fileChan <- f
		}
		close(fileChan)
	}()

	wg.Wait()
}

func (sn *Snapshot) SaveTo(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	enc := gob.NewEncoder(w)
	return enc.Encode(sn)
}

func ReadSnapshotFrom(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	w := bufio.NewReader(f)
	enc := gob.NewDecoder(w)
	var sn Snapshot
	err = enc.Decode(&sn)
	if err != nil {
		return nil, err
	}
	return &sn, nil
}

func MustReadSnapshotFrom(path string) *Snapshot {
	s, err := ReadSnapshotFrom(path)
	if err != nil {
		panic(err)
	}
	return s
}
