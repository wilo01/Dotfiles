package timesheet

import (
	"fmt"
	"time"

	"github.com/dariuszw/hlp/internal/timesheet"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/dariuszw/hlp/pkg/duration"
	"github.com/spf13/cobra"
)

var (
	batchMon     string
	batchTue     string
	batchWed     string
	batchThu     string
	batchFri     string
	batchComment string
	batchDryRun  bool
	batchWeek    int // Week offset (0 = current, -1 = last week)
)

var batchCmd = &cobra.Command{
	Use:   "batch <ticket>",
	Short: "Batch log time for a week",
	Long: `Log time entries for multiple days at once.

Useful for entering a full week's timesheets in one command.

Examples:
  hlp timesheet batch TDT-123 --mon 8h --tue 7h --wed 8h --thu 8h --fri 7h
  hlp timesheet batch TDT-123 --mon 8h --tue 8h --comment "Sprint work"
  hlp timesheet batch TDT-123 --mon 8h --week -1  # Last week's Monday`,
	Args: cobra.ExactArgs(1),
	Run:  runBatch,
}

func init() {
	batchCmd.Flags().StringVar(&batchMon, "mon", "", "Monday hours (e.g., 8h)")
	batchCmd.Flags().StringVar(&batchTue, "tue", "", "Tuesday hours")
	batchCmd.Flags().StringVar(&batchWed, "wed", "", "Wednesday hours")
	batchCmd.Flags().StringVar(&batchThu, "thu", "", "Thursday hours")
	batchCmd.Flags().StringVar(&batchFri, "fri", "", "Friday hours")
	batchCmd.Flags().StringVarP(&batchComment, "comment", "c", "", "comment for all entries")
	batchCmd.Flags().BoolVar(&batchDryRun, "dry-run", false, "preview without posting")
	batchCmd.Flags().IntVar(&batchWeek, "week", 0, "week offset (0=current, -1=last week)")
}

func runBatch(cmd *cobra.Command, args []string) {
	ticket := args[0]

	// Parse day entries
	type dayEntry struct {
		name     string
		duration string
		weekday  time.Weekday
	}

	days := []dayEntry{
		{"Monday", batchMon, time.Monday},
		{"Tuesday", batchTue, time.Tuesday},
		{"Wednesday", batchWed, time.Wednesday},
		{"Thursday", batchThu, time.Thursday},
		{"Friday", batchFri, time.Friday},
	}

	// Filter to only days with values
	var entries []dayEntry
	for _, d := range days {
		if d.duration != "" {
			// Validate duration
			if _, err := duration.Parse(d.duration); err != nil {
				fmt.Println(ui.Error(fmt.Sprintf("Invalid duration for %s: %s", d.name, err.Error())))
				return
			}
			entries = append(entries, d)
		}
	}

	if len(entries) == 0 {
		fmt.Println(ui.Error("No days specified. Use --mon, --tue, etc."))
		return
	}

	// Calculate dates for each day
	now := time.Now()
	// Find the Monday of the current week
	daysFromMonday := int(now.Weekday()) - int(time.Monday)
	if daysFromMonday < 0 {
		daysFromMonday += 7
	}
	monday := now.AddDate(0, 0, -daysFromMonday)

	// Apply week offset
	if batchWeek != 0 {
		monday = monday.AddDate(0, 0, batchWeek*7)
	}

	// Preview
	fmt.Println(ui.Header("Timesheet Batch"))
	fmt.Println()
	fmt.Println(ui.KeyValue("Ticket", ui.Primary.Render(ticket)))
	fmt.Println(ui.KeyValue("Week of", monday.Format("2006-01-02")))
	if batchComment != "" {
		fmt.Println(ui.KeyValue("Comment", batchComment))
	}
	fmt.Println()

	var totalDuration time.Duration
	for _, e := range entries {
		dur, _ := duration.Parse(e.duration)
		totalDuration += dur
		dayDate := monday.AddDate(0, 0, int(e.weekday)-int(time.Monday))
		fmt.Printf("  %s %s %s\n",
			ui.Muted.Render(e.name[:3]),
			dayDate.Format("2006-01-02"),
			ui.Success.Render(duration.Format(dur)))
	}

	fmt.Println()
	fmt.Println(ui.KeyValue("Total", ui.SuccessBold.Render(duration.Format(totalDuration))))
	fmt.Println()

	if batchDryRun {
		fmt.Println(ui.Warning("Dry run - no entries posted"))
		return
	}

	// Load config and create client
	cfg := loadTimesheetConfig()
	if cfg.JiraURL == "" {
		fmt.Println(ui.Error("Timesheet not configured. Run: hlp timesheet setup"))
		return
	}

	client := timesheet.NewClient(cfg)

	// Post entries
	fmt.Println(ui.Info("Posting entries..."))

	successCount := 0
	for _, e := range entries {
		dur, _ := duration.Parse(e.duration)
		dayDate := monday.AddDate(0, 0, int(e.weekday)-int(time.Monday))
		// Set time to 9:00 AM
		startTime := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), 9, 0, 0, 0, dayDate.Location())

		entry := timesheet.WorklogEntry{
			IssueKey:  ticket,
			TimeSpent: duration.ToJiraFormat(dur),
			Started:   startTime,
			Comment:   batchComment,
		}

		fmt.Printf("  %s %s...", e.name[:3], duration.Format(dur))

		if err := client.LogWork(entry); err != nil {
			fmt.Println(" " + ui.Error(err.Error()))
		} else {
			fmt.Println(" " + ui.SuccessMsg("OK"))
			successCount++
		}
	}

	fmt.Println()
	if successCount == len(entries) {
		fmt.Println(ui.SuccessMsg(fmt.Sprintf("All %d entries posted successfully!", successCount)))
	} else {
		fmt.Println(ui.Warning(fmt.Sprintf("%d/%d entries posted", successCount, len(entries))))
	}
}
