package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

// maxFileBytes caps how much of tasks.json is read, so a corrupted or
// runaway file cannot exhaust memory.
const maxFileBytes = 16 << 20

type file struct {
	Tasks []Task `json:"tasks"`
}

// Store is the on-disk task registry. It is not safe for concurrent use within
// a process; across processes, Save serialises via an advisory lock.
type Store struct {
	path  string
	tasks []Task
}

// Load reads the registry, treating a missing file as an empty one so the
// first run needs no setup step.
func Load(path string) (*Store, error) {
	s := &Store{path: path}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}
	if info.Size() > maxFileBytes {
		return nil, fmt.Errorf("%s is %d bytes, above the %d byte limit", path, info.Size(), maxFileBytes)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return s, nil
	}

	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	s.tasks = f.Tasks
	return s, nil
}

// Path is where the registry is persisted.
func (s *Store) Path() string { return s.path }

// All returns every task, ordered by most recently opened first.
func (s *Store) All() []Task {
	out := make([]Task, len(s.tasks))
	copy(out, s.tasks)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].LastOpenedMs > out[j].LastOpenedMs
	})
	return out
}

// Get looks a task up by Jira key, case-insensitively.
func (s *Store) Get(key string) (Task, bool) {
	key = strings.ToUpper(key)
	for _, t := range s.tasks {
		if t.JiraKey == key {
			return t, true
		}
	}
	return Task{}, false
}

// Upsert inserts or replaces a task by Jira key.
func (s *Store) Upsert(t Task) {
	t.JiraKey = strings.ToUpper(t.JiraKey)
	for i, existing := range s.tasks {
		if existing.JiraKey == t.JiraKey {
			s.tasks[i] = t
			return
		}
	}
	s.tasks = append(s.tasks, t)
}

// Delete removes a task by Jira key and reports whether it existed.
func (s *Store) Delete(key string) bool {
	key = strings.ToUpper(key)
	for i, t := range s.tasks {
		if t.JiraKey == key {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return true
		}
	}
	return false
}

// Touch stamps a task as opened now.
func (s *Store) Touch(key string) {
	key = strings.ToUpper(key)
	for i, t := range s.tasks {
		if t.JiraKey == key {
			s.tasks[i].LastOpenedMs = nowMs()
			return
		}
	}
}

// UsedPorts returns every hexer and DB port already claimed by a task, so a new
// environment can pick a free pair.
func (s *Store) UsedPorts() (hexer, db map[int]bool) {
	hexer, db = map[int]bool{}, map[int]bool{}
	for _, t := range s.tasks {
		if t.Hexer.Port != 0 {
			hexer[t.Hexer.Port] = true
		}
		if t.Hexer.DBPort != 0 {
			db[t.Hexer.DBPort] = true
		}
	}
	return hexer, db
}

// Save writes the registry atomically. ~/.config/hlp is a git repo that is
// auto-committed after every command, so a torn write would be committed and
// pushed: hence both the exclusive lock and the write-temp-then-rename.
func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	unlock, err := lockFile(s.path + ".lock")
	if err != nil {
		return err
	}
	defer unlock()

	sorted := make([]Task, len(s.tasks))
	copy(sorted, s.tasks)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].JiraKey < sorted[j].JiraKey })

	data, err := json.MarshalIndent(file{Tasks: sorted}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode tasks: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".tasks-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("failed to replace %s: %w", s.path, err)
	}
	return nil
}

// lockFile takes an advisory exclusive lock, blocking until it is available.
func lockFile(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file %s: %w", path, err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to lock %s: %w", path, err)
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}
