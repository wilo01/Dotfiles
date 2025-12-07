package batch

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/jira"
)

// Entry represents a single worklog entry from CSV
type Entry struct {
	IssueKey  string
	TimeSpent string
	Date      string // DD/MM/YYYY or DD.MM.YYYY
	Comment   string
	Status    string // empty = pending, UPDATED, DONE
	RowNumber int
}

// IsPending returns true if entry needs to be posted
func (e Entry) IsPending() bool {
	return e.Status == ""
}

// IsDone returns true if entry is completed
func (e Entry) IsDone() bool {
	return strings.ToUpper(e.Status) == "DONE"
}

// NeedsProcessing returns true if entry needs processing
func (e Entry) NeedsProcessing() bool {
	return !e.IsDone()
}

// Result represents the result of a single worklog submission
type Result struct {
	Entry        Entry
	Success      bool
	ErrorMessage string
	WorklogID    string
	NewStatus    string
}

// Processor handles batch worklog processing
type Processor struct {
	client      *jira.Client
	defaultTime string
}

// NewProcessor creates a new batch processor
func NewProcessor(client *jira.Client, defaultTime string) *Processor {
	if defaultTime == "" {
		defaultTime = "09:00"
	}
	return &Processor{
		client:      client,
		defaultTime: defaultTime,
	}
}

// DefaultCSVPath returns the default path for the worklogs CSV
func DefaultCSVPath() string {
	return config.GetConfigPath("worklogs.csv")
}

// ParseCSV parses a CSV file into entries
func ParseCSV(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1 // Allow variable fields

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	var entries []Entry
	for i, record := range records {
		rowNum := i + 1

		if len(record) < 3 {
			continue
		}

		// Skip header row
		if rowNum == 1 {
			first := strings.ToLower(record[0])
			if first == "issue_key" || first == "issue" || first == "ticket" {
				continue
			}
		}

		entry := Entry{
			IssueKey:  strings.ToUpper(strings.TrimSpace(record[0])),
			TimeSpent: strings.TrimSpace(record[1]),
			Date:      strings.TrimSpace(record[2]),
			RowNumber: rowNum,
		}

		if len(record) > 3 {
			entry.Comment = strings.TrimSpace(record[3])
		}

		// Check for status in the 5th column or parse from comment
		if len(record) > 4 {
			entry.Status = strings.TrimSpace(record[4])
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// ParsePendingCSV parses only pending entries
func ParsePendingCSV(path string) ([]Entry, error) {
	entries, err := ParseCSV(path)
	if err != nil {
		return nil, err
	}

	var pending []Entry
	for _, e := range entries {
		if e.NeedsProcessing() {
			pending = append(pending, e)
		}
	}
	return pending, nil
}

// ParseDate parses DD/MM/YYYY or DD.MM.YYYY to time.Time
func ParseDate(dateStr, timeStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(timeStr)

	// Try different date formats
	var day, month, year int
	var err error

	if strings.Contains(dateStr, "/") {
		_, err = fmt.Sscanf(dateStr, "%d/%d/%d", &day, &month, &year)
	} else if strings.Contains(dateStr, ".") {
		_, err = fmt.Sscanf(dateStr, "%d.%d.%d", &day, &month, &year)
	} else if strings.Contains(dateStr, "-") {
		// Try YYYY-MM-DD
		_, err = fmt.Sscanf(dateStr, "%d-%d-%d", &year, &month, &day)
	} else {
		return time.Time{}, fmt.Errorf("unknown date format: %s", dateStr)
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", dateStr)
	}

	// Parse time
	hour, minute := 9, 0
	if timeStr != "" {
		fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	}

	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local), nil
}

// Process processes a single entry
func (p *Processor) Process(entry Entry) Result {
	result := Result{Entry: entry}

	// Parse date
	started, err := ParseDate(entry.Date, p.defaultTime)
	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result
	}

	// Create worklog entry
	worklogEntry := jira.WorklogEntry{
		TimeSpent: entry.TimeSpent,
		Started:   started,
		Comment:   entry.Comment,
	}

	// Post to JIRA
	if err := p.client.LogWork(entry.IssueKey, worklogEntry); err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result
	}

	result.Success = true
	result.NewStatus = "DONE"
	return result
}

// ProcessBatch processes all entries
func (p *Processor) ProcessBatch(entries []Entry, dryRun bool, progressFn func(current, total int, result Result)) []Result {
	var results []Result
	total := len(entries)

	for i, entry := range entries {
		var result Result

		if dryRun {
			// Validate only
			_, err := ParseDate(entry.Date, p.defaultTime)
			if err != nil {
				result = Result{
					Entry:        entry,
					Success:      false,
					ErrorMessage: err.Error(),
				}
			} else {
				result = Result{
					Entry:     entry,
					Success:   true,
					NewStatus: "DONE",
				}
			}
		} else {
			result = p.Process(entry)
		}

		results = append(results, result)

		if progressFn != nil {
			progressFn(i+1, total, result)
		}

		// Stop on auth errors
		if !result.Success && strings.Contains(result.ErrorMessage, "401") {
			break
		}
	}

	return results
}

// UpdateCSVStatus updates the CSV file with new statuses
func UpdateCSVStatus(path string, results []Result) error {
	// Build status updates map
	updates := make(map[int]string)
	for _, r := range results {
		if r.Success && r.NewStatus != "" {
			updates[r.Entry.RowNumber] = r.NewStatus
		}
	}

	if len(updates) == 0 {
		return nil
	}

	// Read all lines
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	f.Close()
	if err != nil {
		return err
	}

	// Update statuses
	for i, record := range records {
		rowNum := i + 1
		if newStatus, ok := updates[rowNum]; ok {
			// Ensure we have at least 5 columns
			for len(record) < 5 {
				record = append(record, "")
			}
			record[4] = newStatus
			records[i] = record
		}
	}

	// Write back
	f, err = os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	return writer.WriteAll(records)
}

// AppendEntry appends a new entry to the CSV file
func AppendEntry(path string, issueKey, timeSpent, date, comment string) error {
	if date == "" {
		date = time.Now().Format("02/01/2006")
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open file in append mode
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	return writer.Write([]string{
		strings.ToUpper(issueKey),
		timeSpent,
		date,
		comment,
		"", // Empty status
	})
}
