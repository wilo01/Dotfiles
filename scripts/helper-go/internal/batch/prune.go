package batch

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/pkg/duration"
)

// dateKey is the layout used to bucket entries into calendar days.
const dateKey = "02.01.2006"

// StaleDraft pairs a prunable DRAFT row with the reason it qualified, so the
// confirmation prompt can explain each deletion instead of just listing rows.
type StaleDraft struct {
	Entry  Entry
	Reason string
}

// LoggedTimeByDate sums time per calendar day across entries that represent real
// logged work. DRAFT rows are excluded even when they carry hours: those hours
// have not been posted to JIRA, so counting them would let an unposted draft
// declare its own day "full" and delete its neighbours.
func LoggedTimeByDate(entries []Entry) map[string]time.Duration {
	totals := make(map[string]time.Duration)
	for _, e := range entries {
		if e.IsDraft() {
			continue
		}
		day, err := ParseDate(e.Date, "09:00")
		if err != nil {
			continue
		}
		spent, err := duration.Parse(e.TimeSpent)
		if err != nil {
			continue
		}
		totals[day.Format(dateKey)] += spent
	}
	return totals
}

// DayIsCovered reports whether a day already holds a full expected day of posted
// work, meaning no further worklog can reasonably be added to it.
func DayIsCovered(totals map[string]time.Duration, day time.Time, expectedPerDay time.Duration) bool {
	if expectedPerDay <= 0 {
		return false
	}
	return totals[day.Format(dateKey)] >= expectedPerDay
}

// isPrunableDraft reports whether a row is an empty placeholder that can be
// deleted without losing anything the user put there. It answers "is this row
// disposable", not "is it time to dispose of it" — the timing triggers live in
// FindStaleDrafts.
//
// A row must clear every guard: only DRAFT rows are eligible at all; a draft
// carrying hours is a forgotten status flip that promoteDraftsWithTime owns; a
// draft carrying a comment holds typed notes and is effectively pinned; and
// today's drafts are the live working set.
func isPrunableDraft(e Entry, today time.Time) bool {
	if !e.IsDraft() || e.TimeSpent != "" || e.Comment != "" {
		return false
	}
	day, err := ParseDate(e.Date, "09:00")
	if err != nil {
		return false
	}
	return day.Format(dateKey) != today.Format(dateKey) && day.Before(today)
}

// FindStaleDrafts returns the DRAFT rows that can no longer become worklogs:
// either their day is already fully logged, or the day aged past the retention
// window without ever filling up. Pure by design — now is injected so the
// day-boundary behaviour is testable.
func FindStaleDrafts(entries []Entry, expectedPerDay time.Duration, retentionDays int, now time.Time) []StaleDraft {
	if retentionDays <= 0 {
		retentionDays = config.DefaultDraftRetentionDays
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cutoff := today.AddDate(0, 0, -retentionDays)
	totals := LoggedTimeByDate(entries)

	var stale []StaleDraft
	for _, e := range entries {
		if !isPrunableDraft(e, today) {
			continue
		}
		day, err := ParseDate(e.Date, "09:00")
		if err != nil {
			continue
		}
		day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())

		reason, ok := pruneReason(day, totals, expectedPerDay, cutoff, today)
		if !ok {
			continue
		}
		stale = append(stale, StaleDraft{Entry: e, Reason: reason})
	}
	return stale
}

// pruneReason applies the two timing triggers to an already-disposable draft and
// describes which one fired, for the confirmation prompt.
func pruneReason(day time.Time, totals map[string]time.Duration, expectedPerDay time.Duration, cutoff, today time.Time) (string, bool) {
	if DayIsCovered(totals, day, expectedPerDay) {
		return fmt.Sprintf("day full (%s)", duration.Format(expectedPerDay)), true
	}
	if day.Before(cutoff) {
		age := int(today.Sub(day).Hours() / 24)
		return fmt.Sprintf("aged out (%dd, %s logged)", age, duration.Format(totals[day.Format(dateKey)])), true
	}
	return "", false
}

// RemoveEntriesByRows deletes every given 1-based CSV row in a single
// read-filter-write pass. Looping RemoveEntryByRow would be wrong: each removal
// rewrites the file and shifts every later row number, so the second call would
// delete the wrong line.
func RemoveEntriesByRows(path string, rowNumbers []int) error {
	if len(rowNumbers) == 0 {
		return nil
	}
	doomed := make(map[int]bool, len(rowNumbers))
	for _, n := range rowNumbers {
		doomed[n] = true
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("remove entries: open CSV: %w", err)
	}
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	f.Close()
	if err != nil {
		return fmt.Errorf("remove entries: read CSV: %w", err)
	}

	kept := make([][]string, 0, len(records))
	for i, record := range records {
		if !doomed[i+1] {
			kept = append(kept, record)
		}
	}

	f, err = createForRewrite(path)
	if err != nil {
		return fmt.Errorf("remove entries: write CSV: %w", err)
	}
	defer f.Close()

	return writeWorklogRecords(f, kept)
}
