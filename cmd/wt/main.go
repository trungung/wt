package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/core"
	"github.com/trungung/wt/internal/git"
)

var fromBase string

var rootCmd = &cobra.Command{
	Use:   "wt [branch]",
	Short: "wt is a branch-centric git worktree helper",
	Long:  `A fast, branch-addressable git worktree manager.`,
	Args:  cobra.MaximumNArgs(1),
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
			return err
		}
		fmt.Println(path)
		return nil
	},
}

var execCmd = &cobra.Command{
	Use:   "exec <branch> -- <command...>",
	Short: "run an arbitrary command inside a branch’s worktree",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]
		// Find the '--' delimiter
		dashIndex := -1
		for i, arg := range os.Args {
			if arg == "--" {
				dashIndex = i
				break
			}
		}

		if dashIndex == -1 || dashIndex+1 >= len(os.Args) {
			return fmt.Errorf("missing command after --")
		}

		commandArgs := os.Args[dashIndex+1:]

		path, err := core.FindWorktree(branch)
		if err != nil {
			return err
		}

		c := exec.Command(commandArgs[0], commandArgs[1:]...)
		c.Dir = path
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			return err
		}
		return nil
	},
}

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

			fmt.Println("Select a worktree to remove:")
			for i, wt := range worktrees {
				if i == 0 {
					continue // Skip main worktree
				}
				fmt.Printf("[%d] %s (%s)\n", i, wt.Branch, wt.Path)
			}
			fmt.Print("Selection: ")
			var choice int
			_, err = fmt.Scanf("%d", &choice)
			if err != nil || choice < 1 || choice >= len(worktrees) {
				return fmt.Errorf("invalid selection")
			}
			branch = worktrees[choice].Branch
		} else {
			branch = args[0]
		}

		confirmFn := func(msg string) bool {
			fmt.Printf("%s [y/N]: ", msg)
			var response string
			fmt.Scanf("%s", &response)
			return strings.ToLower(response) == "y"
		}

		return core.RemoveWorktree(branch, forceRemove, confirmFn)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&fromBase, "from", "f", "", "base branch to create from")
	rootCmd.AddCommand(execCmd)
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "r", false, "force removal even if dirty")
	rootCmd.AddCommand(removeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
