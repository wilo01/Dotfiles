package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
