package jira

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var addNoSync bool
var addQuiet bool
var addDryRun bool
var addToSubtask bool

var addCmd = &cobra.Command{
	Use:   "add [ticket-key...]",
	Short: "Add JIRA ticket(s) to worklogs.csv as draft",
	Long: `Add one or more JIRA tickets to worklogs.csv with DRAFT status.

The tickets will be prepared for time logging but not processed by --batch
until you fill in the TimeSpent and change the status from DRAFT to empty.

Sprint sync happens by default on every invocation. Use --no-sync to skip it.

Examples:
  hlp jira add                       # Sync sprint tickets only
  hlp jira add VIS-1234              # Sync sprint + add VIS-1234
  hlp jira add --no-sync VIS-1234    # Add VIS-1234 only (skip sprint sync)
  hlp jira add --dry-run             # Preview sprint sync`,
	Run: runAdd,
}

func init() {
	addCmd.Flags().BoolVar(&addNoSync, "no-sync", false, "Skip sprint sync (only process manual ticket args)")
	addCmd.Flags().BoolVarP(&addQuiet, "quiet", "q", false, "Suppress output unless errors (for cron)")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "Preview changes without modifying files")
	addCmd.Flags().BoolVar(&addToSubtask, "to-subtask", false, "Log time against the subtask instead of the parent (overrides preferences.log_to_subtask)")

	// Deprecated: --sync is now the default behavior
	var addSyncDeprecated bool
	addCmd.Flags().BoolVarP(&addSyncDeprecated, "sync", "s", false, "Deprecated: sprint sync is now the default")
	addCmd.Flags().MarkDeprecated("sync", "sprint sync is now the default; use --no-sync to skip")
}

// ticketKeyPattern validates JIRA ticket format (PROJECT-NUMBER)
var ticketKeyPattern = regexp.MustCompile(`^[A-Z]+-\d+$`)

func runAdd(cmd *cobra.Command, args []string) {
	// Load daily tickets from config
	cfg, cfgErr := config.Load()
	var dailyTickets []string
	logToSubtask := false
	if cfgErr != nil {
		if !addQuiet {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not load config: %v (using defaults)", cfgErr)))
		}
	} else {
		dailyTickets = cfg.Preferences.DailyTickets
		logToSubtask = cfg.Preferences.LogToSubtask
	}
	// CLI flag wins over config / env
	if cmd.Flags().Changed("to-subtask") {
		logToSubtask = addToSubtask
	}

	// --no-sync without ticket keys or daily tickets is a no-op
	if addNoSync && len(args) == 0 && len(dailyTickets) == 0 {
		fmt.Println(ui.Error("--no-sync requires ticket keys (or configure daily_tickets)"))
		fmt.Println(ui.Muted.Render("Examples:"))
		fmt.Println(ui.Muted.Render("  hlp jira add                       # Sync sprint"))
		fmt.Println(ui.Muted.Render("  hlp jira add --no-sync VIS-1234    # Add without sync"))
		return
	}

	// Show dry-run header
	if addDryRun {
		fmt.Println(ui.Warning("[DRY-RUN] Preview mode - no changes will be made"))
		fmt.Println()
	}

	// Get current profile for CSV path
	profile, err := config.GetActiveProfile()
	if err != nil && !addQuiet {
		fmt.Println(ui.Warning(fmt.Sprintf("Could not load profile: %v", err)))
	}
	csvPath := batch.DefaultCSVPathForProfile(profile)

	// Get JIRA client once for all tickets
	client, err := getJiraClient()
	if err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// Parse existing entries for duplicate checking
	todayStr := time.Now().Format("02.01.2006")
	entries, parseErr := batch.ParseCSV(csvPath)
	if parseErr != nil && !errors.Is(parseErr, os.ErrNotExist) {
		fmt.Println(ui.Error(fmt.Sprintf("Worklog CSV is unreadable, aborting so it isn't overwritten: %v", parseErr)))
		fmt.Println(ui.Muted.Render("Fix the CSV (hlp jira edit) and re-run."))
		return
	}

	// Build list of tickets to process
	var ticketKeys []string

	// Fetch sprint tickets (default behavior, skip with --no-sync)
	if !addNoSync {
		if !addQuiet {
			fmt.Println("Fetching assigned tickets...")
		}
		sprintTickets, sprintErr := fetchAssignedTickets(client)
		if sprintErr != nil {
			fmt.Println(ui.Error("Failed to fetch assigned tickets: " + sprintErr.Error()))
			return
		}
		if !addQuiet {
			fmt.Printf("Found %s assigned tickets\n", ui.Success.Render(fmt.Sprintf("%d", len(sprintTickets))))
			fmt.Println()
		}

		for _, t := range sprintTickets {
			ticketKeys = append(ticketKeys, t.Key)
		}

		// Update sync timestamp so other commands (log, status) skip redundant auto-sync
		if syncErr := config.UpdateLastSyncTime(); syncErr != nil && !addQuiet {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not update sync timestamp: %v", syncErr)))
		}
	}

	// Add manual ticket keys
	for _, arg := range args {
		ticketKeys = append(ticketKeys, strings.ToUpper(strings.TrimSpace(arg)))
	}

	// Append daily tickets from config (normalize and validate)
	for _, dt := range dailyTickets {
		upper := strings.ToUpper(strings.TrimSpace(dt))
		if ticketKeyPattern.MatchString(upper) {
			ticketKeys = append(ticketKeys, upper)
		}
	}

	// Deduplicate ticket keys (sprint keys first, case-insensitive)
	ticketKeys = deduplicateKeys(ticketKeys)

	// Track results
	added := 0
	updated := 0
	skipped := 0

	// Process each ticket
	for _, ticketKey := range ticketKeys {
		// Validate ticket key format
		if !ticketKeyPattern.MatchString(ticketKey) {
			if !addQuiet {
				fmt.Printf("%s %s - invalid format (expected PROJECT-123)\n",
					ui.ErrorText.Render("✗"),
					ui.Primary.Render(ticketKey))
			}
			skipped++
			continue
		}

		// Check if ticket exists for today
		// Match on IssueKey (parent-only) OR SubtaskKey to handle both cases
		var existingEntry *batch.Entry
		for i, e := range entries {
			entryDate := strings.Split(e.Date, " ")[0]
			if entryDate != todayStr {
				continue
			}
			if (strings.EqualFold(e.IssueKey, ticketKey) && e.SubtaskKey == "") ||
				strings.EqualFold(e.SubtaskKey, ticketKey) {
				existingEntry = &entries[i]
				break
			}
		}

		// Skip entries that already exist for today (any status except DRAFT which gets updated below)
		if existingEntry != nil && existingEntry.Status != batch.StatusDraft {
			if !addQuiet {
				displayKey := existingEntry.IssueKey
				if existingEntry.SubtaskKey != "" {
					displayKey = existingEntry.IssueKey + " > " + existingEntry.SubtaskKey
				}
				statusDisplay := existingEntry.Status
				if statusDisplay == "" {
					statusDisplay = "PENDING"
				}
				fmt.Printf("%s %s - already exists today (%s %s)\n",
					ui.Muted.Render("⊘"),
					ui.Muted.Render(displayKey),
					ui.Muted.Render(existingEntry.TimeSpent),
					ui.Muted.Render(statusDisplay))
			}
			skipped++
			continue
		}

		// Fetch ticket info (needed for new entries and DRAFT updates)
		ticket, err := client.GetTicket(ticketKey)
		if err != nil {
			if !addQuiet {
				fmt.Printf("%s %s - %s\n",
					ui.ErrorText.Render("✗"),
					ui.Primary.Render(ticketKey),
					ui.Muted.Render(err.Error()))
			}
			skipped++
			continue
		}

		// Resolve subtask relationship
		issueKey := ticket.Key
		subtaskKey := ""
		issueType := ticket.IssueType
		description := ticket.Summary

		if ticket.IsSubtask && ticket.ParentKey != "" {
			parentTicket, err := client.GetTicket(ticket.ParentKey)
			if err != nil {
				fmt.Println(ui.Warning(fmt.Sprintf("Could not fetch parent %s: %v", ticket.ParentKey, err)))
			} else {
				subtaskKey = ticket.Key
				issueKey = parentTicket.Key
				issueType = parentTicket.IssueType
				description = parentTicket.Summary + " > " + ticket.Summary
			}
		}

		displayKey := issueKey
		if subtaskKey != "" {
			displayKey = issueKey + " > " + subtaskKey
		}

		// Skip parent-only entries when parent already has a subtask entry today
		if subtaskKey == "" {
			if trackedVia := findSubtaskKey(entries, issueKey, todayStr); trackedVia != "" {
				if !addQuiet {
					fmt.Printf("%s %s - already tracked via subtask today\n",
						ui.Muted.Render("⊘"),
						ui.Muted.Render(issueKey+" > "+trackedVia))
				}
				skipped++
				continue
			}
		}

		// Handle DRAFT update
		if existingEntry != nil && existingEntry.Status == batch.StatusDraft {
			if !addDryRun {
				err = batch.UpdateEntryDescription(csvPath, existingEntry.RowNumber, description)
				if err != nil {
					if !addQuiet {
						fmt.Printf("%s %s - %s\n",
							ui.ErrorText.Render("✗"),
							ui.Primary.Render(ticketKey),
							ui.Muted.Render(err.Error()))
					}
					skipped++
					continue
				}
			}

			if !addQuiet {
				if addDryRun {
					fmt.Printf("%s %s %s - %s %s\n",
						ui.Primary.Render("↻"),
						ui.Primary.Render(displayKey),
						ui.Muted.Render("["+issueType+"]"),
						ui.Muted.Render(truncateString(description, 40)),
						ui.Muted.Render("(would update)"))
				} else {
					fmt.Printf("%s %s %s - %s %s\n",
						ui.Primary.Render("↻"),
						ui.Primary.Render(displayKey),
						ui.Muted.Render("["+issueType+"]"),
						ui.Muted.Render(truncateString(description, 40)),
						ui.Muted.Render("(updated)"))
				}
			}
			updated++

			if !addDryRun {
				existingEntry.Description = description
			}
			continue
		}

		// Clean up redundant parent-only entry BEFORE prepend (row numbers are still valid)
		if !addDryRun && subtaskKey != "" {
			removeParentOnlyEntry(csvPath, entries, issueKey, todayStr, addQuiet)
		}

		// Prepend to CSV with DRAFT status (skip in dry-run)
		if !addDryRun {
			currentDateTime := time.Now().Format("02.01.2006 15:04")
			subtaskLogInd := "N"
			if logToSubtask && subtaskKey != "" {
				subtaskLogInd = "Y"
			}
			err = batch.PrependEntryWithStatus(
				csvPath,
				issueKey,
				subtaskKey,
				issueType,
				description,
				"", // timeSpent - empty for draft
				currentDateTime,
				"", // comment
				subtaskLogInd,
				batch.StatusDraft,
			)
			if err != nil {
				if !addQuiet {
					fmt.Printf("%s %s - %s\n",
						ui.ErrorText.Render("✗"),
						ui.Primary.Render(ticketKey),
						ui.Muted.Render(err.Error()))
				}
				skipped++
				continue
			}
		}

		// Success
		if !addQuiet {
			if addDryRun {
				fmt.Printf("%s %s %s - %s %s\n",
					ui.Success.Render("✓"),
					ui.Primary.Render(displayKey),
					ui.Muted.Render("["+issueType+"]"),
					ui.Muted.Render(truncateString(description, 40)),
					ui.Muted.Render("(would add)"))
			} else {
				fmt.Printf("%s %s %s - %s\n",
					ui.Success.Render("✓"),
					ui.Primary.Render(displayKey),
					ui.Muted.Render("["+issueType+"]"),
					ui.Muted.Render(truncateString(description, 40)))
			}
		}
		added++

		// Re-read entries for next iteration (skip in dry-run)
		if !addDryRun {
			var rereadErr error
			entries, rereadErr = batch.ParseCSV(csvPath)
			if rereadErr != nil {
				if !addQuiet {
					fmt.Println(ui.Error(fmt.Sprintf("Failed to re-read CSV after write: %v", rereadErr)))
				}
				return
			}
		}
	}

	// Summary (skip if quiet)
	if addQuiet {
		return
	}

	fmt.Println()
	total := added + updated
	if total > 0 {
		summaryParts := []string{}
		if added > 0 {
			if addDryRun {
				summaryParts = append(summaryParts, fmt.Sprintf("%s would be added", ui.Success.Render(fmt.Sprintf("%d", added))))
			} else {
				summaryParts = append(summaryParts, fmt.Sprintf("%s added", ui.Success.Render(fmt.Sprintf("%d", added))))
			}
		}
		if updated > 0 {
			if addDryRun {
				summaryParts = append(summaryParts, fmt.Sprintf("%s would be updated", ui.Primary.Render(fmt.Sprintf("%d", updated))))
			} else {
				summaryParts = append(summaryParts, fmt.Sprintf("%s updated", ui.Primary.Render(fmt.Sprintf("%d", updated))))
			}
		}
		fmt.Println(strings.Join(summaryParts, ", "))

		if addDryRun {
			fmt.Println()
			fmt.Println(ui.Warning("No changes made (dry-run mode)"))
		} else {
			fmt.Println()
			fmt.Println(ui.Muted.Render("Next steps:"))
			fmt.Printf("  1. Edit %s\n", ui.Primary.Render(csvPath))
			fmt.Printf("  2. Fill in TimeSpent (e.g., 2h, 30m)\n")
			fmt.Printf("  3. Change status from DRAFT to empty\n")
			fmt.Printf("  4. Run: %s\n", ui.Primary.Render("hlp jira log --batch"))
		}
	}
	if skipped > 0 {
		fmt.Printf("Skipped %s ticket(s)\n", ui.WarningText.Render(fmt.Sprintf("%d", skipped)))
	}
	if total == 0 && skipped == 0 {
		fmt.Println(ui.Info("No tickets to add"))
	}
}

// RunAutoSync performs sprint sync, returns error if failed.
// If quiet=true, suppresses output except errors.
// TO REVIEW: Ticket processing shares logic with runAdd() - skipped: only 2 occurrences
func RunAutoSync(quiet bool) error {
	// Load config (for log_to_subtask preference; daily tickets loaded later)
	logToSubtask := false
	if cfg, cfgErr := config.Load(); cfgErr == nil {
		logToSubtask = cfg.Preferences.LogToSubtask
	}

	// Get current profile for CSV path
	profile, err := config.GetActiveProfile()
	if err != nil && !quiet {
		fmt.Println(ui.Warning(fmt.Sprintf("Could not load profile: %v", err)))
	}
	csvPath := batch.DefaultCSVPathForProfile(profile)

	// Get JIRA client
	client, err := getJiraClient()
	if err != nil {
		return err
	}

	// Parse existing entries for duplicate checking
	todayStr := time.Now().Format("02.01.2006")
	entries, parseErr := batch.ParseCSV(csvPath)
	if parseErr != nil && !errors.Is(parseErr, os.ErrNotExist) {
		return fmt.Errorf("worklog CSV is unreadable, aborting sync so it isn't overwritten: %w", parseErr)
	}

	// Fetch assigned tickets
	if !quiet {
		fmt.Println("Fetching assigned tickets...")
	}
	sprintTickets, err := fetchAssignedTickets(client)
	if err != nil {
		return fmt.Errorf("failed to fetch assigned tickets: %w", err)
	}

	if !quiet {
		fmt.Printf("Found %d assigned tickets\n", len(sprintTickets))
	}

	// Build combined ticket key list: assigned + daily
	var ticketKeys []string
	for _, t := range sprintTickets {
		ticketKeys = append(ticketKeys, t.Key)
	}

	// Append daily tickets from config (normalize and validate)
	if cfg, cfgErr := config.Load(); cfgErr == nil {
		for _, dt := range cfg.Preferences.DailyTickets {
			upper := strings.ToUpper(strings.TrimSpace(dt))
			if ticketKeyPattern.MatchString(upper) {
				ticketKeys = append(ticketKeys, upper)
			}
		}
	} else if !quiet {
		fmt.Println(ui.Warning(fmt.Sprintf("Could not load config for daily tickets: %v", cfgErr)))
	}
	ticketKeys = deduplicateKeys(ticketKeys)

	// Track results
	added := 0
	updated := 0
	skipped := 0

	// Process each ticket
	for _, ticketKey := range ticketKeys {

		// Check if ticket exists for today
		// Match on IssueKey (parent-only) OR SubtaskKey to handle both cases
		var existingEntry *batch.Entry
		for i, e := range entries {
			entryDate := strings.Split(e.Date, " ")[0]
			if entryDate != todayStr {
				continue
			}
			if (strings.EqualFold(e.IssueKey, ticketKey) && e.SubtaskKey == "") ||
				strings.EqualFold(e.SubtaskKey, ticketKey) {
				existingEntry = &entries[i]
				break
			}
		}

		// Skip entries that already exist for today (any status except DRAFT which gets updated below)
		if existingEntry != nil && existingEntry.Status != batch.StatusDraft {
			skipped++
			continue
		}

		// Fetch ticket info (needed for new entries and DRAFT updates)
		ticket, err := client.GetTicket(ticketKey)
		if err != nil {
			skipped++
			continue
		}

		// Resolve subtask relationship
		issueKey := ticket.Key
		subtaskKey := ""
		issueType := ticket.IssueType
		description := ticket.Summary

		if ticket.IsSubtask && ticket.ParentKey != "" {
			parentTicket, err := client.GetTicket(ticket.ParentKey)
			if err == nil {
				subtaskKey = ticket.Key
				issueKey = parentTicket.Key
				issueType = parentTicket.IssueType
				description = parentTicket.Summary + " > " + ticket.Summary
			}
		}

		displayKey := issueKey
		if subtaskKey != "" {
			displayKey = issueKey + " > " + subtaskKey
		}

		// Skip parent-only entries when parent already has a subtask entry today
		if subtaskKey == "" && findSubtaskKey(entries, issueKey, todayStr) != "" {
			skipped++
			continue
		}

		// Handle DRAFT update
		if existingEntry != nil && existingEntry.Status == batch.StatusDraft {
			err = batch.UpdateEntryDescription(csvPath, existingEntry.RowNumber, description)
			if err != nil {
				skipped++
				continue
			}

			if !quiet {
				fmt.Printf("%s %s %s - %s %s\n",
					ui.Primary.Render("↻"),
					ui.Primary.Render(displayKey),
					ui.Muted.Render("["+issueType+"]"),
					ui.Muted.Render(truncateString(description, 40)),
					ui.Muted.Render("(updated)"))
			}
			updated++
			var rereadErr error
			entries, rereadErr = batch.ParseCSV(csvPath)
			if rereadErr != nil {
				if !quiet {
					fmt.Println(ui.Error(fmt.Sprintf("Failed to re-read CSV: %v", rereadErr)))
				}
				return fmt.Errorf("CSV re-read failed: %w", rereadErr)
			}
			continue
		}

		// Clean up redundant parent-only entry BEFORE prepend (row numbers are still valid)
		if subtaskKey != "" {
			removeParentOnlyEntry(csvPath, entries, issueKey, todayStr, quiet)
		}

		// Prepend to CSV with DRAFT status
		currentDateTime := time.Now().Format("02.01.2006 15:04")
		subtaskLogInd := "N"
		if logToSubtask && subtaskKey != "" {
			subtaskLogInd = "Y"
		}
		err = batch.PrependEntryWithStatus(
			csvPath,
			issueKey,
			subtaskKey,
			issueType,
			description,
			"", // timeSpent - empty for draft
			currentDateTime,
			"", // comment
			subtaskLogInd,
			batch.StatusDraft,
		)
		if err != nil {
			skipped++
			continue
		}

		if !quiet {
			fmt.Printf("%s %s %s - %s\n",
				ui.Success.Render("✓"),
				ui.Primary.Render(displayKey),
				ui.Muted.Render("["+issueType+"]"),
				ui.Muted.Render(truncateString(description, 40)))
		}
		added++
		var rereadErr error
		entries, rereadErr = batch.ParseCSV(csvPath)
		if rereadErr != nil {
			if !quiet {
				fmt.Println(ui.Error(fmt.Sprintf("Failed to re-read CSV: %v", rereadErr)))
			}
			return fmt.Errorf("CSV re-read failed: %w", rereadErr)
		}
	}

	// Summary (only if not quiet and changes made)
	if !quiet && (added > 0 || updated > 0) {
		fmt.Printf("\nAuto-sync: %d added, %d updated, %d skipped\n", added, updated, skipped)
	}

	// Update last sync time
	if syncErr := config.UpdateLastSyncTime(); syncErr != nil && !quiet {
		fmt.Println(ui.Warning(fmt.Sprintf("Could not update sync timestamp: %v", syncErr)))
	}

	return nil
}

// removeParentOnlyEntry removes a today's parent-only DRAFT entry when a subtask entry
// now covers it. Returns true if an entry was removed.
func removeParentOnlyEntry(csvPath string, entries []batch.Entry, issueKey string, todayStr string, quiet bool) bool {
	for _, e := range entries {
		entryDate := strings.Split(e.Date, " ")[0]
		if entryDate == todayStr &&
			strings.EqualFold(e.IssueKey, issueKey) &&
			e.SubtaskKey == "" &&
			e.Status == batch.StatusDraft {
			if !quiet {
				fmt.Printf("%s %s - removed (now tracked via subtask)\n",
					ui.Muted.Render("⊘"),
					ui.Muted.Render(issueKey))
			}
			if err := batch.RemoveEntryByRow(csvPath, e.RowNumber); err != nil {
				if !quiet {
					fmt.Println(ui.Warning(fmt.Sprintf("Could not remove parent-only entry for %s: %v", issueKey, err)))
				}
				return false
			}
			return true
		}
	}
	return false
}

// findSubtaskKey returns the first subtask key tracked under issueKey for today, or "" if none.
// Scoped to today so historical subtask rows don't block parent-only adds on a fresh day.
func findSubtaskKey(entries []batch.Entry, issueKey, todayStr string) string {
	for _, e := range entries {
		entryDate := strings.Split(e.Date, " ")[0]
		if entryDate != todayStr {
			continue
		}
		if strings.EqualFold(e.IssueKey, issueKey) && e.SubtaskKey != "" {
			return e.SubtaskKey
		}
	}
	return ""
}

// deduplicateKeys removes duplicate ticket keys (case-insensitive), preserving order.
// Sprint keys come first (appended before manual args), so they take priority.
func deduplicateKeys(keys []string) []string {
	seen := make(map[string]bool, len(keys))
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		upper := strings.ToUpper(k)
		if !seen[upper] {
			seen[upper] = true
			result = append(result, k)
		}
	}
	return result
}

// IsQuietMode checks if --quiet flag is set on the command or its parents
func IsQuietMode(cmd *cobra.Command) bool {
	quietFlag := cmd.Flag("quiet")
	if quietFlag != nil && quietFlag.Changed {
		return true
	}
	return addQuiet
}
