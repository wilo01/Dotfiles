// Package repopicker provides the EDIT REPOS screen: a multi-select over the
// repos a task enlists, plus its base branch and optional hexer environment.
// It performs no git or tmux work — it returns the requested change and the
// caller applies it.
package repopicker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Repo is one selectable repository.
type Repo struct {
	Name string
	// InTask marks a repo the task already enlists, so it renders pre-checked
	// and unchecking it means "remove".
	InTask bool
	// Branches are the bases this repo can be cut from, the repo default first.
	// Expanding a repo row shows them as a radio group.
	Branches []string
	// Base is the currently chosen base. Empty means the repo default, which is
	// Branches[0].
	Base string
}

// Options configures one run of the picker.
type Options struct {
	Title          string
	Repos          []Repo
	Base           string
	HexerAvailable bool
	HexerEnabled   bool
	HexerModules   []string
	HexerSelected  []string
}

// Result is what the user asked for. Added and Removed are relative to the
// InTask flags the picker was given.
type Result struct {
	Confirmed bool
	// CloneRequest is a repo name or URL to clone. When set, the caller clones
	// it and re-runs the picker; no other field is meaningful.
	CloneRequest string

	Selected []string
	Added    []string
	Removed  []string
	// Bases maps repo name to the base branch chosen for it. A repo absent from
	// the map uses its own default.
	Bases        map[string]string
	Base         string
	Hexer        bool
	HexerModules []string
}

type rowKind int

const (
	rowRepo rowKind = iota
	rowBranch
	rowBase
	rowHexerToggle
	rowHexerModule
	rowClone
	rowSeparator
)

type row struct {
	kind    rowKind
	name    string
	label   string
	checked bool
	inTask  bool
	// repo is the owning repo for a rowBranch.
	repo string
	// expanded tracks whether a rowRepo is showing its branch list.
	expanded bool
	// branches and base carry a rowRepo's choices; base is the selected one.
	branches []string
	base     string
	// seq records when a repo was checked, so the first-selected repo stays
	// first and becomes the task's primary.
	seq int
}

// Model is the bubbletea model for the picker.
type Model struct {
	title  string
	rows   []row
	cursor int

	baseInput  textinput.Model
	cloneInput textinput.Model
	cloning    bool

	seqCounter int
	width      int
	height     int
	offset     int

	confirmed bool
	cancelled bool
	result    Result
}

// New builds the picker model from the discovered repos and current task state.
func New(opts Options) Model {
	selected := map[string]bool{}
	for _, m := range opts.HexerSelected {
		selected[m] = true
	}

	base := textinput.New()
	base.Prompt = ""
	base.SetValue(opts.Base)
	base.CharLimit = 120

	clone := textinput.New()
	clone.Prompt = ""
	clone.Placeholder = "repo name or git URL"
	clone.CharLimit = 200

	m := Model{
		title:      opts.Title,
		baseInput:  base,
		cloneInput: clone,
	}

	// In-task repos first, in their existing order, so the primary repo is
	// stable across edits; the rest alphabetically.
	ordered := make([]Repo, 0, len(opts.Repos))
	for _, r := range opts.Repos {
		if r.InTask {
			ordered = append(ordered, r)
		}
	}
	rest := make([]Repo, 0, len(opts.Repos))
	for _, r := range opts.Repos {
		if !r.InTask {
			rest = append(rest, r)
		}
	}
	sort.Slice(rest, func(i, j int) bool { return rest[i].Name < rest[j].Name })
	ordered = append(ordered, rest...)

	for _, r := range ordered {
		seq := 0
		if r.InTask {
			m.seqCounter++
			seq = m.seqCounter
		}
		base := r.Base
		if base == "" && len(r.Branches) > 0 {
			base = r.Branches[0]
		}
		m.rows = append(m.rows, row{
			kind:     rowRepo,
			name:     r.Name,
			label:    r.Name,
			checked:  r.InTask,
			inTask:   r.InTask,
			seq:      seq,
			branches: r.Branches,
			base:     base,
		})
	}

	m.rows = append(m.rows, row{kind: rowClone, label: "clone a new repo…"})
	m.rows = append(m.rows, row{kind: rowBase, label: "base branch"})

	if opts.HexerAvailable {
		m.rows = append(m.rows, row{kind: rowSeparator, label: "environment"})
		m.rows = append(m.rows, row{kind: rowHexerToggle, label: "add a hexer env (dedicated db + server)", checked: opts.HexerEnabled})
		for _, mod := range opts.HexerModules {
			m.rows = append(m.rows, row{
				kind:    rowHexerModule,
				name:    mod,
				label:   fmt.Sprintf("  hexer app: %s", mod),
				checked: selected[mod],
			})
		}
	}

	m.cursor = m.nextSelectable(-1, 1)
	return m
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

// hexerOn reports whether the hexer toggle is checked, which gates its modules.
func (m Model) hexerOn() bool {
	for _, r := range m.rows {
		if r.kind == rowHexerToggle {
			return r.checked
		}
	}
	return false
}

func (m Model) selectable(i int) bool {
	if i < 0 || i >= len(m.rows) {
		return false
	}
	r := m.rows[i]
	if r.kind == rowSeparator {
		return false
	}
	if r.kind == rowHexerModule && !m.hexerOn() {
		return false
	}
	if r.kind == rowBranch && !m.repoExpanded(r.repo) {
		return false
	}
	return true
}

// repoExpanded reports whether the owning repo row is showing its branches.
func (m Model) repoExpanded(repo string) bool {
	for _, r := range m.rows {
		if r.kind == rowRepo && r.name == repo {
			return r.expanded
		}
	}
	return false
}

// repoRowIndex finds the repo row a branch row belongs to.
func (m Model) repoRowIndex(repo string) int {
	for i, r := range m.rows {
		if r.kind == rowRepo && r.name == repo {
			return i
		}
	}
	return -1
}

// nextSelectable walks from i in the given direction to the next focusable row,
// stopping at the ends rather than wrapping.
func (m Model) nextSelectable(i, dir int) int {
	for j := i + dir; j >= 0 && j < len(m.rows); j += dir {
		if m.selectable(j) {
			return j
		}
	}
	if m.selectable(i) {
		return i
	}
	for j := 0; j < len(m.rows); j++ {
		if m.selectable(j) {
			return j
		}
	}
	return 0
}

// buildResult turns the row state into the diff the caller applies.
func (m Model) buildResult() Result {
	type sel struct {
		name string
		seq  int
	}
	var chosen []sel
	res := Result{Confirmed: true, Base: strings.TrimSpace(m.baseInput.Value()), Bases: map[string]string{}}

	for _, r := range m.rows {
		switch r.kind {
		case rowRepo:
			switch {
			case r.checked:
				chosen = append(chosen, sel{r.name, r.seq})
				if r.base != "" {
					res.Bases[r.name] = r.base
				}
				if !r.inTask {
					res.Added = append(res.Added, r.name)
				}
			case r.inTask:
				res.Removed = append(res.Removed, r.name)
			}
		case rowHexerToggle:
			res.Hexer = r.checked
		case rowHexerModule:
			if r.checked {
				res.HexerModules = append(res.HexerModules, r.name)
			}
		}
	}

	sort.SliceStable(chosen, func(i, j int) bool { return chosen[i].seq < chosen[j].seq })
	for _, c := range chosen {
		res.Selected = append(res.Selected, c.name)
	}
	if !res.Hexer {
		res.HexerModules = nil
	}
	return res
}

// Run shows the picker and blocks until the user confirms or cancels.
func Run(opts Options) (Result, error) {
	prog := tea.NewProgram(New(opts))
	final, err := prog.Run()
	if err != nil {
		return Result{}, fmt.Errorf("repo picker failed: %w", err)
	}
	m, ok := final.(Model)
	if !ok {
		return Result{}, fmt.Errorf("repo picker returned an unexpected model")
	}
	return m.result, nil
}
