// Package task models a unit of work as one Jira ticket worked on across N git
// worktrees, with a resumable Claude session and an optional dedicated hexer
// environment. Persisted as JSON at ~/.config/hlp/tasks.json.
package task

import (
	"crypto/rand"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Repo is one repository enlisted in a task. Source is the main checkout the
// worktree was created from; Base is the default branch resolved at creation
// time, kept so a later hexer provision or refresh uses the same ref.
type Repo struct {
	Name     string `json:"name"`
	Source   string `json:"source"`
	Worktree string `json:"worktree"`
	Base     string `json:"base"`
}

// Hexer describes a task's dedicated TDS environment: an Oracle container plus
// a hexer process serving the task's tds-suite worktree.
type Hexer struct {
	Enabled bool     `json:"enabled"`
	Modules []string `json:"modules"`
	Base    string   `json:"base"`
	Port    int      `json:"port"`
	DBPort  int      `json:"db_port"`
	Host    string   `json:"host"`
}

// Task is keyed by JiraKey: the key is the identity, which is what makes the
// tmux session name and the task directory derivable rather than stored.
type Task struct {
	JiraKey         string `json:"jira_key"`
	Summary         string `json:"summary"`
	TaskRoot        string `json:"task_root"`
	ClaudeSessionID string `json:"claude_session_id"`
	SessionStarted  bool   `json:"session_started"`
	Repos           []Repo `json:"repos"`
	CreatedMs       int64  `json:"created_ms"`
	LastOpenedMs    int64  `json:"last_opened_ms"`
	Hexer           Hexer  `json:"hexer"`
}

// New builds a task with a freshly minted Claude session UUID. The UUID is
// allocated up front so the first launch can pass --session-id and every later
// launch --resume, keeping one conversation per ticket.
func New(key, summary, taskRoot string) (Task, error) {
	sessionID, err := newUUID()
	if err != nil {
		return Task{}, err
	}
	now := nowMs()
	return Task{
		JiraKey:         strings.ToUpper(key),
		Summary:         summary,
		TaskRoot:        taskRoot,
		ClaudeSessionID: sessionID,
		CreatedMs:       now,
		LastOpenedMs:    now,
	}, nil
}

// RepoNames returns the enlisted repo names in stable sorted order.
func (t Task) RepoNames() []string {
	names := make([]string, 0, len(t.Repos))
	for _, r := range t.Repos {
		names = append(names, r.Name)
	}
	sort.Strings(names)
	return names
}

// RepoNamesInOrder returns the enlisted repo names in task order, which is the
// order they were selected and therefore the order of their tmux windows.
func (t Task) RepoNamesInOrder() []string {
	names := make([]string, 0, len(t.Repos))
	for _, r := range t.Repos {
		names = append(names, r.Name)
	}
	return names
}

// Repo returns the named repo and whether it is enlisted.
func (t Task) Repo(name string) (Repo, bool) {
	for _, r := range t.Repos {
		if r.Name == name {
			return r, true
		}
	}
	return Repo{}, false
}

// HasRepo reports whether the repo is enlisted in the task.
func (t Task) HasRepo(name string) bool {
	_, ok := t.Repo(name)
	return ok
}

// AddRepo appends a repo, replacing any existing entry with the same name.
func (t *Task) AddRepo(r Repo) {
	for i, existing := range t.Repos {
		if existing.Name == r.Name {
			t.Repos[i] = r
			return
		}
	}
	t.Repos = append(t.Repos, r)
}

// RemoveRepo drops the named repo and reports whether it was present.
func (t *Task) RemoveRepo(name string) bool {
	for i, r := range t.Repos {
		if r.Name == name {
			t.Repos = append(t.Repos[:i], t.Repos[i+1:]...)
			return true
		}
	}
	return false
}

// SessionName is the tmux session name for the task.
func (t Task) SessionName() string { return t.JiraKey }

// PrimaryRepo is the repo whose tmux window hosts the resumable Claude session.
// Repos are stored in selection order, so the first one is the user's own
// choice of primary rather than an alphabetical accident.
func (t Task) PrimaryRepo() (Repo, bool) {
	if len(t.Repos) == 0 {
		return Repo{}, false
	}
	return t.Repos[0], true
}

func nowMs() int64 { return time.Now().UnixMilli() }

// newUUID generates a random (version 4) UUID without pulling in a dependency.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("failed to generate session id: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
