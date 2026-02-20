package standup

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dariuszw/hlp/internal/ui"
)

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if len(m.days) == 0 {
		return m.renderEmpty()
	}

	header := m.renderHeader()
	dayTabs := m.renderDayTabs()
	footer := m.renderFooter()

	// Render content and set it in viewport
	// TODO: SetContent on value receiver - changes are lost after View() returns.
	// Move SetContent to Update() method where state mutations belong.
	content := m.renderPresentationContent()
	m.viewport.SetContent(content)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		dayTabs,
		m.viewport.View(), // Scrollable content area
		footer,
	)
}

// renderEmpty shows a message when no entries exist
func (m Model) renderEmpty() string {
	msg := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Render("No worklog entries found for the selected date range")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
}

// renderHeader renders the title bar
func (m Model) renderHeader() string {
	today := time.Now().Format("Mon 02 Jan 2006")
	title := titleStyle.Render(fmt.Sprintf("Standup Dashboard - %s", today))

	quitHint := helpKeyStyle.Render("[q]") + helpDescStyle.Render(" quit")

	// Create header with title left, quit hint right
	headerWidth := m.width - 2
	gap := headerWidth - lipgloss.Width(title) - lipgloss.Width(quitHint)
	if gap < 0 {
		gap = 1
	}

	return title + strings.Repeat(" ", gap) + quitHint + "\n"
}

// renderFooter renders the status bar
func (m Model) renderFooter() string {
	if len(m.days) == 0 {
		return "\n" + footerStyle.Render("No entries | [q] quit")
	}

	// Get date range from actual data (newest to oldest)
	newestDay := m.days[0].Date.Format("02 Jan")
	oldestDay := m.days[len(m.days)-1].Date.Format("02 Jan")
	dateRange := fmt.Sprintf("%s - %s", oldestDay, newestDay)

	total := formatDuration(m.getTotalTime())
	dayPos := fmt.Sprintf("Day %d/%d", m.dayIndex+1, len(m.days))

	status := fmt.Sprintf("%s | %s | Total: %s | %d items | [h/l] days [j/k] items [q] quit",
		dateRange, dayPos, total, m.totalItems)

	return "\n" + footerStyle.Render(status)
}

// renderDayTabs renders the day navigation tabs (scrolling window)
func (m Model) renderDayTabs() string {
	if len(m.days) == 0 {
		return tabBarStyle.Render("No entries")
	}

	start, end := m.getVisibleTabRange()
	var tabs []string

	for i := start; i < end; i++ {
		day := m.days[i]
		isActive := i == m.dayIndex

		// Format: "Mon 23 (3)"
		label := fmt.Sprintf("%s(%d)", day.Date.Format("Mon 02"), len(day.Entries))

		// Style based on state
		var tabStyle lipgloss.Style
		if isActive {
			tabStyle = activeTabStyle
		} else {
			tabStyle = inactiveTabStyle
		}

		tabs = append(tabs, tabStyle.Render(label))
	}

	// Show scroll indicators when there are more days
	leftIndicator := "   "
	rightIndicator := "   "

	if start > 0 {
		leftIndicator = helpKeyStyle.Render("[←]") // More newer days to the left
	}
	if end < len(m.days) {
		rightIndicator = helpKeyStyle.Render("[→]") // More older days to the right
	}

	tabBar := leftIndicator + " " + strings.Join(tabs, " │ ") + " " + rightIndicator

	return tabBarStyle.Render(tabBar)
}

// renderListPanel renders the left panel with CURRENT day's items only
func (m Model) renderListPanel(width int) string {
	var sb strings.Builder
	contentHeight := m.height - 8 // Account for header, tabs, footer, borders

	day := m.getCurrentDay()
	if day == nil {
		return listPanelStyle.Width(width).Height(contentHeight).Render("No data")
	}

	// Day header with date and total time
	dayHeader := fmt.Sprintf("%s - %s", day.Day, day.Date.Format("02 Jan"))
	sb.WriteString(dayHeaderStyle.Render(dayHeader))
	sb.WriteString("\n")

	if day.Total > 0 {
		sb.WriteString(timeStyle.Render(fmt.Sprintf("Total: %s", formatDuration(day.Total))))
		sb.WriteString("\n")
	}

	sb.WriteString(divider(width - 4))
	sb.WriteString("\n")

	// Items for current day only
	if len(day.Entries) == 0 {
		sb.WriteString(emptyDayStyle.Render("No entries logged"))
	} else {
		for i, item := range day.Entries {
			isSelected := i == m.itemIndex
			line := m.renderListItem(item, isSelected, width-4)
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	// Apply panel style
	return listPanelStyle.
		Width(width).
		Height(contentHeight).
		Render(sb.String())
}

// renderListItem renders a single item in the list
func (m Model) renderListItem(item StandupItem, selected bool, width int) string {
	// Format: > VIS-1234 (2h 30m)
	prefix := "  "
	if selected {
		prefix = "> "
	}

	timeStr := formatDuration(item.TimeSpent)
	ticketInfo := fmt.Sprintf("%s%s (%s)", prefix, item.Entry.IssueKey, timeStr)

	// Truncate if too long
	if len(ticketInfo) > width {
		ticketInfo = ticketInfo[:width-3] + "..."
	}

	if selected {
		return selectedItemStyle.Render(ticketInfo)
	}
	return itemStyle.Render(ticketInfo)
}

// renderDetailPanel renders the right panel with item details
func (m Model) renderDetailPanel(width int) string {
	contentHeight := m.height - 8 // Account for tabs

	item, day := m.getSelectedItem()

	// Handle empty day
	if item == nil && day != nil {
		return m.renderEmptyDayDetail(day, width, contentHeight)
	}

	if item == nil {
		return detailPanelStyle.
			Width(width).
			Height(contentHeight).
			Render("No item selected")
	}

	var sb strings.Builder

	// Ticket ID
	sb.WriteString(detailTitleStyle.Render(item.Entry.IssueKey))
	sb.WriteString("\n")

	// Description
	if item.Entry.Description != "" {
		desc := wrapText(item.Entry.Description, width-4)
		sb.WriteString(detailValueStyle.Render(desc))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Subtask info (if applicable)
	if item.Entry.SubtaskKey != "" {
		sb.WriteString(detailLabelStyle.Render("Subtask: "))
		sb.WriteString(detailValueStyle.Render(item.Entry.SubtaskKey))
		sb.WriteString("\n")
	}

	// Issue type
	if item.Entry.IssueType != "" {
		sb.WriteString(detailLabelStyle.Render("Type: "))
		sb.WriteString(detailValueStyle.Render(item.Entry.IssueType))
		sb.WriteString("\n")
	}

	// Where worklog goes
	if item.Entry.SubtaskLogInd == "Y" {
		sb.WriteString(detailLabelStyle.Render("Log to: "))
		sb.WriteString(detailValueStyle.Render("Subtask"))
		sb.WriteString("\n")
	}

	// Time and Status
	sb.WriteString(detailLabelStyle.Render("Time: "))
	sb.WriteString(detailValueStyle.Render(formatDuration(item.TimeSpent)))
	sb.WriteString("\n")

	sb.WriteString(detailLabelStyle.Render("Day: "))
	sb.WriteString(detailValueStyle.Render(day.Day + " " + day.Date.Format("02 Jan")))
	sb.WriteString("\n")

	// Status with appropriate styling
	sb.WriteString(detailLabelStyle.Render("Status: "))
	status := item.Entry.Status
	if status == "" {
		status = "PENDING"
	}
	sb.WriteString(ui.FormatStatus(status))
	sb.WriteString("\n")

	sb.WriteString("\n")
	sb.WriteString(divider(width - 4))
	sb.WriteString("\n")

	// Notes (Comment field)
	sb.WriteString(detailLabelStyle.Render("Notes:"))
	sb.WriteString("\n")

	if item.Entry.Comment != "" {
		notes := wrapText(item.Entry.Comment, width-4)
		sb.WriteString(notesStyle.Render(notes))
	} else {
		sb.WriteString(emptyNotesStyle.Render("(no notes)"))
	}

	// Hints at bottom
	sb.WriteString("\n\n")
	hints := helpKeyStyle.Render("[↑/↓]") + helpDescStyle.Render(" items  ") +
		helpKeyStyle.Render("[←/→]") + helpDescStyle.Render(" days")
	sb.WriteString(hints)

	return detailPanelStyle.
		Width(width).
		Height(contentHeight).
		Render(sb.String())
}

// renderEmptyDayDetail renders detail panel for days with no entries
func (m Model) renderEmptyDayDetail(day *DayGroup, width, height int) string {
	var sb strings.Builder

	sb.WriteString(detailTitleStyle.Render(day.Day))
	sb.WriteString("\n")
	sb.WriteString(detailValueStyle.Render(day.Date.Format("Monday, 02 January 2006")))
	sb.WriteString("\n\n")

	sb.WriteString(emptyDayStyle.Render("No work logged for this day"))
	sb.WriteString("\n\n")

	hints := helpKeyStyle.Render("[←/→]") + helpDescStyle.Render(" navigate to other days")
	sb.WriteString(hints)

	return detailPanelStyle.Width(width).Height(height).Render(sb.String())
}

// renderPresentationContent renders the scrollable content for the viewport
func (m Model) renderPresentationContent() string {
	day := m.getCurrentDay()
	if day == nil {
		return "No data"
	}

	var sb strings.Builder
	contentWidth := m.viewport.Width - 2

	// Day header with total time
	sb.WriteString(dayHeaderStyle.Render(fmt.Sprintf("%s - Total: %s",
		day.Day, formatDuration(day.Total))))
	sb.WriteString("\n\n")

	// Handle empty day
	if len(day.Entries) == 0 {
		sb.WriteString(emptyDayStyle.Render("No entries logged for this day"))
	} else {
		// All items for current day
		for i, item := range day.Entries {
			isSelected := i == m.itemIndex
			sb.WriteString(m.renderPresentationItem(i, item, isSelected, contentWidth))
		}
	}

	return sb.String()
}

// renderPresentationItem renders a single item in presentation format
func (m Model) renderPresentationItem(idx int, item StandupItem, selected bool, width int) string {
	var sb strings.Builder

	// Selection indicator
	sel := " "
	if selected {
		sel = ">"
	}

	// Time display
	timeStr := formatDuration(item.TimeSpent)
	if item.TimeSpent == 0 {
		timeStr = "--"
	}

	// Extract action prefix from first line of notes
	actionPrefix, remainingNotes := extractActionPrefix(item.Entry.Comment)

	// Line 1: Number + Issue key + time (aligned)
	// Format: ">  1  VIS-1234 (2h)"
	sb.WriteString(fmt.Sprintf("%s %2d  %s (%s)\n", sel, idx+1, item.Entry.IssueKey, timeStr))

	// Line 2: Title (indented under issue key)
	title := wrapText(item.Entry.Description, width-6)
	for _, line := range strings.Split(title, "\n") {
		sb.WriteString(fmt.Sprintf("      %s\n", line))
	}

	// Action note (if short first line of comment exists)
	if actionPrefix != "" {
		sb.WriteString(fmt.Sprintf("      → %s\n", actionPrefix))
	}

	// Remaining notes (if any)
	if remainingNotes != "" {
		wrapped := wrapText(remainingNotes, width-8)
		for _, line := range strings.Split(wrapped, "\n") {
			sb.WriteString(fmt.Sprintf("        %s\n", line))
		}
	}

	// Add spacing between items
	sb.WriteString("\n")

	if selected {
		return selectedItemStyle.Render(sb.String())
	}
	return sb.String()
}

// extractActionPrefix extracts the first line of a comment as an action prefix
// if it's short enough (< 60 chars), otherwise returns empty and full comment
func extractActionPrefix(comment string) (prefix, remaining string) {
	if comment == "" {
		return "", ""
	}
	lines := strings.SplitN(comment, "\n", 2)
	firstLine := strings.TrimSpace(lines[0])

	// First line is the action prefix if it's short-ish (< 60 chars)
	if len(firstLine) < 60 {
		if len(lines) > 1 {
			return firstLine, strings.TrimSpace(lines[1])
		}
		return firstLine, ""
	}
	return "", comment
}

// formatDuration formats a duration nicely (e.g., "2h 30m")
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return "0m"
}

// wrapText wraps text to fit within width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result []string
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result = append(result, line)
			continue
		}

		// Simple word wrap
		words := strings.Fields(line)
		currentLine := ""

		for _, word := range words {
			if currentLine == "" {
				currentLine = word
			} else if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				result = append(result, currentLine)
				currentLine = word
			}
		}

		if currentLine != "" {
			result = append(result, currentLine)
		}
	}

	return strings.Join(result, "\n")
}
