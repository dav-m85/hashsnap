package cmd

import (
	"log"
	"os"

	"github.com/dav-m85/hashsnap/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create <snapfile> [paths...]",
	Short: "Create a snapshot file (this is the long part)",
	Long:  `Snapshot current working directory by default`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var roots []string
		if len(args) == 1 {
			base, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			roots = append(roots, base)
		} else {
			roots = args[1:]
		}

		// Très rapide !
		snap := core.Snapshot{Root: roots[0]}
		snap.ComputeHashes()

		err := snap.SaveTo(args[0])
		if err != nil {
			log.Fatal("Cannot save:", err)
		}
	},
}
