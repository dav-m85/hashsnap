package hashsnap

import (
	"io/fs"
	"os"
)

var FS fs.FS = os.DirFS(".")
