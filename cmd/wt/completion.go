package main

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed _wt
var zshCompletionScript string

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

func init() {
	rootCmd.AddCommand(completionCmd)
}
