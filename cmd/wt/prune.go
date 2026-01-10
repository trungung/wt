package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/core"
)

var pruneForce bool
var pruneDryRun bool
var pruneFetch bool

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "remove worktrees whose branches are merged into default branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := core.PruneOptions{
			DryRun: pruneDryRun,
			Force:  pruneForce,
			Fetch:  pruneFetch,
		}

		count, candidates, err := core.PruneWorktrees(opts)
		if err != nil {
			return err
		}

		if pruneDryRun {
			if len(candidates) == 0 {
				fmt.Println("No worktrees to prune.")
				return nil
			}
			fmt.Println("Candidates for pruning:")
			for _, b := range candidates {
				fmt.Printf("  %s\n", b)
			}
			fmt.Printf("\nTotal candidates: %d (run without --dry-run to prune)\n", len(candidates))
		} else {
			fmt.Printf("Pruned %d worktrees.\n", count)
		}
		return nil
	},
}

func init() {
	pruneCmd.Flags().BoolVarP(&pruneForce, "force", "f", false, "force removal even if dirty")
	pruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "show what would be removed")
	pruneCmd.Flags().BoolVar(&pruneFetch, "fetch", false, "run git fetch --prune first")
	rootCmd.AddCommand(pruneCmd)
}
