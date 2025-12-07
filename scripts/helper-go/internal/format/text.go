package format

import (
	"regexp"
	"strings"
)

var (
	// Pattern for JIRA ticket at start: VIS-1234, TDT-123, etc.
	// Note: Go regexp doesn't support lookahead, so we handle this in code
	ticketPatternStart = regexp.MustCompile(`(?i)^([a-zA-Z]{2,6})-(\d+)`)
	// Pattern for JIRA ticket anywhere in text
	ticketPatternAny = regexp.MustCompile(`(?i)\b([a-zA-Z]+)-(\d+)\b`)
	// Pattern for special characters to replace with dashes
	specialChars = regexp.MustCompile(`[ !+@#$%^&*(),_.'/:;>\[\]\\-]+`)
	// Pattern for multiple consecutive dashes
	multipleDashes = regexp.MustCompile(`-+`)
)

// Options controls formatting behavior
type Options struct {
	UppercaseTags bool // Uppercase TAG-NUMBER patterns (default: true)
	CleanDashes   bool // Clean up multiple consecutive dashes (default: true)
	StripEdges    bool // Strip leading/trailing dashes (default: true)
}

// DefaultOptions returns the default formatting options
func DefaultOptions() Options {
	return Options{
		UppercaseTags: true,
		CleanDashes:   true,
		StripEdges:    true,
	}
}

// FormatString is the unified string formatting function
func FormatString(text string, opts Options) string {
	text = strings.TrimSpace(text)

	// Convert to lowercase and replace special chars with dashes
	formatted := strings.ToLower(text)
	formatted = strings.ReplaceAll(formatted, "\n", "-")
	formatted = specialChars.ReplaceAllString(formatted, "-")

	if opts.UppercaseTags {
		// Only uppercase JIRA ticket at the START of the string
		// This avoids uppercasing words like "allowed-13", "thing-12", "version-2"
		formatted = ticketPatternStart.ReplaceAllStringFunc(formatted, func(match string) string {
			parts := ticketPatternStart.FindStringSubmatch(match)
			if len(parts) >= 3 {
				return strings.ToUpper(parts[1]) + "-" + parts[2]
			}
			return match
		})
	}

	if opts.CleanDashes {
		formatted = multipleDashes.ReplaceAllString(formatted, "-")
	}

	if opts.StripEdges {
		formatted = strings.Trim(formatted, "-")
	}

	return formatted
}

// JiraBranch generates a git checkout command for a JIRA branch
func JiraBranch(text string) string {
	formatted := FormatString(text, DefaultOptions())
	return "git checkout -b " + formatted
}

// Stash generates a git stash command with formatted message
func Stash(text string) string {
	formatted := FormatString(text, DefaultOptions())
	return "git stash push -u -m " + formatted
}

// Dash converts text to dash-separated format (no tag uppercasing)
func Dash(text string) string {
	opts := DefaultOptions()
	opts.UppercaseTags = false
	return FormatString(text, opts)
}

// PRTitle generates a PR title with TAG-NUMBER uppercased but preserving spaces
// Example: "vis-1234 add new feature" -> "VIS-1234 add new feature"
func PRTitle(text string) string {
	text = strings.TrimSpace(text)

	// Uppercase all JIRA ticket patterns in the text
	formatted := ticketPatternAny.ReplaceAllStringFunc(text, func(match string) string {
		parts := ticketPatternAny.FindStringSubmatch(match)
		if len(parts) >= 3 {
			return strings.ToUpper(parts[1]) + "-" + parts[2]
		}
		return match
	})

	return formatted
}

// Filename generates a clean markdown filename
// Example: "VIS-1234 My Report" -> "vis-1234-my-report.md"
func Filename(text string) string {
	formatted := Dash(text)
	return formatted + ".md"
}

// ExtractTicket extracts a JIRA ticket number from text
// Returns the ticket (e.g., "VIS-1234") and any description after it
func ExtractTicket(text string) (ticket string, description string) {
	text = strings.TrimSpace(text)

	match := ticketPatternStart.FindStringSubmatch(strings.ToLower(text))
	if len(match) >= 3 {
		ticket = strings.ToUpper(match[1]) + "-" + match[2]

		// Extract description (everything after the ticket)
		remaining := text[len(match[0]):]
		remaining = strings.TrimPrefix(remaining, "-")
		remaining = strings.TrimPrefix(remaining, " ")
		description = strings.TrimSpace(remaining)

		return ticket, description
	}

	return "", text
}

// NormalizeTicket ensures ticket is uppercase
func NormalizeTicket(ticket string) string {
	return strings.ToUpper(strings.TrimSpace(ticket))
}
