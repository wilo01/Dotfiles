package task

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeWorktree creates a directory that isWorktree accepts.
func fakeWorktree(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", path, err)
	}
	if err := os.WriteFile(filepath.Join(path, ".git"), []byte("gitdir: /elsewhere\n"), 0o644); err != nil {
		t.Fatalf("write .git in %s: %v", path, err)
	}
}

func driftKinds(drift []Drift) map[DriftKind]string {
	byKind := map[DriftKind]string{}
	for _, d := range drift {
		byKind[d.Kind] = d.Repo
	}
	return byKind
}

func TestReconcileCleanTaskHasNoDrift(t *testing.T) {
	root := filepath.Join(t.TempDir(), "SUITE-1")
	source := filepath.Join(t.TempDir(), "tds-suite")
	fakeWorktree(t, filepath.Join(root, "tds-suite"))
	fakeWorktree(t, source)

	task, _ := New("SUITE-1", "x", root)
	task.AddRepo(Repo{Name: "tds-suite", Source: source, Worktree: filepath.Join(root, "tds-suite")})

	if drift := Reconcile(task); len(drift) != 0 {
		t.Fatalf("expected no drift, got %+v", drift)
	}
}

func TestReconcileDetectsMissingWorktreeAndSource(t *testing.T) {
	root := filepath.Join(t.TempDir(), "SUITE-1")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	task, _ := New("SUITE-1", "x", root)
	task.AddRepo(Repo{Name: "tds-suite", Source: "/nope/tds-suite", Worktree: filepath.Join(root, "tds-suite")})

	byKind := driftKinds(Reconcile(task))
	if byKind[DriftWorktreeMissing] != "tds-suite" {
		t.Error("missing worktree not reported")
	}
	if byKind[DriftSourceMissing] != "tds-suite" {
		t.Error("missing source not reported")
	}
}

func TestReconcileDetectsMissingTaskRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "gone")
	task, _ := New("SUITE-1", "x", root)

	drift := Reconcile(task)
	if len(drift) != 1 {
		t.Fatalf("expected exactly one drift, got %+v", drift)
	}
	if drift[0].Kind != DriftTaskRootMissing || drift[0].Path != root {
		t.Fatalf("expected a task-root-missing drift for %s, got %+v", root, drift[0])
	}
}

func TestReconcileDetectsUntrackedWorktree(t *testing.T) {
	root := filepath.Join(t.TempDir(), "SUITE-1")
	source := filepath.Join(t.TempDir(), "tds-suite")
	fakeWorktree(t, filepath.Join(root, "tds-suite"))
	fakeWorktree(t, filepath.Join(root, "tds-hexer"))
	fakeWorktree(t, source)

	task, _ := New("SUITE-1", "x", root)
	task.AddRepo(Repo{Name: "tds-suite", Source: source, Worktree: filepath.Join(root, "tds-suite")})

	drift := Reconcile(task)
	if len(drift) != 1 || drift[0].Kind != DriftUntracked || drift[0].Repo != "tds-hexer" {
		t.Fatalf("expected one untracked tds-hexer drift, got %+v", drift)
	}
}

func TestReconcileIgnoresNonRepoDirs(t *testing.T) {
	root := filepath.Join(t.TempDir(), "SUITE-1")
	source := filepath.Join(t.TempDir(), "tds-suite")
	fakeWorktree(t, filepath.Join(root, "tds-suite"))
	fakeWorktree(t, source)
	// The seeded .claude dir and a plain notes dir are not worktrees.
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}

	task, _ := New("SUITE-1", "x", root)
	task.AddRepo(Repo{Name: "tds-suite", Source: source, Worktree: filepath.Join(root, "tds-suite")})

	if drift := Reconcile(task); len(drift) != 0 {
		t.Fatalf("non-repo dirs reported as drift: %+v", drift)
	}
}
