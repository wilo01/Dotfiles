package jira

import (
	"testing"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
)

func TestSelectRows(t *testing.T) {
	drafts := []batch.Entry{
		{RowNumber: 2},
		{RowNumber: 5},
		{RowNumber: 9},
	}
	confirmOdd := func(e batch.Entry) bool { return e.RowNumber%2 == 1 }
	confirmNone := func(batch.Entry) bool {
		t.Error("confirmOne must not be called outside individual mode")
		return true
	}

	cases := []struct {
		name   string
		choice string
		want   []int
	}{
		{"y promotes all", "y", []int{2, 5, 9}},
		{"yes promotes all", "yes", []int{2, 5, 9}},
		{"n promotes none", "n", nil},
		{"empty default promotes none", "", nil},
		{"garbage promotes none", "x", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := selectRows(c.choice, drafts, confirmNone)
			if !equalIntSlices(got, c.want) {
				t.Errorf("selectRows(%q) = %v, want %v", c.choice, got, c.want)
			}
		})
	}

	t.Run("i asks per entry", func(t *testing.T) {
		got := selectRows("i", drafts, confirmOdd)
		if !equalIntSlices(got, []int{5, 9}) {
			t.Errorf("selectRows(\"i\") = %v, want [5 9]", got)
		}
	})
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestIsDraftWithTime(t *testing.T) {
	cases := []struct {
		name  string
		entry batch.Entry
		want  bool
	}{
		{"draft with time", batch.Entry{Status: batch.StatusDraft, TimeSpent: "1h"}, true},
		{"draft without time", batch.Entry{Status: batch.StatusDraft, TimeSpent: ""}, false},
		{"draft with whitespace time", batch.Entry{Status: batch.StatusDraft, TimeSpent: "  "}, false},
		{"pending with time", batch.Entry{Status: batch.StatusPending, TimeSpent: "1h"}, false},
		{"done with time", batch.Entry{Status: batch.StatusDone, TimeSpent: "1h"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isDraftWithTime(c.entry); got != c.want {
				t.Errorf("isDraftWithTime(%+v) = %v, want %v", c.entry, got, c.want)
			}
		})
	}
}

func TestStaleRowSet(t *testing.T) {
	stale := []batch.StaleDraft{
		{Entry: batch.Entry{RowNumber: 4}},
		{Entry: batch.Entry{RowNumber: 9}},
	}

	rows := staleRowSet(stale)

	if !rows[4] || !rows[9] {
		t.Errorf("expected rows 4 and 9 marked, got %v", rows)
	}
	if rows[5] {
		t.Errorf("row 5 was not stale but is marked")
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 marked rows, got %d", len(rows))
	}
}

// The deletion prompt shows whole days, so the surrounding DONE rows explain why
// the day qualified — but it must not pull in unaffected days.
func TestEntriesOnDaysOf(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "A", Date: "16.07.2026 09:00", Status: batch.StatusDone},
		{IssueKey: "B", Date: "16.07.2026 09:00", Status: batch.StatusDraft, RowNumber: 2},
		{IssueKey: "C", Date: "17.07.2026 09:00", Status: batch.StatusDone},
	}
	stale := []batch.StaleDraft{{Entry: entries[1]}}

	got := entriesOnDaysOf(entries, stale)

	if len(got) != 2 {
		t.Fatalf("expected both 16.07 rows, got %d: %+v", len(got), got)
	}
	for _, e := range got {
		if e.IssueKey == "C" {
			t.Errorf("17.07 has no stale drafts and must not be included")
		}
	}
}

// The table prints days oldest-first so the most recent day lands at the bottom,
// next to the summary, without scrolling back up.
func TestGroupEntriesByDay_OldestFirst(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "C", Date: "23.07.2026 09:00", TimeSpent: "1h"},
		{IssueKey: "A", Date: "15.07.2026 09:00", TimeSpent: "2h"},
		{IssueKey: "B", Date: "21.07.2026 09:00", TimeSpent: "30m"},
		{IssueKey: "A2", Date: "15.07.2026 14:00", TimeSpent: "3h"},
	}

	groups := groupEntriesByDay(entries, false)

	var order []string
	for _, g := range groups {
		order = append(order, g.date)
	}
	want := []string{"15.07.2026", "21.07.2026", "23.07.2026"}
	if len(order) != len(want) {
		t.Fatalf("got %d groups %v, want %d %v", len(order), order, len(want), want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("group[%d] = %s, want %s", i, order[i], want[i])
		}
	}

	// Entries on the same day stay in input order and their times sum.
	if len(groups[0].entries) != 2 {
		t.Fatalf("15.07 should hold 2 entries, got %d", len(groups[0].entries))
	}
	if groups[0].entries[0].IssueKey != "A" {
		t.Errorf("within-day order changed: got %s first", groups[0].entries[0].IssueKey)
	}
	if want := 5 * time.Hour; groups[0].total != want {
		t.Errorf("15.07 total = %v, want %v", groups[0].total, want)
	}
}

func TestGroupEntriesByDay_NewestFirstFlips(t *testing.T) {
	entries := []batch.Entry{
		{Date: "15.07.2026 09:00"},
		{Date: "23.07.2026 09:00"},
	}

	groups := groupEntriesByDay(entries, true)

	if groups[0].date != "23.07.2026" {
		t.Errorf("newestFirst should put 23.07 first, got %s", groups[0].date)
	}
}

func TestPruneReasonSummary_Deduplicates(t *testing.T) {
	stale := []batch.StaleDraft{
		{Reason: "day full (7h30m)"},
		{Reason: "day full (7h30m)"},
		{Reason: "aged out (15d, 3h logged)"},
	}

	got := pruneReasonSummary(stale)
	want := "day full (7h30m), aged out (15d, 3h logged)"
	if got != want {
		t.Errorf("pruneReasonSummary() = %q, want %q", got, want)
	}
}

// The merged preview must show every row's fate without ever widening the set
// that actually gets posted.
func TestPreviewEntriesWithStale(t *testing.T) {
	pending := []batch.Entry{
		{IssueKey: "P1", Date: "10.08.2026 09:00", RowNumber: 10, TimeSpent: "2h"},
		{IssueKey: "P2", Date: "10.08.2026 09:00", RowNumber: 11, TimeSpent: "1h"},
	}
	all := []batch.Entry{
		{IssueKey: "OLD", Date: "07.08.2026 09:00", RowNumber: 4, Status: batch.StatusDone, TimeSpent: "7h30m"},
		{IssueKey: "DEAD", Date: "07.08.2026 09:00", RowNumber: 5, Status: batch.StatusDraft},
		{IssueKey: "OTHER", Date: "01.08.2026 09:00", RowNumber: 6, Status: batch.StatusDone, TimeSpent: "8h"},
		pending[0], pending[1],
	}
	stale := []batch.StaleDraft{{Entry: all[1], Reason: "day full (7h30m)"}}

	got := previewEntriesWithStale(pending, all, stale)

	rows := make(map[int]int)
	for _, e := range got {
		rows[e.RowNumber]++
	}
	for _, want := range []int{10, 11, 4, 5} {
		if rows[want] != 1 {
			t.Errorf("row %d appears %d times, want exactly 1", want, rows[want])
		}
	}
	if rows[6] != 0 {
		t.Errorf("01.08 has no stale draft and must not be pulled into the preview")
	}
	if len(got) != 4 {
		t.Errorf("expected 4 preview rows, got %d: %+v", len(got), got)
	}
}

func TestPreviewEntriesWithStale_NoStaleReturnsPendingUnchanged(t *testing.T) {
	pending := []batch.Entry{{IssueKey: "P1", RowNumber: 10}}

	got := previewEntriesWithStale(pending, []batch.Entry{{RowNumber: 99}}, nil)

	if len(got) != 1 || got[0].RowNumber != 10 {
		t.Errorf("expected the pending set untouched, got %+v", got)
	}
}

// The whole point of asking the delete question before posting is that posting
// cannot change the answer. LoggedTimeByDate counts every non-DRAFT row's time
// regardless of status, so flipping PENDING to DONE must select the same rows.
// If this fails, the prompt has to move back after the submit loop.
func TestFindStaleDraftsUnaffectedByStatusFlip(t *testing.T) {
	now := time.Date(2026, 8, 10, 15, 0, 0, 0, time.UTC)
	build := func(postedStatus string) []batch.Entry {
		return []batch.Entry{
			{IssueKey: "A", Date: "07.08.2026 09:00", RowNumber: 2, Status: postedStatus, TimeSpent: "7h30m"},
			{IssueKey: "B", Date: "07.08.2026 09:00", RowNumber: 3, Status: batch.StatusDraft},
			{IssueKey: "C", Date: "10.08.2026 09:00", RowNumber: 4, Status: postedStatus, TimeSpent: "1h"},
		}
	}
	expected := 7*time.Hour + 30*time.Minute

	before := batch.FindStaleDrafts(build(batch.StatusPending), expected, 30, now)
	after := batch.FindStaleDrafts(build(batch.StatusDone), expected, 30, now)

	if len(before) != 1 {
		t.Fatalf("expected the 07.08 draft to be stale before posting, got %+v", before)
	}
	if len(before) != len(after) {
		t.Fatalf("stale count changed across posting: %d before, %d after", len(before), len(after))
	}
	for i := range before {
		if before[i].Entry.RowNumber != after[i].Entry.RowNumber {
			t.Errorf("stale row %d changed across posting: %d -> %d",
				i, before[i].Entry.RowNumber, after[i].Entry.RowNumber)
		}
		if before[i].Reason != after[i].Reason {
			t.Errorf("reason changed across posting: %q -> %q", before[i].Reason, after[i].Reason)
		}
	}
}
