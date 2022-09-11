package core

import "os"

// A FileSystem implements access to a collection of named files.
// The elements in a file path are separated by slash ('/', U+002F)
// characters, regardless of host operating system convention.
type FileSystem interface {
	Open(name string) (File, error)
	Remove(name string) error
}

// A File is returned by a FileSystem's Open and Create method and can
// be served by the FileServer implementation.
type File interface {
	Stat() (os.FileInfo, error)
	Readdir(count int) ([]os.FileInfo, error)

	Read([]byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
}
