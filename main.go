// package main

// import (
// 	"crypto/sha1"
// 	"flag"
// 	"fmt"
// 	"os"
// 	"runtime"
// 	"sync"
// 	"time"

// 	"github.com/dav-m85/hashsnap/file"
// 	bar "github.com/schollz/progressbar/v3"
// )

// 2min pour 6Go SSD avec mon quad core

// var mutex = &sync.Mutex{}

package main

import (
	"github.com/dav-m85/hashsnap/cmd"
)

func main() {
	cmd.Execute()
}
