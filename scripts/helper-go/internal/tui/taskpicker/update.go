package taskpicker

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles navigation and filtering. The filter field holds focus, so
// plain characters type into it and navigation uses the arrow keys.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		if size, isSize := msg.(tea.WindowSizeMsg); isSize {
			m.width, m.height = size.Width, size.Height
		}
		return m, nil
	}

	switch key.String() {
	case "esc":
		m.quitting = true
		return m, tea.Quit

	case "ctrl+c":
		if m.filter.Value() != "" {
			m.filter.SetValue("")
			return m.applyFilter(), nil
		}
		m.quitting = true
		return m, tea.Quit

	case "enter":
		if len(m.shown) == 0 {
			return m, nil
		}
		m.chosen = m.shown[m.cursor].Key
		return m, tea.Quit

	case "up", "ctrl+p":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "ctrl+n":
		if m.cursor < len(m.shown)-1 {
			m.cursor++
		}
		return m, nil

	case "home":
		m.cursor = 0
		return m, nil

	case "end":
		m.cursor = max(len(m.shown)-1, 0)
		return m, nil
	}

	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(key)
	return m.applyFilter(), cmd
}
