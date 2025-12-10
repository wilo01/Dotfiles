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
	"github.com/dariuszw/hlp/pkg/duration"
)

// Status constants for worklog entries
const (
	StatusPending = ""
	StatusSync    = "SYNC"
	StatusUpdated = "UPDATED"
	StatusDone    = "DONE"
)

// Entry represents a single worklog entry from CSV
type Entry struct {
	IssueKey    string
	TimeSpent   string
	Date        string // DD/MM/YYYY or DD.MM.YYYY
	Comment     string
	Status      string // empty = pending, SYNC, UPDATED, DONE
	RowNumber   int
	Description string // Ticket summary from JIRA (populated at runtime)
}

// IsPending returns true if entry needs to be posted
func (e Entry) IsPending() bool {
	return e.Status == StatusPending
}

// IsDone returns true if entry is completed
func (e Entry) IsDone() bool {
	return strings.ToUpper(e.Status) == StatusDone
}

// IsUpdated returns true if entry was posted but needs verification
func (e Entry) IsUpdated() bool {
	return strings.ToUpper(e.Status) == StatusUpdated
}

// IsSync returns true if entry matches what's in JIRA
func (e Entry) IsSync() bool {
	return strings.ToUpper(e.Status) == StatusSync
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
	profile     *config.JiraProfile
	mockMode    bool              // For LOCAL profile - simulate without JIRA calls
	currentUser *jira.CurrentUser // Cached current user for ownership checks
}

// ProcessorConfig configures batch processor behavior
type ProcessorConfig struct {
	Client      *jira.Client
	DefaultTime string
	Profile     *config.JiraProfile
	MockMode    bool // For LOCAL profile - simulate without JIRA calls
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

// NewProcessorWithConfig creates a processor with full configuration
func NewProcessorWithConfig(cfg ProcessorConfig) *Processor {
	if cfg.DefaultTime == "" {
		cfg.DefaultTime = "09:00"
	}
	return &Processor{
		client:      cfg.Client,
		defaultTime: cfg.DefaultTime,
		profile:     cfg.Profile,
		mockMode:    cfg.MockMode,
	}
}

// DefaultCSVPath returns the default path for the worklogs CSV
func DefaultCSVPath() string {
	return config.GetConfigPath("worklogs.csv")
}

// DefaultCSVPathForProfile returns the CSV path based on the active profile
// LOCAL profiles use worklogs-local.csv to avoid mixing test data with production
func DefaultCSVPathForProfile(profile *config.JiraProfile) string {
	if profile != nil && profile.IsLocal() {
		return config.GetConfigPath("worklogs-local.csv")
	}
	return config.GetConfigPath("worklogs.csv")
}

// CSV column indices (6-column format)
// Format: issue_key, issue_description, TimeSpent, Date, Comment, Status
const (
	ColIssueKey    = 0
	ColDescription = 1
	ColTimeSpent   = 2
	ColDate        = 3
	ColComment     = 4
	ColStatus      = 5
)

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

		if len(record) < 4 {
			continue
		}

		// Skip header row
		if rowNum == 1 {
			first := strings.ToLower(record[ColIssueKey])
			if first == "issue_key" || first == "issue" || first == "ticket" {
				continue
			}
		}

		entry := Entry{
			IssueKey:    strings.ToUpper(strings.TrimSpace(record[ColIssueKey])),
			Description: strings.TrimSpace(record[ColDescription]),
			TimeSpent:   strings.TrimSpace(record[ColTimeSpent]),
			Date:        strings.TrimSpace(record[ColDate]),
			RowNumber:   rowNum,
		}

		if len(record) > ColComment {
			entry.Comment = strings.TrimSpace(record[ColComment])
		}

		if len(record) > ColStatus {
			entry.Status = strings.TrimSpace(record[ColStatus])
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

// ParseDate parses date string with optional time component
// Formats: DD.MM.YYYY HH:MM, DD.MM.YYYY, DD/MM/YYYY, YYYY-MM-DD
// timeStr parameter is used as fallback if date doesn't contain time
func ParseDate(dateStr, timeStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(timeStr)

	var day, month, year, hour, minute int
	hour, minute = 9, 0 // Default time

	// Check if dateStr contains time component (space followed by HH:MM)
	if idx := strings.Index(dateStr, " "); idx > 0 {
		datePart := dateStr[:idx]
		timePart := strings.TrimSpace(dateStr[idx+1:])

		// Parse time from dateStr
		if _, err := fmt.Sscanf(timePart, "%d:%d", &hour, &minute); err == nil {
			dateStr = datePart // Use only date part for date parsing
		}
	} else if timeStr != "" {
		// Use timeStr parameter as fallback
		fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	}

	// Parse date part
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

	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local), nil
}

// isOwnWorklog checks if a worklog belongs to the current user
func (p *Processor) isOwnWorklog(wl jira.Worklog) bool {
	if p.currentUser == nil {
		user, err := p.client.GetCurrentUser()
		if err != nil {
			return false // Fail safe - don't touch if can't verify
		}
		p.currentUser = user
	}
	return wl.Author == p.currentUser.DisplayName
}

// Process processes a single entry (legacy - use SyncWorklog instead)
func (p *Processor) Process(entry Entry) Result {
	return p.SyncWorklog(entry)
}

// SyncWorklog syncs a worklog entry - compares CSV against JIRA data
// Behavior:
// - If matching time+duration exists in JIRA → SKIP (already exists)
// - If matching time but different duration → DELETE old, CREATE new (UPDATED)
// - If no matching time but other worklogs exist → FAIL (time conflict)
// - If no worklogs for this date → CREATE new (DONE)
func (p *Processor) SyncWorklog(entry Entry) Result {
	result := Result{Entry: entry}

	// 1. Parse date (uses time from CSV if present, otherwise --time flag)
	csvTime, err := ParseDate(entry.Date, p.defaultTime)
	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result
	}

	// 2. Mock mode - skip JIRA calls for LOCAL testing
	if p.mockMode {
		result.Success = true
		result.NewStatus = StatusDone
		result.ErrorMessage = "mock"
		return result
	}

	// 3. Fetch existing worklogs for this issue/date
	existing, err := p.client.GetWorklogsByDate(entry.IssueKey, csvTime)
	if err != nil {
		errStr := err.Error()
		// Check for 404 (issue not found)
		if strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") {
			result.Success = false
			result.ErrorMessage = "issue not found"
			return result
		}
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to fetch worklogs: %v", err)
		return result
	}

	// 3. Parse CSV duration for comparison
	csvDuration, _ := duration.Parse(entry.TimeSpent)
	csvHour, csvMin := csvTime.Hour(), csvTime.Minute()

	// 4. Look for a worklog with matching time
	for _, jiraWL := range existing {
		jiraHour, jiraMin := jiraWL.Started.Hour(), jiraWL.Started.Minute()

		if csvHour == jiraHour && csvMin == jiraMin {
			// Same time - check duration
			if csvDuration == jiraWL.TimeSpent {
				// Already exists with same duration - SKIP
				result.Success = true
				result.NewStatus = StatusDone
				result.ErrorMessage = "already exists"
				return result
			}

			// Same time, different duration - delete old and create new
			// First check if this is our worklog
			if !p.isOwnWorklog(jiraWL) {
				result.Success = false
				result.ErrorMessage = fmt.Sprintf("worklog belongs to %s, skipping", jiraWL.Author)
				return result
			}

			if err := p.client.DeleteWorklog(entry.IssueKey, jiraWL.ID); err != nil {
				result.Success = false
				result.ErrorMessage = fmt.Sprintf("failed to delete existing worklog: %v", err)
				return result
			}

			worklogEntry := jira.WorklogEntry{
				TimeSpent: entry.TimeSpent,
				Started:   csvTime,
				Comment:   entry.Comment,
			}

			if err := p.client.LogWork(entry.IssueKey, worklogEntry); err != nil {
				result.Success = false
				result.ErrorMessage = err.Error()
				return result
			}

			result.Success = true
			result.NewStatus = StatusUpdated
			result.ErrorMessage = "updated duration"
			return result
		}
	}

	// 5. No matching time found - check for conflicts
	if len(existing) > 0 {
		// There are worklogs for this date, but at different times
		// This is a conflict - CSV says one time, JIRA has another
		jiraTime := existing[0].Started
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("time conflict: CSV has %02d:%02d but JIRA has %02d:%02d",
			csvHour, csvMin, jiraTime.Hour(), jiraTime.Minute())
		return result
	}

	// 6. No worklogs for this date - create new
	worklogEntry := jira.WorklogEntry{
		TimeSpent: entry.TimeSpent,
		Started:   csvTime,
		Comment:   entry.Comment,
	}

	if err := p.client.LogWork(entry.IssueKey, worklogEntry); err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result
	}

	result.Success = true
	result.NewStatus = StatusDone
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
					NewStatus: StatusUpdated,
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
			// Ensure we have at least 6 columns (new format)
			for len(record) < 6 {
				record = append(record, "")
			}
			record[ColStatus] = newStatus
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

// UpdateCSVDescriptions updates empty description fields in the CSV
func UpdateCSVDescriptions(path string, descriptions map[string]string) (int, error) {
	if len(descriptions) == 0 {
		return 0, nil
	}

	// Read all lines
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	f.Close()
	if err != nil {
		return 0, err
	}

	// Update descriptions
	updated := 0
	for i, record := range records {
		if len(record) <= ColDescription {
			continue
		}

		issueKey := strings.ToUpper(strings.TrimSpace(record[ColIssueKey]))
		currentDesc := strings.TrimSpace(record[ColDescription])

		// Only update if description is empty and we have a new one
		if currentDesc == "" {
			if newDesc, ok := descriptions[issueKey]; ok && newDesc != "" {
				record[ColDescription] = newDesc
				records[i] = record
				updated++
			}
		}
	}

	if updated == 0 {
		return 0, nil
	}

	// Write back
	f, err = os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	if err := writer.WriteAll(records); err != nil {
		return 0, err
	}

	return updated, nil
}

// AppendEntry appends a new entry to the CSV file
func AppendEntry(path string, issueKey, description, timeSpent, date, comment string) error {
	return AppendEntryWithStatus(path, issueKey, description, timeSpent, date, comment, StatusPending)
}

// AppendEntryWithStatus appends a new entry to the CSV file with a specific status
func AppendEntryWithStatus(path, issueKey, description, timeSpent, date, comment, status string) error {
	if date == "" {
		date = time.Now().Format("02.01.2006") // DD.MM.YYYY format
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
		description,
		timeSpent,
		date,
		comment,
		status,
	})
}

// FindMissingWorklogs compares JIRA worklogs with CSV entries and returns missing ones
func FindMissingWorklogs(jiraWorklogs []jira.Worklog, csvEntries []Entry) []jira.Worklog {
	// Build a set of existing CSV entries keyed by IssueKey+Date+TimeSpent
	existing := make(map[string]bool)
	for _, e := range csvEntries {
		key := buildComparisonKey(e.IssueKey, e.Date, e.TimeSpent)
		existing[key] = true
	}

	var missing []jira.Worklog
	for _, wl := range jiraWorklogs {
		dateStr := wl.Started.Format("02.01.2006 15:04")
		key := buildComparisonKey(wl.IssueKey, dateStr, wl.TimeSpentStr)

		if !existing[key] {
			missing = append(missing, wl)
		}
	}

	return missing
}

// buildComparisonKey creates a unique identifier for worklog comparison
func buildComparisonKey(issueKey, date, timeSpent string) string {
	issueKey = strings.ToUpper(strings.TrimSpace(issueKey))
	date = normalizeDate(date)
	timeSpent = normalizeTimeSpent(timeSpent)
	return fmt.Sprintf("%s|%s|%s", issueKey, date, timeSpent)
}

// normalizeDate converts date to consistent YYYY-MM-DD HH:MM format for comparison
func normalizeDate(date string) string {
	t, err := ParseDate(date, "00:00")
	if err != nil {
		return strings.TrimSpace(date)
	}
	return t.Format("2006-01-02 15:04")
}

// normalizeTimeSpent normalizes time format for comparison
func normalizeTimeSpent(ts string) string {
	d, err := duration.Parse(ts)
	if err != nil {
		return strings.TrimSpace(ts)
	}
	return duration.Format(d)
}

// SortCSVByDate sorts CSV entries by date (newest first) and rewrites the file
func SortCSVByDate(path string) error {
	// Read all records
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

	if len(records) <= 1 {
		return nil // Nothing to sort
	}

	// Check if first row is header
	hasHeader := false
	var header []string
	if len(records) > 0 {
		first := strings.ToLower(records[0][0])
		if first == "issue_key" || first == "issue" || first == "ticket" {
			hasHeader = true
			header = records[0]
			records = records[1:]
		}
	}

	// Sort records by date descending (newest first)
	sortRecordsByDate(records)

	// Write back to file
	f, err = os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write header if present
	if hasHeader {
		if err := writer.Write(header); err != nil {
			return err
		}
	}

	// Write sorted records
	return writer.WriteAll(records)
}

// sortRecordsByDate sorts CSV records by date column (index 2) in descending order
func sortRecordsByDate(records [][]string) {
	for i := 0; i < len(records)-1; i++ {
		for j := i + 1; j < len(records); j++ {
			dateI := parseRecordDate(records[i])
			dateJ := parseRecordDate(records[j])

			// Sort descending (newest first)
			if dateJ.After(dateI) {
				records[i], records[j] = records[j], records[i]
			}
		}
	}
}

// parseRecordDate parses date from CSV record
func parseRecordDate(record []string) time.Time {
	if len(record) <= ColDate {
		return time.Time{}
	}
	t, err := ParseDate(record[ColDate], "00:00")
	if err != nil {
		return time.Time{}
	}
	return t
}
