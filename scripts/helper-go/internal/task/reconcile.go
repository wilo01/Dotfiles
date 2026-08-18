package task

import (
	"os"
	"path/filepath"
	"sort"
)

// DriftKind classifies a disagreement between the registry and the filesystem.
type DriftKind string

const (
	// DriftTaskRootMissing means the whole task directory is gone.
	DriftTaskRootMissing DriftKind = "task-root-missing"
	// DriftWorktreeMissing means a registered repo has no worktree on disk.
	DriftWorktreeMissing DriftKind = "worktree-missing"
	// DriftSourceMissing means the main checkout a worktree was created from is gone.
	DriftSourceMissing DriftKind = "source-missing"
	// DriftUntracked means a worktree exists under the task root that the registry
	// does not know about.
	DriftUntracked DriftKind = "untracked-worktree"
)

// Drift is one discrepancy found by Reconcile.
type Drift struct {
	Kind DriftKind
	Repo string
	Path string
}

// Reconcile compares a task against the filesystem and reports every
// discrepancy. It never repairs anything: the registry and the disk disagreeing
// is a fact the user needs to see, not something to silently paper over.
func Reconcile(t Task) []Drift {
	var drift []Drift

	if !dirExists(t.TaskRoot) {
		drift = append(drift, Drift{Kind: DriftTaskRootMissing, Path: t.TaskRoot})
	}

	tracked := map[string]bool{}
	for _, r := range t.Repos {
		tracked[r.Name] = true
		if !dirExists(r.Worktree) {
			drift = append(drift, Drift{Kind: DriftWorktreeMissing, Repo: r.Name, Path: r.Worktree})
		}
		if !dirExists(r.Source) {
			drift = append(drift, Drift{Kind: DriftSourceMissing, Repo: r.Name, Path: r.Source})
		}
	}

	entries, err := os.ReadDir(t.TaskRoot)
	if err != nil {
		return drift
	}
	var untracked []Drift
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || name[0] == '.' || tracked[name] {
			continue
		}
		path := filepath.Join(t.TaskRoot, name)
		if !isWorktree(path) {
			continue
		}
		untracked = append(untracked, Drift{Kind: DriftUntracked, Repo: name, Path: path})
	}
	sort.Slice(untracked, func(i, j int) bool { return untracked[i].Repo < untracked[j].Repo })

	return append(drift, untracked...)
}

func dirExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// isWorktree reports whether a directory is a git repo or linked worktree, by
// stat-ing its .git entry. Deliberately not `git rev-parse`, which walks up and
// would classify any directory inside an enclosing repo as a repo.
func isWorktree(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}
