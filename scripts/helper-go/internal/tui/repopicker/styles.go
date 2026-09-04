package repopicker

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor = lipgloss.Color("39")
	accentColor  = lipgloss.Color("213")
	mutedColor   = lipgloss.Color("241")
	warnColor    = lipgloss.Color("214")
	textColor    = lipgloss.Color("252")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("232")).
			Background(primaryColor).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().Foreground(textColor)

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor)

	inTaskStyle = lipgloss.NewStyle().Foreground(accentColor)

	disabledStyle = lipgloss.NewStyle().Foreground(mutedColor)

	sectionStyle = lipgloss.NewStyle().Foreground(mutedColor)

	labelStyle = lipgloss.NewStyle().Foreground(mutedColor)

	valueStyle = lipgloss.NewStyle().Foreground(textColor)

	warnStyle = lipgloss.NewStyle().Foreground(warnColor)

	footerStyle = lipgloss.NewStyle().Foreground(mutedColor)

	descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	hintStyle = lipgloss.NewStyle().Foreground(mutedColor).Italic(true)

	mentionedStyle = lipgloss.NewStyle().Foreground(warnColor)
)

func sectionRule(label string, width int) string {
	prefix := "── " + label + " "
	if pad := width - lipgloss.Width(prefix); pad > 0 {
		prefix += strings.Repeat("─", pad)
	}
	return sectionStyle.Render(prefix)
}
