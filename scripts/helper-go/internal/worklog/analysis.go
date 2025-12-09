package worklog

import (
	"sort"
	"time"

	"github.com/dariuszw/hlp/internal/jira"
)

// DayStatus indicates if a day is over, under, or exact
type DayStatus int

const (
	DayStatusExact DayStatus = iota
	DayStatusOver
	DayStatusUnder
)

// DailySummary represents logged time for a single day
type DailySummary struct {
	Date          time.Time
	TotalLogged   time.Duration
	ExpectedHours time.Duration
	Difference    time.Duration // Positive = over, Negative = under
	Status        DayStatus
}

// AnalysisResult contains the full daily breakdown
type AnalysisResult struct {
	Summaries   []DailySummary
	HasWarnings bool
}

// AnalyzeDailyTotals aggregates worklogs by date and compares against expected hours
func AnalyzeDailyTotals(worklogs []jira.Worklog, expectedPerDay time.Duration) *AnalysisResult {
	// Group by date
	byDate := make(map[string]*DailySummary)

	for _, wl := range worklogs {
		dateKey := wl.Started.Format("2006-01-02")

		if _, exists := byDate[dateKey]; !exists {
			byDate[dateKey] = &DailySummary{
				Date:          wl.Started.Truncate(24 * time.Hour),
				ExpectedHours: expectedPerDay,
			}
		}

		byDate[dateKey].TotalLogged += wl.TimeSpent
	}

	// Calculate differences and status
	result := &AnalysisResult{}

	for _, summary := range byDate {
		summary.Difference = summary.TotalLogged - summary.ExpectedHours

		if summary.Difference > 0 {
			summary.Status = DayStatusOver
			result.HasWarnings = true
		} else if summary.Difference < 0 {
			summary.Status = DayStatusUnder
			result.HasWarnings = true
		} else {
			summary.Status = DayStatusExact
		}

		result.Summaries = append(result.Summaries, *summary)
	}

	// Sort by date (newest first)
	sort.Slice(result.Summaries, func(i, j int) bool {
		return result.Summaries[i].Date.After(result.Summaries[j].Date)
	})

	return result
}

// CountByStatus returns counts of over, under, and exact days
func (r *AnalysisResult) CountByStatus() (over, under, exact int) {
	for _, s := range r.Summaries {
		switch s.Status {
		case DayStatusOver:
			over++
		case DayStatusUnder:
			under++
		case DayStatusExact:
			exact++
		}
	}
	return
}
