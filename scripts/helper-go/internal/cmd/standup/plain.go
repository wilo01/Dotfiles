package standup

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
)

// entryFormatOpts controls presentation toggles for formatEntry.
// Color enables lipgloss/ANSI styling; Hyperlinks emits OSC-8 escapes on
// issue keys. Both are off in the publish path so the spreadsheet cell
// receives plain text.
type entryFormatOpts struct {
	Color      bool
	Hyperlinks bool
}

// formatEntry renders one worklog entry to w. Field order, exact text, and
// line breaks match `hlp standup show`'s historic output. The single source
// of truth shared by terminal (`show`) and webhook (`publish`) paths — toggling
// opts is the only difference between them.
//
// Implementation is a thin caller of writeHeaderLine / writeKeyLine /
// writeCommentLine so the merged-group formatter (formatGroup in merge.go)
// can reuse the same writers and stay consistent with single-row output.
func formatEntry(w io.Writer, e batch.Entry, baseURL string, opts entryFormatOpts) {
	writeHeaderLine(w, e, baseURL, opts, true /* includeURL */)
	writeKeyLine(w, e, baseURL, opts, true /* includeSubtask */)
	if strings.TrimSpace(e.Comment) != "" {
		writeCommentLine(w, e, opts, "" /* no prefix */)
	}
}

// writeHeaderLine emits the "Date [Type] TimeSpent Status  URL" line.
// includeURL=false suppresses the trailing URL (used by formatGroup on the
// 2nd+ stacked headers to avoid repeating the same /browse/<key> link).
func writeHeaderLine(w io.Writer, e batch.Entry, baseURL string, opts entryFormatOpts, includeURL bool) {
	d, _ := batch.ParseDate(e.Date, "09:00")

	header := d.Format("02/01/2006")
	if e.IssueType != "" {
		typeLabel := fmt.Sprintf("[%s]", e.IssueType)
		if opts.Color {
			typeLabel = ui.Muted.Render(typeLabel)
		}
		header += " " + typeLabel
	}
	if e.TimeSpent != "" {
		ts := e.TimeSpent
		if opts.Color {
			ts = ui.Success.Render(ts)
		}
		header += " " + ts
	}
	status := e.Status
	if status == "" {
		status = "PENDING"
	}
	if opts.Color {
		header += " " + ui.FormatStatus(status)
	} else {
		header += " " + status
	}
	if includeURL && baseURL != "" {
		jiraURL := baseURL + "/browse/" + e.IssueKey
		if opts.Color {
			jiraURL = ui.Link.Render(jiraURL)
		}
		header += "  " + jiraURL
	}
	fmt.Fprintln(w, header)
}

// writeKeyLine emits the "IssueKey[ > SubtaskKey] Description" line.
// includeSubtask=false is used by formatGroup when collapsing rows with
// differing subtasks under their shared parent IssueKey.
func writeKeyLine(w io.Writer, e batch.Entry, baseURL string, opts entryFormatOpts, includeSubtask bool) {
	issueDisplay := e.IssueKey
	if opts.Color {
		issueDisplay = ui.Primary.Render(e.IssueKey)
	}
	if opts.Hyperlinks && baseURL != "" {
		issueDisplay = ui.Hyperlink(issueDisplay, baseURL+"/browse/"+e.IssueKey)
	}
	keyLine := issueDisplay
	if includeSubtask && e.SubtaskKey != "" {
		subtask := e.SubtaskKey
		if opts.Color {
			subtask = ui.Primary.Render(subtask)
		}
		keyLine = issueDisplay + " > " + subtask
	}
	fmt.Fprintf(w, "%s %s\n", keyLine, e.Description)
}

// writeCommentLine emits a comment, optionally prefixed (used by formatGroup
// to render "  · DD/MM — " before each per-row comment).
func writeCommentLine(w io.Writer, e batch.Entry, opts entryFormatOpts, prefix string) {
	comment := prefix + e.Comment
	if opts.Color {
		comment = ui.Muted.Render(comment)
	}
	fmt.Fprintln(w, comment)
}

// filterStandupEntries applies the standup visibility rule.
// Drop ignored tickets unless the row carries a non-empty Comment ("promoted
// by note"). Used by show, publish, and notify so the rule lives in one place.
func filterStandupEntries(entries []batch.Entry, ignored map[string]bool) []batch.Entry {
	out := make([]batch.Entry, 0, len(entries))
	for _, e := range entries {
		if ignored[e.IssueKey] && strings.TrimSpace(e.Comment) == "" {
			continue
		}
		out = append(out, e)
	}
	return out
}

// applyDayWindow returns entries whose date is within the last `limitDays` days.
// limitDays <= 0 disables the window (returns input unchanged).
func applyDayWindow(entries []batch.Entry, limitDays int) []batch.Entry {
	if limitDays <= 0 {
		return entries
	}
	cutoff := time.Now().AddDate(0, 0, -limitDays)
	out := make([]batch.Entry, 0, len(entries))
	for _, e := range entries {
		d, err := batch.ParseDate(e.Date, "09:00")
		if err == nil && !d.Before(cutoff) {
			out = append(out, e)
		}
	}
	return out
}

// sortEntriesNewestFirst sorts in place; entries with parse errors sink to end.
func sortEntriesNewestFirst(entries []batch.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		di, erri := batch.ParseDate(entries[i].Date, "09:00")
		dj, errj := batch.ParseDate(entries[j].Date, "09:00")
		if erri != nil && errj != nil {
			return false
		}
		if erri != nil {
			return false
		}
		if errj != nil {
			return true
		}
		return di.After(dj)
	})
}

func renderPlainEntries(entries []batch.Entry, limitDays int, baseURL string, merge bool) error {
	sortEntriesNewestFirst(entries)
	entries = applyDayWindow(entries, limitDays)
	entries = filterStandupEntries(entries, getIgnoredTickets())

	if len(entries) == 0 {
		fmt.Println(ui.Warning("No entries found"))
		return nil
	}

	opts := entryFormatOpts{Color: true, Hyperlinks: true}
	if !merge {
		for i, e := range entries {
			formatEntry(os.Stdout, e, baseURL, opts)
			if i < len(entries)-1 {
				fmt.Println()
			}
		}
		return nil
	}

	groups := mergeEntriesByIssueKey(entries)
	for i, g := range groups {
		formatGroup(os.Stdout, g, baseURL, opts)
		if i < len(groups)-1 {
			fmt.Println()
		}
	}
	return nil
}

func getIgnoredTickets() map[string]bool {
	cfg, _ := config.Load()
	ignored := make(map[string]bool)
	for _, ticket := range cfg.Preferences.StandupIgnoredTickets {
		ignored[ticket] = true
	}
	return ignored
}
