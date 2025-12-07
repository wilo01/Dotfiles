package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Pattern matches: 1h, 30m, 2h30m, 1d, 1d4h, etc.
	durationPattern = regexp.MustCompile(`(?i)(\d+)(d|h|m|s)?`)
)

// Parse parses a JIRA-style duration string
// Supported formats: 30m, 2h, 1h30m, 1d, 1d4h, 8h30m, etc.
func Parse(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Find all matches
	matches := durationPattern.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	var total time.Duration
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", match[1])
		}

		unit := "m" // default to minutes
		if len(match) > 2 && match[2] != "" {
			unit = match[2]
		}

		switch unit {
		case "d":
			total += time.Duration(value) * 8 * time.Hour // 1 day = 8 hours in JIRA
		case "h":
			total += time.Duration(value) * time.Hour
		case "m":
			total += time.Duration(value) * time.Minute
		case "s":
			total += time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("unknown unit: %s", unit)
		}
	}

	return total, nil
}

// Format formats a duration as a JIRA-style string
func Format(d time.Duration) string {
	if d == 0 {
		return "0m"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	var parts []string

	if hours >= 8 {
		days := hours / 8
		hours = hours % 8
		parts = append(parts, fmt.Sprintf("%dd", days))
	}

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	if len(parts) == 0 {
		return "0m"
	}

	return strings.Join(parts, "")
}

// ToJiraFormat converts a duration to JIRA API format (e.g., "2h 30m")
func ToJiraFormat(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	var parts []string

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	if len(parts) == 0 {
		return "0m"
	}

	return strings.Join(parts, " ")
}
