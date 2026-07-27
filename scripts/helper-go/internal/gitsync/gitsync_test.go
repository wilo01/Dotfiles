package gitsync

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
}

func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	fullArgs := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", fullArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "user.email", "test@test.local")
	mustGit(t, dir, "config", "user.name", "Test")
	mustGit(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func lastSubject(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(mustGit(t, dir, "log", "-1", "--format=%s"))
}

func TestDirtyWorklogFiles(t *testing.T) {
	cases := []struct {
		name      string
		porcelain string
		want      []string
	}{
		{"empty", "", nil},
		{"modified tracked", " M worklogs.csv\n", []string{"worklogs.csv"}},
		{"untracked local", "?? worklogs-local.csv\n", []string{"worklogs-local.csv"}},
		{"both", " M worklogs.csv\n?? worklogs-local.csv\n", []string{"worklogs.csv", "worklogs-local.csv"}},
		{"staged and modified", "MM worklogs.csv\n", []string{"worklogs.csv"}},
		{"backup file never staged", " M worklogs-bkp.csv\n?? worklogs.csv.bak\n", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dirtyWorklogFiles(tc.porcelain)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("dirtyWorklogFiles(%q) = %v, want %v", tc.porcelain, got, tc.want)
			}
		})
	}
}

func TestNotARepo(t *testing.T) {
	requireGit(t)
	committed, err := CommitAndPushWorklogs(t.TempDir())
	if committed || err != nil {
		t.Errorf("got (%v, %v), want (false, nil)", committed, err)
	}
}

func TestCleanRepo(t *testing.T) {
	requireGit(t)
	dir := initRepo(t)
	writeFile(t, dir, "worklogs.csv", "date,ticket,time\n")
	mustGit(t, dir, "add", "worklogs.csv")
	mustGit(t, dir, "commit", "-m", "initial")

	committed, err := CommitAndPushWorklogs(dir)
	if committed || err != nil {
		t.Errorf("got (%v, %v), want (false, nil)", committed, err)
	}
	if got := lastSubject(t, dir); got != "initial" {
		t.Errorf("unexpected new commit: %q", got)
	}
}

func TestDirtyNoRemote(t *testing.T) {
	requireGit(t)
	dir := initRepo(t)
	writeFile(t, dir, "worklogs.csv", "date,ticket,time\n")
	mustGit(t, dir, "add", "worklogs.csv")
	mustGit(t, dir, "commit", "-m", "initial")
	writeFile(t, dir, "worklogs.csv", "date,ticket,time\n2026-07-15,VIS-1,1h\n")

	committed, err := CommitAndPushWorklogs(dir)
	if !committed {
		t.Error("expected committed == true")
	}
	if err == nil {
		t.Error("expected push error (no remote configured)")
	}
	if got := lastSubject(t, dir); got != "Updating worklogs" {
		t.Errorf("commit subject = %q, want %q", got, "Updating worklogs")
	}
}

func TestUntrackedLocalCSV(t *testing.T) {
	requireGit(t)
	dir := initRepo(t)
	writeFile(t, dir, "worklogs.csv", "date,ticket,time\n")
	mustGit(t, dir, "add", "worklogs.csv")
	mustGit(t, dir, "commit", "-m", "initial")
	writeFile(t, dir, "worklogs-local.csv", "date,ticket,time\n")

	committed, _ := CommitAndPushWorklogs(dir)
	if !committed {
		t.Fatal("expected committed == true")
	}
	shown := mustGit(t, dir, "show", "--name-only", "--format=", "HEAD")
	if !strings.Contains(shown, "worklogs-local.csv") {
		t.Errorf("commit does not contain worklogs-local.csv:\n%s", shown)
	}
}

func TestBackupFilesExcluded(t *testing.T) {
	requireGit(t)
	dir := initRepo(t)
	writeFile(t, dir, "worklogs.csv", "date,ticket,time\n")
	writeFile(t, dir, "worklogs-bkp.csv", "old backup\n")
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial")
	writeFile(t, dir, "worklogs.csv", "date,ticket,time\n2026-07-15,VIS-1,1h\n")
	writeFile(t, dir, "worklogs-bkp.csv", "modified backup\n")

	committed, _ := CommitAndPushWorklogs(dir)
	if !committed {
		t.Fatal("expected committed == true")
	}
	shown := mustGit(t, dir, "show", "--name-only", "--format=", "HEAD")
	if strings.Contains(shown, "worklogs-bkp.csv") {
		t.Errorf("backup file swept into auto-commit:\n%s", shown)
	}
	status := mustGit(t, dir, "status", "--porcelain")
	if !strings.Contains(status, "worklogs-bkp.csv") {
		t.Error("worklogs-bkp.csv should remain dirty after sync")
	}
}

func TestPushSuccess(t *testing.T) {
	requireGit(t)
	dir := initRepo(t)
	bare := t.TempDir()
	mustGit(t, bare, "init", "--bare", "-b", "main")

	writeFile(t, dir, "worklogs.csv", "date,ticket,time\n")
	mustGit(t, dir, "add", "worklogs.csv")
	mustGit(t, dir, "commit", "-m", "initial")
	mustGit(t, dir, "remote", "add", "origin", bare)
	mustGit(t, dir, "push", "-u", "origin", "main")

	writeFile(t, dir, "worklogs.csv", "date,ticket,time\n2026-07-15,VIS-1,1h\n")
	committed, err := CommitAndPushWorklogs(dir)
	if !committed || err != nil {
		t.Fatalf("got (%v, %v), want (true, nil)", committed, err)
	}
	if got := lastSubject(t, bare); got != "Updating worklogs" {
		t.Errorf("remote HEAD subject = %q, want %q", got, "Updating worklogs")
	}
}
