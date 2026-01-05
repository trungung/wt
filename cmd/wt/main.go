package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/config"
	"github.com/trungung/wt/internal/core"
	"github.com/trungung/wt/internal/git"
	"github.com/trungung/wt/internal/ui"
	"github.com/yarlson/tap"
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

			var branches []string
			for i, wt := range worktrees {
				if i == 0 {
					continue // Skip main worktree
				}
				branches = append(branches, wt.Branch)
			}

			branch = ui.PromptAutocomplete("Select a worktree to remove (Tab to complete)", func(input string) []string {
				var filtered []string
				for _, b := range branches {
					if strings.Contains(strings.ToLower(b), strings.ToLower(input)) {
						filtered = append(filtered, b)
					}
				}
				return filtered
			})

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
			return ui.PromptBool(msg, false)
		}

		return core.RemoveWorktree(branch, forceRemove, confirmFn)
	},
}

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

var initYes bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create .wt.config.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := git.GetRepoRoot()
		if err != nil {
			return err
		}

		configPath := config.GetConfigPath(root)
		if _, err := os.Stat(configPath); err == nil {
			fmt.Println(configPath)
			return nil
		}

		detected, err := git.GetDefaultBranch()
		if err != nil && initYes {
			return fmt.Errorf("could not auto-detect default branch for --yes: %w", err)
		}

		cfg := &config.Config{
			DefaultBranch:            detected,
			WorktreePathTemplate:     "$REPO_PATH.wt",
			WorktreeCopyPatterns:     []string{},
			PostCreateCmd:            []string{},
			DeleteBranchWithWorktree: false,
		}

		if !initYes {
			tap.Intro("Initializing .wt.config.json")

			if err != nil {
				fmt.Printf("Warning: %v\n", err)
				cfg.DefaultBranch = ui.PromptRequired("Default branch")
			} else {
				cfg.DefaultBranch = ui.Prompt("Default branch", cfg.DefaultBranch)
			}
			cfg.WorktreePathTemplate = ui.Prompt("Worktree path template", cfg.WorktreePathTemplate)
			cfg.WorktreeCopyPatterns = ui.PromptList("Worktree copy patterns")
			cfg.PostCreateCmd = ui.PromptList("Post-create commands")
			cfg.DeleteBranchWithWorktree = ui.PromptBool("Delete branch with worktree", cfg.DeleteBranchWithWorktree)

			tap.Outro("Configuration generated")
		}

		if err := cfg.Write(root); err != nil {
			return err
		}

		fmt.Println(configPath)
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&fromBase, "from", "f", "", "base branch to create from")
	rootCmd.AddCommand(execCmd)
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "r", false, "force removal even if dirty")
	rootCmd.AddCommand(removeCmd)

	pruneCmd.Flags().BoolVarP(&pruneForce, "force", "f", false, "force removal even if dirty")
	pruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "show what would be removed")
	pruneCmd.Flags().BoolVar(&pruneFetch, "fetch", false, "run git fetch --prune first")
	rootCmd.AddCommand(pruneCmd)

	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "write defaults without prompts")
	rootCmd.AddCommand(initCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
