// Package jira provides JIRA worklog commands for the helper CLI.
package jira

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	internalContext "github.com/dariuszw/hlp/internal/context"
	"github.com/dariuszw/hlp/internal/gitsync"
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
	logSlow     bool   // --slow: Add delays between batch entries
	logSchedule string // --schedule: Wait for scheduled time slot
	logOrder    string // --order: Entry order for slow mode (oldest, newest, random)
	logPrune    bool   // --prune: Delete stale DRAFT placeholder rows
)

// getDescriptionWidth returns dynamic description column width based on terminal
func getDescriptionWidth() int {
	width := ui.GetTerminalWidth()
	fixed := 57 // counter + key + time + date + status + spaces
	descWidth := width - fixed
	if descWidth < 15 {
		descWidth = 15
	}
	if descWidth > 50 {
		descWidth = 50
	}
	return descWidth
}

// getContentWidth returns the total width of a content row
func getContentWidth() int {
	// 2 indent + 3 counter + 20 key + descWidth + 8 time + 12 date + 8 status + spaces
	return 2 + 3 + 20 + getDescriptionWidth() + 8 + 12 + 8 + 4
}

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
  hlp jira log --batch --slow               # Post with random delays (oldest first)
  hlp jira log --batch --slow --order=random  # Shuffle entries randomly
  hlp jira log --batch --slow --order=newest  # Process newest dates first
  hlp jira log --sync                       # Sync worklogs from JIRA (last 7 days)
  hlp jira log --sync --from 2025-12-01     # Sync from specific date
  hlp jira log --sync --dry-run             # Preview sync without changes
  hlp jira log --prune                      # Delete stale DRAFT placeholder rows
  hlp jira log --prune --dry-run            # List stale drafts without deleting

Stale drafts are empty DRAFT rows (no time, no comment) on a past day that is
either already fully logged, or older than preferences.draft_retention_days
(default 7). The prune runs automatically after --batch and --sync.

Scheduled mode (requires --batch, --slow is automatic):
  hlp jira log --batch --schedule           # Next available slot (loops)
  hlp jira log --batch --schedule morning   # Wait for morning slot
  hlp jira log --batch --schedule afternoon # Wait for afternoon slot

Entry order (for --slow mode):
  --order=oldest   Process oldest dates first (default)
  --order=newest   Process newest dates first
  --order=random   Shuffle entries randomly`,
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
	logCmd.Flags().BoolVar(&logSlow, "slow", false, "Add 20s-2min random delays between entries")
	logCmd.Flags().StringVar(&logSchedule, "schedule", "", "Wait for scheduled time: 'morning', 'afternoon', or empty for next slot")
	logCmd.Flag("schedule").NoOptDefVal = "next" // bare --schedule (no value) means "next available slot"
	logCmd.Flags().StringVar(&logOrder, "order", "oldest", "Entry order for slow mode: oldest (default), newest, random")
	logCmd.Flags().BoolVar(&logPrune, "prune", false, "Delete stale DRAFT rows (full days, or aged past retention)")
}

func runLog(cmd *cobra.Command, args []string) {
	if logSync {
		runSyncLog(cmd, args)
		return
	}

	if logPrune {
		runPruneLog()
		return
	}

	// Handle --schedule flag (requires --batch)
	if logSchedule != "" {
		if !logBatch {
			fmt.Println(ui.Error("--schedule requires --batch flag"))
			return
		}
		runScheduledBatchLog(cmd, args)
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
		detector := internalContext.NewDetector()
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

// runPruneLog runs the stale-draft cleanup on its own, outside a batch or sync.
func runPruneLog() {
	_, csvPath := loadBatchProfile()

	found, err := pruneStaleDrafts(csvPath, logDryRun)
	if err != nil {
		fmt.Println(ui.Error("Prune failed: " + err.Error()))
		return
	}
	if found == 0 {
		fmt.Println(ui.Info("No stale drafts to remove"))
		return
	}
	if logDryRun {
		fmt.Println(ui.Warning("Dry run - nothing deleted"))
	}
}

// batchRunConfig controls batch execution behavior
type batchRunConfig struct {
	skipConfirmation bool   // Skip protected profile confirmation (for scheduled mode)
	forceSlow        bool   // Force slow mode regardless of --slow flag (for scheduled mode)
	order            string // Entry order for slow mode (oldest, newest, random)
}

// validateOrderFlag returns an error if the order flag value is invalid
func validateOrderFlag(order string) error {
	validOrders := map[string]bool{"oldest": true, "newest": true, "random": true}
	if !validOrders[order] {
		return fmt.Errorf("invalid --order value '%s'. Valid options: oldest, newest, random", order)
	}
	return nil
}

// loadBatchProfile loads the active profile and displays profile info.
// Returns the profile and CSV path. Falls back to default profile on error.
func loadBatchProfile() (*config.JiraProfile, string) {
	profile, err := config.GetActiveProfile()
	if err != nil {
		fmt.Println(ui.Muted.Render("  (using default profile)"))
	}
	csvPath := batch.DefaultCSVPathForProfile(profile)

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

	fmt.Printf("Loading %s...\n", ui.Primary.Render(csvPath))
	return profile, csvPath
}

func runBatchLog(_ *cobra.Command, _ []string) {
	// Validate --order flag early
	if err := validateOrderFlag(logOrder); err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// DRY-RUN MODE: Handle separately with ALL entries (including DRAFT)
	// This must come BEFORE ParsePendingCSV() to avoid filtering out DRAFT entries
	if logDryRun {
		_, csvPath := loadBatchProfile()

		allEntries, err := batch.ParseCSV(csvPath)
		if err != nil {
			fmt.Println(ui.Error("Failed to parse CSV: " + err.Error()))
			return
		}

		if len(allEntries) == 0 {
			fmt.Println(ui.Info("No entries in CSV"))
			return
		}

		// Count pending vs draft for summary
		pendingCount := 0
		draftCount := 0
		for _, e := range allEntries {
			if e.IsDraft() {
				draftCount++
			} else if e.NeedsProcessing() {
				pendingCount++
			}
		}

		fmt.Printf("Found %s entries (%d pending, %d draft)\n",
			ui.Success.Render(fmt.Sprintf("%d", len(allEntries))),
			pendingCount, draftCount)

		// Get JIRA client for fetching summaries
		client, err := getJiraClient()
		if err != nil {
			fmt.Println(ui.Error(err.Error()))
			return
		}

		// Get unique dates from pending entries AND draft entries (regardless of time)
		relevantDates := make(map[string]bool)
		for _, e := range allEntries {
			if e.NeedsProcessing() || e.IsDraft() {
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

		// Sort entries if slow mode is enabled (for preview of processing order)
		if logSlow && logOrder != "" {
			batch.SortEntries(relevantEntries, batch.EntryOrder(logOrder))
		}

		fmt.Println()
		orderInfo := ""
		if logSlow {
			orderInfo = " " + ui.Muted.Render("[order: "+logOrder+"]")
		}
		// Stale drafts are shown as DELETE inside the table below rather than
		// as a separate list, so every entry's fate is visible in one place.
		stale := batch.FindStaleDrafts(allEntries, getExpectedHoursPerDay(), getDraftRetentionDays(), time.Now())
		markedForDeletion := staleRowSet(stale)

		fmt.Println("Processing worklogs " + ui.Muted.Render("(dry run)") + orderInfo + ":")
		showDryRunGroupedByDay(relevantEntries, getExpectedHoursPerDay(), markedForDeletion)
		fmt.Println()

		// Same per-ticket promotion as the real batch run — fixing statuses
		// here only rewrites the CSV; posting still needs a non-dry run.
		if err := promoteDraftsWithTime(csvPath); err != nil {
			fmt.Println(ui.Warning("Draft promotion failed: " + err.Error()))
		}

		if len(stale) > 0 {
			fmt.Println(ui.Muted.Render(fmt.Sprintf(
				"Run without --dry-run to delete the %d row(s) marked DELETE", len(stale))))
		}

		fmt.Println(ui.Warning("Dry run - no entries posted"))
		return
	}

	// Run batch with normal settings
	runBatchLogCore(batchRunConfig{
		skipConfirmation: logConfirm,
		forceSlow:        false,
		order:            logOrder,
	})
}

// runBatchLogCore contains the shared batch processing logic
func runBatchLogCore(cfg batchRunConfig) {
	profile, csvPath := loadBatchProfile()

	// Offer to promote DRAFT rows that already carry logged time — they
	// would otherwise be silently skipped by ParsePendingCSV below.
	// Skipped under -y/--confirm so automated runs stay non-interactive.
	if !cfg.skipConfirmation {
		if err := promoteDraftsWithTime(csvPath); err != nil {
			fmt.Println(ui.Warning("Draft promotion failed: " + err.Error()))
		}
	}

	// BATCH MODE: Use filtered entries (excludes DRAFT and DONE)
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

	// Sort entries if slow mode is enabled
	slowMode := cfg.forceSlow || logSlow
	if slowMode && cfg.order != "" {
		batch.SortEntries(entries, batch.EntryOrder(cfg.order))
	}

	// Show preview before confirmation
	fmt.Println()
	fmt.Println("Processing worklogs " + ui.Muted.Render("(preview)") + ":")
	showWorklogPreview(entries, getExpectedHoursPerDay())

	// Safety guard for protected profiles (unless confirmation is skipped)
	if !cfg.skipConfirmation {
		if profile != nil && profile.Protected {
			if !ui.ConfirmProtectedProfile(profile.Name, profile.BaseURL, len(entries)) {
				fmt.Println(ui.Info("Operation cancelled"))
				return
			}
		}
	}

	fmt.Println()
	fmt.Println("Processing worklogs:")

	// Track results for summary
	var failedEntries []batch.Result

	// Process batch with profile config (mock mode for LOCAL)
	// Use slow mode if forced or if --slow flag is set
	mockMode := profile != nil && profile.IsLocal()
	processor := batch.NewProcessorWithConfig(batch.ProcessorConfig{
		Client:      client,
		DefaultTime: logTime,
		Profile:     profile,
		MockMode:    mockMode,
	})
	results := processor.ProcessBatch(entries, false, slowMode, func(current, total int, result batch.Result) {
		var status string
		if result.Success {
			if result.ErrorMessage != "" {
				status = ui.FormatStatus(result.NewStatus) + " " + ui.Muted.Render(result.ErrorMessage)
			} else {
				status = ui.FormatStatus(result.NewStatus)
			}
		} else {
			status = ui.FormatStatus("FAILED") + " " + ui.Muted.Render("("+result.ErrorMessage+")")
		}

		descWidth := getDescriptionWidth()
		fmt.Printf("  %s %s %s %s %s %s\n",
			ui.Muted.Render(padRight(fmt.Sprintf("%d", current), 2)),
			ui.Primary.Render(padRight(result.Entry.DisplayKey(), 20)),
			ui.Muted.Render(padRight(truncateString(result.Entry.Description, descWidth), descWidth)),
			ui.Success.Render(padRight(result.Entry.TimeSpent, 8)),
			ui.Muted.Render(padRight(result.Entry.Date, 12)),
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

	// Sweep placeholders last, once the run's entries are marked DONE, so a day
	// this batch just completed is cleaned up in the same pass and the prompt
	// doesn't interrupt the result summary. Skipped when confirmation is
	// suppressed, since scheduled runs have no one to answer it.
	if !cfg.skipConfirmation {
		fmt.Println()
		if _, err := pruneStaleDrafts(csvPath, false); err != nil {
			fmt.Println(ui.Warning("Stale draft prune failed: " + err.Error()))
		}
	}

	// Show daily total warnings for batch
	showBatchDailyWarnings(client, results)
}

// runScheduledBatchLog waits for scheduled time slots and runs batch processing
func runScheduledBatchLog(_ *cobra.Command, _ []string) {
	// Validate --order flag early
	if err := validateOrderFlag(logOrder); err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// Safety: require upfront confirmation for protected profiles
	// This runs once at startup; subsequent loop iterations skip re-prompting.
	profile, err := config.GetActiveProfile()
	if err == nil && profile != nil && profile.Protected {
		fmt.Println(ui.Warning("Scheduled mode will post worklogs to a PROTECTED profile"))
		if !ui.ConfirmProtectedProfile(profile.Name, profile.BaseURL, -1) {
			fmt.Println(ui.Info("Scheduled mode cancelled"))
			return
		}
	}

	// Load schedule configuration
	schedule, err := config.GetScheduleConfig()
	if err != nil {
		fmt.Println(ui.Error("Failed to load schedule config: " + err.Error()))
		return
	}

	if len(schedule.Slots) == 0 {
		fmt.Println(ui.Error("No schedule slots configured. Add slots to config.yaml under preferences.schedule.slots"))
		return
	}

	// Determine target slot name (empty string means "next available")
	targetSlot := ""
	if logSchedule != "next" {
		targetSlot = logSchedule
	}

	// Determine if we should loop
	// If targeting a specific slot, don't loop (single execution)
	// If using "next available", respect config.loop setting
	shouldLoop := schedule.Loop && targetSlot == ""

	// Setup signal handling for graceful Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
			fmt.Println("\n\n" + ui.Info("Interrupted - exiting scheduled mode"))
			cancel()
		case <-ctx.Done():
			// Exit goroutine when context is cancelled (normal completion)
		}
	}()
	defer signal.Stop(sigChan)
	defer cancel()

	for {
		// Calculate next scheduled time
		targetTime, slotIdx, err := config.NextScheduledTime(*schedule, targetSlot)
		if err != nil {
			fmt.Println(ui.Error(err.Error()))
			return
		}

		slot := schedule.Slots[slotIdx]

		// Validate slot is within work hours
		if schedule.WorkHoursStart != "" && schedule.WorkHoursEnd != "" {
			if err := config.ValidateSlotWithinWorkHours(slot.Time, schedule.WorkHoursStart, schedule.WorkHoursEnd); err != nil {
				fmt.Println(ui.Warning(err.Error()))
			}
		}

		// Show schedule header
		fmt.Print(ui.FormatScheduleHeader(slot.Name, targetTime.Format("Mon 02 Jan 15:04"), shouldLoop))

		// Wait until scheduled time
		if !batch.WaitUntilTime(ctx, targetTime, func(remaining time.Duration) {
			fmt.Print(ui.FormatScheduleCountdown(slot.Name, targetTime, remaining))
		}) {
			// Cancelled
			return
		}

		// Clear countdown line
		ui.ClearLine()
		fmt.Println(ui.SuccessMsg("Scheduled time reached - starting batch"))
		fmt.Println()

		// Run batch processing with scheduled mode settings
		runBatchLogCore(batchRunConfig{
			skipConfirmation: true,     // Skip confirmation in scheduled mode
			forceSlow:        true,     // Always use slow mode in scheduled mode
			order:            logOrder, // Use configured order
		})

		// The process-exit sync in cmd.Execute() may be hours away in loop
		// mode, so push each batch's worklog changes immediately.
		gitsync.SyncWorklogs(config.GetConfigDir())

		// Exit loop if not looping or targeting specific slot
		if !shouldLoop {
			fmt.Println()
			fmt.Println(ui.Info("Single scheduled run completed"))
			return
		}

		// Check for cancellation before looping
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println()
			fmt.Println(ui.Info("Waiting for next scheduled slot..."))
			fmt.Println()
		}
	}
}

func runSyncLog(_ *cobra.Command, _ []string) {
	// Load log_to_subtask preference (warn-and-continue mirrors runAdd)
	logToSubtask := false
	if cfg, cfgErr := config.Load(); cfgErr != nil {
		fmt.Println(ui.Warning(fmt.Sprintf("Could not load config: %v (using defaults)", cfgErr)))
	} else {
		logToSubtask = cfg.Preferences.LogToSubtask
	}

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
	csvEntries, parseErr := batch.ParseCSV(csvPath)
	if parseErr != nil && !errors.Is(parseErr, os.ErrNotExist) {
		fmt.Println(ui.Error("Failed to parse CSV: " + parseErr.Error()))
		return
	}

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
					ui.FormatStatusPadded("SYNC", 6),
					ui.Primary.Render(padRight(wl.IssueKey, 12)),
					ui.Muted.Render(padRight("["+detail.IssueType+"]", 12)),
					ui.Muted.Render(padRight(truncateString(detail.Summary, 25), 25)),
					ui.Success.Render(padRight(wl.TimeSpentStr, 8)),
					ui.Muted.Render(wl.Started.Format("02.01.2006 15:04")))
			}
			fmt.Println()
		}

		// Preview DRAFT entries
		fmt.Println("Checking assigned tickets for DRAFT entries " + ui.Muted.Render("(dry run)") + ":")
		sprintTickets, err := fetchAssignedTickets(client)
		if err != nil {
			fmt.Println(ui.Warning("Failed to fetch assigned tickets: " + err.Error()))
		} else {
			entries, parseErr := batch.ParseCSV(csvPath)
			if parseErr != nil && !errors.Is(parseErr, os.ErrNotExist) {
				fmt.Println(ui.Warning("Could not parse CSV for draft preview: " + parseErr.Error()))
			} else {
				if errors.Is(parseErr, os.ErrNotExist) {
					fmt.Println(ui.Muted.Render("  No existing CSV found — showing all assigned tickets as DRAFT"))
				}
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
							ui.FormatStatusPadded("DRAFT", 6),
							ui.Primary.Render(padRight(issueKey, 12)),
							ui.Muted.Render(padRight("["+issueType+"]", 12)),
							ui.Muted.Render(padRight(truncateString(description, 25), 25)),
							ui.Muted.Render("("+lastLoggedDate+")"))
					}
				}

				if draftPreviewCount == 0 {
					fmt.Println(ui.Muted.Render("  No new DRAFT entries needed"))
				}
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
			// Sync path has no subtask context (subtaskKey always ""), so
			// subtaskLogInd is always "N" regardless of logToSubtask.
			if err = batch.AppendEntryWithStatus(csvPath, wl.IssueKey, "", detail.IssueType, detail.Summary, wl.TimeSpentStr, dateStr, wl.Comment, "N", batch.StatusSync); err != nil {
				fmt.Printf("  %s %s - %s\n",
					ui.FormatStatus("FAILED"),
					ui.Primary.Render(wl.IssueKey),
					ui.Muted.Render(err.Error()))
				continue
			}
			addedCount++
			fmt.Printf("  %s %s %s %s %s %s\n",
				ui.FormatStatusPadded("SYNC", 6),
				ui.Primary.Render(padRight(wl.IssueKey, 12)),
				ui.Muted.Render(padRight("["+detail.IssueType+"]", 12)),
				ui.Muted.Render(padRight(truncateString(detail.Summary, 25), 25)),
				ui.Success.Render(padRight(wl.TimeSpentStr, 8)),
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
		entries, parseErr := batch.ParseCSV(csvPath)
		if parseErr != nil {
			fmt.Println(ui.Warning("Could not re-read CSV for descriptions: " + parseErr.Error()))
		} else {
			var keysNeedingDesc []string
			for _, e := range entries {
				if e.Description == "" {
					keysNeedingDesc = append(keysNeedingDesc, e.IssueKey)
				}
			}
			if len(keysNeedingDesc) > 0 {
				fmt.Print("Fetching missing descriptions...")
				descSummaries, descErr := client.GetIssueSummaries(unique(keysNeedingDesc))
				if descErr != nil {
					fmt.Println(" " + ui.Warning("failed: "+descErr.Error()))
				} else {
					updatedCount, updateErr := batch.UpdateCSVDescriptions(csvPath, descSummaries)
					if updateErr != nil {
						fmt.Println(" " + ui.Warning("failed to update: "+updateErr.Error()))
					} else if updatedCount > 0 {
						fmt.Printf(" updated %d entries\n", updatedCount)
					} else {
						fmt.Println(" done")
					}
				}
			}
		}

		fmt.Printf("\nSummary: %s entries synced from JIRA\n",
			ui.Success.Render(fmt.Sprintf("%d", addedCount)))
	}

	// Add assigned tickets as DRAFT entries (at TOP of CSV)
	fmt.Println("\nChecking assigned tickets for DRAFT entries...")
	sprintTickets, err := fetchAssignedTickets(client)
	if err != nil {
		fmt.Println(ui.Warning("Failed to fetch assigned tickets: " + err.Error()))
		return
	}

	// Re-parse CSV to get updated entries after sync
	entries, parseErr := batch.ParseCSV(csvPath)
	if parseErr != nil && !errors.Is(parseErr, os.ErrNotExist) {
		fmt.Println(ui.Error("Failed to re-read CSV for draft entries: " + parseErr.Error()))
		return
	}
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

	// A day that already holds a full day of work needs no new placeholders.
	// Without this guard, sync would re-create exactly the drafts the prune
	// step deletes from that day, and the two would fight every run.
	if lastDay, dayErr := batch.ParseDate(lastLoggedDate, "09:00"); dayErr == nil &&
		batch.DayIsCovered(batch.LoggedTimeByDate(entries), lastDay, getExpectedHoursPerDay()) {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("  %s is fully logged - no new drafts needed", lastLoggedDate)))
		sprintTickets = nil
	}

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
				"",
				draftDateTime,
				"",
				subtaskLogInd,
				batch.StatusDraft,
			)
			if err == nil {
				draftCount++
				fmt.Printf("  %s %s %s %s\n",
					ui.FormatStatusPadded("DRAFT", 6),
					ui.Primary.Render(padRight(issueKey, 12)),
					ui.Muted.Render(padRight("["+issueType+"]", 12)),
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

	fmt.Println()
	if _, err := pruneStaleDrafts(csvPath, false); err != nil {
		fmt.Println(ui.Warning("Stale draft prune failed: " + err.Error()))
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

// fetchAssignedTickets returns the user's tracked tickets using the configurable
// preferences.ticket_fetch_jql. Defaults to all assigned, non-Done tickets so
// non-sprint work (e.g. Escalations board) is included. The literal fallback
// mirrors the config default so the helper is safe even if config.Load() errors.
func fetchAssignedTickets(client *internalJira.Client) ([]internalJira.Ticket, error) {
	jql := "assignee = currentUser() AND status != Done ORDER BY updated DESC"
	if cfg, err := config.Load(); err == nil && cfg.Preferences.TicketFetchJQL != "" {
		jql = cfg.Preferences.TicketFetchJQL
	}
	return client.Search(jql, 50)
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
// Uses rune count for correct handling of multibyte UTF-8 characters
// TO REVIEW: Could move to shared utils (also in add.go) - skipped: only 2 occurrences
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// padRight pads a string to width using rune count for consistent visual alignment
func padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(runes))
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
		expectedStr := "8h"
		if cfg, err := config.Load(); err == nil {
			if cfg.Preferences.ExpectedHoursPerDay != "" {
				expectedStr = cfg.Preferences.ExpectedHoursPerDay
			}
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

// printWorklogRow renders one entry as a table row. Rows already staged for
// deletion report DELETE rather than DRAFT, so the main list is the single place
// that shows what will happen to every entry.
func printWorklogRow(num int, e batch.Entry, markedForDeletion bool) {
	statusDisplay := e.Status
	if statusDisplay == "" {
		statusDisplay = "PENDING"
	}
	timeDisplay := e.TimeSpent
	if timeDisplay == "" {
		timeDisplay = "—"
	}

	rowStyle := ui.Muted
	statusCol := ui.Muted.Render(statusDisplay)

	switch {
	case markedForDeletion:
		statusDisplay = "DELETE"
		rowStyle = ui.ErrorText
		statusCol = ui.FormatStatus(statusDisplay)
	// A DRAFT with hours filled in is almost certainly a forgotten status flip —
	// it would be silently skipped, so paint the whole row in warning color.
	case isDraftWithTime(e):
		rowStyle = ui.WarningText
		statusCol = ui.WarningBold.Render(statusDisplay) + " " + ui.WarningText.Render("⚠")
	case e.Status == batch.StatusDone:
		rowStyle = ui.Success
		statusCol = ui.Success.Render(statusDisplay)
	case e.NeedsProcessing():
		rowStyle = ui.Primary
		statusCol = ui.FormatStatus("PENDING")
	}

	descWidth := getDescriptionWidth()
	fmt.Printf("  %s %s %s %s %s %s\n",
		rowStyle.Render(padRight(fmt.Sprintf("%d", num), 2)),
		rowStyle.Render(padRight(e.DisplayKey(), 20)),
		rowStyle.Render(padRight(truncateString(e.Description, descWidth), descWidth)),
		rowStyle.Render(padRight(timeDisplay, 8)),
		rowStyle.Render(padRight(strings.Split(e.Date, " ")[0], 12)),
		statusCol)
}

// staleRowSet indexes stale drafts by CSV row number so the table renderer can
// mark them without re-running the scan per row.
func staleRowSet(stale []batch.StaleDraft) map[int]bool {
	rows := make(map[int]bool, len(stale))
	for _, s := range stale {
		rows[s.Entry.RowNumber] = true
	}
	return rows
}

// entriesOnDaysOf returns every entry falling on a day that contains at least
// one stale draft, so a deletion prompt shows those rows in their day's context.
func entriesOnDaysOf(entries []batch.Entry, stale []batch.StaleDraft) []batch.Entry {
	days := make(map[string]bool, len(stale))
	for _, s := range stale {
		days[strings.Split(s.Entry.Date, " ")[0]] = true
	}

	var onDays []batch.Entry
	for _, e := range entries {
		if days[strings.Split(e.Date, " ")[0]] {
			onDays = append(onDays, e)
		}
	}
	return onDays
}

// pruneReasonSummary collapses the per-row reasons into one line, since within a
// single run they are nearly always the same handful of values.
func pruneReasonSummary(stale []batch.StaleDraft) string {
	var reasons []string
	seen := make(map[string]bool)
	for _, s := range stale {
		if !seen[s.Reason] {
			seen[s.Reason] = true
			reasons = append(reasons, s.Reason)
		}
	}
	return strings.Join(reasons, ", ")
}

// showDryRunGroupedByDay shows the worklog table followed by a summary line.
func showDryRunGroupedByDay(entries []batch.Entry, expectedHours time.Duration, markedForDeletion map[int]bool) {
	showWorklogTable(entries, expectedHours, markedForDeletion)
	showWorklogSummary(entries, markedForDeletion)
}

// dayGroup collects one calendar day's entries together with its logged total.
type dayGroup struct {
	date    string
	day     time.Time
	entries []batch.Entry
	total   time.Duration
}

// groupEntriesByDay buckets entries per calendar day, oldest day first, so the
// most recent day prints last — next to the summary and any prompt, where it is
// visible without scrolling back up. Sorting is on the parsed date rather than
// the CSV's own order, so the display is stable however the file is arranged.
func groupEntriesByDay(entries []batch.Entry, newestFirst bool) []*dayGroup {
	byDate := make(map[string]*dayGroup)
	var groups []*dayGroup

	for _, e := range entries {
		t, err := batch.ParseDate(e.Date, "09:00")
		if err != nil {
			continue
		}
		dateKey := t.Format("02.01.2006")

		group, exists := byDate[dateKey]
		if !exists {
			group = &dayGroup{date: dateKey, day: t}
			byDate[dateKey] = group
			groups = append(groups, group)
		}
		group.entries = append(group.entries, e)

		dur, _ := duration.Parse(e.TimeSpent)
		group.total += dur
	}

	sort.SliceStable(groups, func(i, j int) bool {
		if newestFirst {
			return groups[j].day.Before(groups[i].day)
		}
		return groups[i].day.Before(groups[j].day)
	})
	return groups
}

// showDayTotal prints the per-day divider, total, and variance from expected.
func showDayTotal(group *dayGroup, expectedHours time.Duration) {
	fmt.Println("  " + ui.Divider(getContentWidth()-2))

	diff := group.total - expectedHours
	var diffStr string
	switch {
	case diff > 0:
		diffStr = ui.WarningText.Render(fmt.Sprintf("(+%s over %s)", duration.Format(diff), duration.Format(expectedHours)))
	case diff < 0:
		diffStr = ui.ErrorText.Render(fmt.Sprintf("(need %s for %s)", duration.Format(-diff), duration.Format(expectedHours)))
	default:
		diffStr = ui.Success.Render("✓")
	}

	// Align total under the time column: 2 + 2 + 1 + 20 + 1 + descWidth + 1
	timeColStart := 27 + getDescriptionWidth()
	fmt.Printf("%s%s %s\n",
		strings.Repeat(" ", timeColStart),
		ui.Success.Render(padRight(duration.Format(group.total), 8)),
		diffStr)
}

// showWorklogTable shows entries grouped by day with per-day totals. Rows in
// markedForDeletion (keyed by CSV row number) render as DELETE.
func showWorklogTable(entries []batch.Entry, expectedHours time.Duration, markedForDeletion map[int]bool) {
	groups := groupEntriesByDay(entries, previewNewestFirst())

	entryNum := 0
	for i, group := range groups {
		for _, e := range group.entries {
			entryNum++
			printWorklogRow(entryNum, e, markedForDeletion[e.RowNumber])
		}
		showDayTotal(group, expectedHours)

		if i < len(groups)-1 {
			fmt.Println()
		}
	}
}

// previewNewestFirst reports whether the day order should be flipped. Days print
// oldest-first by default; --slow --order=newest is the one case where the
// preview must mirror the processing order instead.
func previewNewestFirst() bool {
	return logSlow && logOrder == "newest"
}

// showWorklogSummary counts entries by fate. Rows staged for deletion are
// counted as DELETE, not as drafts, so the totals match the table above.
func showWorklogSummary(entries []batch.Entry, markedForDeletion map[int]bool) {
	pendingCount := 0
	draftCount := 0
	draftWithTimeCount := 0
	deleteCount := 0
	var pendingTime time.Duration

	for _, e := range entries {
		if markedForDeletion[e.RowNumber] {
			deleteCount++
		} else if e.IsDraft() {
			draftCount++
			if isDraftWithTime(e) {
				draftWithTimeCount++
			}
		} else if e.NeedsProcessing() {
			pendingCount++
			dur, _ := duration.Parse(e.TimeSpent)
			pendingTime += dur
		}
	}

	fmt.Println()
	summary := fmt.Sprintf("Summary: %s pending (%s will be posted)",
		ui.Success.Render(fmt.Sprintf("%d", pendingCount)),
		duration.Format(pendingTime))
	if draftCount > 0 {
		summary += fmt.Sprintf(", %s drafts (skipped)", ui.Muted.Render(fmt.Sprintf("%d", draftCount)))
	}
	if deleteCount > 0 {
		summary += fmt.Sprintf(", %s marked DELETE", ui.ErrorText.Render(fmt.Sprintf("%d", deleteCount)))
	}
	fmt.Println(summary)

	if draftWithTimeCount > 0 {
		fmt.Println(ui.Warning(fmt.Sprintf(
			"%d draft(s) have time logged but will NOT be posted — promote them below",
			draftWithTimeCount)))
	}
}

// isDraftWithTime reports a DRAFT row that already has hours filled in —
// the "forgot to flip the status" state the dry-run highlights so the user
// fixes it before the real batch run silently skips it.
func isDraftWithTime(e batch.Entry) bool {
	return e.IsDraft() && strings.TrimSpace(e.TimeSpent) != ""
}

// promoteDraftsWithTime lists DRAFT entries that already carry logged time
// and, after confirmation, flips their status to empty (pending) in the CSV
// so the current batch run posts them. Declining leaves the CSV untouched.
func promoteDraftsWithTime(csvPath string) error {
	allEntries, err := batch.ParseCSV(csvPath)
	if err != nil {
		return err
	}

	var drafts []batch.Entry
	for _, e := range allEntries {
		if isDraftWithTime(e) {
			drafts = append(drafts, e)
		}
	}
	if len(drafts) == 0 {
		return nil
	}

	fmt.Println(ui.Warning(fmt.Sprintf("%d draft(s) have time logged:", len(drafts))))
	descWidth := getDescriptionWidth()
	for _, e := range drafts {
		fmt.Printf("  %s %s %s %s\n",
			ui.WarningText.Render(padRight(e.DisplayKey(), 20)),
			ui.WarningText.Render(padRight(truncateString(e.Description, descWidth), descWidth)),
			ui.WarningBold.Render(padRight(e.TimeSpent, 8)),
			ui.WarningText.Render(strings.Split(e.Date, " ")[0]))
	}

	choice, err := ui.PromptChoice(fmt.Sprintf(
		"Change to PENDING? [%s]: ",
		ui.Muted.Render("y=all / i=individually / N=none")))
	if err != nil {
		choice = ""
	}

	rows := selectRows(choice, drafts, func(e batch.Entry) bool {
		return ui.ConfirmAction(fmt.Sprintf("  Change %s to PENDING (remove DRAFT)?", e.DisplayKey()))
	})
	if len(rows) == 0 {
		fmt.Println(ui.Muted.Render("Drafts left unchanged (still skipped)"))
		return nil
	}

	if err := batch.SetCSVStatusByRows(csvPath, rows, ""); err != nil {
		return err
	}
	fmt.Printf("Changed %s of %d draft(s) to PENDING in the CSV %s\n\n",
		ui.Success.Render(fmt.Sprintf("%d", len(rows))),
		len(drafts),
		ui.Muted.Render("(status only — nothing posted yet)"))
	return nil
}

// selectRows maps a bulk-prompt answer to the CSV rows to act on:
// y/yes selects every entry, i/individual asks confirmOne per entry, and
// anything else (the N default) selects none. Shared by the promote and prune
// flows so both answer the same keys.
func selectRows(choice string, entries []batch.Entry, confirmOne func(batch.Entry) bool) []int {
	var rows []int
	switch choice {
	case "y", "yes":
		for _, e := range entries {
			rows = append(rows, e.RowNumber)
		}
	case "i", "individual":
		for _, e := range entries {
			if confirmOne(e) {
				rows = append(rows, e.RowNumber)
			}
		}
	}
	return rows
}

// getDraftRetentionDays returns how many days a DRAFT placeholder survives on a
// day that never filled up. A missing or nonsensical config value falls back to
// the default rather than pruning everything or nothing.
func getDraftRetentionDays() int {
	cfg, err := config.Load()
	if err != nil || cfg.Preferences.DraftRetentionDays <= 0 {
		return config.DefaultDraftRetentionDays
	}
	return cfg.Preferences.DraftRetentionDays
}

// pruneStaleDrafts deletes DRAFT rows that can no longer become worklogs: their
// day is already fully logged, or it aged past the retention window without
// filling up. Rows carrying time or a comment are never offered — those hold
// something the user typed. In report mode nothing is written, since silently
// deleting rows during a --dry-run would be a surprise.
// Returns how many stale drafts were found, so callers can distinguish "nothing
// to do" from "found some" without re-parsing the CSV.
func pruneStaleDrafts(csvPath string, reportOnly bool) (int, error) {
	entries, err := batch.ParseCSV(csvPath)
	if err != nil {
		return 0, err
	}

	stale := batch.FindStaleDrafts(entries, getExpectedHoursPerDay(), getDraftRetentionDays(), time.Now())
	if len(stale) == 0 {
		return 0, nil
	}

	marked := staleRowSet(stale)

	verb := "will be removed"
	if reportOnly {
		verb = "would be removed"
	}
	fmt.Printf("%s stale draft(s) %s %s:\n",
		ui.ErrorText.Render(fmt.Sprintf("%d", len(stale))),
		verb,
		ui.Muted.Render("(shown as DELETE below)"))
	fmt.Println()

	// Show the affected days in full, not just the doomed rows, so the reason a
	// day qualified (its DONE entries adding up) is visible next to them.
	showWorklogTable(entriesOnDaysOf(entries, stale), getExpectedHoursPerDay(), marked)
	fmt.Println()
	fmt.Println(ui.Muted.Render("Reason: " + pruneReasonSummary(stale)))

	if reportOnly {
		fmt.Println(ui.Muted.Render("(run without --dry-run to delete them)"))
		fmt.Println()
		return len(stale), nil
	}

	choice, err := ui.PromptChoice(fmt.Sprintf(
		"Delete? [%s]: ",
		ui.Muted.Render("y=all / i=individually / N=none")))
	if err != nil {
		choice = ""
	}

	candidates := make([]batch.Entry, len(stale))
	for i, s := range stale {
		candidates[i] = s.Entry
	}
	rows := selectRows(choice, candidates, func(e batch.Entry) bool {
		return ui.ConfirmAction(fmt.Sprintf("  Delete %s (%s)?", e.DisplayKey(), strings.Split(e.Date, " ")[0]))
	})
	if len(rows) == 0 {
		fmt.Println(ui.Muted.Render("Drafts left unchanged"))
		return len(stale), nil
	}

	if err := batch.RemoveEntriesByRows(csvPath, rows); err != nil {
		return len(stale), err
	}
	fmt.Printf("Removed %s of %d stale draft(s) %s\n\n",
		ui.Success.Render(fmt.Sprintf("%d", len(rows))),
		len(stale),
		ui.Muted.Render("(a backup was written to worklogs.csv.autobak)"))
	return len(stale), nil
}

// showWorklogPreview displays pending entries before the confirmation prompt.
func showWorklogPreview(entries []batch.Entry, expectedHours time.Duration) {
	if len(entries) == 0 {
		return
	}
	showWorklogTable(entries, expectedHours, nil)
}
