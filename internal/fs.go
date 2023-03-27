package internal

import (
	"io/fs"
	"os"
)

type OS struct{}

func (OS) Open(name string) (fs.File, error) {
	return os.Open(name)
}
func (OS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (OS) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

// FS Vritual filesystem, can be chrooted kindof with os.DirFS(".")
var FS fs.FS = OS{}
