package dev

import (
	"github.com/spf13/cobra"
)

// DevCmd is the parent command for development utilities
var DevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development utilities",
	Long:  `Development utility commands for text formatting, git operations, and file naming.`,
}

func init() {
	// Add subcommands
	DevCmd.AddCommand(branchCmd)
	DevCmd.AddCommand(stashCmd)
	DevCmd.AddCommand(prCmd)
	DevCmd.AddCommand(dashCmd)
	DevCmd.AddCommand(filenameCmd)
}

// RegisterTopLevel registers dev commands at the top level of the CLI
func RegisterTopLevel(rootCmd *cobra.Command) {
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(stashCmd)
	rootCmd.AddCommand(prCmd)
	rootCmd.AddCommand(dashCmd)
	rootCmd.AddCommand(filenameCmd)
}
