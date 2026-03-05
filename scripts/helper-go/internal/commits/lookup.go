package commits

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// CommitEntry holds parsed data from a Commits.md entry.
type CommitEntry struct {
	Ticket     string
	BranchName string
	Logs       []string
}

// CherryEntry holds the data needed to write a cherry-pick entry to Commits.md.
type CherryEntry struct {
	TargetBranch string
	Ticket       string
	Version      string
	BranchName   string
	Logs         []string
	CherryNote   string
	ChangedFiles []string
	CommitSHA    string
}

// GetLogs returns the logs slice, safe to call on nil CherryEntry receiver.
func (e *CommitEntry) GetLogs() []string {
	if e == nil {
		return nil
	}
	return e.Logs
}

// TicketPattern matches a JIRA ticket key at the start of a string (e.g. "VIS-1234").
var TicketPattern = regexp.MustCompile(`(?i)^([A-Z]+-\d+)`)

// LookupByTicket searches Commits.md backwards for the most recent master entry
// matching the given JIRA ticket. Returns nil, nil if not found.
func LookupByTicket(filePath, ticket string) (*CommitEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read commits file: %w", err)
	}

	entries := splitEntries(string(data))

	// Search backwards (most recent first), prefer entries with logs
	var fallback *CommitEntry
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if !isMasterEntry(entry) {
			continue
		}
		parsed := parseEntry(entry)
		if parsed == nil || !strings.EqualFold(parsed.Ticket, ticket) {
			continue
		}
		// Prefer entries with logs (the original commit, not the merge entry)
		if len(parsed.Logs) > 0 {
			return parsed, nil
		}
		if fallback == nil {
			fallback = parsed
		}
	}

	return fallback, nil
}

// splitEntries splits Commits.md content by date headers (lines starting with "# ").
func splitEntries(content string) []string {
	var entries []string
	var current strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") && current.Len() > 0 {
			entries = append(entries, current.String())
			current.Reset()
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		entries = append(entries, current.String())
	}

	return entries
}

// isMasterEntry checks if an entry has "## master" as the branch section.
func isMasterEntry(entry string) bool {
	scanner := bufio.NewScanner(strings.NewReader(entry))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "## master" {
			return true
		}
	}
	return false
}

// parseEntry extracts ticket, branch name, and logs from a Commits.md entry.
func parseEntry(entry string) *CommitEntry {
	lines := strings.Split(entry, "\n")

	var ticket, branchName string
	var logs []string
	inLogs := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip headers and empty lines for ticket detection
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			if inLogs && trimmed == "" {
				// Empty line after logs — continue (logs section might have gaps)
				continue
			}
			continue
		}

		// Detect "Logs:" section start
		if trimmed == "Logs:" {
			inLogs = true
			continue
		}

		// Detect end of logs section
		if inLogs {
			if strings.HasPrefix(trimmed, "Changed files:") || strings.HasPrefix(trimmed, "- Commit branch") {
				inLogs = false
				continue
			}
			if strings.HasPrefix(trimmed, "- ") {
				logs = append(logs, strings.TrimPrefix(trimmed, "- "))
				continue
			}
			// Non-bullet line in logs section — treat as log content without bullet
			if trimmed != "" {
				logs = append(logs, trimmed)
			}
			continue
		}

		// Extract ticket (first non-header, non-empty line matching ticket pattern)
		if ticket == "" {
			match := TicketPattern.FindStringSubmatch(trimmed)
			if len(match) >= 2 {
				ticket = strings.ToUpper(match[1])
				continue
			}
		}

		// Extract branch name (line after ticket, starts with ticket prefix and has dashes)
		if ticket != "" && branchName == "" &&
			strings.HasPrefix(strings.ToUpper(trimmed), ticket+"-") &&
			!strings.HasPrefix(trimmed, "- ") {
			branchName = trimmed
			continue
		}
	}

	if ticket == "" {
		return nil
	}

	return &CommitEntry{
		Ticket:     ticket,
		BranchName: branchName,
		Logs:       logs,
	}
}

// AppendCherryEntry appends a formatted cherry-pick entry to Commits.md.
func AppendCherryEntry(filePath string, entry CherryEntry) (retErr error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open commits file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && retErr == nil {
			retErr = fmt.Errorf("close commits file: %w", closeErr)
		}
	}()

	now := time.Now()
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("\n# %s\n", now.Format("2006-01-02 15:04")))
	b.WriteString(fmt.Sprintf("## %s\n", entry.TargetBranch))

	// Ticket with version
	if entry.Version != "" {
		b.WriteString(fmt.Sprintf("%s [%s]\n", entry.Ticket, entry.Version))
	} else {
		b.WriteString(fmt.Sprintf("%s\n", entry.Ticket))
	}

	// Branch name
	b.WriteString(fmt.Sprintf("\n%s\n", entry.BranchName))

	// Logs
	b.WriteString("\nLogs:\n")
	for _, log := range entry.Logs {
		b.WriteString(fmt.Sprintf("- %s\n", log))
	}
	if entry.CherryNote != "" {
		b.WriteString(fmt.Sprintf("- %s\n", entry.CherryNote))
	}

	// Changed files
	if len(entry.ChangedFiles) > 0 {
		b.WriteString("\nChanged files: ")
		b.WriteString(strings.Join(entry.ChangedFiles, `\ `))
		b.WriteString(`\ `)
		b.WriteString("\n")
	}

	// Commit hash
	if entry.CommitSHA != "" {
		b.WriteString(fmt.Sprintf("- Commit branch HASH [%s](%s)\n", entry.CommitSHA, entry.BranchName))
	}

	_, err = f.WriteString(b.String())
	return err
}
