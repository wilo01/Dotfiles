package repopicker

import "strings"

// visibleRows is how many form rows are shown before the list scrolls. The repo
// list is long enough (22 repos here) that it must not push the footer off a
// standard terminal.
const visibleRows = 22

// View renders the form.
func (m Model) View() string {
	if m.confirmed || m.cancelled || m.result.CloneRequest != "" {
		return ""
	}

	width := m.width
	if width <= 0 {
		width = 80
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	if m.cloning {
		b.WriteString(labelStyle.Render("  clone a new repo"))
		b.WriteString("\n  ")
		b.WriteString(m.cloneInput.View())
		b.WriteString("\n\n")
		b.WriteString(footerStyle.Render("  ⏎ clone · esc back"))
		return b.String()
	}

	start, end := m.window()
	if start > 0 {
		b.WriteString(disabledStyle.Render("  ↑ more"))
		b.WriteString("\n")
	}
	for i := start; i < end; i++ {
		b.WriteString(m.renderRow(i, width))
		b.WriteString("\n")
	}
	if end < len(m.rows) {
		b.WriteString(disabledStyle.Render("  ↓ more"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if removals := m.pendingRemovals(); len(removals) > 0 {
		b.WriteString(warnStyle.Render("  unchecked on ⏎ = worktree removed and repo dropped: " + strings.Join(removals, ", ")))
		b.WriteString("\n")
	}
	if hint := m.contextHint(); hint != "" {
		b.WriteString(footerStyle.Render("  " + hint))
		b.WriteString("\n")
	}
	b.WriteString(footerStyle.Render("  ↑↓ move · ←→ expand · space toggle · ctrl+c clear · ⏎ ok · esc cancel"))
	return b.String()
}

// window returns the slice of rows to draw, keeping the cursor in view.
func (m Model) window() (int, int) {
	if len(m.rows) <= visibleRows {
		return 0, len(m.rows)
	}
	start := max(m.cursor-visibleRows/2, 0)
	if start+visibleRows > len(m.rows) {
		start = len(m.rows) - visibleRows
	}
	return start, start + visibleRows
}

func (m Model) renderRow(i, width int) string {
	r := m.rows[i]
	focused := i == m.cursor

	if r.kind == rowSeparator {
		return " " + sectionRule(r.label, width-2)
	}

	pointer := "  "
	if focused {
		pointer = selectedItemStyle.Render("▶ ")
	}

	switch r.kind {
	case rowBase:
		field := m.baseInput.View()
		if !focused {
			value := m.baseInput.Value()
			if value == "" {
				value = disabledStyle.Render("(repo default)")
			} else {
				value = valueStyle.Render(value)
			}
			field = value
		}
		return pointer + labelStyle.Render("base branch  ") + field

	case rowClone:
		style := itemStyle
		if focused {
			style = selectedItemStyle
		}
		return pointer + style.Render("[+] "+r.label)

	case rowBranch:
		mark := "( ) "
		if r.checked {
			mark = "(•) "
		}
		style := itemStyle
		if focused {
			style = selectedItemStyle
		}
		return pointer + style.Render(mark+"  ↳ "+r.label)
	}

	mark := "[ ] "
	if r.checked {
		mark = "[x] "
	}

	label := r.label
	if r.kind == rowRepo && len(r.branches) > 1 {
		// The base is only worth showing when it is not the repo default, which
		// is what Branches[0] always is.
		if r.base != "" && r.base != r.branches[0] {
			label += " @ " + r.base
		}
		if !r.expanded {
			label += "  " + disabledStyle.Render("→")
		}
	}
	if r.inTask {
		label += " · in the task (uncheck to remove)"
	}

	style := itemStyle
	switch {
	case r.kind == rowHexerModule && !m.hexerOn():
		style = disabledStyle
	case focused:
		style = selectedItemStyle
	case r.inTask:
		style = inTaskStyle
	}
	return pointer + style.Render(mark+label)
}

// contextHint explains what the arrow keys do on the focused row, since their
// meaning changes between a repo, its branch list and a text field.
func (m Model) contextHint() string {
	r := m.rows[m.cursor]
	switch {
	case r.kind == rowRepo && len(r.branches) > 1 && !r.expanded:
		return "→ to pick a base branch for " + r.name
	case r.kind == rowRepo && r.expanded:
		return "↓ to a branch · space/⏎ picks the base for " + r.name + " · ← collapses"
	case r.kind == rowBranch:
		return "space/⏎ picks the base for " + r.repo + " · ← collapses"
	}
	return ""
}

// pendingRemovals lists in-task repos the user has unchecked, so the
// destructive consequence is visible before ⏎.
func (m Model) pendingRemovals() []string {
	var out []string
	for _, r := range m.rows {
		if r.kind == rowRepo && r.inTask && !r.checked {
			out = append(out, r.name)
		}
	}
	return out
}
