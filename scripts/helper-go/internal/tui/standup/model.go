package standup

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/pkg/duration"
)

// maxVisibleTabs is the maximum number of day tabs visible at once
const maxVisibleTabs = 7

// Model is the bubbletea model for the standup TUI
type Model struct {
	// Data
	days    []DayGroup // Entries grouped by day (only days with entries)
	csvPath string     // Path to CSV for future saving

	// Navigation state (dual-index for tab-style navigation)
	dayIndex   int  // Currently selected day (0 = newest)
	itemIndex  int  // Selected item within current day (-1 if day is empty)
	tabOffset  int  // First visible tab index (for scrolling tabs)
	totalItems int  // Total number of items across all days
	ready      bool // Terminal size received

	// UI dimensions
	width  int
	height int

	// Scrollable content viewport
	viewport viewport.Model

	// Key bindings
	keys keyMap
}

// DayGroup represents entries for a single day
type DayGroup struct {
	Date    time.Time
	Day     string // "THURSDAY"
	Entries []StandupItem
	Total   time.Duration
}

// StandupItem represents a single worklog entry for display
type StandupItem struct {
	Entry     batch.Entry   // Original CSV entry
	TimeSpent time.Duration // Parsed duration
}

// keyMap defines the key bindings
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Quit  key.Binding
	Help  key.Binding
	Enter key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev day"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next day"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
	}
}

// NewModel creates a new standup TUI model
func NewModel(entries []batch.Entry, csvPath string) Model {
	days := groupByDay(entries)

	// Count total items
	total := 0
	for _, d := range days {
		total += len(d.Entries)
	}

	// Start with first item selected (newest day, first item)
	initialItemIndex := 0
	if len(days) == 0 || len(days[0].Entries) == 0 {
		initialItemIndex = -1
	}

	return Model{
		days:       days,
		csvPath:    csvPath,
		dayIndex:   0,
		itemIndex:  initialItemIndex,
		tabOffset:  0,
		totalItems: total,
		viewport:   viewport.New(80, 20), // Will resize on WindowSizeMsg
		keys:       defaultKeyMap(),
	}
}

// groupByDay groups entries by date, sorted newest first
func groupByDay(entries []batch.Entry) []DayGroup {
	// Group by date string
	groups := make(map[string]*DayGroup)

	for _, e := range entries {
		// Parse date
		entryDate, err := batch.ParseDate(e.Date, "00:00")
		if err != nil {
			continue
		}

		dateKey := entryDate.Format("2006-01-02")

		// Parse duration
		dur, _ := duration.Parse(e.TimeSpent)

		item := StandupItem{
			Entry:     e,
			TimeSpent: dur,
		}

		if g, ok := groups[dateKey]; ok {
			g.Entries = append(g.Entries, item)
			g.Total += dur
		} else {
			groups[dateKey] = &DayGroup{
				Date:    entryDate,
				Day:     strings.ToUpper(entryDate.Format("Monday")),
				Entries: []StandupItem{item},
				Total:   dur,
			}
		}
	}

	// Convert to slice and sort by date (newest first)
	var result []DayGroup
	for _, g := range groups {
		result = append(result, *g)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Date.After(result[j].Date)
	})

	return result
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// getSelectedItem returns the currently selected item and its day
func (m Model) getSelectedItem() (*StandupItem, *DayGroup) {
	if len(m.days) == 0 {
		return nil, nil
	}

	day := m.getCurrentDay()
	if day == nil {
		return nil, nil
	}

	// Empty day - return day but nil item
	if len(day.Entries) == 0 || m.itemIndex < 0 {
		return nil, day
	}

	// Clamp itemIndex just in case
	idx := m.itemIndex
	if idx >= len(day.Entries) {
		idx = len(day.Entries) - 1
	}

	return &day.Entries[idx], day
}

// getCurrentDay returns the currently selected day
func (m Model) getCurrentDay() *DayGroup {
	if len(m.days) == 0 || m.dayIndex < 0 || m.dayIndex >= len(m.days) {
		return nil
	}
	return &m.days[m.dayIndex]
}

// getTotalTime calculates total time across all days
func (m Model) getTotalTime() time.Duration {
	var total time.Duration
	for _, d := range m.days {
		total += d.Total
	}
	return total
}

// getVisibleTabRange returns the range of visible tabs for scrolling
func (m Model) getVisibleTabRange() (start, end int) {
	start = m.tabOffset
	end = m.tabOffset + maxVisibleTabs
	if end > len(m.days) {
		end = len(m.days)
	}
	return start, end
}
