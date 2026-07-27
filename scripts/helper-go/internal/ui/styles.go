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
	ColorMuted     = lipgloss.Color("246") // Gray (lighter for visibility)
	ColorHighlight = lipgloss.Color("226") // Yellow
	ColorLink      = lipgloss.Color("75")  // Light blue (hyperlink)
)

// Branch colors for git graph visualization (VSCode Git Graph style)
var BranchColors = []lipgloss.Color{
	lipgloss.Color("#F14C4C"), // Red
	lipgloss.Color("#3B8EEA"), // Blue
	lipgloss.Color("#23D18B"), // Green
	lipgloss.Color("#E5E510"), // Yellow
	lipgloss.Color("#BC3FBC"), // Purple
	lipgloss.Color("#29B8DB"), // Cyan
	lipgloss.Color("#E48B39"), // Orange
	lipgloss.Color("#DDA0DD"), // Plum
}

// BranchStyle returns a style with the color for the given branch index
func BranchStyle(index int) lipgloss.Style {
	color := BranchColors[index%len(BranchColors)]
	return lipgloss.NewStyle().Foreground(color)
}

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
	Link        = lipgloss.NewStyle().Foreground(ColorLink)

	// Combined styles
	Title       = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	Subtitle    = lipgloss.NewStyle().Foreground(ColorMuted)
	Label       = lipgloss.NewStyle().Foreground(ColorMuted)
	Value       = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	Code        = lipgloss.NewStyle().Background(lipgloss.Color("236")).Padding(0, 1)
	SuccessBold = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess)
	WarningBold = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning)
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
