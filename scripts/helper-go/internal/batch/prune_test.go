package batch

import (
	"os"
	"testing"
	"time"
)

// now is a fixed "today" so day-boundary behaviour is deterministic.
var pruneNow = time.Date(2026, 7, 23, 14, 30, 0, 0, time.Local)

func daysAgo(n int) string {
	return pruneNow.AddDate(0, 0, -n).Format("02.01.2006") + " 09:00"
}

const expectedDay = 7*time.Hour + 30*time.Minute

func TestFindStaleDrafts(t *testing.T) {
	cases := []struct {
		name       string
		entries    []Entry
		wantKeys   []string
		wantReason string
	}{
		{
			name: "draft on a full day is pruned immediately",
			entries: []Entry{
				{IssueKey: "DONE-1", Date: daysAgo(1), TimeSpent: "7h30m", Status: StatusDone},
				{IssueKey: "DRAFT-1", Date: daysAgo(1), Status: StatusDraft},
			},
			wantKeys:   []string{"DRAFT-1"},
			wantReason: "day full (7h30m)",
		},
		{
			name: "draft on a partial day inside retention is kept",
			entries: []Entry{
				{IssueKey: "DONE-1", Date: daysAgo(3), TimeSpent: "4h", Status: StatusDone},
				{IssueKey: "DRAFT-1", Date: daysAgo(3), Status: StatusDraft},
			},
			wantKeys: nil,
		},
		{
			name: "same partial day past retention is pruned",
			entries: []Entry{
				{IssueKey: "DONE-1", Date: daysAgo(15), TimeSpent: "3h", Status: StatusDone},
				{IssueKey: "DRAFT-1", Date: daysAgo(15), Status: StatusDraft},
			},
			wantKeys:   []string{"DRAFT-1"},
			wantReason: "aged out (15d, 3h logged)",
		},
		{
			name: "zero-logged day past retention is pruned",
			entries: []Entry{
				{IssueKey: "DRAFT-1", Date: daysAgo(20), Status: StatusDraft},
			},
			wantKeys:   []string{"DRAFT-1"},
			wantReason: "aged out (20d, 0m logged)",
		},
		{
			name: "draft carrying time is never pruned",
			entries: []Entry{
				{IssueKey: "DONE-1", Date: daysAgo(15), TimeSpent: "7h30m", Status: StatusDone},
				{IssueKey: "DRAFT-1", Date: daysAgo(15), TimeSpent: "2h", Status: StatusDraft},
			},
			wantKeys: nil,
		},
		{
			name: "draft carrying a comment is never pruned",
			entries: []Entry{
				{IssueKey: "DONE-1", Date: daysAgo(15), TimeSpent: "7h30m", Status: StatusDone},
				{IssueKey: "DRAFT-1", Date: daysAgo(15), Comment: "chased Sukbir for details", Status: StatusDraft},
			},
			wantKeys: nil,
		},
		{
			name: "today's drafts survive even when today is full",
			entries: []Entry{
				{IssueKey: "DONE-1", Date: daysAgo(0), TimeSpent: "7h30m", Status: StatusDone},
				{IssueKey: "DRAFT-1", Date: daysAgo(0), Status: StatusDraft},
			},
			wantKeys: nil,
		},
		{
			name: "non-draft rows are never pruned",
			entries: []Entry{
				{IssueKey: "DONE-1", Date: daysAgo(30), TimeSpent: "7h30m", Status: StatusDone},
				{IssueKey: "PEND-1", Date: daysAgo(30), Status: StatusPending},
				{IssueKey: "SYNC-1", Date: daysAgo(30), Status: StatusSync},
				{IssueKey: "UPD-1", Date: daysAgo(30), Status: StatusUpdated},
			},
			wantKeys: nil,
		},
		{
			name: "unparseable date is left for the human",
			entries: []Entry{
				{IssueKey: "DRAFT-1", Date: "not-a-date", Status: StatusDraft},
			},
			wantKeys: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FindStaleDrafts(c.entries, expectedDay, 7, pruneNow)

			if len(got) != len(c.wantKeys) {
				t.Fatalf("got %d stale drafts %v, want %d %v", len(got), keysOf(got), len(c.wantKeys), c.wantKeys)
			}
			for i, want := range c.wantKeys {
				if got[i].Entry.IssueKey != want {
					t.Errorf("stale[%d] = %s, want %s", i, got[i].Entry.IssueKey, want)
				}
			}
			if c.wantReason != "" && got[0].Reason != c.wantReason {
				t.Errorf("reason = %q, want %q", got[0].Reason, c.wantReason)
			}
		})
	}
}

// A DRAFT with hours has not been posted to JIRA, so it must not make its day
// look covered and take its neighbours down with it.
func TestFindStaleDrafts_DraftTimeDoesNotCoverTheDay(t *testing.T) {
	entries := []Entry{
		{IssueKey: "DONE-1", Date: daysAgo(2), TimeSpent: "3h30m", Status: StatusDone},
		{IssueKey: "DRAFT-1", Date: daysAgo(2), TimeSpent: "4h", Status: StatusDraft},
		{IssueKey: "DRAFT-2", Date: daysAgo(2), Status: StatusDraft},
	}

	if got := FindStaleDrafts(entries, expectedDay, 7, pruneNow); len(got) != 0 {
		t.Errorf("expected no pruning (day only has 3h30m posted), got %v", keysOf(got))
	}
}

func TestFindStaleDrafts_ZeroRetentionUsesDefault(t *testing.T) {
	entries := []Entry{
		{IssueKey: "DRAFT-1", Date: daysAgo(3), Status: StatusDraft},
		{IssueKey: "DRAFT-2", Date: daysAgo(10), Status: StatusDraft},
	}

	got := FindStaleDrafts(entries, expectedDay, 0, pruneNow)
	if len(got) != 1 || got[0].Entry.IssueKey != "DRAFT-2" {
		t.Errorf("expected only the 10-day-old draft to age out, got %v", keysOf(got))
	}
}

func TestLoggedTimeByDate_ExcludesDrafts(t *testing.T) {
	entries := []Entry{
		{Date: "15.07.2026 09:00", TimeSpent: "4h", Status: StatusDone},
		{Date: "15.07.2026 14:00", TimeSpent: "1h30m", Status: StatusPending},
		{Date: "15.07.2026 16:00", TimeSpent: "2h", Status: StatusDraft},
		{Date: "16.07.2026 09:00", TimeSpent: "", Status: StatusDone},
	}

	totals := LoggedTimeByDate(entries)

	if want := 5*time.Hour + 30*time.Minute; totals["15.07.2026"] != want {
		t.Errorf("15.07 total = %v, want %v", totals["15.07.2026"], want)
	}
	if totals["16.07.2026"] != 0 {
		t.Errorf("16.07 total = %v, want 0", totals["16.07.2026"])
	}
}

// DayIsCovered drives the sync churn guard: when the day sync would prepend new
// drafts to is already full, sync must skip creating them or it re-creates
// exactly what the prune step deletes.
func TestDayIsCovered(t *testing.T) {
	entries := []Entry{
		{Date: "22.07.2026 09:00", TimeSpent: "3h", Status: StatusDone},
		{Date: "22.07.2026 13:00", TimeSpent: "4h30m", Status: StatusPending},
		{Date: "21.07.2026 09:00", TimeSpent: "4h", Status: StatusDone},
	}
	totals := LoggedTimeByDate(entries)

	full, _ := ParseDate("22.07.2026", "09:00")
	partial, _ := ParseDate("21.07.2026", "09:00")
	empty, _ := ParseDate("20.07.2026", "09:00")

	if !DayIsCovered(totals, full, expectedDay) {
		t.Errorf("22.07 has 7h30m posted, expected covered")
	}
	if DayIsCovered(totals, partial, expectedDay) {
		t.Errorf("21.07 has only 4h, expected not covered")
	}
	if DayIsCovered(totals, empty, expectedDay) {
		t.Errorf("20.07 has nothing logged, expected not covered")
	}
	// A zero/unset expectation must never mark every day covered.
	if DayIsCovered(totals, full, 0) {
		t.Errorf("zero expectedPerDay must not report a day as covered")
	}
}

// Regression: looping RemoveEntryByRow would shift row numbers after the first
// delete and take out the wrong lines.
func TestRemoveEntriesByRows_MultipleRowsStayAligned(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "prune_rows_test_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)
	defer os.Remove(tmpPath + ".autobak")

	for _, key := range []string{"VIS-1", "VIS-2", "VIS-3", "VIS-4", "VIS-5"} {
		if err := AppendEntryWithStatus(tmpPath, key, "", "Story", key, "", "06.03.2026 09:00", "", "", StatusDraft); err != nil {
			t.Fatal(err)
		}
	}

	if err := RemoveEntriesByRows(tmpPath, []int{2, 4}); err != nil {
		t.Fatalf("RemoveEntriesByRows failed: %v", err)
	}

	entries, err := ParseCSV(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"VIS-1", "VIS-3", "VIS-5"}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d", len(entries), len(want))
	}
	for i, key := range want {
		if entries[i].IssueKey != key {
			t.Errorf("entry[%d] = %s, want %s", i, entries[i].IssueKey, key)
		}
	}
}

func TestRemoveEntriesByRows_NoRowsIsNoop(t *testing.T) {
	if err := RemoveEntriesByRows("/nonexistent/path.csv", nil); err != nil {
		t.Errorf("expected no-op for empty row list, got %v", err)
	}
}

func keysOf(stale []StaleDraft) []string {
	keys := make([]string, len(stale))
	for i, s := range stale {
		keys[i] = s.Entry.IssueKey
	}
	return keys
}
