package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s failed: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// newRemote builds a bare repo with one commit on the named default branch,
// standing in for origin.
func newRemote(t *testing.T, defaultBranch string) string {
	t.Helper()
	dir := t.TempDir()
	work := filepath.Join(dir, "work")
	bare := filepath.Join(dir, "remote.git")

	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, work, "init", "-q", "-b", defaultBranch)
	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, work, "add", ".")
	git(t, work, "commit", "-qm", "initial")
	git(t, work, "clone", "-q", "--bare", work, bare)
	return bare
}

// newClone clones the remote, optionally stripping origin/HEAD to reproduce the
// smtp-node case where the default branch cannot be read from symbolic-ref.
func newClone(t *testing.T, remote string, withOriginHead bool) string {
	t.Helper()
	dest := filepath.Join(t.TempDir(), "clone")
	out, err := exec.Command("git", "clone", "-q", remote, dest).CombinedOutput()
	if err != nil {
		t.Fatalf("clone failed: %v\n%s", err, out)
	}
	if !withOriginHead {
		exec.Command("git", "-C", dest, "symbolic-ref", "-d", "refs/remotes/origin/HEAD").Run()
	}
	return dest
}

func TestResolveDefaultBaseFromOriginHead(t *testing.T) {
	for _, branch := range []string{"master", "main"} {
		t.Run(branch, func(t *testing.T) {
			clone := newClone(t, newRemote(t, branch), true)

			got, err := resolveDefaultBase(clone)
			if err != nil {
				t.Fatalf("resolveDefaultBase: %v", err)
			}
			if got != branch {
				t.Errorf("expected %q, got %q", branch, got)
			}
		})
	}
}

// The clone has no origin/HEAD, exactly like smtp-node on this machine, which
// the old defaultBase silently answered "master" for.
func TestResolveDefaultBaseWithoutOriginHead(t *testing.T) {
	clone := newClone(t, newRemote(t, "main"), false)

	got, err := resolveDefaultBase(clone)
	if err != nil {
		t.Fatalf("resolveDefaultBase: %v", err)
	}
	if got != "main" {
		t.Errorf("expected the fallback chain to find main, got %q", got)
	}
}

func TestResolveDefaultBaseErrorsWithoutRemote(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "master")

	if _, err := resolveDefaultBase(dir); err == nil {
		t.Fatal("expected an error for a repo with no origin, got nil")
	}
}

func TestAddDetachedWorktreeLeavesNoBranch(t *testing.T) {
	clone := newClone(t, newRemote(t, "master"), true)
	worktree := filepath.Join(t.TempDir(), "SUITE-1", "repo")

	if err := addDetachedWorktree(clone, worktree, "master"); err != nil {
		t.Fatalf("addDetachedWorktree: %v", err)
	}

	head := worktreeHead(worktree)
	if head == "master" || head == "?" {
		t.Errorf("expected a detached HEAD, got %q", head)
	}
	branches := git(t, clone, "branch", "--list")
	if got := len(branches); got == 0 {
		t.Fatal("expected the source clone to still list its own branch")
	}
	if out := git(t, clone, "branch", "--list", "SUITE-1"); out != "" {
		t.Errorf("a branch was created when none was wanted: %q", out)
	}
}

// The whole reason for --detach: git refuses to check out a branch that is
// already checked out elsewhere, and the source clone always holds the default.
func TestAddDetachedWorktreeSucceedsWhileBaseIsCheckedOut(t *testing.T) {
	clone := newClone(t, newRemote(t, "master"), true)
	if head := worktreeHead(clone); head != "master" {
		t.Fatalf("precondition: clone should be on master, got %q", head)
	}

	first := filepath.Join(t.TempDir(), "SUITE-1", "repo")
	second := filepath.Join(t.TempDir(), "SUITE-2", "repo")
	if err := addDetachedWorktree(clone, first, "master"); err != nil {
		t.Fatalf("first worktree: %v", err)
	}
	if err := addDetachedWorktree(clone, second, "master"); err != nil {
		t.Fatalf("second worktree off the same base: %v", err)
	}
}

func TestRefreshSourcePullsLatest(t *testing.T) {
	remote := newRemote(t, "master")
	clone := newClone(t, remote, true)

	upstream := filepath.Join(t.TempDir(), "upstream")
	if out, err := exec.Command("git", "clone", "-q", remote, upstream).CombinedOutput(); err != nil {
		t.Fatalf("clone upstream: %v\n%s", err, out)
	}
	if err := os.WriteFile(filepath.Join(upstream, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, upstream, "add", ".")
	git(t, upstream, "commit", "-qm", "second")
	git(t, upstream, "push", "-q", "origin", "master")

	if _, err := refreshSource(clone, "master"); err != nil {
		t.Fatalf("refreshSource: %v", err)
	}
	if _, err := os.Stat(filepath.Join(clone, "new.txt")); err != nil {
		t.Errorf("refreshSource did not pull the new commit: %v", err)
	}

	worktree := filepath.Join(t.TempDir(), "SUITE-1", "repo")
	if err := addDetachedWorktree(clone, worktree, "master"); err != nil {
		t.Fatalf("addDetachedWorktree: %v", err)
	}
	if _, err := os.Stat(filepath.Join(worktree, "new.txt")); err != nil {
		t.Errorf("worktree was cut from a stale base: %v", err)
	}
}

// A dirty source clone must not block the worktree, because the worktree is cut
// from origin/<base>, not from the local checkout.
func TestRefreshSourceToleratesDirtyClone(t *testing.T) {
	clone := newClone(t, newRemote(t, "master"), true)
	git(t, clone, "checkout", "-q", "-b", "feature/wip")
	if err := os.WriteFile(filepath.Join(clone, "README.md"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := refreshSource(clone, "master"); err != nil {
		t.Fatalf("refreshSource should tolerate a dirty clone, got: %v", err)
	}
	worktree := filepath.Join(t.TempDir(), "SUITE-1", "repo")
	if err := addDetachedWorktree(clone, worktree, "master"); err != nil {
		t.Fatalf("addDetachedWorktree after a dirty refresh: %v", err)
	}
}

func TestDiscoverSourceReposSkipsNonRepos(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"tds-suite", "tds-hexer"} {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "init", "-q")
	}
	for _, name := range []string{"tasks", "smtp-env-tester", ".hidden"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	repos, err := discoverSourceRepos(root)
	if err != nil {
		t.Fatalf("discoverSourceRepos: %v", err)
	}
	if len(repos) != 2 || repos[0].Name != "tds-hexer" || repos[1].Name != "tds-suite" {
		t.Fatalf("expected the two repos in sorted order, got %+v", repos)
	}
}

// tasks/<KEY>/<repo> holds real worktrees; the tasks dir itself must never be
// offered as a source repo even once it contains them.
func TestDiscoverSourceReposSkipsTasksDir(t *testing.T) {
	root := t.TempDir()
	taskRepo := filepath.Join(root, "tasks", "SUITE-1", "tds-suite")
	if err := os.MkdirAll(taskRepo, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tasks", ".git"), []byte("gitdir: /x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repos, err := discoverSourceRepos(root)
	if err != nil {
		t.Fatalf("discoverSourceRepos: %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("tasks dir was offered as a repo: %+v", repos)
	}
}

func TestWorktreeDirtyAndRemove(t *testing.T) {
	clone := newClone(t, newRemote(t, "master"), true)
	worktree := filepath.Join(t.TempDir(), "SUITE-1", "repo")
	if err := addDetachedWorktree(clone, worktree, "master"); err != nil {
		t.Fatal(err)
	}

	if worktreeDirty(worktree) {
		t.Error("a fresh worktree reported dirty")
	}
	if err := os.WriteFile(filepath.Join(worktree, "scratch.txt"), []byte("wip\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !worktreeDirty(worktree) {
		t.Error("an edited worktree reported clean")
	}

	if err := removeWorktree(clone, worktree, false); err == nil {
		t.Error("removing a dirty worktree should be refused without force")
	}
	if err := removeWorktree(clone, worktree, true); err != nil {
		t.Errorf("forced removal failed: %v", err)
	}
	if _, err := os.Stat(worktree); !os.IsNotExist(err) {
		t.Error("worktree still on disk after a forced removal")
	}
}

func TestSeedWorktreeCopiesUntrackedConfig(t *testing.T) {
	source, worktree := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, ".env"), []byte("A=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, ".claude", "settings.local.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	seedWorktree(source, worktree)

	if data, err := os.ReadFile(filepath.Join(worktree, ".env")); err != nil || string(data) != "A=1\n" {
		t.Errorf(".env not seeded: %v %q", err, data)
	}
	if _, err := os.Stat(filepath.Join(worktree, ".claude", "settings.local.json")); err != nil {
		t.Errorf(".claude/settings.local.json not seeded: %v", err)
	}
	if _, err := os.Stat(filepath.Join(worktree, ".env.local")); !os.IsNotExist(err) {
		t.Error("a file absent from the source was invented in the worktree")
	}
}

func TestCloneURLTemplating(t *testing.T) {
	const template = "git@github.com:acreidentity/{{REPO}}.git"
	cases := map[string]string{
		"tds-suite":                          "git@github.com:acreidentity/tds-suite.git",
		"git@github.com:other/thing.git":     "git@github.com:other/thing.git",
		"https://github.com/other/thing.git": "https://github.com/other/thing.git",
	}
	for in, want := range cases {
		if got := cloneURL(in, template); got != want {
			t.Errorf("cloneURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRepoNameFromURL(t *testing.T) {
	cases := map[string]string{
		"git@github.com:acreidentity/tds-suite.git":     "tds-suite",
		"https://github.com/acreidentity/tds-hexer.git": "tds-hexer",
		"https://github.com/acreidentity/tds-hexer":     "tds-hexer",
	}
	for in, want := range cases {
		if got := repoNameFromURL(in); got != want {
			t.Errorf("repoNameFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}

// A stale index.lock makes git refuse to touch the clone. The worktree must
// still be created from the freshly fetched remote ref, and the reason must be
// reported rather than printed into a spinner line.
func TestRefreshSourceReportsStaleLockWithoutFailing(t *testing.T) {
	clone := newClone(t, newRemote(t, "master"), true)
	lock := filepath.Join(clone, ".git", "index.lock")
	if err := os.WriteFile(lock, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	warnings, err := refreshSource(clone, "master")
	if err != nil {
		t.Fatalf("a stale lock must not fail the run: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected one warning, got %v", warnings)
	}
	if !strings.Contains(warnings[0], "rm "+lock) {
		t.Errorf("warning should tell the user how to clear the lock: %q", warnings[0])
	}

	worktree := filepath.Join(t.TempDir(), "SUITE-1", "repo")
	if err := addDetachedWorktree(clone, worktree, "master"); err != nil {
		t.Fatalf("worktree must still be creatable: %v", err)
	}
}

// Warnings are returned, never printed, because the caller holds a spinner.
func TestRefreshSourceReturnsWarningsQuietly(t *testing.T) {
	clone := newClone(t, newRemote(t, "master"), true)
	// The feature branch changes README, and the working copy changes it again
	// without committing, so switching back to master would clobber the edit and
	// git refuses.
	git(t, clone, "checkout", "-q", "-b", "feature/wip")
	if err := os.WriteFile(filepath.Join(clone, "README.md"), []byte("on the branch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, clone, "commit", "-qam", "branch edit")
	if err := os.WriteFile(filepath.Join(clone, "README.md"), []byte("uncommitted\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	warnings, err := refreshSource(clone, "master")
	if err != nil {
		t.Fatalf("refreshSource: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("a refused checkout should be reported")
	}
	if !strings.Contains(warnings[0], "freshly fetched") {
		t.Errorf("warning should say the worktree is still correct: %q", warnings[0])
	}
}

func TestListBaseBranchesFiltersToLongLivedLines(t *testing.T) {
	remote := newRemote(t, "master")
	clone := newClone(t, remote, true)

	for _, branch := range []string{"maintenance/13.3AV", "maintenance/12.1AV", "feature/SUITE-1-something", "bugfix/VI-2-other"} {
		git(t, clone, "checkout", "-q", "-b", branch)
		git(t, clone, "push", "-q", "origin", branch)
	}
	git(t, clone, "checkout", "-q", "master")
	git(t, clone, "fetch", "-q", "origin", "--prune")

	got := listBaseBranches(clone, "master", []string{"maintenance/*", "release/*"})

	if len(got) == 0 || got[0] != "master" {
		t.Fatalf("the repo default must come first, got %v", got)
	}
	if !slices.Contains(got, "maintenance/13.3AV") || !slices.Contains(got, "maintenance/12.1AV") {
		t.Errorf("maintenance lines missing: %v", got)
	}
	for _, name := range got {
		if strings.HasPrefix(name, "feature/") || strings.HasPrefix(name, "bugfix/") {
			t.Errorf("a feature branch was offered as a base: %v", got)
		}
	}
}

// With no patterns configured, only the repo's own default is offered.
func TestListBaseBranchesWithoutPatterns(t *testing.T) {
	clone := newClone(t, newRemote(t, "main"), true)

	if got := listBaseBranches(clone, "main", nil); !reflect.DeepEqual(got, []string{"main"}) {
		t.Errorf("got %v, want [main]", got)
	}
}
