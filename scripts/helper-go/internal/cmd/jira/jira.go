package jira

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/spf13/cobra"
)

// JiraCmd is the parent command for JIRA operations
var JiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "JIRA workflow commands",
	Long:  `Commands for JIRA ticket management, time logging, and workflow automation.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		autoSyncIfNeeded(cmd)
	},
}

func init() {
	JiraCmd.AddCommand(addCmd)
	JiraCmd.AddCommand(configCmd)
	JiraCmd.AddCommand(logCmd)
	JiraCmd.AddCommand(statusCmd)
	JiraCmd.AddCommand(profileCmd)
}

// autoSyncIfNeeded checks conditions and runs auto-sync if appropriate
func autoSyncIfNeeded(cmd *cobra.Command) {
	// Skip for add command - it handles its own sprint sync
	if cmd.Name() == "add" {
		return
	}

	// Skip if --quiet flag passed (cron mode)
	if IsQuietMode(cmd) {
		return
	}

	// Skip if not work hours
	if !config.IsWorkHours() {
		return
	}

	// Skip if already synced today
	if config.SyncedToday() {
		return
	}

	// Check if JIRA is configured before attempting sync
	profile, err := config.GetActiveProfile()
	if err != nil || profile == nil || profile.BaseURL == "" {
		// Silently skip - let the actual command handle this error
		return
	}

	fmt.Println("Auto-syncing sprint tickets...")
	err = RunAutoSync(false)

	if err != nil {
		// Ask user whether to continue
		fmt.Printf("Auto-sync failed: %v\n", err)
		if !promptContinue() {
			os.Exit(1)
		}
	}
	fmt.Println()
}

// promptContinue asks user to continue or abort after sync failure
func promptContinue() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Continue with command? [y/N]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}
