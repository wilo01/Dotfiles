package batch

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEntryExistsForTicket(t *testing.T) {
	entries := []Entry{
		{IssueKey: "VIS-100", Date: "15.12.2025", Status: StatusDraft, Description: "Test ticket"},
		{IssueKey: "VIS-200", Date: "15.12.2025", Status: StatusDone, TimeSpent: "2h", Description: "Done ticket"},
		{IssueKey: "VIS-300", Date: "14.12.2025", Status: "", TimeSpent: "", Description: "Old pending"},
		{IssueKey: "VIS-400", Date: "15.12.2025", Status: StatusSync, Description: "Synced ticket"},
		{IssueKey: "VIS-500", Date: "15.12.2025", Status: StatusDraft, Description: "Parent > Subtask A"},
		{IssueKey: "VIS-600", SubtaskKey: "VIS-601", Date: "15.12.2025", Status: StatusDraft, Description: "Parent task > Child task"},
		{IssueKey: "VIS-700", SubtaskKey: "VIS-701", Date: "15.12.2025", Status: StatusDone, TimeSpent: "1h", Description: "Logged parent > Logged child"},
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
		{
			name:     "Match via SubtaskKey field",
			issueKey: "VIS-601",
			date:     "15.12.2025",
			expected: true,
		},
		{
			name:     "IssueKey still matches when SubtaskKey is present",
			issueKey: "VIS-600",
			date:     "15.12.2025",
			expected: true,
		},
		{
			name:           "Subtask summary check via SubtaskKey match",
			issueKey:       "VIS-601",
			date:           "15.12.2025",
			isSubtask:      true,
			subtaskSummary: "Child task",
			expected:       true,
		},
		{
			name:           "Subtask summary mismatch via SubtaskKey",
			issueKey:       "VIS-601",
			date:           "15.12.2025",
			isSubtask:      true,
			subtaskSummary: "Wrong child",
			expected:       false,
		},
		{
			name:     "Match DONE entry via SubtaskKey",
			issueKey: "VIS-701",
			date:     "01.01.2024",
			expected: true,
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

func TestRemoveEntryByRow(t *testing.T) {
	t.Run("removes middle row", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "remove_row_test_*.csv")
		if err != nil {
			t.Fatal(err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		// Write 3 entries
		_ = AppendEntryWithStatus(tmpPath, "VIS-100", "", "Story", "First", "1h", "06.03.2026 09:00", "", "", StatusDraft)
		_ = AppendEntryWithStatus(tmpPath, "VIS-200", "", "Bug", "Second", "2h", "06.03.2026 10:00", "", "", StatusDraft)
		_ = AppendEntryWithStatus(tmpPath, "VIS-300", "VIS-301", "Task", "Third", "3h", "06.03.2026 11:00", "", "", StatusDraft)

		// Remove row 2 (the VIS-200 entry)
		err = RemoveEntryByRow(tmpPath, 2)
		if err != nil {
			t.Fatalf("RemoveEntryByRow failed: %v", err)
		}

		entries, err := ParseCSV(tmpPath)
		if err != nil {
			t.Fatal(err)
		}

		if len(entries) != 2 {
			t.Fatalf("Expected 2 entries after removal, got %d", len(entries))
		}
		if entries[0].IssueKey != "VIS-100" {
			t.Errorf("Expected first entry VIS-100, got %s", entries[0].IssueKey)
		}
		if entries[1].IssueKey != "VIS-300" {
			t.Errorf("Expected second entry VIS-300, got %s", entries[1].IssueKey)
		}
	})

	t.Run("removes first row", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "remove_row_test_*.csv")
		if err != nil {
			t.Fatal(err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		_ = AppendEntryWithStatus(tmpPath, "VIS-100", "", "Story", "First", "", "06.03.2026 09:00", "", "", StatusDraft)
		_ = AppendEntryWithStatus(tmpPath, "VIS-200", "", "Bug", "Second", "", "06.03.2026 10:00", "", "", StatusDraft)

		err = RemoveEntryByRow(tmpPath, 1)
		if err != nil {
			t.Fatalf("RemoveEntryByRow failed: %v", err)
		}

		entries, err := ParseCSV(tmpPath)
		if err != nil {
			t.Fatal(err)
		}

		if len(entries) != 1 {
			t.Fatalf("Expected 1 entry, got %d", len(entries))
		}
		if entries[0].IssueKey != "VIS-200" {
			t.Errorf("Expected VIS-200, got %s", entries[0].IssueKey)
		}
	})

	t.Run("removes last row", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "remove_row_test_*.csv")
		if err != nil {
			t.Fatal(err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		_ = AppendEntryWithStatus(tmpPath, "VIS-100", "", "Story", "First", "", "06.03.2026 09:00", "", "", StatusDraft)
		_ = AppendEntryWithStatus(tmpPath, "VIS-200", "", "Bug", "Second", "", "06.03.2026 10:00", "", "", StatusDraft)
		_ = AppendEntryWithStatus(tmpPath, "VIS-300", "", "Task", "Third", "", "06.03.2026 11:00", "", "", StatusDraft)

		err = RemoveEntryByRow(tmpPath, 3)
		if err != nil {
			t.Fatalf("RemoveEntryByRow failed: %v", err)
		}

		entries, err := ParseCSV(tmpPath)
		if err != nil {
			t.Fatal(err)
		}

		if len(entries) != 2 {
			t.Fatalf("Expected 2 entries after removing last row, got %d", len(entries))
		}
		if entries[0].IssueKey != "VIS-100" {
			t.Errorf("Expected first entry VIS-100, got %s", entries[0].IssueKey)
		}
		if entries[1].IssueKey != "VIS-200" {
			t.Errorf("Expected second entry VIS-200, got %s", entries[1].IssueKey)
		}
	})

	t.Run("out of range returns error", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "remove_row_test_*.csv")
		if err != nil {
			t.Fatal(err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		_ = AppendEntryWithStatus(tmpPath, "VIS-100", "", "Story", "Only", "", "06.03.2026 09:00", "", "", StatusDraft)

		err = RemoveEntryByRow(tmpPath, 5)
		if err == nil {
			t.Error("Expected error for out-of-range row, got nil")
		}
	})
}

func TestWriteWorklogRecords(t *testing.T) {
	tests := []struct {
		name   string
		record []string
		want   string
	}{
		{
			name:   "header row is not force-quoted",
			record: []string{"issue_key", "subtask_key", "issue_type", "description", "comment", "date", "time_spent", "subtask_log_ind", "status"},
			want:   "issue_key,subtask_key,issue_type,description,comment,date,time_spent,subtask_log_ind,status",
		},
		{
			name:   "empty comment is quoted",
			record: []string{"TDT-26", "", "Story", "Dariusz - adm work", "", "10.06.2026 08:56", "", "N", "DRAFT"},
			want:   `TDT-26,,Story,Dariusz - adm work,"",10.06.2026 08:56,,N,DRAFT`,
		},
		{
			name:   "comment with commas is quoted",
			record: []string{"SUITE-1", "", "Bug", "Some bug", "fixed a, b, and c", "10.06.2026", "1h", "N", "DONE"},
			want:   `SUITE-1,,Bug,Some bug,"fixed a, b, and c",10.06.2026,1h,N,DONE`,
		},
		{
			name:   "embedded quotes in comment are doubled",
			record: []string{"SUITE-2", "", "Bug", "Other bug", `triggers a "go back" action`, "10.06.2026", "1h", "N", "DONE"},
			want:   `SUITE-2,,Bug,Other bug,"triggers a ""go back"" action",10.06.2026,1h,N,DONE`,
		},
		{
			name:   "description with comma still minimally quoted",
			record: []string{"SUITE-3", "", "Story", "Parent > Child, with comma", "note", "10.06.2026", "1h", "N", "DONE"},
			want:   `SUITE-3,,Story,"Parent > Child, with comma","note",10.06.2026,1h,N,DONE`,
		},
		{
			name:   "legacy 7-column row uses minimal quoting",
			record: []string{"OLD-1", "Bug", "Old format", "note", "01.01.2025", "1h", "DONE"},
			want:   "OLD-1,Bug,Old format,note,01.01.2025,1h,DONE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			if err := writeWorklogRecords(&buf, [][]string{tt.record}); err != nil {
				t.Fatalf("writeWorklogRecords failed: %v", err)
			}
			got := strings.TrimRight(buf.String(), "\n")
			if got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestWriteWorklogRecordsRoundTrip(t *testing.T) {
	comment := `1. SUITE-8737 — swipe triggers a "go back" action, includes video, and a comma`

	tmpFile, err := os.CreateTemp("", "csv_roundtrip_test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := AppendEntryWithStatus(tmpPath, "SUITE-8567", "SUITE-8743", "Story", "IWA End to end QA > QA", "1h", "10.06.2026 08:56", comment, "Y", StatusDone); err != nil {
		t.Fatalf("AppendEntryWithStatus failed: %v", err)
	}

	entries, err := ParseCSV(tmpPath)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Comment != comment {
		t.Errorf("Comment round-trip failed:\ngot  %q\nwant %q", entries[0].Comment, comment)
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

func TestWaitWithCountdown_ContextCancellation(t *testing.T) {
	// Test that waitWithCountdown returns false when context is cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	start := time.Now()
	result := waitWithCountdown(ctx, 60) // Would wait 60s if not cancelled
	elapsed := time.Since(start)

	if result != false {
		t.Error("Expected waitWithCountdown to return false when context is cancelled")
	}

	if elapsed > 2*time.Second {
		t.Errorf("Expected immediate return on cancelled context, but took %v", elapsed)
	}
}

func TestWaitWithCountdown_ShortDuration(t *testing.T) {
	// Test that waitWithCountdown completes normally for short duration
	ctx := context.Background()

	start := time.Now()
	result := waitWithCountdown(ctx, 2) // Wait 2 seconds
	elapsed := time.Since(start)

	if result != true {
		t.Error("Expected waitWithCountdown to return true on normal completion")
	}

	// Should take approximately 2 seconds (with some tolerance)
	if elapsed < 1*time.Second || elapsed > 4*time.Second {
		t.Errorf("Expected ~2s duration, got %v", elapsed)
	}
}

func TestWaitWithCountdown_CancelDuringWait(t *testing.T) {
	// Test that waitWithCountdown can be interrupted mid-wait
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 500ms
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result := waitWithCountdown(ctx, 60) // Would wait 60s if not cancelled
	elapsed := time.Since(start)

	if result != false {
		t.Error("Expected waitWithCountdown to return false when cancelled mid-wait")
	}

	// Should return within ~1 second (500ms cancel + up to 1s sleep cycle)
	if elapsed > 2*time.Second {
		t.Errorf("Expected return within ~1s after cancel, but took %v", elapsed)
	}
}

func TestProcessBatch_SlowModeWithMock(t *testing.T) {
	// Test ProcessBatch with slowMode=true in mock mode
	// Mock mode skips actual JIRA calls, so we can test the flow
	processor := &Processor{
		defaultTime: "09:00",
		mockMode:    true,
	}

	entries := []Entry{
		{IssueKey: "TEST-1", Date: "01.01.2025", TimeSpent: "1h", RowNumber: 1},
		{IssueKey: "TEST-2", Date: "01.01.2025", TimeSpent: "2h", RowNumber: 2},
	}

	var progressCalls int
	progressFn := func(current, total int, result Result) {
		progressCalls++
	}

	// Note: This test will take ~40-240 seconds with real delays
	// For unit testing, we test with slowMode=false to verify the parameter is accepted
	results := processor.ProcessBatch(entries, false, false, progressFn)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if progressCalls != 2 {
		t.Errorf("Expected 2 progress calls, got %d", progressCalls)
	}

	for i, r := range results {
		if !r.Success {
			t.Errorf("Entry %d: expected success, got error: %s", i, r.ErrorMessage)
		}
		if r.NewStatus != StatusDone {
			t.Errorf("Entry %d: expected status %s, got %s", i, StatusDone, r.NewStatus)
		}
	}
}

func TestProcessBatch_SlowModeSignature(t *testing.T) {
	// Test that ProcessBatch accepts slowMode parameter
	// This is a compile-time check essentially, but verifies the API
	processor := &Processor{
		defaultTime: "09:00",
		mockMode:    true,
	}

	entries := []Entry{
		{IssueKey: "TEST-1", Date: "01.01.2025", TimeSpent: "1h", RowNumber: 1},
	}

	// Test with slowMode=true (short test with single entry, no delay after last)
	results := processor.ProcessBatch(entries, false, true, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if !results[0].Success {
		t.Errorf("Expected success, got error: %s", results[0].ErrorMessage)
	}
}

func TestProcessBatch_DryRunIgnoresSlowMode(t *testing.T) {
	// Test that dry-run mode works regardless of slowMode setting
	processor := &Processor{
		defaultTime: "09:00",
	}

	entries := []Entry{
		{IssueKey: "TEST-1", Date: "01.01.2025", TimeSpent: "1h", RowNumber: 1},
		{IssueKey: "TEST-2", Date: "01.01.2025", TimeSpent: "2h", RowNumber: 2},
	}

	start := time.Now()
	results := processor.ProcessBatch(entries, true, true, nil) // dryRun=true, slowMode=true
	elapsed := time.Since(start)

	// Dry-run should be fast even with slowMode=true
	if elapsed > 2*time.Second {
		t.Errorf("Dry-run with slowMode should be fast, but took %v", elapsed)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for i, r := range results {
		if !r.Success {
			t.Errorf("Entry %d: expected success in dry-run, got error: %s", i, r.ErrorMessage)
		}
	}
}
