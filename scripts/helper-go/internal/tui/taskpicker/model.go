// Package taskpicker provides the list of existing ticket tasks, so a command
// that needs a key can be run without typing one. It selects; the caller acts.
package taskpicker

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Item is one task in the list.
type Item struct {
	Key     string
	Summary string
	Repos   []string
	// Session is the tmux state: "claude", "idle" or "" when there is none.
	Session string
	// Hexer is the environment's host, empty when the task has none.
	Hexer string
	// LastOpened is a already-formatted relative age, e.g. "2h ago".
	LastOpened string
}

// haystack is what a filter query is matched against.
func (i Item) haystack() string {
	return strings.ToLower(i.Key + " " + i.Summary + " " + strings.Join(i.Repos, " "))
}

// Model is the bubbletea model for the task list.
type Model struct {
	title string
	all   []Item
	shown []Item

	cursor   int
	filter   textinput.Model
	width    int
	height   int
	chosen   string
	quitting bool
}

// New builds the picker over the given tasks, most recently opened first.
func New(title string, items []Item) Model {
	filter := textinput.New()
	filter.Prompt = "/ "
	filter.Placeholder = "filter by key, summary or repo"
	filter.CharLimit = 80
	// Without an explicit width the field renders one character wide and the
	// placeholder is clipped to nothing useful.
	filter.Width = 48
	filter.Focus()

	return Model{title: title, all: items, shown: items, filter: filter}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

// Chosen is the selected task key, empty when the user cancelled.
func (m Model) Chosen() string { return m.chosen }

// applyFilter narrows the list, keeping the cursor on a valid row.
func (m Model) applyFilter() Model {
	query := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if query == "" {
		m.shown = m.all
	} else {
		shown := make([]Item, 0, len(m.all))
		for _, item := range m.all {
			if matchesAllTerms(item.haystack(), query) {
				shown = append(shown, item)
			}
		}
		m.shown = shown
	}
	if m.cursor >= len(m.shown) {
		m.cursor = max(len(m.shown)-1, 0)
	}
	return m
}

// matchesAllTerms requires every whitespace-separated term, so "suite hexer"
// finds a task spanning both without caring about their order.
func matchesAllTerms(haystack, query string) bool {
	for term := range strings.FieldsSeq(query) {
		if !strings.Contains(haystack, term) {
			return false
		}
	}
	return true
}

// Run shows the list and returns the chosen key, or "" if the user cancelled.
func Run(title string, items []Item) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no tasks yet — run: hlp agent start <TICKET-KEY>")
	}
	final, err := tea.NewProgram(New(title, items)).Run()
	if err != nil {
		return "", fmt.Errorf("task picker failed: %w", err)
	}
	m, ok := final.(Model)
	if !ok {
		return "", fmt.Errorf("task picker returned an unexpected model")
	}
	return m.Chosen(), nil
}
