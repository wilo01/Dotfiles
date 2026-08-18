package taskpicker

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		next, _ := m.Update(key(k))
		m = next.(Model)
	}
	return m
}

func typeText(m Model, text string) Model {
	for _, r := range text {
		m = press(m, string(r))
	}
	return m
}

func items() []Item {
	return []Item{
		{Key: "SUITE-9250", Summary: "Biomarin - can't edit message templates", Repos: []string{"tds-suite"}, Session: "claude"},
		{Key: "SUITE-9215", Summary: "Kiosk keyboard caps lock", Repos: []string{"tds-suite", "tds-hexer"}},
		{Key: "VI-3034", Summary: "Fidelity UAT addAlternateHost", Repos: []string{"tds-suite-api"}, Session: "idle"},
	}
}

func shownKeys(m Model) []string {
	out := make([]string, 0, len(m.shown))
	for _, i := range m.shown {
		out = append(out, i.Key)
	}
	return out
}

func TestEnterPicksTheFocusedTask(t *testing.T) {
	m := press(New("OPEN TASK", items()), "enter")

	if m.Chosen() != "SUITE-9250" {
		t.Errorf("chose %q, want the first task", m.Chosen())
	}
}

func TestArrowsMoveTheCursorWithoutWrapping(t *testing.T) {
	m := New("OPEN TASK", items())

	if m = press(m, "up"); m.cursor != 0 {
		t.Errorf("up at the top moved to %d", m.cursor)
	}
	m = press(m, "down", "down", "down", "down")
	if m.cursor != 2 {
		t.Errorf("down past the end moved to %d, want 2", m.cursor)
	}
	if m = press(m, "enter"); m.Chosen() != "VI-3034" {
		t.Errorf("chose %q", m.Chosen())
	}
}

func TestFilterMatchesKeySummaryAndRepo(t *testing.T) {
	cases := map[string][]string{
		"9250":     {"SUITE-9250"},
		"biomarin": {"SUITE-9250"},
		"hexer":    {"SUITE-9215"},
		"vi-":      {"VI-3034"},
		"suite-9":  {"SUITE-9250", "SUITE-9215"},
		// Repo names are searched too, so "suite-" also matches VI-3034 through
		// its tds-suite-api worktree.
		"suite-": {"SUITE-9250", "SUITE-9215", "VI-3034"},
	}
	for query, want := range cases {
		m := typeText(New("OPEN TASK", items()), query)
		if got := shownKeys(m); !reflect.DeepEqual(got, want) {
			t.Errorf("filter %q showed %v, want %v", query, got, want)
		}
	}
}

// Terms are ANDed and order-independent, so a task can be found by naming two
// of its repos or a key plus a word.
func TestFilterRequiresEveryTerm(t *testing.T) {
	m := typeText(New("OPEN TASK", items()), "hexer suite")
	if got := shownKeys(m); !reflect.DeepEqual(got, []string{"SUITE-9215"}) {
		t.Errorf("got %v", got)
	}

	m = typeText(New("OPEN TASK", items()), "biomarin hexer")
	if got := shownKeys(m); len(got) != 0 {
		t.Errorf("terms should be ANDed, got %v", got)
	}
}

func TestFilterIsCaseInsensitive(t *testing.T) {
	m := typeText(New("OPEN TASK", items()), "BIOMARIN")
	if got := shownKeys(m); !reflect.DeepEqual(got, []string{"SUITE-9250"}) {
		t.Errorf("got %v", got)
	}
}

func TestFilterKeepsTheCursorInRange(t *testing.T) {
	m := press(New("OPEN TASK", items()), "down", "down")
	if m.cursor != 2 {
		t.Fatalf("precondition: cursor = %d", m.cursor)
	}

	m = typeText(m, "biomarin")
	if m.cursor != 0 {
		t.Errorf("cursor left dangling at %d after filtering to one row", m.cursor)
	}
	if m = press(m, "enter"); m.Chosen() != "SUITE-9250" {
		t.Errorf("chose %q", m.Chosen())
	}
}

func TestEnterOnAnEmptyFilterDoesNothing(t *testing.T) {
	m := typeText(New("OPEN TASK", items()), "zzzz")
	if got := shownKeys(m); len(got) != 0 {
		t.Fatalf("expected no matches, got %v", got)
	}

	m = press(m, "enter")
	if m.Chosen() != "" {
		t.Errorf("picked %q with nothing shown", m.Chosen())
	}
}

func TestCtrlCClearsTheFilterBeforeCancelling(t *testing.T) {
	m := typeText(New("OPEN TASK", items()), "biomarin")

	m = press(m, "ctrl+c")
	if got := len(shownKeys(m)); got != 3 {
		t.Errorf("first ctrl+c should clear the filter, %d rows shown", got)
	}
	if m.quitting {
		t.Error("first ctrl+c quit instead of clearing")
	}

	if m = press(m, "ctrl+c"); !m.quitting {
		t.Error("second ctrl+c did not cancel")
	}
	if m.Chosen() != "" {
		t.Errorf("cancelling returned %q", m.Chosen())
	}
}

func TestEscCancels(t *testing.T) {
	m := press(New("OPEN TASK", items()), "esc")

	if m.Chosen() != "" {
		t.Errorf("esc returned %q", m.Chosen())
	}
}

func TestRunRefusesAnEmptyList(t *testing.T) {
	if _, err := Run("OPEN TASK", nil); err == nil {
		t.Fatal("expected an error with no tasks")
	}
}
