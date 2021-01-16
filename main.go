package main

// 'gdp' duplicate file finder sketch to test speed of go vs e.g. rmlint

import (
	"bufio"
	"crypto/sha1"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// gd file struct
type File struct {
	path string
	size int64
	sum  hash.Hash // TODO: use pointed instead
}

type Group struct {
	files []*File
	fsize int64
}

const FIRST int64 = 4096 * 4
const MULT int64 = 8

var files = make([]*File, 0)
var groups = make([]*Group, 0)
var mutex = &sync.Mutex{}

func collectFiles(fp string, info os.FileInfo, err error) error {
	// collect files visited by filepath.Walk()
	// excludes hidden files and dirs
	// excludes zero-size files

	if err != nil {
		log.Println(err)
		return nil
	}

	if strings.HasPrefix(info.Name(), ".") {
		// hidden file or folder; skip file, don't recurse directory:
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	if info.IsDir() {
		// don't store directory entry
		return nil
	}

	if !info.Mode().IsRegular() {
		// ignore symlinks
		return nil
	}

	if info.Size() > 0 {
		// collect file
		files = append(files, &File{path: fp, size: info.Size(), sum: sha1.New()})
	}
	return nil
}

func HashFile(file *File, start int64, bytes int64) error {
	// hashes file from 'start' for 'bytes' bytes

	// open input file
	fi, err := os.Open(file.path)
	if err != nil {
		return err
	}
	defer fi.Close()

	// seek to start position
	_, err = fi.Seek(start, 0)
	if err != nil {
		return err
	}

	// use bufio to copy the required number of bytes to the hasher
	written, err := io.CopyN(bufio.NewWriter(file.sum), bufio.NewReader(fi), bytes)
	if err != nil && err != io.EOF {
		return err
	}

	if written != bytes {
		return errors.New("Read unexpected number of bytes")
	}

	return nil
}

func StealHash(file *File) ([sha1.Size]byte, error) {
	// grabs a copy of the sha1 hash result

	var hash [sha1.Size]byte

	sum := file.sum.Sum(nil)
	for i := range sum {
		hash[i] = sum[i]
	}

	return hash, nil
}

// BySize implements sort.Interface for []File based on
// the size field.
type BySize []*File

func (f BySize) Len() int           { return len(f) }
func (f BySize) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f BySize) Less(i, j int) bool { return f[i].size < f[j].size }

func groupBySize() error {
	// groups files[] by size and stores groups with size > 1 in groups[] slice

	// first sort into size order
	sort.Sort(BySize(files))

	var size int64 = -1 // size of current group
	var group *Group    // = new(Group)

	for _, file := range files {
		if file.size != size {
			// start new group
			group = new(Group)
			size = file.size
			group.fsize = size
		}
		group.files = append(group.files, file)
		if len(group.files) == 2 {
			groups = append(groups, group)
		}

	}

	return nil
}

func printGroup(group *Group) {
	// outputs file paths from a Group
	mutex.Lock()
	for _, file := range group.files {
		fmt.Println(file.path)
	}
	fmt.Println()
	mutex.Unlock()
}

func hashGroup(group *Group, start int64, end int64, wg *sync.WaitGroup) {
	// incrementally hash files in group, split into
	// matching subgroups and recursively hash those
	defer wg.Done()

	// don't hash past end of file
	if end > group.fsize {
		end = group.fsize
	}

	// iterate over group.files, mapping by [partial] hash
	var matches = make(map[[sha1.Size]byte]*Group)
	for _, file := range group.files {

		// do the hashing work
		err := HashFile(file, start, end-start)
		if err != nil {
			log.Println(err)
			continue
		}
		hash, err := StealHash(file)
		if err != nil {
			log.Println(err)
			continue
		}

		// check for matching hash
		match, ok := matches[hash]
		if ok {
			// matching group found; add this file to existing group
			match.files = append(match.files, file)
		} else {
			// create new group in map
			matches[hash] = &Group{[]*File{file}, file.size}
		}

	}

	// find all matched groups with 2 or more files
	for _, grp := range matches {

		if len(grp.files) < 2 {
			continue
		}

		if end == group.fsize {
			// duplicate file set (or hash collision potentially)
			printGroup(grp)
		} else {
			// partially matched files; recurse next hash increment
			wg.Add(1)
			hashGroup(grp, end, end*MULT, wg)
		}
	}
}

func findMatches() error {
	// find matching file sets in groups[]
	var wg sync.WaitGroup
	for _, group := range groups {
		wg.Add(1)
		hashGroup(group, 0, FIRST, &wg)
	}
	wg.Wait()
	return nil
}

func main() {

	if len(os.Args) > 3 {
		fmt.Printf("USAGE : %s target_directory [threads]\n", os.Args[0])
		os.Exit(0)
	}

	// get the target directory
	dir, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// set the default max number of threads
	if len(os.Args) >= 3 {
		// set the user-input max number of threads
		threads, err := strconv.Atoi(os.Args[2])
		if err == nil {
			runtime.GOMAXPROCS(threads)
		}
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// collect files under dir
	err = filepath.Walk(dir, collectFiles)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// group files by size
	err = groupBySize()

	// find matches and output results
	err = findMatches()

}
