package jira

import (
	"fmt"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/context"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/dariuszw/hlp/pkg/duration"
	"github.com/spf13/cobra"
)

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
		t, err := time.Parse("2006-01-02", logDate)
		if err != nil {
			fmt.Println(ui.Error("Invalid date format. Use YYYY-MM-DD"))
			return
		}
		startTime = t
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
}

func runBatchLog(cmd *cobra.Command, args []string) {
	csvPath := batch.DefaultCSVPath()

	// Show current profile
	profile, _ := config.GetActiveProfile()
	if profile != nil {
		modeStr := ui.Success.Render(" [READ/WRITE]")
		if profile.Protected {
			modeStr = ui.WarningText.Render(" [PROTECTED]")
		}
		fmt.Printf("Profile: %s%s\n", ui.Primary.Render(profile.Name), modeStr)
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

	// Get JIRA client (needed for both dry-run and actual processing)
	client, err := getJiraClient()
	if err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// Fetch ticket summaries for descriptions
	uniqueKeys := getUniqueTicketKeys(entries)
	summaries, _ := client.GetIssueSummaries(uniqueKeys)
	for i := range entries {
		if summary, ok := summaries[entries[i].IssueKey]; ok {
			entries[i].Description = summary
		}
	}

	if logDryRun {
		// Preview mode - show entries without processing
		fmt.Println("Processing worklogs " + ui.Muted.Render("(dry run)") + ":")
		for i, e := range entries {
			fmt.Printf("  %s %s %s %s %s ... %s\n",
				ui.Muted.Render(fmt.Sprintf("%d/%d", i+1, len(entries))),
				ui.Primary.Render(fmt.Sprintf("%-12s", e.IssueKey)),
				ui.Muted.Render(fmt.Sprintf("%-30s", truncateString(e.Description, 30))),
				ui.Success.Render(fmt.Sprintf("%-8s", e.TimeSpent)),
				ui.Muted.Render(fmt.Sprintf("%-12s", e.Date)),
				ui.Muted.Render("PENDING"))
		}
		fmt.Println()
		fmt.Println(ui.Warning("Dry run - no entries posted"))
		return
	}

	fmt.Println("Processing worklogs:")

	// Track results for summary
	var failedEntries []batch.Result

	// Process batch
	processor := batch.NewProcessor(client, logTime)
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
}

func runSyncLog(cmd *cobra.Command, args []string) {
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

	// Load existing CSV entries
	csvPath := batch.DefaultCSVPath()
	csvEntries, _ := batch.ParseCSV(csvPath) // Empty if file doesn't exist

	// Find missing entries
	missing := batch.FindMissingWorklogs(jiraWorklogs, csvEntries)

	if len(missing) == 0 {
		fmt.Println(ui.Info("CSV is up to date - no new worklogs found"))
		return
	}

	fmt.Printf("Found %s new worklogs to add\n",
		ui.Success.Render(fmt.Sprintf("%d", len(missing))))

	// Fetch issue summaries for descriptions
	fmt.Print("Fetching issue summaries...")
	uniqueKeys := getUniqueWorklogKeys(missing)
	summaries, _ := client.GetIssueSummaries(uniqueKeys)
	fmt.Println(" done")
	fmt.Println()

	// Preview (if dry-run) or add entries
	if logDryRun {
		fmt.Println("New entries to add " + ui.Muted.Render("(dry run)") + ":")
		for _, wl := range missing {
			desc := summaries[wl.IssueKey]
			fmt.Printf("  %s %s %s %s %s\n",
				ui.Primary.Render("SYNC"),
				ui.Primary.Render(fmt.Sprintf("%-12s", wl.IssueKey)),
				ui.Muted.Render(fmt.Sprintf("%-30s", truncateString(desc, 30))),
				ui.Success.Render(fmt.Sprintf("%-8s", wl.TimeSpentStr)),
				ui.Muted.Render(wl.Started.Format("02.01.2006 15:04")))
		}
		fmt.Println()
		fmt.Println(ui.Warning("Dry run - no changes made"))
		return
	}

	// Append new entries with SYNC status
	fmt.Println("Adding new entries:")
	addedCount := 0
	for _, wl := range missing {
		dateStr := wl.Started.Format("02.01.2006 15:04") // Include time for proper sorting
		desc := summaries[wl.IssueKey]
		if err := batch.AppendEntryWithStatus(csvPath, wl.IssueKey, desc, wl.TimeSpentStr, dateStr, wl.Comment, batch.StatusSync); err != nil {
			fmt.Printf("  %s %s - %s\n",
				ui.ErrorText.Render("FAILED"),
				ui.Primary.Render(wl.IssueKey),
				ui.Muted.Render(err.Error()))
			continue
		}
		addedCount++
		fmt.Printf("  %s %s %s %s %s\n",
			ui.Success.Render("SYNC"),
			ui.Primary.Render(fmt.Sprintf("%-12s", wl.IssueKey)),
			ui.Muted.Render(fmt.Sprintf("%-30s", truncateString(desc, 30))),
			ui.Success.Render(fmt.Sprintf("%-8s", wl.TimeSpentStr)),
			ui.Muted.Render(dateStr))
	}

	// Sort CSV by date
	fmt.Println("\nSorting CSV by date...")
	if err := batch.SortCSVByDate(csvPath); err != nil {
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
