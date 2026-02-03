package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/git"
)

var completionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "generate completion script for the specified shell",
	Long: `Generate shell completion scripts.

To load completions (standalone, without shell-setup):

Zsh:
  source <(wt completion zsh)

Bash:
  source <(wt completion bash)

Fish:
  wt completion fish | source

Note: If you use 'eval "$(wt shell-setup)"', completions are already included.
`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"zsh", "bash", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "bash":
			return rootCmd.GenBashCompletionV2(os.Stdout, true)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return cmd.Help()
		}
	},
}

// completeBranches returns all local branches for shell completion.
func completeBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	branches, err := git.ListLocalBranches()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return branches, cobra.ShellCompDirectiveNoFileComp
}

// completeWorktreeBranches returns branches that have active worktrees.
func completeWorktreeBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var branches []string
	for i, wt := range worktrees {
		if i == 0 {
			continue // Skip main worktree
		}
		branches = append(branches, wt.Branch)
	}
	return branches, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Register dynamic completions for the root command (wt <branch>)
	rootCmd.ValidArgsFunction = completeBranches

	// Register dynamic completions for remove command
	removeCmd.ValidArgsFunction = completeWorktreeBranches

	// Register completion for --from flag on root command
	_ = rootCmd.RegisterFlagCompletionFunc("from", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		branches, err := git.ListLocalBranches()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return branches, cobra.ShellCompDirectiveNoFileComp
	})
}
