package duration

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Duration
		expectError bool
	}{
		// Basic units
		{"30m", "30m", 30 * time.Minute, false},
		{"2h", "2h", 2 * time.Hour, false},
		{"45s", "45s", 45 * time.Second, false},

		// Days (1d = 8h JIRA workday)
		{"1d equals 8h", "1d", 8 * time.Hour, false},
		{"2d equals 16h", "2d", 16 * time.Hour, false},

		// Combined durations
		{"1d4h", "1d4h", 12 * time.Hour, false},
		{"2h30m", "2h30m", 2*time.Hour + 30*time.Minute, false},
		{"1d4h30m", "1d4h30m", 12*time.Hour + 30*time.Minute, false},

		// Case insensitive
		{"uppercase 2H30M", "2H30M", 2*time.Hour + 30*time.Minute, false},
		{"mixed case 1D4h", "1D4h", 12 * time.Hour, false},

		// Whitespace handling
		{"leading spaces", "  2h", 2 * time.Hour, false},
		{"trailing spaces", "2h  ", 2 * time.Hour, false},
		{"both spaces", "  2h  ", 2 * time.Hour, false},

		// Default unit (minutes when no unit specified)
		{"bare number defaults to minutes", "30", 30 * time.Minute, false},

		// Error cases
		{"empty string", "", 0, true},
		{"invalid format", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got %v", tt.input, result)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		// Zero
		{"zero duration", 0, "0m"},

		// Minutes only
		{"30 minutes", 30 * time.Minute, "30m"},
		{"45 minutes", 45 * time.Minute, "45m"},

		// Hours only
		{"2 hours", 2 * time.Hour, "2h"},
		{"7 hours", 7 * time.Hour, "7h"},

		// Hours and minutes
		{"2h30m", 2*time.Hour + 30*time.Minute, "2h30m"},
		{"1h15m", 1*time.Hour + 15*time.Minute, "1h15m"},

		// Day conversion edge cases (8h = 1d)
		{"8h becomes 1d", 8 * time.Hour, "1d"},
		{"16h becomes 2d", 16 * time.Hour, "2d"},
		{"24h becomes 3d", 24 * time.Hour, "3d"},

		// Day + hours
		{"10h becomes 1d2h", 10 * time.Hour, "1d2h"},
		{"12h becomes 1d4h", 12 * time.Hour, "1d4h"},
		{"9h becomes 1d1h", 9 * time.Hour, "1d1h"},

		// Full combinations
		{"1d4h30m", 12*time.Hour + 30*time.Minute, "1d4h30m"},
		{"2d2h15m", 18*time.Hour + 15*time.Minute, "2d2h15m"},

		// Only minutes (less than an hour)
		{"59 minutes", 59 * time.Minute, "59m"},

		// Edge case: exactly 1 day with minutes
		{"1d30m", 8*time.Hour + 30*time.Minute, "1d30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Format(tt.input)
			if result != tt.expected {
				t.Errorf("Format(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToJiraFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		// Zero
		{"zero duration", 0, "0m"},

		// Single units
		{"2 hours", 2 * time.Hour, "2h"},
		{"30 minutes", 30 * time.Minute, "30m"},

		// Combined with space separator (API format)
		{"2h30m with space", 2*time.Hour + 30*time.Minute, "2h 30m"},
		{"8h (no day conversion)", 8 * time.Hour, "8h"},
		{"12h30m", 12*time.Hour + 30*time.Minute, "12h 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJiraFormat(tt.input)
			if result != tt.expected {
				t.Errorf("ToJiraFormat(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseAndFormat_RoundTrip(t *testing.T) {
	// Test that Parse and Format work together correctly
	tests := []struct {
		input    string
		expected string
	}{
		{"30m", "30m"},
		{"2h", "2h"},
		{"2h30m", "2h30m"},
		{"1d", "1d"},       // 8h parses to 8h, formats to 1d
		{"1d4h", "1d4h"},   // 12h parses to 12h, formats to 1d4h
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parsed, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.input, err)
			}
			result := Format(parsed)
			if result != tt.expected {
				t.Errorf("Parse(%q) -> Format = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
