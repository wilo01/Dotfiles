package jira

import (
	"fmt"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile [name]",
	Short: "Manage JIRA connection profiles",
	Long: `Switch between JIRA profiles (localhost, production, etc.)

Examples:
  hlp jira profile              # Show current profile
  hlp jira profile local        # Switch to localhost
  hlp jira profile acre         # Switch to production (Acre)
  hlp jira profile --list       # List all profiles`,
	Args: cobra.MaximumNArgs(1),
	Run:  runProfile,
}

var (
	profileList bool
	profileAdd  string
)

func init() {
	profileCmd.Flags().BoolVarP(&profileList, "list", "l", false, "List all profiles")
	profileCmd.Flags().StringVarP(&profileAdd, "add", "a", "", "Add new profile (name)")
}

func runProfile(cmd *cobra.Command, args []string) {
	// Initialize profiles if needed
	if err := config.InitializeProfiles(); err != nil {
		fmt.Println(ui.Warning("Could not initialize profiles: " + err.Error()))
	}

	// List profiles
	if profileList {
		listProfiles()
		return
	}

	// Show current profile if no args
	if len(args) == 0 {
		showCurrentProfile()
		return
	}

	// Switch profile
	profileName := args[0]
	switchProfile(profileName)
}

func showCurrentProfile() {
	profile, err := config.GetActiveProfile()
	if err != nil {
		fmt.Println(ui.Error("Failed to get profile: " + err.Error()))
		return
	}

	modeStr := "[READ/WRITE]"
	modeColor := ui.Success
	if profile.Protected {
		modeStr = "[PROTECTED]"
		modeColor = ui.WarningText
	}

	fmt.Printf("Current profile: %s (%s) %s\n",
		ui.Primary.Render(profile.Name),
		ui.Muted.Render(profile.BaseURL),
		modeColor.Render(modeStr))

	if len(profile.Projects) > 0 {
		fmt.Printf("Projects: %s\n", ui.Muted.Render(strings.Join(profile.Projects, ", ")))
	}
}

func listProfiles() {
	profiles, err := config.LoadProfiles()
	if err != nil {
		fmt.Println(ui.Error("Failed to load profiles: " + err.Error()))
		return
	}

	fmt.Println("Available profiles:")
	for name, profile := range profiles.Profiles {
		activeMarker := "  "
		if name == profiles.ActiveProfile {
			activeMarker = ui.Success.Render("→ ")
		}

		modeStr := ""
		if profile.Protected {
			modeStr = ui.WarningText.Render(" [PROTECTED]")
		}

		fmt.Printf("%s%s %s%s\n",
			activeMarker,
			ui.Primary.Render(fmt.Sprintf("%-12s", name)),
			ui.Muted.Render(profile.BaseURL),
			modeStr)
	}
}

func switchProfile(name string) {
	// Get current profile for comparison
	current, _ := config.GetActiveProfile()
	currentName := ""
	if current != nil {
		currentName = current.Name
	}

	if currentName == name {
		fmt.Printf("%s Already on profile: %s\n",
			ui.WarningText.Render("!"),
			ui.Primary.Render(name))
		return
	}

	if err := config.SetActiveProfile(name); err != nil {
		fmt.Println(ui.Error("Failed to switch profile: " + err.Error()))
		return
	}

	// Get new profile info
	profile, _ := config.GetActiveProfile()
	modeStr := "[READ/WRITE]"
	modeColor := ui.Success
	if profile != nil && profile.Protected {
		modeStr = "[PROTECTED]"
		modeColor = ui.WarningText
	}

	fmt.Printf("%s Switched profile: %s → %s\n",
		ui.Success.Render("✓"),
		ui.Muted.Render(currentName),
		ui.Primary.Render(name))

	if profile != nil {
		fmt.Printf("  URL: %s %s\n",
			ui.Muted.Render(profile.BaseURL),
			modeColor.Render(modeStr))
	}
}
