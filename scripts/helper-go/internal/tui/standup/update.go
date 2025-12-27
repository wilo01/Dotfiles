package standup

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Resize viewport (leave room for header/tabs/footer)
		headerHeight := 4 // header + tabs
		footerHeight := 2
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - headerHeight - footerHeight

		m.ready = true
		return m, nil
	}

	// Let viewport handle its own messages (mouse scroll, etc.)
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// handleKeyMsg processes keyboard input
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Left):
		return m.moveDayLeft(), nil

	case key.Matches(msg, m.keys.Right):
		return m.moveDayRight(), nil

	case key.Matches(msg, m.keys.Up):
		return m.moveItemUp(), nil

	case key.Matches(msg, m.keys.Down):
		return m.moveItemDown(), nil
	}

	return m, nil
}

// moveDayLeft moves visually left (to newer days = lower index since index 0 is newest)
func (m Model) moveDayLeft() Model {
	if m.dayIndex > 0 {
		m.dayIndex--
		m.itemIndex = m.clampItemIndex(m.dayIndex, m.itemIndex)

		// Scroll tabs if selected day is before visible range
		if m.dayIndex < m.tabOffset {
			m.tabOffset = m.dayIndex
		}

		// Reset viewport to top when switching days
		m.viewport.SetYOffset(0)
	}
	return m
}

// moveDayRight moves visually right (to older days = higher index)
func (m Model) moveDayRight() Model {
	if m.dayIndex < len(m.days)-1 {
		m.dayIndex++
		m.itemIndex = m.clampItemIndex(m.dayIndex, m.itemIndex)

		// Scroll tabs if selected day is beyond visible range
		if m.dayIndex >= m.tabOffset+maxVisibleTabs {
			m.tabOffset = m.dayIndex - maxVisibleTabs + 1
		}

		// Reset viewport to top when switching days
		m.viewport.SetYOffset(0)
	}
	return m
}

// moveItemUp moves cursor up within current day only
func (m Model) moveItemUp() Model {
	if m.itemIndex > 0 {
		m.itemIndex--
		m.scrollToItem()
	}
	return m
}

// moveItemDown moves cursor down within current day only
func (m Model) moveItemDown() Model {
	day := m.getCurrentDay()
	if day == nil {
		return m
	}

	if m.itemIndex < len(day.Entries)-1 {
		m.itemIndex++
		m.scrollToItem()
	}
	return m
}

// scrollToItem scrolls viewport to keep selected item visible
func (m *Model) scrollToItem() {
	// Each item is approximately 4 lines tall (number+key, title, spacing)
	itemHeight := 4
	headerLines := 3 // Day header + blank line

	targetLine := headerLines + (m.itemIndex * itemHeight)

	// Scroll if item is above visible area
	if targetLine < m.viewport.YOffset {
		m.viewport.SetYOffset(targetLine)
	}

	// Scroll if item is below visible area
	if targetLine >= m.viewport.YOffset+m.viewport.Height-itemHeight {
		m.viewport.SetYOffset(targetLine - m.viewport.Height + itemHeight + 2)
	}
}

// clampItemIndex ensures itemIndex is valid for given day
// Preserves position when possible, resets to 0 or -1 otherwise
func (m Model) clampItemIndex(dayIdx, itemIdx int) int {
	if dayIdx < 0 || dayIdx >= len(m.days) {
		return -1
	}

	dayEntries := len(m.days[dayIdx].Entries)
	if dayEntries == 0 {
		return -1 // Empty day
	}

	if itemIdx < 0 {
		return 0 // Was on empty day, now on first item
	}

	if itemIdx >= dayEntries {
		return dayEntries - 1 // Clamp to last item
	}

	return itemIdx // Preserve position
}
