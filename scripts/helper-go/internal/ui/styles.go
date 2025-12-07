package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	ColorPrimary   = lipgloss.Color("39")  // Cyan
	ColorSecondary = lipgloss.Color("213") // Pink
	ColorSuccess   = lipgloss.Color("82")  // Green
	ColorWarning   = lipgloss.Color("214") // Orange
	ColorError     = lipgloss.Color("196") // Red
	ColorMuted     = lipgloss.Color("241") // Gray
	ColorHighlight = lipgloss.Color("226") // Yellow
)

// Styles
var (
	// Text styles
	Bold      = lipgloss.NewStyle().Bold(true)
	Italic    = lipgloss.NewStyle().Italic(true)
	Underline = lipgloss.NewStyle().Underline(true)

	// Semantic styles
	Primary     = lipgloss.NewStyle().Foreground(ColorPrimary)
	Secondary   = lipgloss.NewStyle().Foreground(ColorSecondary)
	Success     = lipgloss.NewStyle().Foreground(ColorSuccess)
	WarningText = lipgloss.NewStyle().Foreground(ColorWarning)
	ErrorText   = lipgloss.NewStyle().Foreground(ColorError)
	Muted       = lipgloss.NewStyle().Foreground(ColorMuted)
	Highlight   = lipgloss.NewStyle().Foreground(ColorHighlight)

	// Combined styles
	Title       = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	Subtitle    = lipgloss.NewStyle().Foreground(ColorMuted)
	Label       = lipgloss.NewStyle().Foreground(ColorMuted)
	Value       = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	Code        = lipgloss.NewStyle().Background(lipgloss.Color("236")).Padding(0, 1)
	SuccessBold = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess)
	ErrorBold   = lipgloss.NewStyle().Bold(true).Foreground(ColorError)

	// Box styles
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(0, 1)

	SuccessBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(0, 1)

	ErrorBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(0, 1)
)

// Logo returns the ASCII art logo
func Logo() string {
	logo := `
  _     _
 | |__ | |_ __
 | '_ \| | '_ \
 | | | | | |_) |
 |_| |_|_| .__/
         |_|    `
	return Primary.Render(logo)
}

// StatusIcon returns an icon for the given status
func StatusIcon(success bool) string {
	if success {
		return Success.Render("✓")
	}
	return ErrorText.Render("✗")
}

// Spinner characters for progress indication
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
