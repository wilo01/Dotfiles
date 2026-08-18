package repopicker

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles key input. Text fields keep focus for character keys, so
// toggling and navigation only apply when the cursor is on a checkbox row.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.cloning {
			return m.updateCloning(msg)
		}
		return m.updateForm(msg)
	}
	return m, nil
}

func (m Model) updateCloning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.cloning = false
		m.cloneInput.Blur()
		m.cloneInput.SetValue("")
		return m, nil

	case "ctrl+c":
		if m.cloneInput.Value() != "" {
			m.cloneInput.SetValue("")
			return m, nil
		}
		m.cancelled = true
		return m, tea.Quit

	case "enter":
		value := strings.TrimSpace(m.cloneInput.Value())
		if value == "" {
			m.cloning = false
			m.cloneInput.Blur()
			return m, nil
		}
		m.result = Result{CloneRequest: value}
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.cloneInput, cmd = m.cloneInput.Update(msg)
	return m, cmd
}

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	onText := m.rows[m.cursor].kind == rowBase

	switch msg.String() {
	case "esc", "q":
		if onText && msg.String() == "q" {
			break // "q" is a literal character while editing the base branch
		}
		m.cancelled = true
		m.result = Result{}
		return m, tea.Quit

	case "ctrl+c":
		if onText && m.baseInput.Value() != "" {
			m.baseInput.SetValue("")
			return m, nil
		}
		m.cancelled = true
		m.result = Result{}
		return m, tea.Quit

	case "enter":
		switch m.rows[m.cursor].kind {
		case rowClone:
			m.cloning = true
			m.baseInput.Blur()
			return m, m.cloneInput.Focus()
		case rowBranch:
			// Enter on a branch picks it, rather than submitting the whole form
			// on what reads as a selection.
			return m.toggle(), nil
		}
		m.confirmed = true
		m.result = m.buildResult()
		return m, tea.Quit

	case "up", "shift+tab":
		return m.moveCursor(-1), nil

	case "down", "tab":
		return m.moveCursor(1), nil

	case "right":
		if onText {
			break // the cursor belongs to the branch-name field
		}
		return m.expand(true), nil

	case "left":
		if onText {
			break
		}
		return m.expand(false), nil

	case " ":
		if onText {
			break // a space belongs in the branch name
		}
		return m.toggle(), nil
	}

	if onText {
		var cmd tea.Cmd
		m.baseInput, cmd = m.baseInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) moveCursor(dir int) Model {
	m.cursor = m.nextSelectable(m.cursor, dir)
	if m.rows[m.cursor].kind == rowBase {
		m.baseInput.Focus()
	} else {
		m.baseInput.Blur()
	}
	return m
}

// expand opens or closes the focused repo's branch list. Closing from inside
// the list jumps back to the owning repo, so the cursor never lands on a hidden
// row.
func (m Model) expand(open bool) Model {
	r := m.rows[m.cursor]

	switch r.kind {
	case rowRepo:
		if len(r.branches) < 2 {
			return m // nothing to choose between
		}
		if open == r.expanded {
			return m
		}
		if open {
			m = m.insertBranchRows(m.cursor)
		} else {
			m = m.removeBranchRows(r.name)
		}
	case rowBranch:
		if open {
			return m
		}
		owner := m.repoRowIndex(r.repo)
		m = m.removeBranchRows(r.repo)
		if owner >= 0 {
			m.cursor = owner
		}
	}
	return m
}

// insertBranchRows materializes a repo's branch radio group directly beneath it.
func (m Model) insertBranchRows(at int) Model {
	repo := m.rows[at]
	rows := make([]row, 0, len(repo.branches))
	for _, b := range repo.branches {
		rows = append(rows, row{
			kind:    rowBranch,
			name:    b,
			label:   b,
			repo:    repo.name,
			checked: b == repo.base,
		})
	}

	expanded := make([]row, 0, len(m.rows)+len(rows))
	expanded = append(expanded, m.rows[:at+1]...)
	expanded = append(expanded, rows...)
	expanded = append(expanded, m.rows[at+1:]...)

	m.rows = expanded
	m.rows[at].expanded = true
	return m
}

func (m Model) removeBranchRows(repo string) Model {
	kept := make([]row, 0, len(m.rows))
	for _, r := range m.rows {
		if r.kind == rowBranch && r.repo == repo {
			continue
		}
		if r.kind == rowRepo && r.name == repo {
			r.expanded = false
		}
		kept = append(kept, r)
	}
	m.rows = kept
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	return m
}

// toggle flips the focused checkbox. Checking a repo stamps it with the next
// sequence number so selection order — and therefore the primary repo — is
// preserved; unchecking clears it.
func (m Model) toggle() Model {
	r := &m.rows[m.cursor]
	switch r.kind {
	case rowRepo:
		r.checked = !r.checked
		if r.checked {
			m.seqCounter++
			r.seq = m.seqCounter
		} else {
			r.seq = 0
		}
	case rowBranch:
		// A radio group: picking one base clears the others, and choosing a base
		// implies wanting the repo.
		repo, base := r.repo, r.name
		for i := range m.rows {
			switch {
			case m.rows[i].kind == rowBranch && m.rows[i].repo == repo:
				m.rows[i].checked = m.rows[i].name == base
			case m.rows[i].kind == rowRepo && m.rows[i].name == repo:
				m.rows[i].base = base
				if !m.rows[i].checked {
					m.rows[i].checked = true
					m.seqCounter++
					m.rows[i].seq = m.seqCounter
				}
			}
		}
		return m
	case rowHexerToggle, rowHexerModule:
		r.checked = !r.checked
	default:
		return m
	}

	// Turning hexer off can hide the row the cursor sits on.
	if !m.selectable(m.cursor) {
		m.cursor = m.nextSelectable(m.cursor, -1)
	}
	return m
}
