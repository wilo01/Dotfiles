// Package standup provides the TUI dashboard for standup preparation
package standup

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	primaryColor   = lipgloss.Color("39")  // Cyan
	secondaryColor = lipgloss.Color("213") // Pink
	successColor   = lipgloss.Color("82")  // Green
	warningColor   = lipgloss.Color("214") // Orange
	mutedColor     = lipgloss.Color("241") // Gray
	highlightColor = lipgloss.Color("226") // Yellow
	borderColor    = lipgloss.Color("62")  // Purple-ish
)

// Base styles
var (
	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	// Day headers in list
	dayHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlightColor).
			MarginTop(1)

	// List items
	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				Background(lipgloss.Color("236"))

	// Time display
	timeStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Detail panel
	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	notesStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Italic(true)

	emptyNotesStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Footer
	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Borders
	listPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	detailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(0, 1)

	// Presentation view (full-width, all items visible)
	presentationPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(1, 2)

	// Status indicators
	statusDoneStyle = lipgloss.NewStyle().
			Foreground(successColor)

	statusPendingStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	// Day tabs bar
	tabBarStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("232")).
			Background(primaryColor).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 1)

	emptyTabStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Padding(0, 1)

	// Empty day message
	emptyDayStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)
)

// Divider creates a horizontal divider
func divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Repeat("─", width))
}
