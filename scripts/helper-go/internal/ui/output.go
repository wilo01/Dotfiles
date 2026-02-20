package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dariuszw/hlp/internal/worklog"
	"github.com/dariuszw/hlp/pkg/duration"
	"golang.org/x/term"
)

// GetTerminalWidth returns the current terminal width, or 80 as fallback
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < 40 {
		return 80
	}
	return width
}

// Error formats an error message
func Error(msg string) string {
	icon := ErrorText.Render("✗")
	return fmt.Sprintf("%s %s", icon, ErrorText.Render(msg))
}

// Warning formats a warning message
func Warning(msg string) string {
	icon := WarningText.Render("⚠")
	return fmt.Sprintf("%s %s", icon, WarningText.Render(msg))
}

// SuccessMsg formats a success message
func SuccessMsg(msg string) string {
	icon := Success.Render("✓")
	return fmt.Sprintf("%s %s", icon, Success.Render(msg))
}

// Info formats an info message
func Info(msg string) string {
	icon := Primary.Render("ℹ")
	return fmt.Sprintf("%s %s", icon, msg)
}

// KeyValue formats a key-value pair
func KeyValue(key, value string) string {
	return fmt.Sprintf("%s %s", Label.Render(key+":"), Value.Render(value))
}

// KeyValuePadded formats a key-value pair with padding
func KeyValuePadded(key, value string, width int) string {
	padding := width - len(key)
	if padding < 1 {
		padding = 1
	}
	return fmt.Sprintf("%s%s%s", Label.Render(key+":"), strings.Repeat(" ", padding), Value.Render(value))
}

// Header formats a section header
func Header(text string) string {
	line := strings.Repeat("─", len(text)+4)
	return fmt.Sprintf("%s\n%s\n%s",
		Muted.Render(line),
		Title.Render("  "+text+"  "),
		Muted.Render(line))
}

// Divider returns a horizontal divider
func Divider(width int) string {
	return Muted.Render(strings.Repeat("─", width))
}

// Bullet formats a bulleted item
func Bullet(text string) string {
	return fmt.Sprintf("  %s %s", Muted.Render("•"), text)
}

// NumberedItem formats a numbered list item
func NumberedItem(num int, text string) string {
	return fmt.Sprintf("  %s %s", Muted.Render(fmt.Sprintf("%d.", num)), text)
}

// CodeBlock formats text as a code block
func CodeBlock(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, Code.Render(line))
	}
	return strings.Join(result, "\n")
}

// Table formats data as a simple table
func Table(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	for i, h := range headers {
		sb.WriteString(headerStyle.Render(padRight(h, widths[i])))
		if i < len(headers)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// Separator
	for i, w := range widths {
		sb.WriteString(Muted.Render(strings.Repeat("─", w)))
		if i < len(widths)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// Rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				sb.WriteString(padRight(cell, widths[i]))
				if i < len(row)-1 {
					sb.WriteString("  ")
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// ProgressBar creates a simple text-based progress bar
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %d/%d", Primary.Render(bar), current, total)
}

// ConfirmAction prompts user for Y/N confirmation
// Returns true if user confirms, false otherwise
func ConfirmAction(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [y/N]: ", message)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y" || input == "yes"
}

// ConfirmProtectedProfile shows a warning and prompts for confirmation
func ConfirmProtectedProfile(profileName, baseURL string, entryCount int) bool {
	fmt.Println()
	fmt.Println(WarningText.Render("WARNING: Protected Profile"))
	fmt.Println(Divider(50))
	fmt.Printf("  Profile:  %s\n", Primary.Render(profileName))
	fmt.Printf("  URL:      %s\n", Muted.Render(baseURL))
	fmt.Printf("  Entries:  %s\n", Success.Render(fmt.Sprintf("%d", entryCount)))
	fmt.Println(Divider(50))
	fmt.Println()

	return ConfirmAction(WarningText.Render("Proceed with batch operation?"))
}

// DailyBreakdown formats the daily time summary with warnings
func DailyBreakdown(summaries []worklog.DailySummary, expectedStr string) string {
	if len(summaries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(Header("Daily Time Summary"))
	sb.WriteString("\n\n")

	for _, day := range summaries {
		dateStr := day.Date.Format("Mon 02 Jan")
		loggedStr := duration.Format(day.TotalLogged)

		var statusIcon, statusText string

		switch day.Status {
		case worklog.DayStatusOver:
			statusIcon = WarningText.Render("^")
			diffStr := "+" + duration.Format(day.Difference)
			statusText = WarningText.Render(diffStr)
		case worklog.DayStatusUnder:
			statusIcon = ErrorText.Render("v")
			diffStr := "-" + duration.Format(-day.Difference)
			statusText = ErrorText.Render(diffStr)
		default:
			statusIcon = Success.Render("=")
			statusText = Success.Render("OK")
		}

		sb.WriteString(fmt.Sprintf("  %s  %s  %s / %s  %s\n",
			Muted.Render(dateStr),
			statusIcon,
			loggedStr,
			Muted.Render(expectedStr),
			statusText))
	}

	return sb.String()
}

// DailyWarningsSummary returns a summary line for daily totals
func DailyWarningsSummary(result *worklog.AnalysisResult) string {
	if !result.HasWarnings {
		return ""
	}

	overCount, underCount, _ := result.CountByStatus()

	var parts []string
	if overCount > 0 {
		parts = append(parts, WarningText.Render(fmt.Sprintf("%d day(s) over", overCount)))
	}
	if underCount > 0 {
		parts = append(parts, ErrorText.Render(fmt.Sprintf("%d day(s) under", underCount)))
	}

	return Warning(strings.Join(parts, ", "))
}

// FormatDailyWarning formats a single-day warning message
func FormatDailyWarning(date time.Time, totalLogged, expected time.Duration) string {
	diff := totalLogged - expected
	dateStr := date.Format("Mon 02 Jan")
	expectedStr := duration.Format(expected)

	if diff > 0 {
		return Warning(fmt.Sprintf(
			"Daily total for %s: %s (expected %s, over by %s)",
			dateStr,
			duration.Format(totalLogged),
			expectedStr,
			duration.Format(diff)))
	} else if diff < 0 {
		return Warning(fmt.Sprintf(
			"Daily total for %s: %s (expected %s, under by %s)",
			dateStr,
			duration.Format(totalLogged),
			expectedStr,
			duration.Format(-diff)))
	}

	return SuccessMsg(fmt.Sprintf(
		"Daily total for %s: %s (exactly %s)",
		dateStr,
		duration.Format(totalLogged),
		expectedStr))
}

// FormatScheduleCountdown formats the schedule wait countdown line
func FormatScheduleCountdown(slotName string, targetTime time.Time, remaining time.Duration) string {
	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60
	seconds := int(remaining.Seconds()) % 60

	var timeStr string
	if hours > 0 {
		timeStr = fmt.Sprintf("%dh%02dm%02ds", hours, minutes, seconds)
	} else if minutes > 0 {
		timeStr = fmt.Sprintf("%dm%02ds", minutes, seconds)
	} else {
		timeStr = fmt.Sprintf("%ds", seconds)
	}

	return fmt.Sprintf("\r  %s Next run: %s (%s) in %s (Ctrl+C to cancel)",
		SpinnerFrames[0],
		Primary.Render(slotName),
		Muted.Render(targetTime.Format("15:04")),
		Success.Render(timeStr))
}

// FormatScheduleHeader formats the schedule mode header
func FormatScheduleHeader(slotName, targetTime string, loop bool) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(Header("Scheduled Batch Mode"))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("  Slot:    %s\n", Primary.Render(slotName)))
	sb.WriteString(fmt.Sprintf("  Target:  %s\n", Primary.Render(targetTime)))
	if loop {
		sb.WriteString(fmt.Sprintf("  Mode:    %s\n", Muted.Render("loop (will continue after run)")))
	} else {
		sb.WriteString(fmt.Sprintf("  Mode:    %s\n", Muted.Render("single run")))
	}
	sb.WriteString("\n")
	return sb.String()
}

// ClearLine clears the current line in the terminal
func ClearLine() {
	width := GetTerminalWidth()
	fmt.Print("\r" + strings.Repeat(" ", width) + "\r")
}
