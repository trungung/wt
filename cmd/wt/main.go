package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/config"
	"github.com/trungung/wt/internal/core"
	"github.com/trungung/wt/internal/git"
	"github.com/trungung/wt/internal/ui"
)

//go:embed _wt
var zshCompletionScript string

var version = "0.0.1"

var fromBase string

var rootCmd = &cobra.Command{
	Use:     "wt [branch]",
	Short:   "wt is a branch-centric git worktree helper",
	Long:    `A fast, branch-addressable git worktree manager.`,
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
			var copyPatterns string
			var postCmds string

			fmt.Println("Initializing .wt.config.json")

			err := huh.NewInput().
				Title("Default branch").
				Value(&cfg.DefaultBranch).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("this field is required")
					}
					return nil
				}).
				Run()
			if err != nil {
				return err
			}
			fmt.Printf("Default branch: %s\n\n", cfg.DefaultBranch)

			err = huh.NewInput().
				Title("Worktree path template").
				Value(&cfg.WorktreePathTemplate).
				Run()
			if err != nil {
				return err
			}
			fmt.Printf("Path template: %s\n\n", cfg.WorktreePathTemplate)

			err = huh.NewInput().
				Title("Worktree copy patterns (comma separated)").
				Placeholder(strings.Join(cfg.WorktreeCopyPatterns, ", ")).
				Value(&copyPatterns).
				Run()
			if err != nil {
				return err
			}
			cfg.WorktreeCopyPatterns = splitPromptList(copyPatterns)
			fmt.Printf("Copy patterns: [%s]\n\n", strings.Join(cfg.WorktreeCopyPatterns, ", "))

			err = huh.NewInput().
				Title("Post-create commands (comma separated)").
				Placeholder(strings.Join(cfg.PostCreateCmd, ", ")).
				Value(&postCmds).
				Run()
			if err != nil {
				return err
			}
			cfg.PostCreateCmd = splitPromptList(postCmds)
			fmt.Printf("Post-create commands: [%s]\n\n", strings.Join(cfg.PostCreateCmd, ", "))

			err = huh.NewConfirm().
				Title("Delete branch with worktree?").
				Value(&cfg.DeleteBranchWithWorktree).
				Run()
			if err != nil {
				return err
			}
			fmt.Printf("Delete branch with worktree: %t\n\n", cfg.DeleteBranchWithWorktree)
		}

		if err := cfg.Write(root); err != nil {
			return err
		}

		fmt.Println(configPath)
		return nil
	},
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "check project health",
	RunE: func(cmd *cobra.Command, args []string) error {
		checks, hasError := core.RunHealthCheck()
		for _, c := range checks {
			fmt.Printf("[%s] %s: %s\n", c.Level, c.Name, c.Message)
		}
		if hasError {
			os.Exit(1)
		}
		return nil
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "generate completion script for the specified shell",
	Long: `To load completions:

Zsh:
  # Add to your ~/.zshrc:
  source <(wt completion zsh)

  # Or, if completions are not loading, you may need to add:
  fpath=(~/.zsh/completions $fpath)
  autoload -Uz _wt
  compdef _wt wt
`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"zsh"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "zsh":
			fmt.Print(zshCompletionScript)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
		return nil
	},
}

func splitPromptList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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

	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(completionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
