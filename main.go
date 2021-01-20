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
	"fmt"
	"log"
	"os"

	"github.com/dav-m85/hashsnap/core"
	"github.com/integrii/flaggy"
)

func main() {
	flaggy.SetName("hashsnap")
	flaggy.SetDescription("A snapshot manipulator to ease deduplication across filesystems")

	// Declare variables and their defaults
	// var stringFlagF = "defaultValueF"
	// var intFlagT = 3
	// var boolFlagB bool

	var local string

	// Create the subcommand
	createCmd := flaggy.NewSubcommand("create")
	createCmd.Description = "Create a snapshot file"
	flaggy.AttachSubcommand(createCmd, 1)
	createCmd.AddPositionalValue(&local, "file", 1, true, "Output file")

	dedupCmd := flaggy.NewSubcommand("dedup")
	flaggy.AttachSubcommand(dedupCmd, 1)
	dedupCmd.AddPositionalValue(&local, "file", 1, true, "Input file")

	// Add a flag to the subcommand
	// createCmd.String(&stringFlagF, "t", "testFlag", "A test string flag")

	// add a global bool flag for fun
	//flaggy.Bool(&boolFlagB, "y", "yes", "A sample boolean flag")

	//  the base subcommand to the parser at position 1

	// Declare variables and their defaults

	// Parse the subcommand and all flags
	flaggy.Parse()

	// we can check if a subcommand was used easily
	if createCmd.Used {
		var roots []string
		base, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		roots = append(roots, base)

		// Très rapide !
		snap := core.Snapshot{Root: base}
		snap.ComputeHashes()

		err = snap.SaveTo(local)
		if err != nil {
			log.Fatal("Cannot save:", err)
		}
	}

	if dedupCmd.Used {
		// Très rapide !
		snap, err := core.ReadSnapshotFrom(local)

		if err != nil {
			log.Fatal("Cannot read:", err)
		}

		fmt.Println("%#v", snap)
		snap.Dedup()
	}
	// fmt.Println(flaggy.TrailingArguments[0:])
}
