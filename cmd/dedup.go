package cmd

import (
	"fmt"
	"log"

	"github.com/dav-m85/hashsnap/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dedupCmd)
}

var dedupCmd = &cobra.Command{
	Use:   "dedup <snapfile>",
	Short: "...",
	// Long:  `Snapshot current working directory by default`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Très rapide !
		snap, err := core.ReadSnapshotFrom(args[0])

		if err != nil {
			log.Fatal("Cannot read:", err)
		}

		fmt.Println("%#v", snap)

		// check for matching hash
		// matches := make(map[[sha1.Size]byte]*Group)

		// for _, f := range snap.files {
		// 	match, ok := matches[f.hash]
		// 	if ok {
		// 		// matching group found; add this file to existing group
		// 		match.files = append(match.files, f)
		// 	} else {
		// 		// create new group in map
		// 		matches[f.hash] = &Group{[]*File{f}, f.size}
		// 	}
		// }

		// for _, group := range matches {
		// 	if len(group.files) > 1 {
		// 		fmt.Println("Duplicates\n", group.files)
		// 	}
		// }
	},
}
