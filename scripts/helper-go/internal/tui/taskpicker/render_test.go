package taskpicker

import (
	"os"
	"strings"
	"testing"
)

// TestRenderPreview prints the picker as it would appear, so the layout can be
// eyeballed without a terminal. Run with -v to see it.
func TestRenderPreview(t *testing.T) {
	m := New("OPEN TASK", []Item{
		{Key: "SUITE-9250", Summary: "Biomarin - can't edit message templates",
			Repos: []string{"tds-suite"}, Session: "claude", LastOpened: "3m ago"},
		{Key: "SUITE-9215", Summary: "[SWIFT] Access Log portlet not restricting role permissions",
			Repos: []string{"tds-suite", "tds-hexer"}, Session: "idle", LastOpened: "2h ago"},
		{Key: "VI-3034", Summary: "Fidelity UAT: addAlternateHost API issue",
			Repos: []string{"tds-suite-api", "tds-suite"}, Hexer: "vi-3034.acrid.dev", LastOpened: "3d ago"},
	})
	m.width = 120

	view := m.View()
	if !strings.Contains(view, "SUITE-9250") {
		t.Fatal("task keys missing from the view")
	}
	if os.Getenv("VERBOSE_RENDER") != "" {
		t.Log("\n" + view)
	}
}
