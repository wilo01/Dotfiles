package agent

import (
	"os/exec"
	"reflect"
	"testing"
)

// isolatedTmux points every tmux call in the test at a throwaway server. Fedora's
// tmux 3.7 can crash the whole server, which would take the user's real sessions
// with it, so tests must never touch the default socket.
func isolatedTmux(t *testing.T) string {
	t.Helper()
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}
	socket := "hlptest-" + t.Name()
	t.Setenv("HLP_TMUX_SOCKET", socket)
	t.Cleanup(func() {
		exec.Command("tmux", "-L", socket, "kill-server").Run()
	})
	return socket
}

func TestNewSessionWithWindowNamesTheFirstWindow(t *testing.T) {
	isolatedTmux(t)
	dir := t.TempDir()

	if err := newSessionWithWindow("SUITE-1", "tds-suite", dir); err != nil {
		t.Fatalf("newSessionWithWindow: %v", err)
	}

	if got := listWindowNames("SUITE-1"); !reflect.DeepEqual(got, []string{"tds-suite"}) {
		t.Errorf("expected one window named tds-suite, got %v", got)
	}
	if got := windowPath("SUITE-1", "tds-suite"); got != dir {
		t.Errorf("window cwd = %q, want %q", got, dir)
	}
}

func TestNewAndKillWindow(t *testing.T) {
	isolatedTmux(t)
	first, second := t.TempDir(), t.TempDir()

	if err := newSessionWithWindow("SUITE-1", "tds-suite", first); err != nil {
		t.Fatal(err)
	}
	if err := newWindow("SUITE-1", "tds-hexer", second); err != nil {
		t.Fatalf("newWindow: %v", err)
	}

	if !windowExists("SUITE-1", "tds-hexer") {
		t.Fatal("added window not found")
	}
	if got := windowPath("SUITE-1", "tds-hexer"); got != second {
		t.Errorf("added window cwd = %q, want %q", got, second)
	}

	if err := killWindow("SUITE-1", "tds-hexer"); err != nil {
		t.Fatalf("killWindow: %v", err)
	}
	if windowExists("SUITE-1", "tds-hexer") {
		t.Error("window still present after kill")
	}
}

func TestSyncWindowsAddsAndRemoves(t *testing.T) {
	isolatedTmux(t)
	suite, hexer, api := t.TempDir(), t.TempDir(), t.TempDir()

	if err := newSessionWithWindow("SUITE-1", "tds-suite", suite); err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{"tds-suite": true, "tds-hexer": true, "tds-api-server": true}

	want := []windowSpec{{"tds-suite", suite}, {"tds-hexer", hexer}, {"tds-api-server", api}}
	if err := syncWindows("SUITE-1", want, known); err != nil {
		t.Fatalf("syncWindows add: %v", err)
	}
	for _, name := range []string{"tds-suite", "tds-hexer", "tds-api-server"} {
		if !windowExists("SUITE-1", name) {
			t.Errorf("window %s was not created", name)
		}
	}

	want = []windowSpec{{"tds-suite", suite}}
	if err := syncWindows("SUITE-1", want, known); err != nil {
		t.Fatalf("syncWindows remove: %v", err)
	}
	if got := listWindowNames("SUITE-1"); !reflect.DeepEqual(got, []string{"tds-suite"}) {
		t.Errorf("expected only tds-suite to survive, got %v", got)
	}
}

// A window the user split off by hand is not a repo window and must survive a
// repo being dropped from the task.
func TestSyncWindowsLeavesUnknownWindowsAlone(t *testing.T) {
	isolatedTmux(t)
	suite := t.TempDir()

	if err := newSessionWithWindow("SUITE-1", "tds-suite", suite); err != nil {
		t.Fatal(err)
	}
	if err := newWindow("SUITE-1", "notes", t.TempDir()); err != nil {
		t.Fatal(err)
	}

	known := map[string]bool{"tds-suite": true}
	if err := syncWindows("SUITE-1", nil, known); err != nil {
		t.Fatalf("syncWindows: %v", err)
	}

	if windowExists("SUITE-1", "tds-suite") {
		t.Error("a known repo window should have been removed")
	}
	if !windowExists("SUITE-1", "notes") {
		t.Error("a hand-made window was removed")
	}
}

func TestSyncWindowsIsIdempotent(t *testing.T) {
	isolatedTmux(t)
	suite, hexer := t.TempDir(), t.TempDir()

	if err := newSessionWithWindow("SUITE-1", "tds-suite", suite); err != nil {
		t.Fatal(err)
	}
	want := []windowSpec{{"tds-suite", suite}, {"tds-hexer", hexer}}
	known := map[string]bool{"tds-suite": true, "tds-hexer": true}

	for range 3 {
		if err := syncWindows("SUITE-1", want, known); err != nil {
			t.Fatalf("syncWindows: %v", err)
		}
	}
	if got := listWindowNames("SUITE-1"); len(got) != 2 {
		t.Errorf("repeated syncs changed the window set: %v", got)
	}
}
