// Package standupgen provides standup report generation from worklog entries
package standupgen

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/pkg/duration"
)

// Report represents a standup report
type Report struct {
	StartDate  time.Time
	EndDate    time.Time
	Entries    []ReportEntry
	TotalTime  time.Duration
	TaskCount  int
	DayCount   int
}

// ReportEntry represents a single ticket in the report
type ReportEntry struct {
	IssueKey    string
	Description string
	TimeSpent   time.Duration
	Comments    []string // Multiple worklogs may have different comments
	Date        time.Time
}

// BuildReport creates a Report from batch entries
func BuildReport(entries []batch.Entry, start, end time.Time) *Report {
	report := &Report{
		StartDate: start,
		EndDate:   end,
	}

	// Group entries by IssueKey
	grouped := make(map[string]*ReportEntry)
	days := make(map[string]bool)

	for _, e := range entries {
		// Parse duration
		dur, err := duration.Parse(e.TimeSpent)
		if err != nil {
			continue // Skip invalid entries
		}

		// Parse date for tracking unique days
		entryDate, _ := batch.ParseDate(e.Date, "00:00")
		dayKey := entryDate.Format("2006-01-02")
		days[dayKey] = true

		key := e.IssueKey
		if existing, ok := grouped[key]; ok {
			// Add to existing entry
			existing.TimeSpent += dur
			if e.Comment != "" && !contains(existing.Comments, e.Comment) {
				existing.Comments = append(existing.Comments, e.Comment)
			}
		} else {
			// Create new entry
			grouped[key] = &ReportEntry{
				IssueKey:    e.IssueKey,
				Description: e.Description,
				TimeSpent:   dur,
				Comments:    filterEmpty([]string{e.Comment}),
				Date:        entryDate,
			}
		}

		report.TotalTime += dur
	}

	// Convert map to slice and sort by issue key
	for _, entry := range grouped {
		report.Entries = append(report.Entries, *entry)
	}
	sort.Slice(report.Entries, func(i, j int) bool {
		return report.Entries[i].IssueKey < report.Entries[j].IssueKey
	})

	report.TaskCount = len(report.Entries)
	report.DayCount = len(days)

	return report
}

// GenerateMarkdown generates a markdown standup report
func GenerateMarkdown(report *Report) string {
	var sb strings.Builder

	// Header with date range
	dateRange := formatDateRange(report.StartDate, report.EndDate)
	sb.WriteString(fmt.Sprintf("# Daily Standup - %s\n\n", dateRange))

	// Section header
	if report.DayCount == 1 {
		sb.WriteString("## What I did\n\n")
	} else {
		sb.WriteString("## What I did (last " + fmt.Sprintf("%d", report.DayCount) + " days)\n\n")
	}

	// Entries
	if len(report.Entries) == 0 {
		sb.WriteString("_No logged work for this period_\n\n")
	} else {
		for _, entry := range report.Entries {
			// Ticket header with description
			if entry.Description != "" {
				sb.WriteString(fmt.Sprintf("### %s - %s\n", entry.IssueKey, entry.Description))
			} else {
				sb.WriteString(fmt.Sprintf("### %s\n", entry.IssueKey))
			}

			// Time spent
			sb.WriteString(fmt.Sprintf("- **Time**: %s\n", formatDuration(entry.TimeSpent)))

			// Comments (work done)
			for _, comment := range entry.Comments {
				sb.WriteString(fmt.Sprintf("- %s\n", comment))
			}

			sb.WriteString("\n")
		}
	}

	// Summary line
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("**Total**: %s | **Tasks**: %d",
		formatDuration(report.TotalTime),
		report.TaskCount))

	if report.DayCount > 1 {
		sb.WriteString(fmt.Sprintf(" | **Days**: %d", report.DayCount))
	}

	sb.WriteString("\n")

	return sb.String()
}

// formatDateRange formats the date range for the header
func formatDateRange(start, end time.Time) string {
	if start.Equal(end) {
		return start.Format("Mon 02 Jan 2006")
	}
	// Check if same month
	if start.Month() == end.Month() && start.Year() == end.Year() {
		return fmt.Sprintf("%s - %s",
			start.Format("Mon 02"),
			end.Format("Mon 02 Jan 2006"))
	}
	return fmt.Sprintf("%s - %s",
		start.Format("Mon 02 Jan"),
		end.Format("Mon 02 Jan 2006"))
}

// formatDuration formats duration in a human-readable way (e.g., "3h 30m")
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", minutes)
}

// contains checks if a string is in a slice
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// filterEmpty removes empty strings from a slice
func filterEmpty(slice []string) []string {
	var result []string
	for _, s := range slice {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}
