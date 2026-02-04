package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/core"
	"github.com/trungung/wt/internal/git"
)

var version = "0.0.5"

var fromBase string

var rootCmd = &cobra.Command{
	Use:   "wt [branch]",
	Short: "wt is a branch-centric git worktree helper",
	Long: `A fast, branch-addressable git worktree manager.

Commands:
  wt                 List all worktrees
  wt <branch>        Ensure worktree exists for branch (creates if needed)
  wt cd <branch>     Create worktree and navigate to it (requires shell-setup)
  wt init            Create .wt.config.json
  wt remove <branch> Remove worktree
  wt prune           Remove merged worktrees
  wt health          Check configuration
  wt shell-setup     Generate shell wrapper and completions
`,
	Version: version,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// wt: list worktrees
			worktrees, err := git.ListWorktrees()
			if err != nil {
				return err
			}
			for _, wt := range worktrees {
				fmt.Printf("%s\t%s\n", wt.Branch, wt.Path)
			}
			return nil
		}

		// wt <branch>: ensure worktree
		branch := args[0]
		path, err := core.EnsureWorktree(branch, fromBase)
		if err != nil {
			var rbErr *core.RollbackError
			if errors.As(err, &rbErr) {
				fmt.Fprintf(os.Stderr, "Error: %v\n", rbErr.OriginalErr)
				fmt.Fprintf(os.Stderr, "Rollback status: %s\n", rbErr.RollbackStatus)
				os.Exit(1)
			}
			return err
		}
		fmt.Println(path)
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&fromBase, "from", "f", "", "base branch to create from")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
