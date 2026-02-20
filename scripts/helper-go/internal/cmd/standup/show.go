package standup

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	standuptui "github.com/dariuszw/hlp/internal/tui/standup"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	showTUI  bool
	showDays int
)

func init() {
	showCmd.Flags().BoolVar(&showTUI, "tui", false, "Launch interactive TUI dashboard")
	showCmd.Flags().IntVar(&showDays, "days", 0, "Limit to last N days (0=all)")
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show standup notes from worklog",
	Long: `Display worklog entries for standup meetings.

Default output is plain text (spreadsheet-friendly format).
Use --tui for interactive dashboard.

Examples:
  hlp standup show              # Plain text output (default)
  hlp standup show --days 3     # Last 3 days only
  hlp standup show --tui        # Interactive TUI dashboard`,
	RunE: runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	// Determine CSV path based on active profile
	profile, err := config.GetActiveProfile()
	if err != nil {
		fmt.Println(ui.Muted.Render("  (using default profile)"))
	}
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

	// TUI mode (opt-in)
	if showTUI {
		model := standuptui.NewModel(entries, csvPath)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	}

	// Plain text output (smart default: Monday=3 days, otherwise=1 day)
	days := showDays
	if days == 0 {
		if time.Now().Weekday() == time.Monday {
			days = 3
		} else {
			days = 1
		}
	}
	var baseURL string
	if profile != nil {
		baseURL = profile.BaseURL
	}
	return renderPlainEntries(entries, days, baseURL)
}
