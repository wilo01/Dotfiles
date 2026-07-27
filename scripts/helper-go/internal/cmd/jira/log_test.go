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
