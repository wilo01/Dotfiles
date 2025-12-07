package context

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var ticketPattern = regexp.MustCompile(`(?i)(vis|tdt|task|bug|feat)-(\d+)`)

// WorkContext holds information about the current work context
type WorkContext struct {
	Repository   string
	Branch       string
	Ticket       string
	Description  string
	SessionStart time.Time
}

// Detector detects work context from the environment
type Detector struct {
	logDir string
}

// NewDetector creates a new context detector
func NewDetector() *Detector {
	home, _ := os.UserHomeDir()
	return &Detector{
		logDir: filepath.Join(home, "Dev", "Private", "AI", "logger"),
	}
}

// Detect detects the current work context
func (d *Detector) Detect() WorkContext {
	ctx := WorkContext{
		SessionStart: time.Now(),
	}

	// Try to detect from git
	ctx.Repository = d.getRepository()
	ctx.Branch = d.getBranch()

	// Try to extract ticket from various sources
	ticket, desc := d.detectTicket()
	ctx.Ticket = ticket
	ctx.Description = desc

	return ctx
}

// DetectTicket detects just the ticket number
func (d *Detector) DetectTicket() (ticket string, description string) {
	return d.detectTicket()
}

// detectTicket tries to detect a ticket from various sources
func (d *Detector) detectTicket() (ticket string, description string) {
	// 1. Check current working directory
	cwd, _ := os.Getwd()
	if ticket, desc := extractTicketFromPath(cwd); ticket != "" {
		return ticket, desc
	}

	// 2. Check git branch
	if branch := d.getBranch(); branch != "" {
		if ticket, desc := extractTicketFromPath(branch); ticket != "" {
			return ticket, desc
		}
	}

	// 3. Check environment variables
	for _, env := range []string{"CLAUDE_PROJECT_DIR", "PWD", "OLDPWD"} {
		if path := os.Getenv(env); path != "" {
			if ticket, desc := extractTicketFromPath(path); ticket != "" {
				return ticket, desc
			}
		}
	}

	// 4. Check recent log files
	if ticket, desc := d.checkRecentLogFiles(); ticket != "" {
		return ticket, desc
	}

	return "", ""
}

// extractTicketFromPath extracts a JIRA ticket from a path string
func extractTicketFromPath(path string) (ticket string, description string) {
	match := ticketPattern.FindStringSubmatch(path)
	if len(match) >= 3 {
		ticket = strings.ToUpper(match[1]) + "-" + match[2]

		// Try to extract description from the path
		idx := strings.Index(strings.ToLower(path), strings.ToLower(match[0]))
		if idx >= 0 {
			remaining := path[idx+len(match[0]):]
			remaining = strings.TrimPrefix(remaining, "-")
			remaining = strings.TrimPrefix(remaining, "/")

			// Take first path component as description
			parts := strings.SplitN(remaining, "/", 2)
			if len(parts) > 0 && parts[0] != "" {
				description = strings.ReplaceAll(parts[0], "-", " ")
				description = strings.Title(description)
			}
		}

		return ticket, description
	}
	return "", ""
}

// checkRecentLogFiles looks for recent VIS-*.md files in the log directory
func (d *Detector) checkRecentLogFiles() (ticket string, description string) {
	if d.logDir == "" {
		return "", ""
	}

	entries, err := os.ReadDir(d.logDir)
	if err != nil {
		return "", ""
	}

	// Find VIS-*.md files
	type fileInfo struct {
		name    string
		modTime time.Time
	}
	var visFiles []fileInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(strings.ToUpper(name), "VIS-") || !strings.HasSuffix(name, ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		visFiles = append(visFiles, fileInfo{name: name, modTime: info.ModTime()})
	}

	if len(visFiles) == 0 {
		return "", ""
	}

	// Sort by modification time (newest first)
	sort.Slice(visFiles, func(i, j int) bool {
		return visFiles[i].modTime.After(visFiles[j].modTime)
	})

	// Parse the most recent file
	name := visFiles[0].name
	return extractTicketFromPath(name)
}

// getRepository returns the current git repository name
func (d *Detector) getRepository() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	path := strings.TrimSpace(string(out))
	return filepath.Base(path)
}

// getBranch returns the current git branch
func (d *Detector) getBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getLastCommitMessage returns the last commit message
func (d *Detector) getLastCommitMessage() string {
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
