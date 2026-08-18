package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileIsEmpty(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "tasks.json"))
	if err != nil {
		t.Fatalf("Load on a missing file returned an error: %v", err)
	}
	if got := len(s.All()); got != 0 {
		t.Fatalf("expected 0 tasks, got %d", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	task, err := New("suite-8602", "Update jspdf.js", "/tasks/SUITE-8602")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	task.AddRepo(Repo{Name: "tds-suite", Source: "/branches/tds-suite", Worktree: "/tasks/SUITE-8602/tds-suite", Base: "master"})
	task.Hexer = Hexer{Enabled: true, Modules: []string{"safe"}, Base: "master", Port: 3104, DBPort: 1535, Host: "suite-8602.acrid.dev"}
	s.Upsert(task)

	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, ok := reloaded.Get("SUITE-8602")
	if !ok {
		t.Fatal("task missing after reload")
	}
	if got.JiraKey != "SUITE-8602" {
		t.Errorf("key not upper-cased on Upsert: got %q", got.JiraKey)
	}
	if got.ClaudeSessionID != task.ClaudeSessionID {
		t.Errorf("session id changed: %q vs %q", got.ClaudeSessionID, task.ClaudeSessionID)
	}
	if len(got.Repos) != 1 || got.Repos[0].Name != "tds-suite" {
		t.Errorf("repos not round-tripped: %+v", got.Repos)
	}
	if got.Hexer.Port != 3104 || got.Hexer.DBPort != 1535 {
		t.Errorf("hexer ports not round-tripped: %+v", got.Hexer)
	}
}

func TestGetIsCaseInsensitive(t *testing.T) {
	s := &Store{}
	task, _ := New("SUITE-1", "x", "/tasks/SUITE-1")
	s.Upsert(task)

	for _, key := range []string{"SUITE-1", "suite-1", "Suite-1"} {
		if _, ok := s.Get(key); !ok {
			t.Errorf("Get(%q) missed the task", key)
		}
	}
}

func TestUpsertReplacesRatherThanDuplicates(t *testing.T) {
	s := &Store{}
	first, _ := New("SUITE-1", "old summary", "/tasks/SUITE-1")
	s.Upsert(first)

	second := first
	second.Summary = "new summary"
	s.Upsert(second)

	if got := len(s.All()); got != 1 {
		t.Fatalf("expected 1 task after re-upsert, got %d", got)
	}
	got, _ := s.Get("SUITE-1")
	if got.Summary != "new summary" {
		t.Errorf("summary not replaced: %q", got.Summary)
	}
}

func TestDeleteReportsPresence(t *testing.T) {
	s := &Store{}
	task, _ := New("SUITE-1", "x", "/tasks/SUITE-1")
	s.Upsert(task)

	if !s.Delete("suite-1") {
		t.Error("Delete returned false for an existing task")
	}
	if s.Delete("suite-1") {
		t.Error("Delete returned true for an already-deleted task")
	}
}

func TestUsedPortsSkipsZero(t *testing.T) {
	s := &Store{}
	withEnv, _ := New("SUITE-1", "x", "/a")
	withEnv.Hexer = Hexer{Enabled: true, Port: 3104, DBPort: 1535}
	noEnv, _ := New("SUITE-2", "x", "/b")
	s.Upsert(withEnv)
	s.Upsert(noEnv)

	hexer, db := s.UsedPorts()
	if !hexer[3104] || !db[1535] {
		t.Errorf("claimed ports missing: hexer=%v db=%v", hexer, db)
	}
	if hexer[0] || db[0] {
		t.Error("unset ports were reported as claimed")
	}
}

func TestSaveIsAtomicallyReplaced(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	s, _ := Load(path)
	task, _ := New("SUITE-1", "x", "/a")
	s.Upsert(task)
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if name := e.Name(); name != "tasks.json" && name != "tasks.json.lock" {
			t.Errorf("Save left a stray file behind: %s", name)
		}
	}
}

func TestAddAndRemoveRepo(t *testing.T) {
	task, _ := New("SUITE-1", "x", "/a")
	task.AddRepo(Repo{Name: "tds-suite", Base: "master"})
	task.AddRepo(Repo{Name: "tds-hexer", Base: "master"})
	task.AddRepo(Repo{Name: "tds-suite", Base: "main"})

	if len(task.Repos) != 2 {
		t.Fatalf("re-adding a repo duplicated it: %+v", task.RepoNames())
	}
	if r, _ := task.Repo("tds-suite"); r.Base != "main" {
		t.Errorf("re-adding did not replace the entry: base=%q", r.Base)
	}
	if !task.RemoveRepo("tds-hexer") || task.RemoveRepo("tds-hexer") {
		t.Error("RemoveRepo did not report presence correctly")
	}
}

func TestPrimaryRepoIsSelectionOrder(t *testing.T) {
	task, _ := New("SUITE-1", "x", "/a")
	if _, ok := task.PrimaryRepo(); ok {
		t.Error("PrimaryRepo reported a repo on an empty task")
	}
	task.AddRepo(Repo{Name: "zeta"})
	task.AddRepo(Repo{Name: "alpha"})

	primary, ok := task.PrimaryRepo()
	if !ok || primary.Name != "zeta" {
		t.Errorf("expected the first selected repo, got %+v", primary)
	}
	if names := task.RepoNames(); names[0] != "alpha" {
		t.Errorf("RepoNames should be sorted, got %v", names)
	}
}

func TestNewUUIDIsWellFormedAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for range 100 {
		id, err := newUUID()
		if err != nil {
			t.Fatalf("newUUID: %v", err)
		}
		if len(id) != 36 || id[14] != '4' {
			t.Fatalf("malformed v4 uuid: %q", id)
		}
		if seen[id] {
			t.Fatalf("duplicate uuid: %q", id)
		}
		seen[id] = true
	}
}
