package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/core"
	"github.com/trungung/wt/internal/git"
	"github.com/trungung/wt/internal/ui"
)

var forceRemove bool

var removeCmd = &cobra.Command{
	Use:   "remove [branch]",
	Short: "remove a worktree and optionally delete its branch",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var branch string
		if len(args) == 0 {
			// Interactive selection
			worktrees, err := git.ListWorktrees()
			if err != nil {
				return err
			}
			if len(worktrees) <= 1 {
				return fmt.Errorf("no other worktrees to remove")
			}

			var branches []string
			for i, wt := range worktrees {
				if i == 0 {
					continue // Skip main worktree
				}
				branches = append(branches, wt.Branch)
			}

			err = huh.NewSelect[string]().
				Title("Select a worktree to remove").
				Options(huh.NewOptions(branches...)...).
				Value(&branch).
				Run()
			if err != nil {
				return err
			}

			if branch == "" {
				return fmt.Errorf("no worktree selected")
			}

			// Validate that the returned branch actually exists as a worktree
			found := false
			for _, b := range branches {
				if b == branch {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid worktree selected: %s", branch)
			}
		} else {
			branch = args[0]
		}

		confirmFn := func(msg string) bool {
			result, err := ui.PromptBoolWithError(msg, false)
			if err != nil {
				// If prompt fails (e.g., Ctrl+C), treat as declined
				return false
			}
			return result
		}

		return core.RemoveWorktree(branch, forceRemove, confirmFn)
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "r", false, "force removal even if dirty")
	rootCmd.AddCommand(removeCmd)
}
