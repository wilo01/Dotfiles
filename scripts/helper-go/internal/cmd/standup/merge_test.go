package standup

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dariuszw/hlp/internal/batch"
)

func TestMergeEntriesByIssueKey_groupsSameKey(t *testing.T) {
	entries := []batch.Entry{
		{IssueKey: "SUITE-8291", Date: "27/04/2026", Status: "DRAFT", Comment: "investigating"},
		{IssueKey: "SUITE-8291", Date: "24/04/2026", Status: "DONE", TimeSpent: "30m"},
	}
	groups := mergeEntriesByIssueKey(entries)
	if got, want := len(groups), 1; got != want {
		t.Fatalf("len(groups) = %d, want %d", got, want)
	}
	if got, want := len(groups[0].Members), 2; got != want {
		t.Fatalf("len(Members) = %d, want %d", got, want)
	}
	// Members preserve input order (newest first because input is sorted newest first).
	if groups[0].Members[0].Date != "27/04/2026" {
		t.Errorf("Members[0].Date = %q, want %q", groups[0].Members[0].Date, "27/04/2026")
	}
	if groups[0].Members[1].Date != "24/04/2026" {
		t.Errorf("Members[1].Date = %q, want %q", groups[0].Members[1].Date, "24/04/2026")
	}
}

func TestMergeEntriesByIssueKey_keepsOrderAcrossGroups(t *testing.T) {
	// Sorted newest-first; SUITE-A is encountered first because its newest row is newest overall.
	entries := []batch.Entry{
		{IssueKey: "SUITE-A", Date: "27/04/2026"},
		{IssueKey: "SUITE-B", Date: "26/04/2026"},
		{IssueKey: "SUITE-A", Date: "25/04/2026"},
		{IssueKey: "SUITE-C", Date: "24/04/2026"},
	}
	groups := mergeEntriesByIssueKey(entries)
	if got, want := len(groups), 3; got != want {
		t.Fatalf("len(groups) = %d, want %d", got, want)
	}
	wantOrder := []string{"SUITE-A", "SUITE-B", "SUITE-C"}
	for i, g := range groups {
		if g.Members[0].IssueKey != wantOrder[i] {
			t.Errorf("groups[%d].Members[0].IssueKey = %q, want %q", i, g.Members[0].IssueKey, wantOrder[i])
		}
	}
}

func TestMergeEntriesByIssueKey_subtaskKeyIgnored(t *testing.T) {
	// Parent-only merge key — subtask differences collapse into parent group.
	entries := []batch.Entry{
		{IssueKey: "SUITE-8107", SubtaskKey: "", Date: "26/04/2026"},
		{IssueKey: "SUITE-8107", SubtaskKey: "SUITE-8219", Date: "25/04/2026"},
	}
	groups := mergeEntriesByIssueKey(entries)
	if got, want := len(groups), 1; got != want {
		t.Fatalf("len(groups) = %d, want %d (subtask should not split groups)", got, want)
	}
	if got, want := len(groups[0].Members), 2; got != want {
		t.Errorf("len(Members) = %d, want %d", got, want)
	}
}

func TestMergeEntriesByIssueKey_emptyInput(t *testing.T) {
	if got := mergeEntriesByIssueKey(nil); got != nil {
		t.Errorf("mergeEntriesByIssueKey(nil) = %v, want nil", got)
	}
	if got := mergeEntriesByIssueKey([]batch.Entry{}); got != nil {
		t.Errorf("mergeEntriesByIssueKey([]) = %v, want nil", got)
	}
}

// TestFormatGroup_singleMember_matchesFormatEntry is a regression guard:
// for a 1-member group, formatGroup must emit byte-identical output to
// formatEntry so existing single-row standup output is unaffected by the
// merge feature.
func TestFormatGroup_singleMember_matchesFormatEntry(t *testing.T) {
	e := batch.Entry{
		IssueKey:    "SUITE-4478",
		IssueType:   "Story",
		Description: "Build view archive list popup and its functionalities",
		Date:        "27/04/2026",
		Status:      "DRAFT",
	}
	baseURL := "https://acre-identity.atlassian.net"
	opts := entryFormatOpts{}

	var direct, viaGroup bytes.Buffer
	formatEntry(&direct, e, baseURL, opts)
	formatGroup(&viaGroup, entryGroup{Members: []batch.Entry{e}}, baseURL, opts)

	if direct.String() != viaGroup.String() {
		t.Errorf("formatGroup(size=1) diverged from formatEntry\n--- formatEntry ---\n%s\n--- formatGroup ---\n%s", direct.String(), viaGroup.String())
	}
}

// TestFormatGroup_singleMember_preservesSubtaskOnKeyLine confirms that the
// "SUITE-X > SUITE-Y" subtask suffix on the description line is kept for
// 1-member groups (it's only dropped when ≥2 rows merge under one parent).
func TestFormatGroup_singleMember_preservesSubtaskOnKeyLine(t *testing.T) {
	e := batch.Entry{
		IssueKey:    "SUITE-8107",
		SubtaskKey:  "SUITE-8219",
		IssueType:   "Bug",
		Description: "Guest Dashboard issue",
		Date:        "27/04/2026",
		Status:      "DRAFT",
	}
	var buf bytes.Buffer
	formatGroup(&buf, entryGroup{Members: []batch.Entry{e}}, "", entryFormatOpts{})

	if !strings.Contains(buf.String(), "SUITE-8107 > SUITE-8219 Guest Dashboard issue") {
		t.Errorf("expected single-member group to keep subtask suffix; got:\n%s", buf.String())
	}
}

func TestFormatGroup_stackedHeadersAndPrefixedComments(t *testing.T) {
	members := []batch.Entry{
		{
			IssueKey:    "SUITE-8291",
			IssueType:   "Bug",
			Description: "Visitor booked did not get invite email -260413-292337",
			Date:        "27/04/2026",
			Status:      "DRAFT",
			Comment:     "still investigating",
		},
		{
			IssueKey:    "SUITE-8291",
			IssueType:   "Bug",
			Description: "Visitor booked did not get invite email -260413-292337",
			Date:        "24/04/2026",
			TimeSpent:   "30m",
			Status:      "DONE",
			Comment:     "Initial fix deployed",
		},
	}
	var buf bytes.Buffer
	formatGroup(&buf, entryGroup{Members: members}, "https://acre-identity.atlassian.net", entryFormatOpts{})
	got := buf.String()

	want := "27/04/2026 [Bug] DRAFT  https://acre-identity.atlassian.net/browse/SUITE-8291\n" +
		"24/04/2026 [Bug] 30m DONE\n" +
		"SUITE-8291 Visitor booked did not get invite email -260413-292337\n" +
		"  · 27/04 — still investigating\n" +
		"  · 24/04 — Initial fix deployed\n"

	if got != want {
		t.Errorf("unexpected merged-group output\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatGroup_skipsEmptyComments(t *testing.T) {
	members := []batch.Entry{
		{IssueKey: "SUITE-8291", IssueType: "Bug", Description: "X", Date: "27/04/2026", Status: "DRAFT", Comment: "still investigating"},
		{IssueKey: "SUITE-8291", IssueType: "Bug", Description: "X", Date: "24/04/2026", Status: "DONE", TimeSpent: "30m", Comment: ""},
	}
	var buf bytes.Buffer
	formatGroup(&buf, entryGroup{Members: members}, "", entryFormatOpts{})

	out := buf.String()
	commentLines := strings.Count(out, "  · ")
	if commentLines != 1 {
		t.Errorf("expected exactly 1 comment line (empty comment skipped), got %d in:\n%s", commentLines, out)
	}
}

// TestFormatEntry_whitespaceOnlyCommentSuppressed locks in formatEntry's
// TrimSpace check: a row whose Comment is only whitespace must not emit a
// blank comment line. Mirrors the rule used by filterStandupEntries.
func TestFormatEntry_whitespaceOnlyCommentSuppressed(t *testing.T) {
	e := batch.Entry{
		IssueKey:    "SUITE-9000",
		IssueType:   "Task",
		Description: "Whitespace check",
		Date:        "30/04/2026",
		Status:      "DRAFT",
		Comment:     "   \t  \n",
	}
	var buf bytes.Buffer
	formatEntry(&buf, e, "", entryFormatOpts{})

	got := buf.String()
	want := "30/04/2026 [Task] DRAFT\nSUITE-9000 Whitespace check\n"
	if got != want {
		t.Errorf("whitespace-only Comment must not produce a comment line\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestGroupCommentPrefix_dateOnly(t *testing.T) {
	got := groupCommentPrefix(batch.Entry{Date: "27/04/2026", Status: "DRAFT"})
	want := "  · 27/04 — "
	if got != want {
		t.Errorf("groupCommentPrefix = %q, want %q (status must not appear)", got, want)
	}
}
