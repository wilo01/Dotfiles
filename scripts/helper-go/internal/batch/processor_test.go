package batch

import (
	"os"
	"strings"
	"testing"
)

func TestEntryExistsForTicket(t *testing.T) {
	entries := []Entry{
		{IssueKey: "VIS-100", Date: "15.12.2025", Status: StatusDraft, Description: "Test ticket"},
		{IssueKey: "VIS-200", Date: "15.12.2025", Status: StatusDone, TimeSpent: "2h", Description: "Done ticket"},
		{IssueKey: "VIS-300", Date: "14.12.2025", Status: "", TimeSpent: "", Description: "Old pending"},
		{IssueKey: "VIS-400", Date: "15.12.2025", Status: StatusSync, Description: "Synced ticket"},
		{IssueKey: "VIS-500", Date: "15.12.2025", Status: StatusDraft, Description: "Parent > Subtask A"},
	}

	tests := []struct {
		name           string
		issueKey       string
		date           string
		isSubtask      bool
		subtaskSummary string
		expected       bool
	}{
		{
			name:     "DRAFT entry exists for date",
			issueKey: "VIS-100",
			date:     "15.12.2025",
			expected: true,
		},
		{
			name:     "DRAFT entry exists regardless of date (prevents duplicate drafts)",
			issueKey: "VIS-100",
			date:     "14.12.2025",
			expected: true,
		},
		{
			name:     "DONE entry exists regardless of date",
			issueKey: "VIS-200",
			date:     "01.01.2024",
			expected: true,
		},
		{
			name:     "Entry with TimeSpent exists regardless of date",
			issueKey: "VIS-200",
			date:     "01.01.2024",
			expected: true,
		},
		{
			name:     "SYNC entry exists regardless of date",
			issueKey: "VIS-400",
			date:     "01.01.2024",
			expected: true,
		},
		{
			name:     "Pending entry (no status) with different date - not exists",
			issueKey: "VIS-300",
			date:     "15.12.2025",
			expected: false,
		},
		{
			name:     "Pending entry (no status) same date - exists",
			issueKey: "VIS-300",
			date:     "14.12.2025",
			expected: true,
		},
		{
			name:     "Non-existent ticket",
			issueKey: "VIS-999",
			date:     "15.12.2025",
			expected: false,
		},
		{
			name:     "Case insensitive key match",
			issueKey: "vis-100",
			date:     "15.12.2025",
			expected: true,
		},
		{
			name:           "Subtask matches with summary in description",
			issueKey:       "VIS-500",
			date:           "15.12.2025",
			isSubtask:      true,
			subtaskSummary: "Subtask A",
			expected:       true,
		},
		{
			name:           "Subtask does not match with different summary",
			issueKey:       "VIS-500",
			date:           "15.12.2025",
			isSubtask:      true,
			subtaskSummary: "Subtask B",
			expected:       false,
		},
		{
			name:           "Non-subtask ignores summary matching",
			issueKey:       "VIS-100",
			date:           "15.12.2025",
			isSubtask:      false,
			subtaskSummary: "Whatever",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EntryExistsForTicket(entries, tt.issueKey, tt.date, tt.isSubtask, tt.subtaskSummary)
			if result != tt.expected {
				t.Errorf("EntryExistsForTicket(%q, %q, %v, %q) = %v, want %v",
					tt.issueKey, tt.date, tt.isSubtask, tt.subtaskSummary, result, tt.expected)
			}
		})
	}
}

func TestCSVQuotingWithCommas(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantQuoted  bool
	}{
		{
			name:        "Description with commas in brackets",
			description: "DSS - Confirmation on auto access group assignment > [Code Review]-[13.1AV,13.0AV,12.1AV]",
			wantQuoted:  true,
		},
		{
			name:        "Simple description without commas",
			description: "Simple task description",
			wantQuoted:  false,
		},
		{
			name:        "Description with embedded quotes",
			description: `Say "hello" to the world`,
			wantQuoted:  true,
		},
		{
			name:        "Parent child combined description with comma",
			description: "Parent Task > Child, with comma in name",
			wantQuoted:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile, err := os.CreateTemp("", "csv_quote_test_*.csv")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			tmpPath := tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpPath)

			// Write entry using AppendEntryWithStatus
			err = AppendEntryWithStatus(tmpPath, "TEST-1", "", "Bug", tt.description, "1h", "23.12.2025 09:00", "", "", StatusDraft)
			if err != nil {
				t.Fatalf("AppendEntryWithStatus failed: %v", err)
			}

			// Read raw file content
			rawContent, err := os.ReadFile(tmpPath)
			if err != nil {
				t.Fatalf("Failed to read temp file: %v", err)
			}
			rawStr := string(rawContent)

			// Check if description is quoted in raw CSV
			if tt.wantQuoted {
				// For fields with commas or quotes, csv.Writer wraps in double quotes
				// Embedded quotes become ""
				if !strings.Contains(rawStr, `"`) {
					t.Errorf("Expected quoted field in raw CSV for description with special chars, got: %s", rawStr)
				}
			}

			// Round-trip: parse back and verify data preserved exactly
			entries, err := ParseCSV(tmpPath)
			if err != nil {
				t.Fatalf("ParseCSV failed: %v", err)
			}

			if len(entries) != 1 {
				t.Fatalf("Expected 1 entry, got %d", len(entries))
			}

			if entries[0].Description != tt.description {
				t.Errorf("Round-trip failed: got %q, want %q", entries[0].Description, tt.description)
			}
		})
	}
}
