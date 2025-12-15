package git

import (
	"github.com/spf13/cobra"
)

// GitCmd is the parent command for git utilities
var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git utilities",
	Long:  `Commands for git operations, branch comparison, and repository analysis.`,
}

func init() {
	GitCmd.AddCommand(compareCmd)
}
