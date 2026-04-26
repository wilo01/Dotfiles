package standup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/spf13/cobra"
)

var (
	notifyDays    int
	notifyTimeout int
)

func init() {
	notifyCmd.Flags().IntVar(&notifyDays, "days", 0, "Limit to last N days (0=smart default: Mon=3, else=1)")
	notifyCmd.Flags().IntVar(&notifyTimeout, "timeout", 30000, "Notification expire timeout in ms (0=persistent)")
}

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Desktop notification with standup preview",
	Long: `Pop up a desktop notification (notify-send) showing how many entries
will be in the next standup publish. Use this on a T-30 timer to remind
yourself to review/edit entries before auto-publish fires.

Examples:
  hlp standup notify           # Smart-default range (Mon=3 days, else=1)
  hlp standup notify --days 3  # Last 3 days`,
	RunE: runNotify,
}

func runNotify(cmd *cobra.Command, args []string) error {
	profile, _ := config.GetActiveProfile()
	csvPath := batch.DefaultCSVPathForProfile(profile)
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return fmt.Errorf("worklog CSV not found: %s", csvPath)
	}

	entries, err := batch.ParseCSV(csvPath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	days := smartDefaultDays(notifyDays)

	// Apply the same pipeline as publish so the count and preview match what
	// will actually get sent to the sheet.
	sortEntriesNewestFirst(entries)
	entries = applyDayWindow(entries, days)
	entries = filterStandupEntries(entries, getIgnoredTickets())

	count := len(entries)
	keys := uniqueIssueKeys(entries, 8)

	word := "entries"
	if count == 1 {
		word = "entry"
	}
	title := fmt.Sprintf("Standup digest: %d %s ready", count, word)
	body := fmt.Sprintf("Window: last %d day(s). Run `hlp standup publish` to publish now.", days)
	if len(keys) > 0 {
		body += "\n" + strings.Join(keys, ", ")
	}

	notifySend := exec.Command("notify-send",
		"--app-name=hlp",
		"--icon=appointment-soon",
		fmt.Sprintf("--expire-time=%d", notifyTimeout),
		"--urgency=normal",
		title,
		body,
	)
	if err := notifySend.Run(); err != nil {
		return fmt.Errorf("notify-send: %w", err)
	}
	fmt.Printf("notified: %s\n", title)
	return nil
}

// uniqueIssueKeys returns up to max distinct IssueKeys preserving order.
func uniqueIssueKeys(entries []batch.Entry, max int) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, max)
	for _, e := range entries {
		if seen[e.IssueKey] {
			continue
		}
		seen[e.IssueKey] = true
		out = append(out, e.IssueKey)
		if len(out) >= max {
			break
		}
	}
	return out
}

