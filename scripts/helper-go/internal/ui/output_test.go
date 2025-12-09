package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/dariuszw/hlp/internal/worklog"
)

func TestFormatDailyWarning_Over(t *testing.T) {
	date := time.Date(2025, 12, 9, 0, 0, 0, 0, time.Local)
	totalLogged := 10 * time.Hour
	expected := 8 * time.Hour

	result := FormatDailyWarning(date, totalLogged, expected)

	if !strings.Contains(result, "over by") {
		t.Error("Expected 'over by' in warning message")
	}
	// 10h formats as "1d2h" (8h = 1d in JIRA)
	if !strings.Contains(result, "1d2h") {
		t.Errorf("Expected '1d2h' total in message, got: %s", result)
	}
	if !strings.Contains(result, "2h") {
		t.Error("Expected '2h' difference in message")
	}
}

func TestFormatDailyWarning_Under(t *testing.T) {
	date := time.Date(2025, 12, 9, 0, 0, 0, 0, time.Local)
	totalLogged := 6 * time.Hour
	expected := 8 * time.Hour

	result := FormatDailyWarning(date, totalLogged, expected)

	if !strings.Contains(result, "under by") {
		t.Error("Expected 'under by' in warning message")
	}
	if !strings.Contains(result, "6h") {
		t.Error("Expected '6h' total in message")
	}
	if !strings.Contains(result, "2h") {
		t.Error("Expected '2h' difference in message")
	}
}

func TestFormatDailyWarning_Exact(t *testing.T) {
	date := time.Date(2025, 12, 9, 0, 0, 0, 0, time.Local)
	totalLogged := 8 * time.Hour
	expected := 8 * time.Hour

	result := FormatDailyWarning(date, totalLogged, expected)

	if !strings.Contains(result, "exactly") {
		t.Error("Expected 'exactly' in success message")
	}
}

func TestDailyBreakdown(t *testing.T) {
	summaries := []worklog.DailySummary{
		{
			Date:          time.Date(2025, 12, 9, 0, 0, 0, 0, time.Local),
			TotalLogged:   10 * time.Hour,
			ExpectedHours: 8 * time.Hour,
			Difference:    2 * time.Hour,
			Status:        worklog.DayStatusOver,
		},
		{
			Date:          time.Date(2025, 12, 8, 0, 0, 0, 0, time.Local),
			TotalLogged:   6 * time.Hour,
			ExpectedHours: 8 * time.Hour,
			Difference:    -2 * time.Hour,
			Status:        worklog.DayStatusUnder,
		},
	}

	result := DailyBreakdown(summaries, "8h")

	if !strings.Contains(result, "Daily Time Summary") {
		t.Error("Expected header in breakdown")
	}
	// 10h formats as "1d2h" (8h = 1d in JIRA)
	if !strings.Contains(result, "1d2h") {
		t.Errorf("Expected '1d2h' in breakdown for over day, got: %s", result)
	}
	if !strings.Contains(result, "6h") {
		t.Error("Expected '6h' in breakdown for under day")
	}
}

func TestDailyWarningsSummary(t *testing.T) {
	result := &worklog.AnalysisResult{
		HasWarnings: true,
		Summaries: []worklog.DailySummary{
			{Status: worklog.DayStatusOver},
			{Status: worklog.DayStatusUnder},
			{Status: worklog.DayStatusUnder},
		},
	}

	summary := DailyWarningsSummary(result)

	if !strings.Contains(summary, "1 day(s) over") {
		t.Error("Expected '1 day(s) over' in summary")
	}
	if !strings.Contains(summary, "2 day(s) under") {
		t.Error("Expected '2 day(s) under' in summary")
	}
}

func TestDailyWarningsSummary_NoWarnings(t *testing.T) {
	result := &worklog.AnalysisResult{
		HasWarnings: false,
		Summaries:   []worklog.DailySummary{},
	}

	summary := DailyWarningsSummary(result)

	if summary != "" {
		t.Errorf("Expected empty string for no warnings, got: %s", summary)
	}
}
