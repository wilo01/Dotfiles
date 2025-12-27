// Package jira provides JIRA worklog commands for the helper CLI.
package jira

import (
	"fmt"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/context"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/dariuszw/hlp/internal/worklog"
	"github.com/dariuszw/hlp/pkg/duration"
	"github.com/spf13/cobra"
)

const dayDuration = 24 * time.Hour

var (
	logTicket   string
	logComment  string
	logDate     string
	logTime     string
	logBatch    bool
	logDryRun   bool
	logSync     bool
	logFromDate string
	logToDate   string
	logConfirm  bool
)

var logCmd = &cobra.Command{
	Use:   "log [duration] [description]",
	Short: "Log work time to a JIRA ticket",
	Long: `Log work time to a JIRA ticket.

Duration format: 30m, 2h, 1h30m, 1d, etc.

Examples:
  hlp jira log 2h "Fixed login bug"
  hlp jira log 2h -t VIS-1234 "Code review"
  hlp jira log 30m                          # Uses auto-detected ticket
  hlp jira log --batch                      # Process CSV file
  hlp jira log --batch --dry-run            # Preview batch without posting
  hlp jira log --sync                       # Sync worklogs from JIRA (last 7 days)
  hlp jira log --sync --from 2025-12-01     # Sync from specific date
  hlp jira log --sync --dry-run             # Preview sync without changes`,
	Args: cobra.MaximumNArgs(2),
	Run:  runLog,
}

func init() {
	logCmd.Flags().StringVarP(&logTicket, "ticket", "t", "", "JIRA ticket key (auto-detects if not provided)")
	logCmd.Flags().StringVarP(&logComment, "comment", "c", "", "Work log comment")
	logCmd.Flags().StringVar(&logDate, "date", "", "Date for worklog (YYYY-MM-DD, default: today)")
	logCmd.Flags().StringVar(&logTime, "time", "09:00", "Start time (HH:MM)")
	logCmd.Flags().BoolVarP(&logBatch, "batch", "b", false, "Process batch CSV file")
	logCmd.Flags().BoolVar(&logDryRun, "dry-run", false, "Preview batch without posting")
	logCmd.Flags().BoolVar(&logSync, "sync", false, "Sync worklogs from JIRA to CSV")
	logCmd.Flags().StringVar(&logFromDate, "from", "", "Start date for sync (YYYY-MM-DD, default: 7 days ago)")
	logCmd.Flags().StringVar(&logToDate, "to", "", "End date for sync (YYYY-MM-DD, default: today)")
	logCmd.Flags().BoolVarP(&logConfirm, "confirm", "y", false, "Skip confirmation prompt for protected profiles")
}

func runLog(cmd *cobra.Command, args []string) {
	if logSync {
		runSyncLog(cmd, args)
		return
	}

	if logBatch {
		runBatchLog(cmd, args)
		return
	}

	// Validate we have duration
	if len(args) < 1 {
		fmt.Println(ui.Error("Duration is required. Example: hlp jira log 2h"))
		return
	}

	durationStr := args[0]
	var description string
	if len(args) > 1 {
		description = args[1]
	}
	if description == "" {
		description = logComment
	}

	// Parse duration
	dur, err := duration.Parse(durationStr)
	if err != nil {
		fmt.Println(ui.Error("Invalid duration format: " + err.Error()))
		return
	}

	// Get ticket
	ticket := logTicket
	if ticket == "" {
		detector := context.NewDetector()
		ticket, _ = detector.DetectTicket()
	}

	if ticket == "" {
		fmt.Println(ui.Error("Could not detect ticket. Use -t to specify."))
		return
	}

	// Parse date/time
	var startTime time.Time
	if logDate != "" {
		startTime, err = time.Parse("2006-01-02", logDate)
		if err != nil {
			fmt.Println(ui.Error("Invalid date format. Use YYYY-MM-DD"))
			return
		}
	} else {
		startTime = time.Now()
	}

	// Parse time
	if logTime != "" {
		parts := []int{9, 0}
		fmt.Sscanf(logTime, "%d:%d", &parts[0], &parts[1])
		startTime = time.Date(
			startTime.Year(), startTime.Month(), startTime.Day(),
			parts[0], parts[1], 0, 0, startTime.Location(),
		)
	}

	// Get JIRA client
	client, err := getJiraClient()
	if err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// Log work
	fmt.Printf("%s Logging %s to %s...\n",
		ui.SpinnerFrames[0],
		duration.Format(dur),
		ui.Primary.Render(ticket))

	entry := internalJira.WorklogEntry{
		TimeSpent: duration.ToJiraFormat(dur),
		Started:   startTime,
		Comment:   description,
	}

	if err := client.LogWork(ticket, entry); err != nil {
		fmt.Println(ui.Error("Failed to log work: " + err.Error()))
		return
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Logged %s to %s", duration.Format(dur), ticket)))
	if description != "" {
		fmt.Println(ui.Muted.Render("  Comment: " + description))
	}

	// Show daily total warning
	showDailyWarning(client, startTime)
}

func runBatchLog(_ *cobra.Command, _ []string) {
	// Get current profile for CSV path and auto-create decision
	profile, err := config.GetActiveProfile()
	if err != nil {
		fmt.Println(ui.Muted.Render("  (using default profile)"))
	}
	csvPath := batch.DefaultCSVPathForProfile(profile)

	// Show current profile with CSV info
	if profile != nil {
		modeStr := ui.Success.Render(" [READ/WRITE]")
		csvInfo := ui.Muted.Render(" → worklogs.csv")
		if profile.Protected {
			modeStr = ui.WarningText.Render(" [PROTECTED]")
		}
		if profile.IsLocal() {
			csvInfo = ui.Muted.Render(" → worklogs-local.csv")
			modeStr = ui.Primary.Render(" [MOCK]")
		}
		fmt.Printf("Profile: %s%s%s\n", ui.Primary.Render(profile.Name), modeStr, csvInfo)
	}

	// Show loading message
	fmt.Printf("Loading %s...\n", ui.Primary.Render(csvPath))

	// Parse CSV
	entries, err := batch.ParsePendingCSV(csvPath)
	if err != nil {
		fmt.Println(ui.Error("Failed to parse CSV: " + err.Error()))
		return
	}

	if len(entries) == 0 {
		fmt.Println(ui.Info("No pending entries found"))
		return
	}

	fmt.Printf("Found %s pending entries\n", ui.Success.Render(fmt.Sprintf("%d", len(entries))))

	// Get JIRA client (needed for preview and processing)
	client, err := getJiraClient()
	if err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// Fetch ticket summaries for descriptions (needed for preview)
	uniqueKeys := getUniqueTicketKeys(entries)
	summaries, err := client.GetIssueSummaries(uniqueKeys)
	if err != nil {
		fmt.Println(ui.Warning("Could not fetch ticket summaries: " + err.Error()))
		summaries = make(map[string]string)
	}
	for i := range entries {
		if summary, ok := summaries[entries[i].IssueKey]; ok {
			entries[i].Description = summary
		}
	}

	// Show preview before confirmation (only for non-dry-run)
	if !logDryRun {
		fmt.Println()
		fmt.Println("Processing worklogs " + ui.Muted.Render("(preview)") + ":")
		showWorklogPreview(entries, getExpectedHoursPerDay())
	}

	// Safety guard for protected profiles
	if !logDryRun && !logConfirm {
		if profile != nil && profile.Protected {
			if !ui.ConfirmProtectedProfile(profile.Name, profile.BaseURL, len(entries)) {
				fmt.Println(ui.Info("Operation cancelled"))
				return
			}
		}
	}

	fmt.Println()

	if logDryRun {
		// Preview mode - load ALL entries (including DONE) for complete daily totals
		allEntries, _ := batch.ParseCSV(csvPath)

		// Get unique dates from pending entries AND draft entries with time
		relevantDates := make(map[string]bool)
		for _, e := range entries {
			if t, err := batch.ParseDate(e.Date, "09:00"); err == nil {
				relevantDates[t.Format("2006-01-02")] = true
			}
		}
		// Also include dates with DRAFT entries that have time logged
		for _, e := range allEntries {
			if e.IsDraft() && e.TimeSpent != "" {
				if t, err := batch.ParseDate(e.Date, "09:00"); err == nil {
					relevantDates[t.Format("2006-01-02")] = true
				}
			}
		}

		// Filter all entries to only include relevant dates
		var relevantEntries []batch.Entry
		for _, e := range allEntries {
			if t, err := batch.ParseDate(e.Date, "09:00"); err == nil {
				if relevantDates[t.Format("2006-01-02")] {
					relevantEntries = append(relevantEntries, e)
				}
			}
		}

		// Fetch descriptions for all relevant entries
		allKeys := getUniqueTicketKeys(relevantEntries)
		allSummaries, _ := client.GetIssueSummaries(allKeys)
		for i := range relevantEntries {
			if summary, ok := allSummaries[relevantEntries[i].IssueKey]; ok {
				relevantEntries[i].Description = summary
			}
		}

		fmt.Println("Processing worklogs " + ui.Muted.Render("(dry run)") + ":")
		showDryRunGroupedByDay(relevantEntries, getExpectedHoursPerDay())
		fmt.Println()
		fmt.Println(ui.Warning("Dry run - no entries posted"))
		return
	}

	fmt.Println("Processing worklogs:")

	// Track results for summary
	var failedEntries []batch.Result

	// Process batch with profile config (mock mode for LOCAL)
	mockMode := profile != nil && profile.IsLocal()
	processor := batch.NewProcessorWithConfig(batch.ProcessorConfig{
		Client:      client,
		DefaultTime: logTime,
		Profile:     profile,
		MockMode:    mockMode,
	})
	results := processor.ProcessBatch(entries, false, func(current, total int, result batch.Result) {
		var status string
		if result.Success {
			if result.ErrorMessage != "" {
				status = ui.Success.Render(result.NewStatus) + " " + ui.Muted.Render(result.ErrorMessage)
			} else {
				status = ui.Success.Render(result.NewStatus)
			}
		} else {
			status = ui.ErrorText.Render("FAILED") + " " + ui.Muted.Render("("+result.ErrorMessage+")")
		}

		fmt.Printf("  %s %s %s %s %s ... %s\n",
			ui.Muted.Render(fmt.Sprintf("%d/%d", current, total)),
			ui.Primary.Render(fmt.Sprintf("%-12s", result.Entry.IssueKey)),
			ui.Muted.Render(fmt.Sprintf("%-30s", truncateString(result.Entry.Description, 30))),
			ui.Success.Render(fmt.Sprintf("%-8s", result.Entry.TimeSpent)),
			ui.Muted.Render(fmt.Sprintf("%-12s", result.Entry.Date)),
			status)
	})

	// Update CSV
	if err := batch.UpdateCSVStatus(csvPath, results); err != nil {
		fmt.Println(ui.Warning("Failed to update CSV: " + err.Error()))
	}

	// Count results
	doneCount := 0
	updatedCount := 0
	failedCount := 0

	for _, r := range results {
		if !r.Success {
			failedCount++
			failedEntries = append(failedEntries, r)
		} else if r.NewStatus == batch.StatusDone {
			doneCount++
		} else if r.NewStatus == batch.StatusUpdated {
			updatedCount++
		}
	}

	// Summary
	fmt.Printf("\nSummary: %s DONE, %s UPDATED, %s failed\n",
		ui.Success.Render(fmt.Sprintf("%d", doneCount)),
		ui.Primary.Render(fmt.Sprintf("%d", updatedCount)),
		ui.ErrorText.Render(fmt.Sprintf("%d", failedCount)))

	if doneCount > 0 || updatedCount > 0 {
		fmt.Println(ui.Success.Render("CSV updated"))
	}

	// Show failed entries
	if len(failedEntries) > 0 {
		fmt.Println("\n" + ui.ErrorText.Render("Failed entries:"))
		for _, r := range failedEntries {
			fmt.Printf("  %s %s - %s\n",
				ui.Muted.Render(fmt.Sprintf("Row %d:", r.Entry.RowNumber)),
				ui.Primary.Render(r.Entry.IssueKey),
				ui.ErrorText.Render(r.ErrorMessage))
		}
	}

	// Show daily total warnings for batch
	showBatchDailyWarnings(client, results)
}

func runSyncLog(_ *cobra.Command, _ []string) {
	// Determine date range (default: last 7 days)
	toDate := time.Now()
	fromDate := toDate.AddDate(0, 0, -7)

	if logFromDate != "" {
		t, err := time.Parse("2006-01-02", logFromDate)
		if err != nil {
			fmt.Println(ui.Error("Invalid --from date. Use YYYY-MM-DD"))
			return
		}
		fromDate = t
	}

	if logToDate != "" {
		t, err := time.Parse("2006-01-02", logToDate)
		if err != nil {
			fmt.Println(ui.Error("Invalid --to date. Use YYYY-MM-DD"))
			return
		}
		toDate = t
	}

	// Get JIRA client
	client, err := getJiraClient()
	if err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// Fetch worklogs from JIRA
	fmt.Printf("Searching for issues updated since %s...\n",
		ui.Primary.Render(fromDate.Format("2006-01-02")))

	jiraWorklogs, err := client.FetchUserWorklogs(fromDate, toDate, func(current, total int, issueKey string) {
		fmt.Printf("\r  Fetching worklogs... (%d/%d) %s",
			current, total, ui.Muted.Render(issueKey))
	})

	if err != nil {
		fmt.Println() // Clear progress line
		fmt.Println(ui.Error("Failed to fetch worklogs: " + err.Error()))
		return
	}

	fmt.Println() // Clear progress line
	fmt.Printf("Found %s worklogs in JIRA (date range: %s to %s)\n",
		ui.Success.Render(fmt.Sprintf("%d", len(jiraWorklogs))),
		ui.Muted.Render(fromDate.Format("2006-01-02")),
		ui.Muted.Render(toDate.Format("2006-01-02")))

	// Load existing CSV entries using profile-aware path
	profile, _ := config.GetActiveProfile()
	csvPath := batch.DefaultCSVPathForProfile(profile)
	csvEntries, _ := batch.ParseCSV(csvPath) // Empty if file doesn't exist

	// Find missing worklogs
	missing := batch.FindMissingWorklogs(jiraWorklogs, csvEntries)
	addedCount := 0

	// Fetch issue details if there are missing worklogs
	var details map[string]internalJira.IssueDetails
	if len(missing) > 0 {
		fmt.Printf("Found %s new worklogs to add\n",
			ui.Success.Render(fmt.Sprintf("%d", len(missing))))
		fmt.Print("Fetching issue details...")
		uniqueKeys := getUniqueWorklogKeys(missing)
		details, _ = client.GetIssueDetails(uniqueKeys)
		fmt.Println(" done")
		fmt.Println()
	} else {
		fmt.Println(ui.Info("CSV is up to date - no new worklogs found"))
	}

	// Handle dry-run: preview both SYNC and DRAFT entries
	if logDryRun {
		// Preview sync entries
		if len(missing) > 0 {
			fmt.Println("New SYNC entries to add " + ui.Muted.Render("(dry run)") + ":")
			for _, wl := range missing {
				detail := details[wl.IssueKey]
				fmt.Printf("  %s %s %s %s %s %s\n",
					ui.Primary.Render("SYNC"),
					ui.Primary.Render(fmt.Sprintf("%-12s", wl.IssueKey)),
					ui.Muted.Render("["+detail.IssueType+"]"),
					ui.Muted.Render(fmt.Sprintf("%-25s", truncateString(detail.Summary, 25))),
					ui.Success.Render(fmt.Sprintf("%-8s", wl.TimeSpentStr)),
					ui.Muted.Render(wl.Started.Format("02.01.2006 15:04")))
			}
			fmt.Println()
		}

		// Preview DRAFT entries
		fmt.Println("Checking sprint tickets for DRAFT entries " + ui.Muted.Render("(dry run)") + ":")
		sprintTickets, err := client.SearchSprintTickets()
		if err != nil {
			fmt.Println(ui.Warning("Failed to fetch sprint tickets: " + err.Error()))
		} else {
			entries, _ := batch.ParseCSV(csvPath)
			draftPreviewCount := 0

			// Find last logged date (same logic as actual run)
			lastLoggedDate := time.Now().Format("02.01.2006")
			for _, e := range entries {
				if e.TimeSpent != "" || e.Status == batch.StatusDone ||
					e.Status == batch.StatusSync || e.Status == batch.StatusUpdated {
					lastLoggedDate = strings.Split(e.Date, " ")[0]
					break
				}
			}

			for _, ticket := range sprintTickets {
				// For sub-tasks: show that it will use parent key
				issueKey := ticket.Key
				issueType := ticket.IssueType
				description := ticket.Summary

				if ticket.IsSubtask && ticket.ParentKey != "" {
					issueKey = ticket.ParentKey
					description = "... > " + ticket.Summary // Preview shows parent will be fetched
				}

				exists := batch.EntryExistsForTicket(entries, issueKey, lastLoggedDate, ticket.IsSubtask, ticket.Summary)

				if !exists {
					draftPreviewCount++
					fmt.Printf("  %s %s %s %s %s\n",
						ui.Muted.Render("DRAFT"),
						ui.Primary.Render(fmt.Sprintf("%-12s", issueKey)),
						ui.Muted.Render("["+issueType+"]"),
						ui.Muted.Render(truncateString(description, 25)),
						ui.Muted.Render("("+lastLoggedDate+")"))
				}
			}

			if draftPreviewCount == 0 {
				fmt.Println(ui.Muted.Render("  No new DRAFT entries needed"))
			}
		}

		fmt.Println()
		fmt.Println(ui.Warning("Dry run - no changes made"))
		return
	}

	// Actual sync: Append new entries with SYNC status
	if len(missing) > 0 {
		fmt.Println("Adding new entries:")
		for _, wl := range missing {
			dateStr := wl.Started.Format("02.01.2006 15:04")
			detail := details[wl.IssueKey]
			if err = batch.AppendEntryWithStatus(csvPath, wl.IssueKey, "", detail.IssueType, detail.Summary, wl.TimeSpentStr, dateStr, wl.Comment, "", batch.StatusSync); err != nil {
				fmt.Printf("  %s %s - %s\n",
					ui.ErrorText.Render("FAILED"),
					ui.Primary.Render(wl.IssueKey),
					ui.Muted.Render(err.Error()))
				continue
			}
			addedCount++
			fmt.Printf("  %s %s %s %s %s %s\n",
				ui.Success.Render("SYNC"),
				ui.Primary.Render(fmt.Sprintf("%-12s", wl.IssueKey)),
				ui.Muted.Render("["+detail.IssueType+"]"),
				ui.Muted.Render(fmt.Sprintf("%-25s", truncateString(detail.Summary, 25))),
				ui.Success.Render(fmt.Sprintf("%-8s", wl.TimeSpentStr)),
				ui.Muted.Render(dateStr))
		}

		// Sort CSV by date
		fmt.Println("\nSorting CSV by date...")
		if err = batch.SortCSVByDate(csvPath); err != nil {
			fmt.Println(ui.Warning("Failed to sort CSV: " + err.Error()))
		} else {
			fmt.Println(ui.Success.Render("CSV sorted (newest first)"))
		}

		// Fill missing descriptions for existing entries
		entries, _ := batch.ParseCSV(csvPath)
		var keysNeedingDesc []string
		for _, e := range entries {
			if e.Description == "" {
				keysNeedingDesc = append(keysNeedingDesc, e.IssueKey)
			}
		}
		if len(keysNeedingDesc) > 0 {
			fmt.Print("Fetching missing descriptions...")
			descSummaries, _ := client.GetIssueSummaries(unique(keysNeedingDesc))
			updatedCount, _ := batch.UpdateCSVDescriptions(csvPath, descSummaries)
			if updatedCount > 0 {
				fmt.Printf(" updated %d entries\n", updatedCount)
			} else {
				fmt.Println(" done")
			}
		}

		fmt.Printf("\nSummary: %s entries synced from JIRA\n",
			ui.Success.Render(fmt.Sprintf("%d", addedCount)))
	}

	// Add sprint tickets as DRAFT entries (at TOP of CSV)
	fmt.Println("\nChecking sprint tickets for DRAFT entries...")
	sprintTickets, err := client.SearchSprintTickets()
	if err != nil {
		fmt.Println(ui.Warning("Failed to fetch sprint tickets: " + err.Error()))
		return
	}

	// Re-parse CSV to get updated entries after sync
	entries, _ := batch.ParseCSV(csvPath)
	draftCount := 0

	// Find last logged date from CSV (excluding DRAFT entries)
	// CSV is sorted newest first, so first non-DRAFT entry is the most recent
	lastLoggedDate := time.Now().Format("02.01.2006") // fallback to today
	for _, e := range entries {
		if e.TimeSpent != "" || e.Status == batch.StatusDone ||
			e.Status == batch.StatusSync || e.Status == batch.StatusUpdated {
			lastLoggedDate = strings.Split(e.Date, " ")[0]
			break
		}
	}
	draftDateTime := lastLoggedDate + " 09:00"

	for _, ticket := range sprintTickets {
		// Determine issue key, subtask key, type, and description
		// For sub-tasks: use parent key/type and combined description, store subtask key
		issueKey := ticket.Key
		subtaskKey := ""
		issueType := ticket.IssueType
		description := ticket.Summary

		if ticket.IsSubtask && ticket.ParentKey != "" {
			// Fetch parent details
			parentTicket, err := client.GetTicket(ticket.ParentKey)
			if err != nil {
				fmt.Println(ui.Warning(fmt.Sprintf("Could not fetch parent %s: %v", ticket.ParentKey, err)))
			} else {
				subtaskKey = ticket.Key // Store original subtask key
				issueKey = parentTicket.Key
				issueType = parentTicket.IssueType
				description = parentTicket.Summary + " > " + ticket.Summary
			}
		}

		// Check if entry already exists (for the issue key we'll use, not original sub-task key)
		exists := batch.EntryExistsForTicket(entries, issueKey, lastLoggedDate, ticket.IsSubtask, ticket.Summary)

		if !exists {
			err = batch.PrependEntryWithStatus(
				csvPath,
				issueKey,
				subtaskKey,
				issueType,
				description,
				"",
				draftDateTime,
				"",
				"",
				batch.StatusDraft,
			)
			if err == nil {
				draftCount++
				fmt.Printf("  %s %s %s %s\n",
					ui.Muted.Render("DRAFT"),
					ui.Primary.Render(fmt.Sprintf("%-12s", issueKey)),
					ui.Muted.Render("["+issueType+"]"),
					ui.Muted.Render(truncateString(description, 35)))
			}
		}
	}

	if draftCount > 0 {
		fmt.Printf("\nAdded %s DRAFT entries from sprint (at top of CSV)\n",
			ui.Success.Render(fmt.Sprintf("%d", draftCount)))
	} else {
		fmt.Println(ui.Muted.Render("No new DRAFT entries needed"))
	}

	// Always enrich and restructure existing entries (even when no new worklogs)
	fmt.Print("\nChecking for entries needing metadata update...")
	enriched, restructured, err := batch.EnrichAndRestructureCSV(csvPath, client)
	if err != nil {
		fmt.Println(ui.Warning(" " + err.Error()))
	} else if enriched > 0 || restructured > 0 {
		fmt.Printf(" updated %d, restructured %d subtasks\n", enriched, restructured)
	} else {
		fmt.Println(" all entries up to date")
	}
}

func getJiraClient() (*internalJira.Client, error) {
	// Get active profile for profile-aware token retrieval
	profile, err := config.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	if profile == nil || profile.BaseURL == "" {
		return nil, fmt.Errorf("JIRA not configured. Run: hlp jira config")
	}

	// Get token for the specific profile (checks MCP .env, then keyring fallback)
	credMgr := config.NewCredentialManager()
	token, err := credMgr.GetJiraTokenForProfile(profile.Name)
	if err != nil {
		return nil, fmt.Errorf("API token not found for profile '%s'. Check ~/.claude/mcp-servers/jira-profiles/jira-%s.env", profile.Name, profile.Name)
	}

	return internalJira.NewClient(profile.BaseURL, profile.Email, token), nil
}

// getUniqueTicketKeys returns unique ticket keys from entries
func getUniqueTicketKeys(entries []batch.Entry) []string {
	seen := make(map[string]bool)
	var keys []string
	for _, e := range entries {
		if !seen[e.IssueKey] {
			seen[e.IssueKey] = true
			keys = append(keys, e.IssueKey)
		}
	}
	return keys
}

// getUniqueWorklogKeys returns unique issue keys from worklogs
func getUniqueWorklogKeys(worklogs []internalJira.Worklog) []string {
	seen := make(map[string]bool)
	var keys []string
	for _, wl := range worklogs {
		if !seen[wl.IssueKey] {
			seen[wl.IssueKey] = true
			keys = append(keys, wl.IssueKey)
		}
	}
	return keys
}

// truncateString truncates a string to maxLen, adding "..." if truncated
// TO REVIEW: Could move to shared utils (also in add.go) - skipped: only 2 occurrences
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// unique returns unique strings from a slice
// TO REVIEW: Could move to shared utils (also in processor.go) - skipped: only 2 occurrences
func unique(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// getExpectedHoursPerDay returns the expected hours per day from config
func getExpectedHoursPerDay() time.Duration {
	cfg, err := config.Load()
	if err != nil {
		return 8 * time.Hour
	}

	expectedStr := cfg.Preferences.ExpectedHoursPerDay
	if expectedStr == "" {
		return 8 * time.Hour
	}

	expectedDur, err := duration.Parse(expectedStr)
	if err != nil {
		return 8 * time.Hour
	}

	return expectedDur
}

// showDailyWarning fetches daily total and shows warning if over/under expected hours
func showDailyWarning(client *internalJira.Client, logDate time.Time) {
	expectedDur := getExpectedHoursPerDay()

	// Fetch all worklogs for this day
	fromDate := time.Date(logDate.Year(), logDate.Month(), logDate.Day(), 0, 0, 0, 0, logDate.Location())
	toDate := fromDate.Add(dayDuration)

	worklogs, err := client.FetchUserWorklogs(fromDate, toDate, nil)
	if err != nil {
		return // Silently skip if fetch fails
	}

	// Calculate daily total
	var totalLogged time.Duration
	targetDate := logDate.Format("2006-01-02")
	for _, wl := range worklogs {
		if wl.Started.Format("2006-01-02") == targetDate {
			totalLogged += wl.TimeSpent
		}
	}

	// Show warning
	fmt.Println()
	fmt.Println(ui.FormatDailyWarning(logDate, totalLogged, expectedDur))
}

// showBatchDailyWarnings shows per-day breakdown with over/under warnings for batch mode
func showBatchDailyWarnings(client *internalJira.Client, results []batch.Result) {
	// Get unique dates from successfully processed entries
	dates := getUniqueDatesFromResults(results)
	if len(dates) == 0 {
		return
	}

	expectedDur := getExpectedHoursPerDay()

	// Find date range
	minDate, maxDate := dates[0], dates[0]
	for _, d := range dates {
		if d.Before(minDate) {
			minDate = d
		}
		if d.After(maxDate) {
			maxDate = d
		}
	}

	// Fetch existing worklogs for date range
	worklogs, err := client.FetchUserWorklogs(minDate, maxDate.Add(dayDuration), nil)
	if err != nil {
		return // Silently skip if fetch fails
	}

	// Analyze
	analysis := worklog.AnalyzeDailyTotals(worklogs, expectedDur)

	if analysis.HasWarnings {
		cfg, _ := config.Load()
		expectedStr := cfg.Preferences.ExpectedHoursPerDay
		if expectedStr == "" {
			expectedStr = "8h"
		}

		fmt.Print(ui.DailyBreakdown(analysis.Summaries, expectedStr))
		fmt.Println(ui.DailyWarningsSummary(analysis))
	}
}

// getUniqueDatesFromResults extracts unique dates from processed results
func getUniqueDatesFromResults(results []batch.Result) []time.Time {
	seen := make(map[string]bool)
	var dates []time.Time

	for _, r := range results {
		if !r.Success {
			continue
		}

		t, err := batch.ParseDate(r.Entry.Date, "09:00")
		if err != nil {
			continue
		}

		dateKey := t.Format("2006-01-02")
		if !seen[dateKey] {
			seen[dateKey] = true
			dates = append(dates, t.Truncate(dayDuration))
		}
	}

	return dates
}

// showDryRunGroupedByDay shows entries grouped by day with totals
func showDryRunGroupedByDay(entries []batch.Entry, expectedHours time.Duration) {
	// Group entries by date
	type dayGroup struct {
		date    string
		entries []batch.Entry
		total   time.Duration
	}

	byDate := make(map[string]*dayGroup)
	dateOrder := []string{}

	for _, e := range entries {
		t, err := batch.ParseDate(e.Date, "09:00")
		if err != nil {
			continue
		}
		dateKey := t.Format("02.01.2006")

		dur, _ := duration.Parse(e.TimeSpent)

		if _, exists := byDate[dateKey]; !exists {
			byDate[dateKey] = &dayGroup{date: dateKey}
			dateOrder = append(dateOrder, dateKey)
		}
		byDate[dateKey].entries = append(byDate[dateKey].entries, e)
		byDate[dateKey].total += dur
	}

	// Print entries grouped by day
	entryNum := 0
	totalEntries := len(entries)

	for _, dateKey := range dateOrder {
		group := byDate[dateKey]

		// Print entries for this day
		for _, e := range group.entries {
			entryNum++
			// Check if entry is skipped (DONE/SYNC/UPDATED/DRAFT - grayed out) or PENDING (normal)
			isSkipped := e.Status == batch.StatusDone || e.Status == batch.StatusUpdated || e.Status == batch.StatusSync || e.Status == batch.StatusDraft
			if isSkipped {
				// Skipped entries - all grayed out
				statusDisplay := e.Status
				if statusDisplay == "" {
					statusDisplay = "PENDING"
				}
				fmt.Printf("  %s %s %s %s %s ... %s\n",
					ui.Muted.Render(fmt.Sprintf("%d/%d", entryNum, totalEntries)),
					ui.Muted.Render(fmt.Sprintf("%-12s", e.IssueKey)),
					ui.Muted.Render(fmt.Sprintf("%-30s", truncateString(e.Description, 30))),
					ui.Muted.Render(fmt.Sprintf("%-8s", e.TimeSpent)),
					ui.Muted.Render(fmt.Sprintf("%-12s", e.Date)),
					ui.Muted.Render(statusDisplay))
			} else {
				// PENDING entries - normal colors
				fmt.Printf("  %s %s %s %s %s ... %s\n",
					ui.Muted.Render(fmt.Sprintf("%d/%d", entryNum, totalEntries)),
					ui.Primary.Render(fmt.Sprintf("%-12s", e.IssueKey)),
					ui.Muted.Render(fmt.Sprintf("%-30s", truncateString(e.Description, 30))),
					ui.Success.Render(fmt.Sprintf("%-8s", e.TimeSpent)),
					ui.Muted.Render(fmt.Sprintf("%-12s", e.Date)),
					ui.Muted.Render("PENDING"))
			}
		}

		// Print separator and total for this day
		fmt.Println("  " + ui.Muted.Render("─────────────────────────────────────────────────────────────────────────────────────"))

		// Calculate difference from expected
		diff := group.total - expectedHours
		var diffStr string
		if diff > 0 {
			diffStr = ui.WarningText.Render(fmt.Sprintf("(+%s over %s)", duration.Format(diff), duration.Format(expectedHours)))
		} else if diff < 0 {
			diffStr = ui.ErrorText.Render(fmt.Sprintf("(need %s for %s)", duration.Format(-diff), duration.Format(expectedHours)))
		} else {
			diffStr = ui.Success.Render("✓")
		}

		// Align total under time_spent column
		// Entry format: "  X/X TICKET-KEY    DESCRIPTION                    TIME     DATE..."
		// Positions:     2   4   12           30                             8
		// Total before time: 2 + 4 + 12 + 1 + 30 + 1 = 50
		totalStr := duration.Format(group.total)
		fmt.Printf("  %s%s%s    %s\n",
			ui.Muted.Render(fmt.Sprintf("%-10s", dateKey)),
			"                                      ", // 38 spaces to align with time column
			ui.Success.Render(fmt.Sprintf("%-8s", totalStr)),
			diffStr)

		// Add blank line between days (except for last day)
		if dateKey != dateOrder[len(dateOrder)-1] {
			fmt.Println()
		}
	}
}

// showWorklogPreview displays pending entries before confirmation prompt
// TO REVIEW: Similar to showDryRunGroupedByDay() - skipped: only 2 occurrences
func showWorklogPreview(entries []batch.Entry, expectedHours time.Duration) {
	if len(entries) == 0 {
		return
	}

	// Group entries by date
	type dayGroup struct {
		date    string
		entries []batch.Entry
		total   time.Duration
	}

	byDate := make(map[string]*dayGroup)
	dateOrder := []string{}

	for _, e := range entries {
		t, err := batch.ParseDate(e.Date, "09:00")
		if err != nil {
			continue
		}
		dateKey := t.Format("02.01.2006")

		dur, _ := duration.Parse(e.TimeSpent)

		if _, exists := byDate[dateKey]; !exists {
			byDate[dateKey] = &dayGroup{date: dateKey}
			dateOrder = append(dateOrder, dateKey)
		}
		byDate[dateKey].entries = append(byDate[dateKey].entries, e)
		byDate[dateKey].total += dur
	}

	// Print entries grouped by day
	entryNum := 0
	totalEntries := len(entries)

	for _, dateKey := range dateOrder {
		group := byDate[dateKey]

		// Print entries for this day
		for _, e := range group.entries {
			entryNum++

			// Format date - show time only if present in the Date field
			dateDisplay := e.Date

			fmt.Printf("  %s %s %s %s %s ... %s\n",
				ui.Muted.Render(fmt.Sprintf("%d/%d", entryNum, totalEntries)),
				ui.Primary.Render(fmt.Sprintf("%-12s", e.IssueKey)),
				ui.Muted.Render(fmt.Sprintf("%-30s", truncateString(e.Description, 30))),
				ui.Success.Render(fmt.Sprintf("%-8s", e.TimeSpent)),
				ui.Muted.Render(fmt.Sprintf("%-16s", dateDisplay)),
				ui.Muted.Render("PENDING"))
		}

		// Print separator and total for this day
		fmt.Println("  " + ui.Muted.Render("─────────────────────────────────────────────────────────────────────────────────────"))

		// Calculate difference from expected
		diff := group.total - expectedHours
		var diffStr string
		if diff > 0 {
			diffStr = ui.WarningText.Render(fmt.Sprintf("(+%s over %s)", duration.Format(diff), duration.Format(expectedHours)))
		} else if diff < 0 {
			diffStr = ui.ErrorText.Render(fmt.Sprintf("(need %s for %s)", duration.Format(-diff), duration.Format(expectedHours)))
		} else {
			diffStr = ui.Success.Render("✓")
		}

		totalStr := duration.Format(group.total)
		fmt.Printf("  %s%s%s    %s\n",
			ui.Muted.Render(fmt.Sprintf("%-10s", dateKey)),
			"                                      ", // 38 spaces to align with time column
			ui.Success.Render(fmt.Sprintf("%-8s", totalStr)),
			diffStr)

		// Add blank line between days (except for last day)
		if dateKey != dateOrder[len(dateOrder)-1] {
			fmt.Println()
		}
	}
}
