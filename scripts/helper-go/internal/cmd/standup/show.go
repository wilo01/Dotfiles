package standup

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	standuptui "github.com/dariuszw/hlp/internal/tui/standup"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Interactive standup dashboard",
	Long: `Launch an interactive TUI dashboard to view and prepare standup notes.

Navigate through your worklog entries grouped by day, view details,
and prepare for your standup meeting.

Examples:
  hlp standup show           # Show all history with scrolling tabs`,
	RunE: runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	// Determine CSV path based on active profile
	profile, _ := config.GetActiveProfile()
	csvPath := batch.DefaultCSVPathForProfile(profile)

	// Check if CSV exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return fmt.Errorf("worklog CSV not found: %s", csvPath)
	}

	// Parse all entries from CSV
	entries, err := batch.ParseCSV(csvPath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println(ui.Warning("No entries found in worklog CSV"))
		return nil
	}

	// Create and run TUI with ALL entries (no date filtering - explore full history)
	model := standuptui.NewModel(entries, csvPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
