package timesheet

import (
	"fmt"
	"time"

	"github.com/dariuszw/hlp/internal/context"
	"github.com/dariuszw/hlp/internal/timesheet"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/dariuszw/hlp/pkg/duration"
	"github.com/spf13/cobra"
)

var (
	logComment string
	logDate    string
	logTime    string
)

var logCmd = &cobra.Command{
	Use:   "log <ticket> <duration>",
	Short: "Log time to a ticket via Timesheet plugin",
	Long: `Log work time to a Jira ticket using the Timesheet Tracking plugin.

Duration format: 30m, 2h, 1h30m, 1d, etc.

Examples:
  hlp timesheet log TDT-123 2h
  hlp timesheet log TDT-123 2h --comment "Fixed bug"
  hlp timesheet log TDT-123 8h --date 2025-12-01
  hlp timesheet log TDT-123 30m --time 14:30`,
	Args: cobra.RangeArgs(1, 2),
	Run:  runLog,
}

func init() {
	logCmd.Flags().StringVarP(&logComment, "comment", "c", "", "work log comment")
	logCmd.Flags().StringVar(&logDate, "date", "", "date for worklog (YYYY-MM-DD, default: today)")
	logCmd.Flags().StringVar(&logTime, "time", "", "start time (HH:MM, default: now)")
}

func runLog(cmd *cobra.Command, args []string) {
	// Get ticket and duration
	var ticket, durationStr string

	if len(args) == 2 {
		ticket = args[0]
		durationStr = args[1]
	} else if len(args) == 1 {
		// Could be just duration with auto-detected ticket
		durationStr = args[0]
		detector := context.NewDetector()
		ticket, _ = detector.DetectTicket()
	}

	if ticket == "" {
		fmt.Println(ui.Error("Ticket is required. Usage: hlp timesheet log <ticket> <duration>"))
		return
	}

	if durationStr == "" {
		fmt.Println(ui.Error("Duration is required. Usage: hlp timesheet log <ticket> <duration>"))
		return
	}

	// Validate duration
	dur, err := duration.Parse(durationStr)
	if err != nil {
		fmt.Println(ui.Error("Invalid duration format: " + err.Error()))
		return
	}

	// Parse date/time
	startTime := time.Now()

	if logDate != "" {
		t, err := time.Parse("2006-01-02", logDate)
		if err != nil {
			fmt.Println(ui.Error("Invalid date format. Use YYYY-MM-DD"))
			return
		}
		startTime = time.Date(t.Year(), t.Month(), t.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, startTime.Location())
	}

	if logTime != "" {
		var hour, minute int
		if _, err := fmt.Sscanf(logTime, "%d:%d", &hour, &minute); err != nil {
			fmt.Println(ui.Error("Invalid time format. Use HH:MM"))
			return
		}
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(),
			hour, minute, 0, 0, startTime.Location())
	}

	// Load timesheet config
	cfg := loadTimesheetConfig()
	if cfg.JiraURL == "" {
		fmt.Println(ui.Error("Timesheet not configured. Run: hlp timesheet setup"))
		return
	}

	// Create client and log work
	client := timesheet.NewClient(cfg)

	entry := timesheet.WorklogEntry{
		IssueKey:  ticket,
		TimeSpent: duration.ToJiraFormat(dur),
		Started:   startTime,
		Comment:   logComment,
	}

	fmt.Printf("%s Logging %s to %s via Timesheet plugin...\n",
		ui.SpinnerFrames[0],
		duration.Format(dur),
		ui.Primary.Render(ticket))

	if err := client.LogWork(entry); err != nil {
		fmt.Println(ui.Error("Failed to log work: " + err.Error()))
		return
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Logged %s to %s", duration.Format(dur), ticket)))
	if logComment != "" {
		fmt.Println(ui.Muted.Render("  Comment: " + logComment))
	}
}
