package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/trungung/wt/internal/core"
)

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

func init() {
	rootCmd.AddCommand(healthCmd)
}
