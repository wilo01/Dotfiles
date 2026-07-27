package standup

import (
	"testing"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
)

// daysAgo formats a date n days before now in the CSV's DD/MM/YYYY format.
// Tests use offsets well inside/outside the window (1 vs 10 for a 3-day
// window) so the 09:00 parse-time anchor can't cause boundary flakes.
func daysAgo(n int) string {
	return time.Now().AddDate(0, 0, -n).Format("02/01/2006")
}

func issueKeys(entries []batch.Entry) []string {
	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		keys = append(keys, e.IssueKey)
	}
	return keys
}

func TestFilterTicketsWithComments_dropsUncommentedTickets(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "ZZT-1", Date: daysAgo(0), Comment: "made progress"},
		{IssueKey: "ZZT-1", Date: daysAgo(1), Comment: ""},
		{IssueKey: "ZZT-2", Date: daysAgo(0), Comment: ""},
		{IssueKey: "ZZT-2", Date: daysAgo(1), Comment: ""},
	}
	got := filterTicketsWithComments(entries)

	want := []string{"ZZT-1", "ZZT-1"}
	gotKeys := issueKeys(got)
	if len(gotKeys) != len(want) {
		t.Fatalf("kept %d entries %v, want %d %v (one comment keeps ALL of that ticket's rows; uncommented ticket fully dropped)", len(gotKeys), gotKeys, len(want), want)
	}
	for i := range want {
		if gotKeys[i] != want[i] {
			t.Errorf("entry[%d].IssueKey = %q, want %q", i, gotKeys[i], want[i])
		}
	}
}

func TestFilterTicketsWithComments_whitespaceCommentDoesNotCount(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "ZZT-3", Date: daysAgo(0), Comment: "   \t "},
	}
	if got := filterTicketsWithComments(entries); len(got) != 0 {
		t.Errorf("whitespace-only comment must not keep a ticket; kept %v", issueKeys(got))
	}
}

func TestFilterTicketsWithComments_subtaskCommentKeepsParentGroup(t *testing.T) {
	// Comment sits on a subtask row; the parent-keyed sibling row must survive
	// because merging groups on IssueKey only (SubtaskKey ignored).
	entries := []batch.Entry{
		{IssueKey: "ZZT-4", SubtaskKey: "", Date: daysAgo(0), Comment: ""},
		{IssueKey: "ZZT-4", SubtaskKey: "ZZT-5", Date: daysAgo(1), Comment: "subtask note"},
	}
	got := filterTicketsWithComments(entries)
	if len(got) != 2 {
		t.Errorf("expected both rows of ZZT-4 kept via subtask comment, got %v", issueKeys(got))
	}
}

func TestFilterTicketsWithComments_emptyInput(t *testing.T) {
	if got := filterTicketsWithComments(nil); len(got) != 0 {
		t.Errorf("filterTicketsWithComments(nil) kept %v, want none", issueKeys(got))
	}
}

// TestPrepareStandupEntries_commentOutsideWindowDropsTicket verifies the
// comment filter runs AFTER applyDayWindow: a comment older than the window
// must not keep the ticket's recent uncommented rows visible.
func TestPrepareStandupEntries_commentOutsideWindowDropsTicket(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "ZZT-6", Date: daysAgo(10), Comment: "old note, outside window"},
		{IssueKey: "ZZT-6", Date: daysAgo(1), Comment: ""},
	}
	got := prepareStandupEntries(entries, 3, false)
	if len(got) != 0 {
		t.Errorf("comment outside the 3-day window must not keep the ticket; kept %v", issueKeys(got))
	}
}

func TestPrepareStandupEntries_includeAllBypassesCommentFilter(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "ZZT-7", Date: daysAgo(1), Comment: ""},
	}
	got := prepareStandupEntries(entries, 3, true)
	if len(got) != 1 {
		t.Errorf("includeAll=true must keep uncommented tickets; kept %v", issueKeys(got))
	}
}

func TestPrepareStandupEntries_doesNotMutateInput(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "ZZT-8", Date: daysAgo(2), Comment: "older"},
		{IssueKey: "ZZT-9", Date: daysAgo(0), Comment: "newer"},
	}
	prepareStandupEntries(entries, 3, false)
	if entries[0].IssueKey != "ZZT-8" || entries[1].IssueKey != "ZZT-9" {
		t.Errorf("caller slice reordered: %v", issueKeys(entries))
	}
}

func TestBuildStandupBlob_countsReflectCommentFilter(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "ZZT-10", Description: "commented ticket", Date: daysAgo(0), Comment: "note one"},
		{IssueKey: "ZZT-10", Description: "commented ticket", Date: daysAgo(1), Comment: ""},
		{IssueKey: "ZZT-11", Description: "silent ticket", Date: daysAgo(0), Comment: ""},
	}
	blob, count, groupCount := buildStandupBlob(entries, 3, "", true, false, entryFormatOpts{})
	if count != 2 {
		t.Errorf("entry_count = %d, want 2 (both ZZT-10 rows, ZZT-11 dropped)", count)
	}
	if groupCount != 1 {
		t.Errorf("group_count = %d, want 1", groupCount)
	}
	if blob == "" {
		t.Error("blob is empty, want rendered ZZT-10 group")
	}
}
