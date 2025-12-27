package standup

import (
	"fmt"
	"os"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/standupgen"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	generateDays   int
	generateDate   string
	generateOutput string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate standup report from worklog data",
	Long: `Generate a markdown standup report from your worklog CSV.

Default behavior:
  - Monday: Shows Friday's work (skips weekend)
  - Tue-Fri: Shows last 2 days of work

Examples:
  hlp standup generate              # Smart default (2 days, or Friday on Monday)
  hlp standup generate --days 3     # Last 3 days
  hlp standup generate --date 2025-12-23  # Specific date only
  hlp standup generate -o standup.md     # Save to file`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().IntVarP(&generateDays, "days", "d", 0, "Number of days to include (0 = smart default)")
	generateCmd.Flags().StringVar(&generateDate, "date", "", "Specific date (YYYY-MM-DD)")
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Output file path (default: stdout)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Determine CSV path based on active profile
	profile, _ := config.GetActiveProfile()
	csvPath := batch.DefaultCSVPathForProfile(profile)

	// Check if CSV exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return fmt.Errorf("worklog CSV not found: %s", csvPath)
	}

	// Parse all entries from CSV
	entries, err := batch.ParseCSV(csvPath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println(ui.Warning("No entries found in worklog CSV"))
		return nil
	}

	// Determine date range
	startDate, endDate := getDateRange()

	// Filter entries by date range
	filtered := filterEntriesByDateRange(entries, startDate, endDate)

	if len(filtered) == 0 {
		fmt.Println(ui.Warning(fmt.Sprintf("No entries found for %s to %s",
			startDate.Format("02 Jan 2006"),
			endDate.Format("02 Jan 2006"))))
		return nil
	}

	// Generate report
	report := standupgen.BuildReport(filtered, startDate, endDate)
	markdown := standupgen.GenerateMarkdown(report)

	// Output
	if generateOutput != "" {
		if err := os.WriteFile(generateOutput, []byte(markdown), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Println(ui.SuccessMsg(fmt.Sprintf("Report saved to %s", generateOutput)))
	} else {
		fmt.Println(markdown)
	}

	return nil
}

// getDateRange determines the date range based on flags or smart defaults
func getDateRange() (start, end time.Time) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// Specific date requested
	if generateDate != "" {
		parsed, err := time.Parse("2006-01-02", generateDate)
		if err != nil {
			// Try other formats
			parsed, err = time.Parse("02.01.2006", generateDate)
			if err != nil {
				fmt.Println(ui.Warning(fmt.Sprintf("Invalid date format: %s, using today", generateDate)))
				parsed = today
			}
		}
		// For specific date, show just that day
		return parsed, parsed
	}

	// Days flag specified
	if generateDays > 0 {
		start = today.AddDate(0, 0, -generateDays)
		end = today.AddDate(0, 0, -1) // Up to yesterday
		return start, end
	}

	// Smart default: 2 days, or Friday on Monday
	return getSmartDefaultRange(today)
}

// getSmartDefaultRange returns the default date range based on day of week
// Monday: Friday (skip weekend) = 3 days back to 1 day back
// Tue-Fri: Yesterday and day before = 2 days back to 1 day back
func getSmartDefaultRange(today time.Time) (start, end time.Time) {
	end = today.AddDate(0, 0, -1) // Yesterday

	switch today.Weekday() {
	case time.Monday:
		// On Monday, go back to Friday (skip Sat/Sun)
		start = today.AddDate(0, 0, -3) // Friday
	case time.Saturday:
		// On Saturday, include Thu & Fri
		start = today.AddDate(0, 0, -2)
	case time.Sunday:
		// On Sunday, include Thu & Fri
		start = today.AddDate(0, 0, -3)
		end = today.AddDate(0, 0, -2)
	default:
		// Tue-Fri: last 2 days
		start = today.AddDate(0, 0, -2)
	}

	return start, end
}

// filterEntriesByDateRange filters entries to include only those within the date range
func filterEntriesByDateRange(entries []batch.Entry, start, end time.Time) []batch.Entry {
	var filtered []batch.Entry

	for _, e := range entries {
		entryDate, err := batch.ParseDate(e.Date, "00:00")
		if err != nil {
			continue
		}

		// Normalize to just date (no time)
		entryDay := time.Date(entryDate.Year(), entryDate.Month(), entryDate.Day(), 0, 0, 0, 0, time.Local)
		startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
		endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.Local)

		// Include if entry is within range (inclusive)
		if (entryDay.Equal(startDay) || entryDay.After(startDay)) &&
			(entryDay.Equal(endDay) || entryDay.Before(endDay)) {
			filtered = append(filtered, e)
		}
	}

	return filtered
}
