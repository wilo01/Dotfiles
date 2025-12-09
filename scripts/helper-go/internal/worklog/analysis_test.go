package worklog

import (
	"testing"
	"time"

	"github.com/dariuszw/hlp/internal/jira"
)

func TestAnalyzeDailyTotals_ExactMatch(t *testing.T) {
	worklogs := []jira.Worklog{
		{Started: time.Date(2025, 12, 9, 9, 0, 0, 0, time.Local), TimeSpent: 8 * time.Hour},
	}

	result := AnalyzeDailyTotals(worklogs, 8*time.Hour)

	if result.HasWarnings {
		t.Error("Expected no warnings for exact match")
	}

	if len(result.Summaries) != 1 {
		t.Errorf("Expected 1 summary, got %d", len(result.Summaries))
	}

	if result.Summaries[0].Status != DayStatusExact {
		t.Errorf("Expected DayStatusExact, got %d", result.Summaries[0].Status)
	}
}

func TestAnalyzeDailyTotals_Over(t *testing.T) {
	worklogs := []jira.Worklog{
		{Started: time.Date(2025, 12, 9, 9, 0, 0, 0, time.Local), TimeSpent: 10 * time.Hour},
	}

	result := AnalyzeDailyTotals(worklogs, 8*time.Hour)

	if !result.HasWarnings {
		t.Error("Expected warnings for over-logged day")
	}

	if result.Summaries[0].Status != DayStatusOver {
		t.Errorf("Expected DayStatusOver, got %d", result.Summaries[0].Status)
	}

	expectedDiff := 2 * time.Hour
	if result.Summaries[0].Difference != expectedDiff {
		t.Errorf("Expected difference of %v, got %v", expectedDiff, result.Summaries[0].Difference)
	}
}

func TestAnalyzeDailyTotals_Under(t *testing.T) {
	worklogs := []jira.Worklog{
		{Started: time.Date(2025, 12, 9, 9, 0, 0, 0, time.Local), TimeSpent: 6 * time.Hour},
	}

	result := AnalyzeDailyTotals(worklogs, 8*time.Hour)

	if !result.HasWarnings {
		t.Error("Expected warnings for under-logged day")
	}

	if result.Summaries[0].Status != DayStatusUnder {
		t.Errorf("Expected DayStatusUnder, got %d", result.Summaries[0].Status)
	}

	expectedDiff := -2 * time.Hour
	if result.Summaries[0].Difference != expectedDiff {
		t.Errorf("Expected difference of %v, got %v", expectedDiff, result.Summaries[0].Difference)
	}
}

func TestAnalyzeDailyTotals_MultipleDays(t *testing.T) {
	worklogs := []jira.Worklog{
		// Day 1: 10h (over)
		{Started: time.Date(2025, 12, 9, 9, 0, 0, 0, time.Local), TimeSpent: 6 * time.Hour},
		{Started: time.Date(2025, 12, 9, 15, 0, 0, 0, time.Local), TimeSpent: 4 * time.Hour},
		// Day 2: 6h (under)
		{Started: time.Date(2025, 12, 8, 9, 0, 0, 0, time.Local), TimeSpent: 6 * time.Hour},
		// Day 3: 8h (exact)
		{Started: time.Date(2025, 12, 7, 9, 0, 0, 0, time.Local), TimeSpent: 8 * time.Hour},
	}

	result := AnalyzeDailyTotals(worklogs, 8*time.Hour)

	if !result.HasWarnings {
		t.Error("Expected warnings for days that are not exact")
	}

	if len(result.Summaries) != 3 {
		t.Errorf("Expected 3 summaries, got %d", len(result.Summaries))
	}

	over, under, exact := result.CountByStatus()
	if over != 1 {
		t.Errorf("Expected 1 over day, got %d", over)
	}
	if under != 1 {
		t.Errorf("Expected 1 under day, got %d", under)
	}
	if exact != 1 {
		t.Errorf("Expected 1 exact day, got %d", exact)
	}
}

func TestAnalyzeDailyTotals_MultipleEntriesSameDay(t *testing.T) {
	worklogs := []jira.Worklog{
		{Started: time.Date(2025, 12, 9, 9, 0, 0, 0, time.Local), TimeSpent: 2 * time.Hour},
		{Started: time.Date(2025, 12, 9, 11, 0, 0, 0, time.Local), TimeSpent: 3 * time.Hour},
		{Started: time.Date(2025, 12, 9, 14, 0, 0, 0, time.Local), TimeSpent: 3 * time.Hour},
	}

	result := AnalyzeDailyTotals(worklogs, 8*time.Hour)

	if len(result.Summaries) != 1 {
		t.Errorf("Expected 1 summary (same day aggregated), got %d", len(result.Summaries))
	}

	if result.Summaries[0].TotalLogged != 8*time.Hour {
		t.Errorf("Expected total of 8h, got %v", result.Summaries[0].TotalLogged)
	}

	if result.Summaries[0].Status != DayStatusExact {
		t.Error("Expected exact match for 8h logged")
	}
}

func TestAnalyzeDailyTotals_CustomExpectedHours(t *testing.T) {
	worklogs := []jira.Worklog{
		{Started: time.Date(2025, 12, 9, 9, 0, 0, 0, time.Local), TimeSpent: 7*time.Hour + 30*time.Minute},
	}

	// Test with 7.5h expected (like user mentioned)
	result := AnalyzeDailyTotals(worklogs, 7*time.Hour+30*time.Minute)

	if result.HasWarnings {
		t.Error("Expected no warnings for exact match with 7h30m")
	}

	if result.Summaries[0].Status != DayStatusExact {
		t.Errorf("Expected DayStatusExact, got %d", result.Summaries[0].Status)
	}
}

func TestCountByStatus(t *testing.T) {
	result := &AnalysisResult{
		Summaries: []DailySummary{
			{Status: DayStatusOver},
			{Status: DayStatusOver},
			{Status: DayStatusUnder},
			{Status: DayStatusExact},
			{Status: DayStatusExact},
			{Status: DayStatusExact},
		},
	}

	over, under, exact := result.CountByStatus()

	if over != 2 {
		t.Errorf("Expected 2 over, got %d", over)
	}
	if under != 1 {
		t.Errorf("Expected 1 under, got %d", under)
	}
	if exact != 3 {
		t.Errorf("Expected 3 exact, got %d", exact)
	}
}
