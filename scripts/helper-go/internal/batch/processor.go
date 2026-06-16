package batch

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/dariuszw/hlp/pkg/duration"
)

// Status constants for worklog entries
const (
	StatusPending = ""
	StatusSync    = "SYNC"
	StatusUpdated = "UPDATED"
	StatusDone    = "DONE"
	StatusDraft   = "DRAFT" // Entry not ready for batch processing

	defaultStartTime = "09:00" // Default worklog start time
)

// Entry represents a single worklog entry from CSV
type Entry struct {
	IssueKey    string // Parent/main issue key
	SubtaskKey  string // Sub-task key if applicable (empty for non-subtasks)
	IssueType   string // Story, Bug, Task, etc.
	Description string // Ticket summary from JIRA (combined for subtasks)
	Comment     string
	Date        string // DD/MM/YYYY or DD.MM.YYYY
	TimeSpent   string
	SubtaskLogInd string // Y = log to subtask, N/empty = log to parent
	Status      string // empty = pending, SYNC, UPDATED, DONE
	RowNumber   int    // Internal - row number in CSV
}

// IsPending returns true if entry needs to be posted
func (e Entry) IsPending() bool {
	return e.Status == StatusPending
}

// IsDone returns true if entry is completed
func (e Entry) IsDone() bool {
	return strings.ToUpper(e.Status) == StatusDone
}

// IsDraft returns true if entry is a draft (not ready for batch)
func (e Entry) IsDraft() bool {
	return strings.ToUpper(e.Status) == StatusDraft
}

// IsUpdated returns true if entry was posted but needs verification
func (e Entry) IsUpdated() bool {
	return strings.ToUpper(e.Status) == StatusUpdated
}

// IsSync returns true if entry matches what's in JIRA
func (e Entry) IsSync() bool {
	return strings.ToUpper(e.Status) == StatusSync
}

// DisplayKey returns the display key for the entry
// For subtasks: "ParentKey > SubtaskKey", otherwise just IssueKey
func (e Entry) DisplayKey() string {
	if e.SubtaskKey != "" {
		return e.IssueKey + " > " + e.SubtaskKey
	}
	return e.IssueKey
}

// NeedsProcessing returns true if entry needs processing
func (e Entry) NeedsProcessing() bool {
	return !e.IsDone() && !e.IsDraft()
}

// GetLoggingTarget returns the key where worklog should be posted
func (e Entry) GetLoggingTarget() string {
	if strings.ToUpper(e.SubtaskLogInd) == "Y" && e.SubtaskKey != "" {
		return e.SubtaskKey
	}
	return e.IssueKey
}

// EntryExistsForTicket checks if an entry exists for the given ticket on the specified date.
// Matches on both IssueKey and SubtaskKey to handle subtask entries correctly.
// For subtasks, it also checks if the description contains the subtask summary.
func EntryExistsForTicket(entries []Entry, ticketKey, date string, isSubtask bool, subtaskSummary string) bool {
	for _, e := range entries {
		keyMatches := strings.EqualFold(e.IssueKey, ticketKey) ||
			(e.SubtaskKey != "" && strings.EqualFold(e.SubtaskKey, ticketKey))
		if !keyMatches {
			continue
		}
		entryDate := strings.Split(e.Date, " ")[0]
		descriptionMatches := !isSubtask || strings.Contains(e.Description, subtaskSummary)

		if (entryDate == date && descriptionMatches) ||
			e.TimeSpent != "" ||
			e.Status == StatusDone ||
			e.Status == StatusSync ||
			e.Status == StatusUpdated ||
			(e.Status == StatusDraft && descriptionMatches) {
			return true
		}
	}
	return false
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
		defaultTime = defaultStartTime
	}
	return &Processor{
		client:      client,
		defaultTime: defaultTime,
	}
}

// NewProcessorWithConfig creates a processor with full configuration
func NewProcessorWithConfig(cfg ProcessorConfig) *Processor {
	if cfg.DefaultTime == "" {
		cfg.DefaultTime = defaultStartTime
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

// CSV column indices (9-column format)
// Format: issue_key, subtask_key, issue_type, description, comment, date, time_spent, log_to, status
const (
	ColIssueKey    = 0
	ColSubtaskKey  = 1
	ColIssueType   = 2
	ColDescription = 3
	ColComment     = 4
	ColDate        = 5
	ColTimeSpent   = 6
	ColSubtaskLogInd = 7
	ColStatus      = 8
)

// forceQuote CSV-encodes a field, always wrapping it in double quotes.
func forceQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// encodeField CSV-encodes a field, quoting only when the content requires it.
func encodeField(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return forceQuote(s)
	}
	return s
}

// isHeaderRecord reports whether a record is the CSV header row.
func isHeaderRecord(record []string) bool {
	if len(record) == 0 {
		return false
	}
	first := strings.ToLower(record[0])
	return first == "issue_key" || first == "issue" || first == "ticket"
}

// writeWorklogRecords writes CSV records to w, always quoting the comment
// column (ColComment) in 9-column data rows — even when empty — so the
// comment field is visually consistent and safe for naive downstream parsers.
// The header row and legacy short rows use standard minimal quoting.
// Use this instead of csv.Writer for the worklogs file.
func writeWorklogRecords(w io.Writer, records [][]string) error {
	for _, record := range records {
		fields := make([]string, len(record))
		forceComment := len(record) >= 9 && !isHeaderRecord(record)
		for i, field := range record {
			if forceComment && i == ColComment {
				fields[i] = forceQuote(field)
			} else {
				fields[i] = encodeField(field)
			}
		}
		if _, err := fmt.Fprintln(w, strings.Join(fields, ",")); err != nil {
			return err
		}
	}
	return nil
}

// ParseCSV parses a CSV file into entries
// Supports 9-column (new), 7-column, and old 6-column formats for backward compatibility
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
			first := strings.ToLower(record[0])
			if first == "issue_key" || first == "issue" || first == "ticket" {
				continue
			}
		}

		// Detect format by column count
		if len(record) >= 9 {
			// New 9-column format: issue_key, subtask_key, issue_type, description, comment, date, time_spent, log_to, status
			entry := Entry{
				IssueKey:    strings.ToUpper(strings.TrimSpace(record[ColIssueKey])),
				SubtaskKey:  strings.ToUpper(strings.TrimSpace(record[ColSubtaskKey])),
				IssueType:   strings.TrimSpace(record[ColIssueType]),
				Description: strings.TrimSpace(record[ColDescription]),
				Comment:     strings.TrimSpace(record[ColComment]),
				Date:        strings.TrimSpace(record[ColDate]),
				TimeSpent:   strings.TrimSpace(record[ColTimeSpent]),
				SubtaskLogInd: strings.ToUpper(strings.TrimSpace(record[ColSubtaskLogInd])),
				Status:      strings.TrimSpace(record[ColStatus]),
				RowNumber:   rowNum,
			}
			entries = append(entries, entry)
		} else if len(record) >= 7 {
			// 7-column format: issue_key, issue_type, description, comment, date, time_spent, status
			// (no subtask_key or log_to - backward compatible)
			entry := Entry{
				IssueKey:    strings.ToUpper(strings.TrimSpace(record[0])),
				IssueType:   strings.TrimSpace(record[1]),
				Description: strings.TrimSpace(record[2]),
				Comment:     strings.TrimSpace(record[3]),
				Date:        strings.TrimSpace(record[4]),
				TimeSpent:   strings.TrimSpace(record[5]),
				Status:      strings.TrimSpace(record[6]),
				RowNumber:   rowNum,
			}
			entries = append(entries, entry)
		} else {
			// Old 6-column format: issue_key, description, time_spent, date, comment, status
			entry := Entry{
				IssueKey:    strings.ToUpper(strings.TrimSpace(record[0])),
				Description: strings.TrimSpace(record[1]),
				TimeSpent:   strings.TrimSpace(record[2]),
				Date:        strings.TrimSpace(record[3]),
				RowNumber:   rowNum,
			}
			if len(record) > 4 {
				entry.Comment = strings.TrimSpace(record[4])
			}
			if len(record) > 5 {
				entry.Status = strings.TrimSpace(record[5])
			}
			entries = append(entries, entry)
		}
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

	// Determine the target issue key based on log_to setting
	targetKey := entry.GetLoggingTarget()

	// 1. Parse date (uses time from CSV if present, otherwise --time flag)
	csvTime, err := ParseDate(entry.Date, p.defaultTime)
	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result
	}

	// 2. Reject entries with no time spent before any JIRA call
	if strings.TrimSpace(entry.TimeSpent) == "" {
		result.Success = false
		result.ErrorMessage = "no time spent specified"
		return result
	}

	// 3. Mock mode - skip JIRA calls for LOCAL testing
	if p.mockMode {
		result.Success = true
		result.NewStatus = StatusDone
		result.ErrorMessage = "mock"
		return result
	}

	// 4. Fetch existing worklogs for this issue/date
	existing, err := p.client.GetWorklogsByDate(targetKey, csvTime)
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

	// 5. Parse CSV duration for comparison
	csvDuration, _ := duration.Parse(entry.TimeSpent)
	csvHour, csvMin := csvTime.Hour(), csvTime.Minute()

	// 6. Look for a worklog with matching time
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

			if err := p.client.DeleteWorklog(targetKey, jiraWL.ID); err != nil {
				result.Success = false
				result.ErrorMessage = fmt.Sprintf("failed to delete existing worklog: %v", err)
				return result
			}

			worklogEntry := jira.WorklogEntry{
				TimeSpent: entry.TimeSpent,
				Started:   csvTime,
				Comment:   entry.Comment,
			}

			if err := p.client.LogWork(targetKey, worklogEntry); err != nil {
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

	// 7. No matching time found - check for conflicts
	if len(existing) > 0 {
		// There are worklogs for this date, but at different times
		// This is a conflict - CSV says one time, JIRA has another
		jiraTime := existing[0].Started
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("time conflict: CSV has %02d:%02d but JIRA has %02d:%02d",
			csvHour, csvMin, jiraTime.Hour(), jiraTime.Minute())
		return result
	}

	// 8. No worklogs for this date - create new
	worklogEntry := jira.WorklogEntry{
		TimeSpent: entry.TimeSpent,
		Started:   csvTime,
		Comment:   entry.Comment,
	}

	if err := p.client.LogWork(targetKey, worklogEntry); err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result
	}

	result.Success = true
	result.NewStatus = StatusDone
	return result
}

// ProcessBatch processes all entries
// slowMode adds random delays (20-120s) between entries to make timestamps appear more organic
func (p *Processor) ProcessBatch(entries []Entry, dryRun bool, slowMode bool, progressFn func(current, total int, result Result)) []Result {
	var results []Result
	total := len(entries)

	// Setup signal handling for graceful Ctrl+C (only in slow mode)
	ctx, cancel := context.WithCancel(context.Background())
	if slowMode {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			select {
			case <-sigChan:
				cancel()
			case <-ctx.Done():
				// Exit goroutine when context is cancelled (normal completion)
			}
		}()
		defer signal.Stop(sigChan)
	}
	defer cancel()

	for i, entry := range entries {
		// Check for cancellation
		select {
		case <-ctx.Done():
			fmt.Println("\n  Interrupted - stopping batch (already posted entries remain)")
			return results
		default:
		}

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

		// Add delay between entries in slow mode (not after last one, not on failure, not in dry-run)
		if slowMode && !dryRun && result.Success && i < len(entries)-1 {
			delay := 20 + rand.Intn(101) // 20-120 seconds
			if !waitWithCountdown(ctx, delay) {
				fmt.Println("\n  Interrupted - stopping batch (already posted entries remain)")
				return results
			}
		}
	}

	return results
}

// waitWithCountdown displays a countdown and waits, returning false if cancelled
func waitWithCountdown(ctx context.Context, seconds int) bool {
	for remaining := seconds; remaining > 0; remaining-- {
		select {
		case <-ctx.Done():
			fmt.Println()
			return false
		default:
			mins := remaining / 60
			secs := remaining % 60
			if mins > 0 {
				fmt.Printf("\r  Waiting %dm%02ds before next entry... (Ctrl+C to cancel)", mins, secs)
			} else {
				fmt.Printf("\r  Waiting %ds before next entry... (Ctrl+C to cancel)    ", secs)
			}
			time.Sleep(1 * time.Second)
		}
	}
	// Clear the countdown line
	ui.ClearLine()
	return true
}

// WaitUntilTime waits until the target time, showing a countdown.
// The progressFn is called immediately, then every second with the remaining duration.
// Returns false if cancelled via context, true if target time was reached.
func WaitUntilTime(ctx context.Context, target time.Time, progressFn func(remaining time.Duration)) bool {
	// Immediate first display (before ticker starts)
	remaining := time.Until(target)
	if remaining <= 0 {
		return true
	}
	if progressFn != nil {
		progressFn(remaining)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			remaining = time.Until(target)
			if remaining <= 0 {
				return true
			}
			if progressFn != nil {
				progressFn(remaining)
			}
		}
	}
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
			// Ensure we have at least 9 columns (new format)
			if len(record) < 9 {
				fmt.Printf("Warning: Row %d has %d columns, padding to 9\n", rowNum, len(record))
				for len(record) < 9 {
					record = append(record, "")
				}
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

	return writeWorklogRecords(f, records)
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

	if err := writeWorklogRecords(f, records); err != nil {
		return 0, err
	}

	return updated, nil
}

// AppendEntry appends a new entry to the CSV file
func AppendEntry(path string, issueKey, subtaskKey, issueType, description, timeSpent, date, comment string) error {
	return AppendEntryWithStatus(path, issueKey, subtaskKey, issueType, description, timeSpent, date, comment, "", StatusPending)
}

// AppendEntryWithStatus appends a new entry to the CSV file with a specific status
// 9-column format: issue_key, subtask_key, issue_type, description, comment, date, time_spent, subtask_log_ind, status
func AppendEntryWithStatus(path, issueKey, subtaskKey, issueType, description, timeSpent, date, comment, subtaskLogInd, status string) error {
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

	// Default: log to parent. Callers opt into subtask logging via explicit "Y".
	if subtaskLogInd == "" {
		subtaskLogInd = "N"
	}

	return writeWorklogRecords(f, [][]string{{
		strings.ToUpper(issueKey),
		strings.ToUpper(subtaskKey),
		issueType,
		description,
		comment,
		date,
		timeSpent,
		strings.ToUpper(subtaskLogInd),
		status,
	}})
}

// PrependEntryWithStatus adds a new entry at the top of the CSV file (after header)
// 9-column format: issue_key, subtask_key, issue_type, description, comment, date, time_spent, subtask_log_ind, status
func PrependEntryWithStatus(path, issueKey, subtaskKey, issueType, description, timeSpent, date, comment, subtaskLogInd, status string) error {
	if date == "" {
		date = time.Now().Format("02.01.2006")
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Read existing file
	var records [][]string
	if f, err := os.Open(path); err == nil {
		reader := csv.NewReader(f)
		reader.FieldsPerRecord = -1
		records, _ = reader.ReadAll()
		f.Close()
	}

	// Default: log to parent. Callers opt into subtask logging via explicit "Y".
	if subtaskLogInd == "" {
		subtaskLogInd = "N"
	}

	// Create new entry (9-column format)
	newEntry := []string{
		strings.ToUpper(issueKey),
		strings.ToUpper(subtaskKey),
		issueType,
		description,
		comment,
		date,
		timeSpent,
		strings.ToUpper(subtaskLogInd),
		status,
	}

	// Insert after header (or at position 0 if no header)
	var result [][]string
	if len(records) > 0 {
		// Check if first row is header
		first := strings.ToLower(records[0][0])
		if first == "issue_key" || first == "issue" || first == "ticket" {
			// Has header - insert after it
			result = append(result, records[0])
			result = append(result, newEntry)
			result = append(result, records[1:]...)
		} else {
			// No header - insert at top
			result = append(result, newEntry)
			result = append(result, records...)
		}
	} else {
		result = [][]string{newEntry}
	}

	// Write back
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return writeWorklogRecords(f, result)
}

// RemoveEntryByRow removes a CSV row by its 1-based row number
func RemoveEntryByRow(path string, rowNumber int) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("remove entry: open CSV: %w", err)
	}

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	f.Close()
	if err != nil {
		return fmt.Errorf("remove entry: read CSV: %w", err)
	}

	idx := rowNumber - 1
	if idx < 0 || idx >= len(records) {
		return fmt.Errorf("row %d out of range (1-%d)", rowNumber, len(records))
	}

	records = append(records[:idx], records[idx+1:]...)

	f, err = os.Create(path)
	if err != nil {
		return fmt.Errorf("remove entry: write CSV: %w", err)
	}
	defer f.Close()

	return writeWorklogRecords(f, records)
}

// UpdateEntryDescription updates the description of an existing entry by row number
func UpdateEntryDescription(path string, rowNumber int, newDescription string) error {
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

	// Update the description at the specified row
	if rowNumber > 0 && rowNumber <= len(records) {
		idx := rowNumber - 1 // Convert to 0-based index
		if len(records[idx]) > ColDescription {
			records[idx][ColDescription] = newDescription
		}
	}

	// Write back
	f, err = os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return writeWorklogRecords(f, records)
}

// FindMissingWorklogs compares JIRA worklogs with CSV entries and returns missing ones
func FindMissingWorklogs(jiraWorklogs []jira.Worklog, csvEntries []Entry) []jira.Worklog {
	// Build a set of existing CSV entries keyed by IssueKey+Date+TimeSpent
	// When SubtaskLogInd=Y, use SubtaskKey as the comparison key (worklog was logged to subtask)
	existing := make(map[string]bool)
	for _, e := range csvEntries {
		// Use the key where worklog was actually logged
		issueKey := e.IssueKey
		if strings.EqualFold(e.SubtaskLogInd, "Y") && e.SubtaskKey != "" {
			issueKey = e.SubtaskKey // Worklog was logged to subtask
		}
		key := buildComparisonKey(issueKey, e.Date, e.TimeSpent)
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

	// Write header (if present) followed by sorted records
	if hasHeader {
		records = append([][]string{header}, records...)
	}
	return writeWorklogRecords(f, records)
}

// sortRecordsByDate sorts CSV records by date column (index 2) in descending order
func sortRecordsByDate(records [][]string) {
	sort.Slice(records, func(i, j int) bool {
		dateI := parseRecordDate(records[i])
		dateJ := parseRecordDate(records[j])
		// Sort descending (newest first)
		return dateJ.Before(dateI)
	})
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

// unique returns unique strings from a slice
func unique(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// EntryOrder defines the order for processing entries
type EntryOrder string

const (
	OrderOldest EntryOrder = "oldest" // Process oldest dates first (ascending)
	OrderNewest EntryOrder = "newest" // Process newest dates first (descending)
	OrderRandom EntryOrder = "random" // Shuffle entries randomly
)

// SortEntries sorts entries according to the specified order
// Used with --slow mode to control processing order
func SortEntries(entries []Entry, order EntryOrder) {
	switch order {
	case OrderRandom:
		rand.Shuffle(len(entries), func(i, j int) {
			entries[i], entries[j] = entries[j], entries[i]
		})
	case OrderNewest:
		sort.SliceStable(entries, func(i, j int) bool {
			dateI, _ := ParseDate(entries[i].Date, "00:00")
			dateJ, _ := ParseDate(entries[j].Date, "00:00")
			return dateJ.Before(dateI) // Descending (newest first)
		})
	case OrderOldest:
		fallthrough
	default:
		sort.SliceStable(entries, func(i, j int) bool {
			dateI, _ := ParseDate(entries[i].Date, "00:00")
			dateJ, _ := ParseDate(entries[j].Date, "00:00")
			return dateI.Before(dateJ) // Ascending (oldest first)
		})
	}
}

// EnrichAndRestructureCSV updates empty descriptions/types AND restructures subtask entries
// Returns (enrichedCount, restructuredCount, error)
func EnrichAndRestructureCSV(path string, client *jira.Client) (int, int, error) {
	entries, err := ParseCSV(path)
	if err != nil {
		return 0, 0, err
	}

	// Collect keys needing lookup (empty desc/type OR potential subtasks without SubtaskKey)
	var keysToFetch []string
	for _, e := range entries {
		if e.Description == "" || e.IssueType == "" {
			keysToFetch = append(keysToFetch, e.IssueKey)
		}
		// For entries that might be subtasks stored in IssueKey (SubtaskKey is empty)
		if e.SubtaskKey == "" && e.IssueKey != "" {
			keysToFetch = append(keysToFetch, e.IssueKey)
		}
	}

	if len(keysToFetch) == 0 {
		return 0, 0, nil
	}

	// Fetch issue details
	details, err := client.GetIssueDetails(unique(keysToFetch))
	if err != nil {
		return 0, 0, err
	}

	// Collect parent keys for subtasks that need restructuring
	var parentKeys []string
	for _, detail := range details {
		if detail.IsSubtask && detail.ParentKey != "" {
			parentKeys = append(parentKeys, detail.ParentKey)
		}
	}

	// Fetch parent details
	var parentDetails map[string]jira.IssueDetails
	if len(parentKeys) > 0 {
		parentDetails, _ = client.GetIssueDetails(unique(parentKeys))
	}

	// Update CSV
	return updateCSVWithRestructure(path, details, parentDetails)
}

// updateCSVWithRestructure updates the CSV file with enrichment and restructuring
func updateCSVWithRestructure(path string, details map[string]jira.IssueDetails, parentDetails map[string]jira.IssueDetails) (int, int, error) {
	// Read all records
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	f.Close()
	if err != nil {
		return 0, 0, err
	}

	enriched := 0
	restructured := 0

	for i, record := range records {
		if len(record) < 9 {
			continue
		}

		// Skip header
		if i == 0 {
			first := strings.ToLower(record[0])
			if first == "issue_key" || first == "issue" || first == "ticket" {
				continue
			}
		}

		issueKey := strings.ToUpper(strings.TrimSpace(record[ColIssueKey]))
		subtaskKey := strings.ToUpper(strings.TrimSpace(record[ColSubtaskKey]))
		currentDesc := strings.TrimSpace(record[ColDescription])
		currentType := strings.TrimSpace(record[ColIssueType])

		detail, hasDetail := details[issueKey]

		// Check if this entry needs restructuring (IssueKey is a subtask but SubtaskKey is empty)
		if hasDetail && detail.IsSubtask && detail.ParentKey != "" && subtaskKey == "" {
			// Restructure: move issueKey to SubtaskKey, use parent as IssueKey
			parentDetail, hasParent := parentDetails[detail.ParentKey]
			if hasParent {
				record[ColSubtaskKey] = issueKey                                                            // Original key becomes subtask
				record[ColIssueKey] = detail.ParentKey                                                      // Parent becomes main key
				record[ColIssueType] = parentDetail.IssueType                                               // Parent's type
				record[ColDescription] = parentDetail.Summary + " > " + detail.Summary // Combined description
				record[ColSubtaskLogInd] = "Y"                                                              // Worklog was logged to subtask
				records[i] = record
				restructured++
				continue // Skip normal enrichment since we just did full restructure
			}
		}

		// Normal enrichment: fill empty Description and IssueType
		updated := false
		if currentDesc == "" && hasDetail && detail.Summary != "" {
			record[ColDescription] = detail.Summary
			updated = true
		}
		if currentType == "" && hasDetail && detail.IssueType != "" {
			record[ColIssueType] = detail.IssueType
			updated = true
		}
		if updated {
			records[i] = record
			enriched++
		}
	}

	if enriched == 0 && restructured == 0 {
		return 0, 0, nil
	}

	// Write back
	f, err = os.Create(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	if err := writeWorklogRecords(f, records); err != nil {
		return 0, 0, err
	}

	return enriched, restructured, nil
}
