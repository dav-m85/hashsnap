package ncore

import (
	"fmt"
	"os"
	"path/filepath"
)

/*
 https://en.wikipedia.org/wiki/Breadth-first_search
1  procedure BFS(G, root) is
2      let Q be a queue
3      label root as discovered
4      Q.enqueue(root)
5      while Q is not empty do
6          v := Q.dequeue()
7          if v is the goal then
8              return v
9          for all edges from v to w in G.adjacentEdges(v) do
10              if w is not labeled as discovered then
11                  label w as discovered
12                  Q.enqueue(w)
*/
func walker(nodeChan chan<- *Node, root string) {
	var q []*Node
	var id uint64 = 1
	var rootNode = &Node{
		ID:       id,
		ParentID: 0, // No parent
		parent:   nil,
		Path:     root,
	}

	q = append(q, rootNode)
	for len(q) > 0 {
		node := q[0]
		q = q[1:]

		// Send to discovered
		nodeChan <- node

		// Walk deeper
		if node.Path != "" { // isDir
			names, err := readDirNames(node.Path)
			if err != nil {
				panic(err)
			}
			for _, name := range names {
				filename := filepath.Join(node.Path, name)
				info, err := lstat(filename)
				if err != nil {
					fmt.Println(err)
					continue
				}

				// For dirs, ignore symlinks and zero-size files
				if !info.IsDir() {
					if !info.Mode().IsRegular() {
						continue
					}
					if info.Size() == 0 {
						continue
					}
				}

				id++

				child := &Node{
					ID:       id,
					parent:   node,
					ParentID: node.ID,
				}

				if info.IsDir() {
					child.Path = filename
				} else {
					child.Name = name
					child.Size = uint64(info.Size())
				}

				q = append(q, child)
			}
		}
	}

	close(nodeChan)
}

var lstat = os.Lstat // for testing

// readDirNames reads the directory named by dirname and returns
// a list of directory entries.
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return names, nil
}
