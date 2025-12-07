package timesheet

import (
	"github.com/spf13/cobra"
)

// TimesheetCmd is the parent command for timesheet operations
var TimesheetCmd = &cobra.Command{
	Use:   "timesheet",
	Short: "Jira Cloud Timesheet Tracking plugin commands",
	Long: `Commands for logging time via the Jira Cloud Timesheet Tracking plugin.

This uses GraphQL to communicate directly with the Timesheet Forge extension,
bypassing the standard JIRA REST API.

Note: Requires session cookies from an authenticated browser session.`,
}

func init() {
	TimesheetCmd.AddCommand(setupCmd)
	TimesheetCmd.AddCommand(logCmd)
	TimesheetCmd.AddCommand(batchCmd)
}
