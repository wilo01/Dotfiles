package taskpicker

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// visibleRows caps the list so the filter and footer stay on screen.
const visibleRows = 15

var (
	primaryColor = lipgloss.Color("39")
	accentColor  = lipgloss.Color("213")
	mutedColor   = lipgloss.Color("241")
	activeColor  = lipgloss.Color("82")
	textColor    = lipgloss.Color("252")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("232")).
			Background(primaryColor).
			Padding(0, 1)

	keyStyle         = lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	selectedKeyStyle = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	summaryStyle     = lipgloss.NewStyle().Foreground(textColor)
	metaStyle        = lipgloss.NewStyle().Foreground(mutedColor)
	liveStyle        = lipgloss.NewStyle().Foreground(activeColor)
	footerStyle      = lipgloss.NewStyle().Foreground(mutedColor)
)

// View renders the filter, the list and the footer.
func (m Model) View() string {
	if m.quitting || m.chosen != "" {
		return ""
	}

	width := m.width
	if width <= 0 {
		width = 100
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n  ")
	b.WriteString(m.filter.View())
	b.WriteString("\n\n")

	if len(m.shown) == 0 {
		b.WriteString(metaStyle.Render("  no task matches that filter"))
		b.WriteString("\n\n")
		b.WriteString(footerStyle.Render("  ctrl+c clears the filter · esc cancel"))
		return b.String()
	}

	start, end := m.window()
	if start > 0 {
		b.WriteString(metaStyle.Render("  ↑ more"))
		b.WriteString("\n")
	}
	keyWidth := m.keyWidth()
	for i := start; i < end; i++ {
		b.WriteString(m.renderRow(i, keyWidth, width))
		b.WriteString("\n")
	}
	if end < len(m.shown) {
		b.WriteString(metaStyle.Render("  ↓ more"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("  ↑↓ move · type to filter · ⏎ open · ctrl+c clear · esc cancel"))
	return b.String()
}

// keyWidth aligns the summaries against the longest ticket key on screen.
func (m Model) keyWidth() int {
	widest := 0
	for _, item := range m.shown {
		widest = max(widest, lipgloss.Width(item.Key))
	}
	return widest
}

func (m Model) window() (int, int) {
	if len(m.shown) <= visibleRows {
		return 0, len(m.shown)
	}
	start := max(m.cursor-visibleRows/2, 0)
	if start+visibleRows > len(m.shown) {
		start = len(m.shown) - visibleRows
	}
	return start, start + visibleRows
}

func (m Model) renderRow(i, keyWidth, width int) string {
	item := m.shown[i]
	focused := i == m.cursor

	pointer := "  "
	style := keyStyle
	if focused {
		pointer = selectedKeyStyle.Render("▶ ")
		style = selectedKeyStyle
	}

	key := style.Render(item.Key + strings.Repeat(" ", max(keyWidth-lipgloss.Width(item.Key), 0)))
	line := pointer + key + "  " + summaryStyle.Render(truncate(item.Summary, width-keyWidth-34))

	if meta := m.meta(item); meta != "" {
		line += "  " + meta
	}
	return line
}

// meta is the trailing status: a live session first, since that is what tells
// you whether a ticket is already in progress.
func (m Model) meta(item Item) string {
	var parts []string
	switch item.Session {
	case "claude":
		parts = append(parts, liveStyle.Render("● claude"))
	case "idle":
		parts = append(parts, metaStyle.Render("○ idle"))
	}
	if len(item.Repos) > 0 {
		parts = append(parts, metaStyle.Render(strings.Join(item.Repos, ",")))
	}
	if item.Hexer != "" {
		parts = append(parts, metaStyle.Render("hexer"))
	}
	if item.LastOpened != "" {
		parts = append(parts, metaStyle.Render(item.LastOpened))
	}
	return strings.Join(parts, metaStyle.Render(" · "))
}

func truncate(s string, limit int) string {
	if limit < 8 {
		limit = 8
	}
	if lipgloss.Width(s) <= limit {
		return s
	}
	return s[:limit-1] + "…"
}
