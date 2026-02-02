package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/config"
	"github.com/trungung/wt/internal/git"
)

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

		if !initYes {
			fmt.Println()
			fmt.Println("💡 Tip: Enable seamless navigation with:")
			fmt.Println("   wt shell-setup zsh >> ~/.zshrc && source ~/.zshrc")
			fmt.Println()
		}
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "write defaults without prompts")
	rootCmd.AddCommand(initCmd)
}
