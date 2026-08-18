package repopicker

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	if s == " " {
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// press feeds keys through Update and returns the resulting model.
func press(m Model, keys ...string) Model {
	for _, k := range keys {
		next, _ := m.Update(key(k))
		m = next.(Model)
	}
	return m
}

// cursorTo moves the cursor onto the named repo row.
func cursorTo(t *testing.T, m Model, name string) Model {
	t.Helper()
	for i, r := range m.rows {
		if r.kind == rowRepo && r.name == name {
			m.cursor = i
			return m
		}
	}
	t.Fatalf("no repo row named %q", name)
	return m
}

func baseOpts() Options {
	return Options{
		Title: "EDIT REPOS · SUITE-8602",
		Repos: []Repo{
			{Name: "tds-suite", InTask: true},
			{Name: "tds-hexer"},
			{Name: "tds-api-server"},
		},
		Base:           "master",
		HexerAvailable: true,
		HexerModules:   []string{"safe", "kiosk"},
	}
}

func TestInTaskReposArePreCheckedAndFirst(t *testing.T) {
	m := New(baseOpts())

	if m.rows[0].kind != rowRepo || m.rows[0].name != "tds-suite" {
		t.Fatalf("in-task repo is not first: %+v", m.rows[0])
	}
	if !m.rows[0].checked || !m.rows[0].inTask {
		t.Error("in-task repo is not pre-checked")
	}
	if m.rows[1].name != "tds-api-server" || m.rows[2].name != "tds-hexer" {
		t.Errorf("remaining repos not sorted: %q %q", m.rows[1].name, m.rows[2].name)
	}
}

func TestConfirmWithNoChangesIsANoOpDiff(t *testing.T) {
	m := press(New(baseOpts()), "enter")

	if !m.result.Confirmed {
		t.Fatal("result not confirmed")
	}
	if len(m.result.Added) != 0 || len(m.result.Removed) != 0 {
		t.Errorf("expected an empty diff, got added=%v removed=%v", m.result.Added, m.result.Removed)
	}
	if !reflect.DeepEqual(m.result.Selected, []string{"tds-suite"}) {
		t.Errorf("selection changed: %v", m.result.Selected)
	}
	if m.result.Base != "master" {
		t.Errorf("base not carried through: %q", m.result.Base)
	}
}

func TestCheckingARepoAddsIt(t *testing.T) {
	m := cursorTo(t, New(baseOpts()), "tds-hexer")
	m = press(m, " ", "enter")

	if !reflect.DeepEqual(m.result.Added, []string{"tds-hexer"}) {
		t.Errorf("expected tds-hexer added, got %v", m.result.Added)
	}
	if len(m.result.Removed) != 0 {
		t.Errorf("unexpected removals: %v", m.result.Removed)
	}
}

func TestUncheckingAnInTaskRepoRemovesIt(t *testing.T) {
	m := cursorTo(t, New(baseOpts()), "tds-suite")
	m = press(m, " ", "enter")

	if !reflect.DeepEqual(m.result.Removed, []string{"tds-suite"}) {
		t.Errorf("expected tds-suite removed, got %v", m.result.Removed)
	}
	if len(m.result.Selected) != 0 {
		t.Errorf("expected an empty selection, got %v", m.result.Selected)
	}
}

// The first-selected repo becomes the task's primary, so selection order has to
// survive rather than collapsing to alphabetical.
func TestSelectionOrderIsPreserved(t *testing.T) {
	opts := baseOpts()
	opts.Repos = []Repo{{Name: "tds-suite"}, {Name: "tds-hexer"}, {Name: "tds-api-server"}}
	m := New(opts)

	m = cursorTo(t, m, "tds-hexer")
	m = press(m, " ")
	m = cursorTo(t, m, "tds-api-server")
	m = press(m, " ")
	m = cursorTo(t, m, "tds-suite")
	m = press(m, " ", "enter")

	want := []string{"tds-hexer", "tds-api-server", "tds-suite"}
	if !reflect.DeepEqual(m.result.Selected, want) {
		t.Errorf("selection order lost: got %v, want %v", m.result.Selected, want)
	}
}

func TestReSelectingMovesRepoToTheEndOfOrder(t *testing.T) {
	opts := baseOpts()
	opts.Repos = []Repo{{Name: "a"}, {Name: "b"}}
	m := New(opts)

	m = cursorTo(t, m, "a")
	m = press(m, " ")
	m = cursorTo(t, m, "b")
	m = press(m, " ")
	m = cursorTo(t, m, "a")
	m = press(m, " ", " ", "enter")

	want := []string{"b", "a"}
	if !reflect.DeepEqual(m.result.Selected, want) {
		t.Errorf("got %v, want %v", m.result.Selected, want)
	}
}

func TestHexerModulesOnlyCountWhenHexerIsEnabled(t *testing.T) {
	m := New(baseOpts())

	// Modules are unreachable while the toggle is off.
	for i, r := range m.rows {
		if r.kind == rowHexerModule && m.selectable(i) {
			t.Fatal("a hexer module was selectable with hexer off")
		}
	}

	var toggle int
	for i, r := range m.rows {
		if r.kind == rowHexerToggle {
			toggle = i
		}
	}
	m.cursor = toggle
	m = press(m, " ", "down", " ", "enter")

	if !m.result.Hexer {
		t.Fatal("hexer not enabled")
	}
	if !reflect.DeepEqual(m.result.HexerModules, []string{"safe"}) {
		t.Errorf("expected the safe module, got %v", m.result.HexerModules)
	}
}

func TestDisablingHexerDropsItsModules(t *testing.T) {
	opts := baseOpts()
	opts.HexerEnabled = true
	opts.HexerSelected = []string{"safe", "kiosk"}
	m := New(opts)

	var toggle int
	for i, r := range m.rows {
		if r.kind == rowHexerToggle {
			toggle = i
		}
	}
	m.cursor = toggle
	m = press(m, " ", "enter")

	if m.result.Hexer {
		t.Error("hexer still reported as enabled")
	}
	if m.result.HexerModules != nil {
		t.Errorf("modules survived the toggle: %v", m.result.HexerModules)
	}
}

func TestEscCancelsWithoutADiff(t *testing.T) {
	m := cursorTo(t, New(baseOpts()), "tds-hexer")
	m = press(m, " ", "esc")

	if m.result.Confirmed {
		t.Error("cancelled picker reported as confirmed")
	}
	if len(m.result.Added) != 0 {
		t.Errorf("cancelled picker returned a diff: %v", m.result.Added)
	}
}

func TestCloneRequestShortCircuits(t *testing.T) {
	m := New(baseOpts())
	for i, r := range m.rows {
		if r.kind == rowClone {
			m.cursor = i
		}
	}
	m = press(m, "enter")
	if !m.cloning {
		t.Fatal("enter on the clone row did not open the clone field")
	}

	m = press(m, "t", "d", "s", "-", "c", "p", "p", "enter")
	if m.result.CloneRequest != "tds-cpp" {
		t.Errorf("clone request not captured: %q", m.result.CloneRequest)
	}
	if m.result.Confirmed {
		t.Error("a clone request should not confirm the form")
	}
}

func TestSpaceIsLiteralInTheBaseBranchField(t *testing.T) {
	m := New(baseOpts())
	for i, r := range m.rows {
		if r.kind == rowBase {
			m.cursor = i
			m.baseInput.Focus()
		}
	}
	m.baseInput.SetValue("")
	m = press(m, "a", " ", "b")

	if got := m.baseInput.Value(); got != "a b" {
		t.Errorf("space was swallowed as a toggle: %q", got)
	}
}

func TestPendingRemovalsAreSurfaced(t *testing.T) {
	m := cursorTo(t, New(baseOpts()), "tds-suite")
	m = press(m, " ")

	if got := m.pendingRemovals(); !reflect.DeepEqual(got, []string{"tds-suite"}) {
		t.Errorf("pending removal not surfaced: %v", got)
	}
}

func branchOpts() Options {
	return Options{
		Title: "EDIT REPOS · SUITE-1",
		Repos: []Repo{
			{Name: "tds-suite", Branches: []string{"master", "maintenance/13.3AV", "maintenance/13.2AV"}},
			{Name: "smtp-node", Branches: []string{"main"}},
		},
	}
}

func TestRepoCollapsedByDefault(t *testing.T) {
	m := New(branchOpts())

	for _, r := range m.rows {
		if r.kind == rowBranch {
			t.Fatal("branch rows are visible before expanding")
		}
	}
}

func TestExpandRevealsBranchesAndDefaultIsSelected(t *testing.T) {
	m := cursorTo(t, New(branchOpts()), "tds-suite")
	m = press(m, "right")

	var branches []string
	selected := ""
	for _, r := range m.rows {
		if r.kind == rowBranch && r.repo == "tds-suite" {
			branches = append(branches, r.name)
			if r.checked {
				selected = r.name
			}
		}
	}
	want := []string{"master", "maintenance/13.3AV", "maintenance/13.2AV"}
	if !reflect.DeepEqual(branches, want) {
		t.Errorf("branches = %v, want %v", branches, want)
	}
	if selected != "master" {
		t.Errorf("default base not pre-selected, got %q", selected)
	}
}

// A repo with only its default branch has nothing to choose between.
func TestSingleBranchRepoDoesNotExpand(t *testing.T) {
	m := cursorTo(t, New(branchOpts()), "smtp-node")
	m = press(m, "right")

	for _, r := range m.rows {
		if r.kind == rowBranch {
			t.Fatal("a single-branch repo expanded")
		}
	}
}

func TestPickingABranchSetsTheBaseAndSelectsTheRepo(t *testing.T) {
	m := cursorTo(t, New(branchOpts()), "tds-suite")
	// right expands, two downs reach the second branch, space picks it, and left
	// collapses back onto the repo so enter submits rather than re-picking.
	m = press(m, "right", "down", "down", " ", "left", "enter")

	if got := m.result.Bases["tds-suite"]; got != "maintenance/13.3AV" {
		t.Errorf("base = %q, want maintenance/13.3AV", got)
	}
	// Choosing a base implies wanting the repo.
	if !reflect.DeepEqual(m.result.Selected, []string{"tds-suite"}) {
		t.Errorf("selected = %v, want [tds-suite]", m.result.Selected)
	}
}

func TestBranchSelectionIsARadioGroup(t *testing.T) {
	m := cursorTo(t, New(branchOpts()), "tds-suite")
	m = press(m, "right", "down", "down", " ", "down", " ")

	checked := 0
	for _, r := range m.rows {
		if r.kind == rowBranch && r.checked {
			checked++
		}
	}
	if checked != 1 {
		t.Errorf("expected exactly one base checked, got %d", checked)
	}
	if m = press(m, "left", "enter"); m.result.Bases["tds-suite"] != "maintenance/13.2AV" {
		t.Errorf("last pick did not win: %q", m.result.Bases["tds-suite"])
	}
}

func TestCollapseReturnsCursorToTheRepo(t *testing.T) {
	m := cursorTo(t, New(branchOpts()), "tds-suite")
	m = press(m, "right", "down", "left")

	r := m.rows[m.cursor]
	if r.kind != rowRepo || r.name != "tds-suite" {
		t.Errorf("cursor landed on %v/%q after collapsing", r.kind, r.name)
	}
	for _, row := range m.rows {
		if row.kind == rowBranch {
			t.Fatal("branch rows survived the collapse")
		}
	}
}

// Two repos on different lines is the whole point: one ticket investigated
// across master and a maintenance branch at once.
func TestDifferentBasesPerRepo(t *testing.T) {
	opts := branchOpts()
	opts.Repos[1] = Repo{Name: "tds-hexer", Branches: []string{"master", "maintenance/13.3AV"}}
	m := New(opts)

	m = cursorTo(t, m, "tds-suite")
	m = press(m, " ") // keep master
	m = cursorTo(t, m, "tds-hexer")
	m = press(m, "right", "down", "down", " ", "left", "enter")

	if got := m.result.Bases["tds-suite"]; got != "master" {
		t.Errorf("tds-suite base = %q, want master", got)
	}
	if got := m.result.Bases["tds-hexer"]; got != "maintenance/13.3AV" {
		t.Errorf("tds-hexer base = %q, want maintenance/13.3AV", got)
	}
}
