package standup

import (
	"fmt"
	"sort"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
)

func renderPlainEntries(entries []batch.Entry, limitDays int) error {
	// Sort newest first (entries with parse errors go to end)
	sort.Slice(entries, func(i, j int) bool {
		di, erri := batch.ParseDate(entries[i].Date, "09:00")
		dj, errj := batch.ParseDate(entries[j].Date, "09:00")
		if erri != nil && errj != nil {
			return false
		}
		if erri != nil {
			return false // i goes after j
		}
		if errj != nil {
			return true // i goes before j
		}
		return di.After(dj)
	})

	// Filter by days if specified
	if limitDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -limitDays)
		filtered := []batch.Entry{}
		for _, e := range entries {
			d, err := batch.ParseDate(e.Date, "09:00")
			if err == nil && !d.Before(cutoff) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	// Filter out ignored tickets from config
	ignored := getIgnoredTickets()
	filtered := []batch.Entry{}
	for _, e := range entries {
		if !ignored[e.IssueKey] {
			filtered = append(filtered, e)
		}
	}
	entries = filtered

	if len(entries) == 0 {
		fmt.Println(ui.Warning("No entries found"))
		return nil
	}

	for i, e := range entries {
		printPlainEntry(e)
		if i < len(entries)-1 {
			fmt.Println(ui.Divider(ui.GetTerminalWidth()))
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

func printPlainEntry(e batch.Entry) {
	d, _ := batch.ParseDate(e.Date, "09:00")

	// Line 1: Date [IssueType] Time Status (each with own color)
	header := ui.Muted.Render(d.Format("02/01/2006"))
	if e.IssueType != "" {
		header += " " + ui.Muted.Render(fmt.Sprintf("[%s]", e.IssueType))
	}
	if e.TimeSpent != "" {
		header += " " + ui.Success.Render(e.TimeSpent)
	}
	if e.Status != "" {
		header += " " + formatStatus(e.Status)
	}
	fmt.Println(header)

	// Line 2: Key + Description (one line)
	keyLine := e.IssueKey
	if e.SubtaskKey != "" {
		keyLine = e.IssueKey + " > " + e.SubtaskKey
	}
	fmt.Printf("%s %s\n", ui.Primary.Render(keyLine), e.Description)

	// Line 4: Comment (only if present)
	if e.Comment != "" {
		fmt.Println(ui.Muted.Render(e.Comment))
	}
}

func formatStatus(status string) string {
	switch status {
	case "DONE", "SYNC", "UPDATED":
		return ui.Success.Render(status)
	case "DRAFT":
		return ui.Muted.Render(status)
	default:
		return ui.SuccessBold.Render(status)
	}
}
