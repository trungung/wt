package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/core"
)

var execCmd = &cobra.Command{
	Use:   "exec <branch> -- <command...>",
	Short: "run an arbitrary command inside a branch's worktree",
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

func init() {
	rootCmd.AddCommand(execCmd)
}
