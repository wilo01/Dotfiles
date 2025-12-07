package jira

import (
	"github.com/spf13/cobra"
)

// JiraCmd is the parent command for JIRA operations
var JiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "JIRA workflow commands",
	Long:  `Commands for JIRA ticket management, time logging, and workflow automation.`,
}

func init() {
	JiraCmd.AddCommand(configCmd)
	JiraCmd.AddCommand(logCmd)
	JiraCmd.AddCommand(statusCmd)
}
