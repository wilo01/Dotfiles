package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/task"
	"github.com/dariuszw/hlp/internal/ui"
)

// workspace builds a repos root holding real clones, plus an empty tasks root,
// mirroring ~/tds-branch-opener/branches and .../branches/tasks.
func workspace(t *testing.T, repos ...string) *config.Config {
	t.Helper()
	root := t.TempDir()

	for _, name := range repos {
		remote := newRemote(t, "master")
		dest := filepath.Join(root, name)
		if err := runGit(t, "clone", "-q", remote, dest); err != nil {
			t.Fatalf("clone %s: %v", name, err)
		}
	}

	cfg := config.Default()
	cfg.Agent.ReposRoot = root
	cfg.Agent.TasksRoot = filepath.Join(root, "tasks")
	cfg.Agent.Hexer.Enabled = false
	return cfg
}

func runGit(t *testing.T, args ...string) error {
	t.Helper()
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		t.Logf("git %v: %s", args, out)
	}
	return err
}

func newStore(t *testing.T) *task.Store {
	t.Helper()
	s, err := task.Load(filepath.Join(t.TempDir(), "tasks.json"))
	if err != nil {
		t.Fatalf("task.Load: %v", err)
	}
	return s
}

// confirmWith drives the y/N prompts for the duration of a test.
func confirmWith(t *testing.T, answer string) {
	t.Helper()
	restore := ui.SetInput(strings.NewReader(strings.Repeat(answer+"\n", 10)))
	t.Cleanup(restore)
}

func TestApplyTaskCreatesNestedDetachedWorktrees(t *testing.T) {
	cfg := workspace(t, "tds-suite", "tds-hexer")
	store := newStore(t)

	got, err := applyTask(cfg, store, applyOptions{
		Key:      "SUITE-8602",
		Summary:  "Update jspdf.js",
		Selected: []string{"tds-suite", "tds-hexer"},
	})
	if err != nil {
		t.Fatalf("applyTask: %v", err)
	}

	wantRoot := filepath.Join(cfg.Agent.TasksRoot, "SUITE-8602")
	if got.TaskRoot != wantRoot {
		t.Errorf("task root = %q, want %q", got.TaskRoot, wantRoot)
	}

	// Every repo nests, including when there is only one, so adding a second
	// later never has to move the first.
	for _, name := range []string{"tds-suite", "tds-hexer"} {
		wt := filepath.Join(wantRoot, name)
		if _, err := os.Stat(wt); err != nil {
			t.Errorf("%s worktree missing at %s: %v", name, wt, err)
		}
		if head := worktreeHead(wt); !strings.HasPrefix(head, "detached") {
			t.Errorf("%s should be detached, got %q", name, head)
		}
	}
	if got.ClaudeSessionID == "" {
		t.Error("no claude session id minted")
	}
	if got.SessionStarted {
		t.Error("a fresh task must not be marked as started")
	}
}

func TestApplyTaskCreatesNoBranch(t *testing.T) {
	cfg := workspace(t, "tds-suite")
	store := newStore(t)

	if _, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}}); err != nil {
		t.Fatalf("applyTask: %v", err)
	}

	source := filepath.Join(cfg.Agent.ReposRoot, "tds-suite")
	branches := git(t, source, "branch", "--list")
	if strings.Contains(branches, "SUITE-1") {
		t.Errorf("a branch was created for the ticket: %q", branches)
	}
}

func TestApplyTaskIsIdempotent(t *testing.T) {
	cfg := workspace(t, "tds-suite")
	store := newStore(t)
	opts := applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}}

	first, err := applyTask(cfg, store, opts)
	if err != nil {
		t.Fatalf("first applyTask: %v", err)
	}
	second, err := applyTask(cfg, store, opts)
	if err != nil {
		t.Fatalf("second applyTask: %v", err)
	}

	if first.ClaudeSessionID != second.ClaudeSessionID {
		t.Error("re-running start minted a new claude session")
	}
	if len(second.Repos) != 1 {
		t.Errorf("repos duplicated: %+v", second.RepoNamesInOrder())
	}
}

func TestApplyTaskAddsARepoToAnExistingTask(t *testing.T) {
	cfg := workspace(t, "tds-suite", "tds-hexer")
	store := newStore(t)

	if _, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}}); err != nil {
		t.Fatal(err)
	}
	got, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite", "tds-hexer"}})
	if err != nil {
		t.Fatalf("applyTask: %v", err)
	}

	if want := []string{"tds-suite", "tds-hexer"}; !reflect.DeepEqual(got.RepoNamesInOrder(), want) {
		t.Errorf("repos = %v, want %v", got.RepoNamesInOrder(), want)
	}
	if _, err := os.Stat(filepath.Join(got.TaskRoot, "tds-hexer")); err != nil {
		t.Errorf("added worktree missing: %v", err)
	}
}

func TestApplyTaskRemovesADroppedRepo(t *testing.T) {
	cfg := workspace(t, "tds-suite", "tds-hexer")
	store := newStore(t)
	confirmWith(t, "y")

	created, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite", "tds-hexer"}})
	if err != nil {
		t.Fatal(err)
	}
	dropped := filepath.Join(created.TaskRoot, "tds-hexer")

	got, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}})
	if err != nil {
		t.Fatalf("applyTask: %v", err)
	}

	if got.HasRepo("tds-hexer") {
		t.Error("dropped repo still enlisted")
	}
	if _, err := os.Stat(dropped); !os.IsNotExist(err) {
		t.Errorf("dropped worktree still on disk: %v", err)
	}
	if _, err := os.Stat(filepath.Join(got.TaskRoot, "tds-suite")); err != nil {
		t.Errorf("kept worktree was removed: %v", err)
	}
}

// Removing a repo with uncommitted work must fail loudly rather than discard it.
func TestApplyTaskRefusesToRemoveDirtyWorktree(t *testing.T) {
	cfg := workspace(t, "tds-suite", "tds-hexer")
	store := newStore(t)
	confirmWith(t, "y")

	created, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite", "tds-hexer"}})
	if err != nil {
		t.Fatal(err)
	}
	dirty := filepath.Join(created.TaskRoot, "tds-hexer")
	if err := os.WriteFile(filepath.Join(dirty, "wip.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}})
	if err == nil {
		t.Fatal("expected removal of a dirty worktree to be refused")
	}
	if !strings.Contains(err.Error(), "force") {
		t.Errorf("error should point at --force, got: %v", err)
	}
	if _, statErr := os.Stat(dirty); statErr != nil {
		t.Errorf("dirty worktree was removed anyway: %v", statErr)
	}
}

func TestApplyTaskForceRemovesDirtyWorktree(t *testing.T) {
	cfg := workspace(t, "tds-suite", "tds-hexer")
	store := newStore(t)
	confirmWith(t, "y")

	created, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite", "tds-hexer"}})
	if err != nil {
		t.Fatal(err)
	}
	dirty := filepath.Join(created.TaskRoot, "tds-hexer")
	if err := os.WriteFile(filepath.Join(dirty, "wip.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}, Force: true}); err != nil {
		t.Fatalf("forced removal: %v", err)
	}
	if _, err := os.Stat(dirty); !os.IsNotExist(err) {
		t.Error("forced removal left the worktree behind")
	}
}

func TestApplyTaskDryRunTouchesNothing(t *testing.T) {
	cfg := workspace(t, "tds-suite")
	store := newStore(t)

	got, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}, DryRun: true})
	if err != nil {
		t.Fatalf("applyTask: %v", err)
	}

	if _, err := os.Stat(got.TaskRoot); !os.IsNotExist(err) {
		t.Error("dry-run created the task root")
	}
	if _, ok := store.Get("SUITE-1"); ok {
		t.Error("dry-run persisted the task")
	}
}

func TestApplyTaskRejectsUnknownRepo(t *testing.T) {
	cfg := workspace(t, "tds-suite")
	store := newStore(t)

	_, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite", "nope"}})
	if err == nil {
		t.Fatal("expected an error for an unknown repo")
	}
	// The whole run must fail rather than half-build the task.
	if _, statErr := os.Stat(filepath.Join(cfg.Agent.TasksRoot, "SUITE-1", "tds-suite")); statErr == nil {
		t.Error("a worktree was created despite the failure")
	}
}

func TestApplyTaskSelectionOrderSetsPrimary(t *testing.T) {
	cfg := workspace(t, "tds-suite", "tds-hexer")
	store := newStore(t)

	got, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-hexer", "tds-suite"}})
	if err != nil {
		t.Fatal(err)
	}

	primary, ok := got.PrimaryRepo()
	if !ok || primary.Name != "tds-hexer" {
		t.Errorf("primary = %+v, want tds-hexer", primary)
	}
}

// The end-to-end shape the user asked for: pick repos, get one tmux window per
// repo in the right worktree, then add and drop a repo and have the windows
// follow.
func TestEnsureSessionBuildsAndReconcilesWindows(t *testing.T) {
	isolatedTmux(t)
	cfg := workspace(t, "tds-suite", "tds-hexer", "tds-api-server")
	store := newStore(t)
	confirmWith(t, "y")

	created, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite", "tds-hexer"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureSession(cfg, store, created, launchOptions{Manual: true, NoAttach: true}); err != nil {
		t.Fatalf("ensureSession: %v", err)
	}

	if got := listWindowNames("SUITE-1"); !reflect.DeepEqual(got, []string{"tds-suite", "tds-hexer"}) {
		t.Fatalf("windows = %v, want one per repo in task order", got)
	}
	if got := windowPath("SUITE-1", "tds-hexer"); got != filepath.Join(created.TaskRoot, "tds-hexer") {
		t.Errorf("window cwd = %q", got)
	}

	// The launch flips SessionStarted, so a reopen resumes rather than recreates.
	reloaded, _ := store.Get("SUITE-1")
	if !reloaded.SessionStarted {
		t.Error("session not marked as started after launch")
	}

	added, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite", "tds-hexer", "tds-api-server"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureSession(cfg, store, added, launchOptions{Manual: true, NoAttach: true}); err != nil {
		t.Fatal(err)
	}
	if !windowExists("SUITE-1", "tds-api-server") {
		t.Error("window not added for the new repo")
	}

	dropped, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureSession(cfg, store, dropped, launchOptions{Manual: true, NoAttach: true}); err != nil {
		t.Fatal(err)
	}
	if got := listWindowNames("SUITE-1"); !reflect.DeepEqual(got, []string{"tds-suite"}) {
		t.Errorf("windows after drop = %v, want just tds-suite", got)
	}
}

func TestSeedTaskRootCopiesSharedConfig(t *testing.T) {
	cfg := workspace(t, "tds-suite")
	store := newStore(t)

	if err := os.WriteFile(filepath.Join(cfg.Agent.ReposRoot, "CLAUDE.md"), []byte("# rules\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	claudeDir := filepath.Join(cfg.Agent.ReposRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(got.TaskRoot, "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md not seeded into the task root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(got.TaskRoot, ".claude", "settings.local.json")); err != nil {
		t.Errorf(".claude not seeded into the task root: %v", err)
	}
}

// Attaching is a convenience layered on top of work that already succeeded, so
// a session that cannot be attached to must not fail the command.
func TestEnsureSessionSurvivesAFailedAttach(t *testing.T) {
	isolatedTmux(t)
	cfg := workspace(t, "tds-suite")
	store := newStore(t)

	created, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}})
	if err != nil {
		t.Fatal(err)
	}

	// Attach is requested, but the test has no terminal to hand over.
	if err := ensureSession(cfg, store, created, launchOptions{Manual: true}); err != nil {
		t.Fatalf("a failed attach must not fail the run: %v", err)
	}
	if !sessionExists("SUITE-1") {
		t.Error("session was not created")
	}
	if !windowExists("SUITE-1", "tds-suite") {
		t.Error("window was not created")
	}
}

// The point of per-repo bases: one ticket investigated across master in one
// repo and a maintenance line in another, simultaneously.
func TestApplyTaskUsesPerRepoBases(t *testing.T) {
	cfg := workspace(t, "tds-suite", "tds-hexer")
	store := newStore(t)

	// Give tds-hexer a maintenance line on its origin.
	hexerSource := filepath.Join(cfg.Agent.ReposRoot, "tds-hexer")
	git(t, hexerSource, "checkout", "-q", "-b", "maintenance/13.3AV")
	if err := os.WriteFile(filepath.Join(hexerSource, "release.txt"), []byte("13.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, hexerSource, "add", "-A")
	git(t, hexerSource, "commit", "-qm", "release line")
	git(t, hexerSource, "push", "-q", "origin", "maintenance/13.3AV")
	git(t, hexerSource, "checkout", "-q", "master")

	got, err := applyTask(cfg, store, applyOptions{
		Key:      "SUITE-1",
		Selected: []string{"tds-suite", "tds-hexer"},
		Bases:    map[string]string{"tds-hexer": "maintenance/13.3AV"},
	})
	if err != nil {
		t.Fatalf("applyTask: %v", err)
	}

	suite, _ := got.Repo("tds-suite")
	if suite.Base != "master" {
		t.Errorf("tds-suite base = %q, want master", suite.Base)
	}
	hexer, _ := got.Repo("tds-hexer")
	if hexer.Base != "maintenance/13.3AV" {
		t.Errorf("tds-hexer base = %q, want maintenance/13.3AV", hexer.Base)
	}
	if _, err := os.Stat(filepath.Join(hexer.Worktree, "release.txt")); err != nil {
		t.Errorf("tds-hexer worktree was not cut from the maintenance line: %v", err)
	}
	if _, err := os.Stat(filepath.Join(suite.Worktree, "release.txt")); !os.IsNotExist(err) {
		t.Error("tds-suite worktree picked up the maintenance line")
	}
}

// Re-adding a repo on a different base would silently keep the old checkout, so
// it is refused with the way out spelled in the message.
func TestApplyTaskRefusesToSilentlyChangeBase(t *testing.T) {
	cfg := workspace(t, "tds-suite")
	store := newStore(t)

	if _, err := applyTask(cfg, store, applyOptions{Key: "SUITE-1", Selected: []string{"tds-suite"}}); err != nil {
		t.Fatal(err)
	}
	_, err := applyTask(cfg, store, applyOptions{
		Key:      "SUITE-1",
		Selected: []string{"tds-suite"},
		Bases:    map[string]string{"tds-suite": "other"},
	})
	if err == nil {
		t.Fatal("expected changing the base of an existing worktree to be refused")
	}
	if !strings.Contains(err.Error(), "already checked out") {
		t.Errorf("unhelpful error: %v", err)
	}
}

// A failing environment must not strand the worktrees: they exist on disk, so
// the task has to be recorded or status cannot see them and rm cannot clean up.
func TestApplyTaskPersistsWhenHexerConfigFails(t *testing.T) {
	cfg := workspace(t, "tds-hexer")
	cfg.Agent.Hexer.Enabled = true
	// The env serves the tds-suite worktree, which this task does not enlist.
	cfg.Agent.Hexer.TDSRepo = "tds-suite"
	store := newStore(t)

	got, err := applyTask(cfg, store, applyOptions{
		Key:          "SUITE-1",
		Selected:     []string{"tds-hexer"},
		Hexer:        true,
		HexerModules: []string{"safe"},
	})
	if err == nil {
		t.Fatal("expected the misconfigured environment to surface")
	}
	if !strings.Contains(err.Error(), "worktrees are ready") {
		t.Errorf("error should say the worktrees survived: %v", err)
	}

	saved, ok := store.Get("SUITE-1")
	if !ok {
		t.Fatal("task was not recorded, its worktrees are now orphaned")
	}
	if !saved.HasRepo("tds-hexer") {
		t.Error("the created worktree is not recorded on the task")
	}
	if saved.Hexer.Enabled {
		t.Error("a rejected environment must not be recorded as enabled")
	}
	if _, statErr := os.Stat(got.TaskRoot); statErr != nil {
		t.Errorf("task root missing: %v", statErr)
	}
}

// Configuring an environment allocates its ports and host but must not spend
// minutes provisioning: that happens in the session's own window.
func TestApplyTaskConfiguresHexerWithoutProvisioning(t *testing.T) {
	cfg := workspace(t, "tds-suite")
	cfg.Agent.Hexer.Enabled = true
	cfg.Agent.Hexer.TDSRepo = "tds-suite"
	cfg.Agent.Hexer.HexerDir = filepath.Join(t.TempDir(), "absent")
	cfg.Agent.Hexer.TDSDir = filepath.Join(t.TempDir(), "absent")
	store := newStore(t)

	started := time.Now()
	got, err := applyTask(cfg, store, applyOptions{
		Key:          "SUITE-1",
		Selected:     []string{"tds-suite"},
		Hexer:        true,
		HexerModules: []string{"safe"},
	})
	if err != nil {
		t.Fatalf("applyTask: %v", err)
	}
	// A real provision starts an Oracle container; anything near that would mean
	// the work did not move to the background.
	if elapsed := time.Since(started); elapsed > 20*time.Second {
		t.Errorf("applyTask took %s — provisioning is still inline", elapsed)
	}

	if !got.Hexer.Enabled {
		t.Error("hexer not marked enabled")
	}
	if got.Hexer.Port < cfg.Agent.Hexer.HexerPortMin || got.Hexer.Port > cfg.Agent.Hexer.HexerPortMax {
		t.Errorf("hexer port %d outside the configured range", got.Hexer.Port)
	}
	if got.Hexer.DBPort == 0 {
		t.Error("no db port allocated")
	}
	if got.Hexer.Host != "suite-1."+cfg.Agent.Hexer.HostnameSuffix {
		t.Errorf("host = %q", got.Hexer.Host)
	}
	if got.Hexer.Base != "master" {
		t.Errorf("base = %q, want the tds-suite base", got.Hexer.Base)
	}
}

func TestHexerUpCommandIsRunnableAndQuoted(t *testing.T) {
	cfg := workspace(t, "tds-suite")
	cfg.Agent.Hexer.Enabled = true
	cfg.Agent.Hexer.TDSRepo = "tds-suite"
	store := newStore(t)

	got, err := applyTask(cfg, store, applyOptions{
		Key: "SUITE-1", Selected: []string{"tds-suite"},
		Hexer: true, HexerModules: []string{"safe"},
	})
	if err != nil {
		t.Fatal(err)
	}

	command, err := hexerUpCommand(cfg, got)
	if err != nil {
		t.Fatalf("hexerUpCommand: %v", err)
	}
	for _, want := range []string{"hexer-task.sh", "up suite-1", "--hexer-port", "--module safe:safe:"} {
		if !strings.Contains(command, want) {
			t.Errorf("command missing %q:\n%s", want, command)
		}
	}
	if !strings.Contains(command, "HEXER_TASK_HEXER_DIR") {
		t.Error("machine paths are not passed through the environment")
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"plain":           "plain",
		"/no/spaces/here": "/no/spaces/here",
		"":                "''",
		"has space":       "'has space'",
		"semi;colon":      "'semi;colon'",
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %s, want %s", in, got, want)
		}
	}
	if got := shellQuote("it's"); got != `'it'\''s'` {
		t.Errorf("shellQuote(\"it's\") = %s", got)
	}
}
