package hexer

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseProgress(t *testing.T) {
	cases := []struct {
		line string
		step string
		pct  int
		msg  string
		ok   bool
	}{
		{"step=clone pct=10 msg=cloning tds-suite", "clone", 10, "cloning tds-suite", true},
		{"step=db pct=100 msg=", "db", 100, "", true},
		{"step=db pct=55", "db", 55, "", true},
		{"▶ some human line", "", 0, "", false},
		{"step=db pct=notanumber msg=x", "", 0, "", false},
		{"", "", 0, "", false},
	}
	for _, c := range cases {
		step, pct, msg, ok := parseProgress(c.line)
		if ok != c.ok || step != c.step || pct != c.pct || msg != c.msg {
			t.Errorf("parseProgress(%q) = (%q,%d,%q,%v), want (%q,%d,%q,%v)",
				c.line, step, pct, msg, ok, c.step, c.pct, c.msg, c.ok)
		}
	}
}

// A message containing "pct=" or spaces must survive intact, since it is free
// text written for a human.
func TestParseProgressKeepsMessageIntact(t *testing.T) {
	_, _, msg, ok := parseProgress("step=x pct=1 msg=waiting for db pct=high, be patient")
	if !ok {
		t.Fatal("line not parsed")
	}
	if msg != "waiting for db pct=high, be patient" {
		t.Errorf("message mangled: %q", msg)
	}
}

func TestEnsureScriptWritesAndIsIdempotent(t *testing.T) {
	dir := t.TempDir()

	path, err := EnsureScript(dir)
	if err != nil {
		t.Fatalf("EnsureScript: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("script is not executable: %v", info.Mode())
	}

	first, _ := os.ReadFile(path)
	if len(first) == 0 {
		t.Fatal("embedded script is empty")
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureScript(dir); err != nil {
		t.Fatalf("second EnsureScript: %v", err)
	}
}

// A hlp upgrade ships a new script; a stale copy on disk must be replaced.
func TestEnsureScriptReplacesStaleCopy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hexer-task.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho stale\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureScript(dir); err != nil {
		t.Fatalf("EnsureScript: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) == "#!/bin/sh\necho stale\n" {
		t.Error("stale script was not replaced")
	}
	if len(got) != len(script) {
		t.Errorf("materialized script is %d bytes, embedded is %d", len(got), len(script))
	}
}

func TestAllocatePortSkipsClaimed(t *testing.T) {
	used := map[int]bool{3100: true, 3101: true}

	port, err := AllocatePort(3100, 3199, used)
	if err != nil {
		t.Fatalf("AllocatePort: %v", err)
	}
	if port != 3102 {
		t.Errorf("expected 3102, got %d", port)
	}
}

func TestAllocatePortExhausted(t *testing.T) {
	used := map[int]bool{3100: true, 3101: true}

	if _, err := AllocatePort(3100, 3101, used); err == nil {
		t.Error("expected an error when the range is exhausted")
	}
}

func TestHostname(t *testing.T) {
	if got := Hostname("SUITE-8377", "acrid.dev"); got != "suite-8377.acrid.dev" {
		t.Errorf("Hostname = %q", got)
	}
}

func TestEffectiveModulesFallsBackToDefaults(t *testing.T) {
	var cfg Config
	if got := cfg.ModuleNames(); !reflect.DeepEqual(got, []string{"safe", "kiosk"}) {
		t.Errorf("default module names = %v", got)
	}

	cfg.Modules = []Module{{Name: "only", Route: "r", Static: "s", RT: "t"}}
	if got := cfg.ModuleNames(); !reflect.DeepEqual(got, []string{"only"}) {
		t.Errorf("configured modules ignored: %v", got)
	}
}

func TestModulesByName(t *testing.T) {
	var cfg Config

	mods, err := cfg.ModulesByName([]string{"kiosk"})
	if err != nil {
		t.Fatalf("ModulesByName: %v", err)
	}
	if len(mods) != 1 || mods[0].Static != "source/ui-kiosk" {
		t.Errorf("wrong module resolved: %+v", mods)
	}
	if got := mods[0].arg(); got != "kiosk:kiosk:source/ui-kiosk:source/server/rtkiosk" {
		t.Errorf("module arg = %q", got)
	}

	if _, err := cfg.ModulesByName([]string{"nope"}); err == nil {
		t.Error("expected an error for an unknown app name")
	}
}
