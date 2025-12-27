// Package standup provides commands for daily standup report generation
package standup

import (
	"github.com/spf13/cobra"
)

// StandupCmd is the parent command for standup operations
var StandupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Daily standup report commands",
	Long:  `Commands for generating and managing daily standup reports from worklog data.`,
}

func init() {
	StandupCmd.AddCommand(generateCmd)
	StandupCmd.AddCommand(showCmd)
}
