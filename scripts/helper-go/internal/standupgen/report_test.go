package standupgen

import (
	"strings"
	"testing"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
)

func TestBuildReport(t *testing.T) {
	tests := []struct {
		name      string
		entries   []batch.Entry
		startDate time.Time
		endDate   time.Time
		wantTasks int
		wantDays  int
		wantTime  time.Duration
	}{
		{
			name: "single entry",
			entries: []batch.Entry{
				{
					IssueKey:    "VIS-123",
					Description: "Fix auth bug",
					TimeSpent:   "2h",
					Date:        "23.12.2025",
					Comment:     "Fixed token refresh",
				},
			},
			startDate: time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			endDate:   time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			wantTasks: 1,
			wantDays:  1,
			wantTime:  2 * time.Hour,
		},
		{
			name: "multiple entries same ticket",
			entries: []batch.Entry{
				{
					IssueKey:    "VIS-123",
					Description: "Fix auth bug",
					TimeSpent:   "2h",
					Date:        "23.12.2025",
					Comment:     "Morning work",
				},
				{
					IssueKey:    "VIS-123",
					Description: "Fix auth bug",
					TimeSpent:   "1h30m",
					Date:        "23.12.2025",
					Comment:     "Afternoon work",
				},
			},
			startDate: time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			endDate:   time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			wantTasks: 1, // Should be grouped
			wantDays:  1,
			wantTime:  3*time.Hour + 30*time.Minute,
		},
		{
			name: "multiple different tickets",
			entries: []batch.Entry{
				{
					IssueKey:    "VIS-123",
					Description: "Fix auth bug",
					TimeSpent:   "2h",
					Date:        "23.12.2025",
				},
				{
					IssueKey:    "VIS-456",
					Description: "Code review",
					TimeSpent:   "1h",
					Date:        "23.12.2025",
				},
			},
			startDate: time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			endDate:   time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			wantTasks: 2,
			wantDays:  1,
			wantTime:  3 * time.Hour,
		},
		{
			name: "multiple days",
			entries: []batch.Entry{
				{
					IssueKey:    "VIS-123",
					Description: "Fix auth bug",
					TimeSpent:   "2h",
					Date:        "23.12.2025",
				},
				{
					IssueKey:    "VIS-456",
					Description: "Code review",
					TimeSpent:   "1h",
					Date:        "24.12.2025",
				},
			},
			startDate: time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			endDate:   time.Date(2025, 12, 24, 0, 0, 0, 0, time.Local),
			wantTasks: 2,
			wantDays:  2,
			wantTime:  3 * time.Hour,
		},
		{
			name:      "empty entries",
			entries:   []batch.Entry{},
			startDate: time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			endDate:   time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
			wantTasks: 0,
			wantDays:  0,
			wantTime:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := BuildReport(tt.entries, tt.startDate, tt.endDate)

			if report.TaskCount != tt.wantTasks {
				t.Errorf("TaskCount = %d, want %d", report.TaskCount, tt.wantTasks)
			}
			if report.DayCount != tt.wantDays {
				t.Errorf("DayCount = %d, want %d", report.DayCount, tt.wantDays)
			}
			if report.TotalTime != tt.wantTime {
				t.Errorf("TotalTime = %v, want %v", report.TotalTime, tt.wantTime)
			}
		})
	}
}

func TestGenerateMarkdown(t *testing.T) {
	report := &Report{
		StartDate: time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
		Entries: []ReportEntry{
			{
				IssueKey:    "VIS-123",
				Description: "Fix auth bug",
				TimeSpent:   2 * time.Hour,
				Comments:    []string{"Fixed token refresh"},
			},
		},
		TotalTime: 2 * time.Hour,
		TaskCount: 1,
		DayCount:  1,
	}

	md := GenerateMarkdown(report)

	// Check that markdown contains expected elements
	expectedParts := []string{
		"# Daily Standup",
		"## What I did",
		"### VIS-123 - Fix auth bug",
		"**Time**: 2h",
		"Fixed token refresh",
		"**Total**: 2h",
		"**Tasks**: 1",
	}

	for _, part := range expectedParts {
		if !strings.Contains(md, part) {
			t.Errorf("Markdown missing expected part: %q\nGot:\n%s", part, md)
		}
	}
}

func TestGenerateMarkdownMultipleDays(t *testing.T) {
	report := &Report{
		StartDate: time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2025, 12, 24, 0, 0, 0, 0, time.Local),
		Entries: []ReportEntry{
			{
				IssueKey:    "VIS-123",
				Description: "Fix auth bug",
				TimeSpent:   4 * time.Hour,
				Comments:    []string{},
			},
		},
		TotalTime: 4 * time.Hour,
		TaskCount: 1,
		DayCount:  2,
	}

	md := GenerateMarkdown(report)

	// Should include "Days" in summary when multiple days
	if !strings.Contains(md, "**Days**: 2") {
		t.Errorf("Markdown should include Days count for multi-day reports.\nGot:\n%s", md)
	}

	// Should indicate multiple days in header
	if !strings.Contains(md, "last 2 days") {
		t.Errorf("Markdown header should mention multiple days.\nGot:\n%s", md)
	}
}

func TestGenerateMarkdownNoDescription(t *testing.T) {
	report := &Report{
		StartDate: time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
		Entries: []ReportEntry{
			{
				IssueKey:    "VIS-123",
				Description: "", // No description
				TimeSpent:   1 * time.Hour,
				Comments:    []string{},
			},
		},
		TotalTime: 1 * time.Hour,
		TaskCount: 1,
		DayCount:  1,
	}

	md := GenerateMarkdown(report)

	// Should just show issue key without dash separator
	if strings.Contains(md, "VIS-123 - \n") || strings.Contains(md, "VIS-123 -\n") {
		t.Errorf("Should not have trailing dash when no description.\nGot:\n%s", md)
	}
	if !strings.Contains(md, "### VIS-123\n") {
		t.Errorf("Should have issue key alone when no description.\nGot:\n%s", md)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{1 * time.Hour, "1h"},
		{30 * time.Minute, "30m"},
		{1*time.Hour + 30*time.Minute, "1h 30m"},
		{8 * time.Hour, "8h"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestCommentsDeduplication(t *testing.T) {
	entries := []batch.Entry{
		{
			IssueKey:    "VIS-123",
			Description: "Test",
			TimeSpent:   "1h",
			Date:        "23.12.2025",
			Comment:     "Same comment",
		},
		{
			IssueKey:    "VIS-123",
			Description: "Test",
			TimeSpent:   "1h",
			Date:        "23.12.2025",
			Comment:     "Same comment", // Duplicate
		},
		{
			IssueKey:    "VIS-123",
			Description: "Test",
			TimeSpent:   "1h",
			Date:        "23.12.2025",
			Comment:     "Different comment",
		},
	}

	report := BuildReport(entries,
		time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local),
		time.Date(2025, 12, 23, 0, 0, 0, 0, time.Local))

	if len(report.Entries) != 1 {
		t.Fatalf("Expected 1 grouped entry, got %d", len(report.Entries))
	}

	// Should have 2 unique comments, not 3
	if len(report.Entries[0].Comments) != 2 {
		t.Errorf("Expected 2 unique comments, got %d: %v",
			len(report.Entries[0].Comments), report.Entries[0].Comments)
	}
}
