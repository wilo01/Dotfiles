// Package standup provides commands for daily standup report generation
package standup

import (
	"time"

	"github.com/spf13/cobra"
)

// smartDefaultDays returns the day window when caller passed 0:
// Monday=3 (covers Friday), other weekdays=1 (yesterday).
// Pre-set values (>0) pass through unchanged.
func smartDefaultDays(d int) int {
	if d > 0 {
		return d
	}
	if time.Now().Weekday() == time.Monday {
		return 3
	}
	return 1
}

// StandupCmd is the parent command for standup operations
var StandupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Daily standup report commands",
	Long:  `Commands for generating and managing daily standup reports from worklog data.`,
}

func init() {
	StandupCmd.AddCommand(generateCmd)
	StandupCmd.AddCommand(showCmd)
	StandupCmd.AddCommand(publishCmd)
	StandupCmd.AddCommand(notifyCmd)
}
