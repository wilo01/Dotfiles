package repopicker

import (
	"os"
	"strings"
	"testing"
)

const longDescription = `Customer reports that message templates cannot be edited on the 13.1 line.
Editing the body and saving results in a NULL value being written to the database.

Steps to reproduce:
1. Log into the backoffice as an administrator
2. Navigate to Configuration > Message Templates
3. Open any template and change the body text
4. Save

Expected: the new body is persisted.
Actual: the body column is set to NULL and the template renders empty.

The route appears to bind a phantom variable in messageTemplateEdit.apex, and the
same pattern shows up in the hexer routing layer. The kiosk app is unaffected.

Note this only reproduces on maintenance/13.1AV, not on master.`

func descOpts() Options {
	return Options{
		Title:       "EDIT REPOS · SUITE-9250 · Biomarin - can't edit message templates",
		Description: longDescription,
		Repos: []Repo{
			{Name: "tds-suite", Branches: []string{"master", "maintenance/13.1AV"}},
			{Name: "tds-hexer", Mentioned: true, Branches: []string{"master"}},
			{Name: "tds-kiosk-chrome-app", Branches: []string{"master"}},
		},
		HexerAvailable: true,
		HexerModules:   []string{"safe", "kiosk"},
	}
}

func TestDescriptionPreviewIsTruncatedWithAPointer(t *testing.T) {
	m := New(descOpts())
	m.width = 100

	view := m.View()
	if !strings.Contains(view, "Customer reports that message templates") {
		t.Error("description preview missing")
	}
	if strings.Contains(view, "not on master") {
		t.Error("the whole description was rendered inline")
	}
	if !strings.Contains(view, "d for the full description") {
		t.Error("no pointer to the full text")
	}
	if !strings.Contains(view, "d description") {
		t.Error("footer does not mention the description key")
	}
}

func TestNoDescriptionMeansNoPanel(t *testing.T) {
	opts := descOpts()
	opts.Description = ""
	m := New(opts)
	m.width = 100

	view := m.View()
	if strings.Contains(view, "full description") || strings.Contains(view, "d description") {
		t.Error("description affordances shown with no description")
	}
}

func TestDPressOpensAndClosesTheOverlay(t *testing.T) {
	m := New(descOpts())
	m.width = 100

	m = press(m, "d")
	if !m.showDescription {
		t.Fatal("d did not open the overlay")
	}
	if !strings.Contains(m.View(), "d or esc back") {
		t.Error("overlay footer missing")
	}

	if m = press(m, "esc"); m.showDescription {
		t.Error("esc did not close the overlay")
	}
	if m.cancelled {
		t.Error("closing the overlay cancelled the picker")
	}
}

func TestOverlayScrollsToTheEnd(t *testing.T) {
	m := New(descOpts())
	m.width = 100

	m = press(m, "d")
	for range 10 {
		m = press(m, "pgdown")
	}
	if !strings.Contains(m.View(), "not on master") {
		t.Error("could not scroll to the end of the description")
	}
}

// The overlay must not mutate the selection underneath it.
func TestOverlayKeysDoNotTouchTheRepoList(t *testing.T) {
	m := cursorTo(t, New(descOpts()), "tds-suite")
	m = press(m, "d", "down", " ", "down", " ", "esc")

	for _, r := range m.rows {
		if r.kind == rowRepo && r.checked {
			t.Fatalf("%s got checked while the overlay was open", r.name)
		}
	}
}

func TestOverlayScrollClampsAtTheTop(t *testing.T) {
	m := New(descOpts())
	m.width = 100

	m = press(m, "d", "up", "up", "up")
	if m.descriptionTop != 0 {
		t.Errorf("scrolled above the first line: %d", m.descriptionTop)
	}
}

// d is a literal character while the base-branch field has focus.
func TestDIsLiteralInTheBaseField(t *testing.T) {
	m := New(descOpts())
	for i, r := range m.rows {
		if r.kind == rowBase {
			m.cursor = i
			m.baseInput.Focus()
		}
	}
	m.baseInput.SetValue("")
	m = press(m, "d", "e", "v")

	if m.showDescription {
		t.Fatal("d opened the overlay while typing a branch name")
	}
	if got := m.baseInput.Value(); got != "dev" {
		t.Errorf("base field = %q, want dev", got)
	}
}

func TestMentionedRepoIsTagged(t *testing.T) {
	m := New(descOpts())
	m.width = 100

	view := m.View()
	if !strings.Contains(view, "tds-hexer · mentioned") {
		t.Error("mentioned repo not tagged")
	}
	if strings.Contains(view, "tds-kiosk-chrome-app · mentioned") {
		t.Error("an unmentioned repo was tagged")
	}
}

func TestWrapPreservesParagraphs(t *testing.T) {
	lines := wrap("one two three\n\nfour five", 20)

	blank := false
	for _, l := range lines {
		if l == "" {
			blank = true
		}
		if len(l) > 20 {
			t.Errorf("line exceeds width: %q", l)
		}
	}
	if !blank {
		t.Errorf("paragraph break was collapsed: %v", lines)
	}
}

func TestRenderDescriptionPreview(t *testing.T) {
	m := New(descOpts())
	m.width = 100
	if os.Getenv("VERBOSE_RENDER") != "" {
		t.Log("\n" + m.View())
		t.Log("\n" + press(m, "d").View())
	}
}
