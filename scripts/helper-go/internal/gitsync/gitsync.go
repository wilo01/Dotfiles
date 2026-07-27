package gitsync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/ui"
)

const commitMessage = "Updating worklogs"

const pushTimeout = 30 * time.Second

// worklogFiles are the only files this package will ever stage. The config
// repo also holds worklogs-bkp.csv, *.bak and *.pre-restore-* snapshots that
// must never be swept into an automatic commit, so no globs.
var worklogFiles = []string{"worklogs.csv", "worklogs-local.csv"}

// SyncWorklogs commits and pushes worklog CSV changes in configDir.
// Best-effort: prints a warning to stderr on failure, never returns an error,
// and never affects the exit code of the command that triggered it.
func SyncWorklogs(configDir string) {
	committed, err := CommitAndPushWorklogs(configDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Warning("Worklog git sync: "+err.Error()))
		return
	}
	if committed {
		fmt.Println(ui.Muted.Render("✓ worklogs synced to git"))
	}
}

// CommitAndPushWorklogs is the testable core. committed reports whether a
// commit was created; err is non-nil when commit or push failed (a failed
// push still leaves committed == true — the commit exists locally and the
// next successful push delivers it).
func CommitAndPushWorklogs(dir string) (committed bool, err error) {
	if _, statErr := os.Stat(filepath.Join(dir, ".git")); statErr != nil {
		return false, nil
	}

	statusArgs := append([]string{"status", "--porcelain", "--"}, worklogFiles...)
	status, err := gitRun(dir, statusArgs...)
	if err != nil {
		return false, err
	}

	files := dirtyWorklogFiles(status)
	if len(files) == 0 {
		return false, nil
	}

	if _, err := gitRun(dir, append([]string{"add", "--"}, files...)...); err != nil {
		return false, err
	}
	if _, err := gitRun(dir, "commit", "-m", commitMessage); err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), pushTimeout)
	defer cancel()
	pushCmd := exec.CommandContext(ctx, "git", "-C", dir, "push")
	if out, err := pushCmd.CombinedOutput(); err != nil {
		return true, fmt.Errorf("committed locally but push failed (will retry on next run): %s", firstLine(out))
	}

	return true, nil
}

// dirtyWorklogFiles parses `git status --porcelain` output and returns the
// file names that should be staged for the auto-commit.
//
// Porcelain format, one line per file:
//
//	XY <path>          e.g. " M worklogs.csv"
//	?? <path>          untracked (a freshly created worklogs-local.csv)
//	XY <old> -> <new>  renames (won't occur with our fixed pathspec)
//
// The status pathspec already limits output to worklogFiles, but this is the
// safety-critical filter for what gets committed automatically, so it must
// only ever return exact names from worklogFiles — defence in depth against
// a future pathspec change accidentally sweeping backup files into commits.
func dirtyWorklogFiles(porcelain string) []string {
	var files []string
	for line := range strings.SplitSeq(porcelain, "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSuffix(line[3:], "\r")
		if slices.Contains(worklogFiles, path) {
			files = append(files, path)
		}
	}
	return files
}

func gitRun(dir string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", fullArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s", args[0], firstLine(out))
	}
	return string(out), nil
}

func firstLine(out []byte) string {
	s := strings.TrimSpace(string(out))
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
