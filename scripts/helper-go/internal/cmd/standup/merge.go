package standup

import (
	"fmt"
	"io"
	"strings"

	"github.com/dariuszw/hlp/internal/batch"
)

// entryGroup is one or more CSV rows merged by IssueKey. Members are stored
// in the order they were encountered; because grouping happens after
// sortEntriesNewestFirst, that order is newest-first.
type entryGroup struct {
	Members []batch.Entry
}

// mergeEntriesByIssueKey collapses rows sharing IssueKey into groups. Order
// across groups is preserved (a group's position is determined by its first
// member, i.e. its newest row when input is sorted newest-first).
//
// SubtaskKey is intentionally ignored: rows like "SUITE-8107" and
// "SUITE-8107 > SUITE-8219" merge into one group under the parent IssueKey.
// This is the user-confirmed merge semantics for `hlp standup show / publish`.
func mergeEntriesByIssueKey(entries []batch.Entry) []entryGroup {
	if len(entries) == 0 {
		return nil
	}
	idx := make(map[string]int, len(entries))
	groups := make([]entryGroup, 0, len(entries))
	for _, e := range entries {
		key := e.IssueKey
		if pos, ok := idx[key]; ok {
			groups[pos].Members = append(groups[pos].Members, e)
			continue
		}
		idx[key] = len(groups)
		groups = append(groups, entryGroup{Members: []batch.Entry{e}})
	}
	return groups
}

// formatGroup renders an entryGroup. Single-member groups delegate to
// formatEntry verbatim, so existing single-row output is byte-for-byte
// identical to the pre-merge behavior.
//
// Multi-member groups emit:
//
//	<header line per member, newest first>
//	<one IssueKey + Description line>     (subtask suffix dropped — see merge key choice)
//	<comment line per member that has a non-empty Comment, prefixed with "  · DD/MM — ">
func formatGroup(w io.Writer, g entryGroup, baseURL string, opts entryFormatOpts) {
	if len(g.Members) == 0 {
		return
	}
	if len(g.Members) == 1 {
		formatEntry(w, g.Members[0], baseURL, opts)
		return
	}

	// Only the newest header keeps the URL — subsequent rows would just
	// repeat /browse/<same-key>, which is noise.
	for i, e := range g.Members {
		writeHeaderLine(w, e, baseURL, opts, i == 0 /* includeURL */)
	}
	// Use the newest member (Members[0]) for the shared description line.
	// Subtask suffix is omitted because, with parent-only merging, members
	// may carry different subtasks (or none).
	writeKeyLine(w, g.Members[0], baseURL, opts, false /* includeSubtask */)

	for _, e := range g.Members {
		if strings.TrimSpace(e.Comment) == "" {
			continue
		}
		writeCommentLine(w, e, opts, groupCommentPrefix(e))
	}
}

// groupCommentPrefix returns "  · DD/MM — " for a member of a merged group.
// Status is intentionally omitted — it already appears on each stacked header
// line for that row, so duplicating it on the comment is noise.
func groupCommentPrefix(e batch.Entry) string {
	d, err := batch.ParseDate(e.Date, "09:00")
	dateStr := e.Date
	if err == nil {
		dateStr = d.Format("02/01")
	}
	return fmt.Sprintf("  · %s — ", dateStr)
}
