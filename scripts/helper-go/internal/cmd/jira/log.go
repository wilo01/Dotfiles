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
	"github.com/spf13/viper"
)

var (
	logTicket  string
	logComment string
	logDate    string
	logTime    string
	logBatch   bool
	logDryRun  bool
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
  hlp jira log --batch --dry-run            # Preview batch without posting`,
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
}

func runLog(cmd *cobra.Command, args []string) {
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

	// Parse CSV
	entries, err := batch.ParsePendingCSV(csvPath)
	if err != nil {
		fmt.Println(ui.Error("Failed to parse CSV: " + err.Error()))
		return
	}

	if len(entries) == 0 {
		fmt.Println(ui.Info("No pending entries in " + csvPath))
		return
	}

	fmt.Println(ui.Header("Batch Worklog"))
	fmt.Println()
	fmt.Printf("Found %d pending entries\n\n", len(entries))

	// Preview entries
	for _, e := range entries {
		fmt.Printf("  %s %s %s %s\n",
			ui.Primary.Render(e.IssueKey),
			ui.Success.Render(e.TimeSpent),
			ui.Muted.Render(e.Date),
			ui.Muted.Render(e.Comment))
	}
	fmt.Println()

	if logDryRun {
		fmt.Println(ui.Warning("Dry run - no entries posted"))
		return
	}

	// Get JIRA client
	client, err := getJiraClient()
	if err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// Process batch
	processor := batch.NewProcessor(client, logTime)
	results := processor.ProcessBatch(entries, false, func(current, total int, result batch.Result) {
		if result.Success {
			fmt.Printf("  %s %s %s\n",
				ui.StatusIcon(true),
				result.Entry.IssueKey,
				result.Entry.TimeSpent)
		} else {
			fmt.Printf("  %s %s: %s\n",
				ui.StatusIcon(false),
				result.Entry.IssueKey,
				result.ErrorMessage)
		}
	})

	// Update CSV
	if err := batch.UpdateCSVStatus(csvPath, results); err != nil {
		fmt.Println(ui.Warning("Failed to update CSV: " + err.Error()))
	}

	// Summary
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	fmt.Println()
	if successCount == len(entries) {
		fmt.Println(ui.SuccessMsg(fmt.Sprintf("All %d entries posted!", successCount)))
	} else {
		fmt.Println(ui.Warning(fmt.Sprintf("%d/%d entries posted", successCount, len(entries))))
	}
}

func getJiraClient() (*internalJira.Client, error) {
	baseURL := viper.GetString("jira.base_url")
	email := viper.GetString("jira.email")

	if baseURL == "" || email == "" {
		return nil, fmt.Errorf("JIRA not configured. Run: hlp jira config")
	}

	// Get token from keyring or environment
	credMgr := config.NewCredentialManager()
	token, err := credMgr.GetJiraToken()
	if err != nil {
		return nil, fmt.Errorf("API token not found. Run: hlp jira config")
	}

	return internalJira.NewClient(baseURL, email, token), nil
}
