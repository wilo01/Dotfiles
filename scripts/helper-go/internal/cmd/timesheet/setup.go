package timesheet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/timesheet"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure Timesheet plugin authentication",
	Long: `Configure authentication for the Jira Cloud Timesheet Tracking plugin.

You need to provide session cookies from an authenticated browser session:
1. Log in to Jira Cloud in your browser
2. Open DevTools (F12) → Application → Cookies
3. Copy the values for:
   - tenant.session.token
   - atlassian.xsrf.token
   - JSESSIONID

These tokens expire periodically and will need to be refreshed.`,
	Run: runSetup,
}

func runSetup(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config if present
	existingCfg := loadTimesheetConfig()

	// Prompt for Jira URL
	fmt.Print(ui.Label.Render("Jira Cloud URL (e.g., https://company.atlassian.net)"))
	if existingCfg.JiraURL != "" {
		fmt.Printf(" [%s]: ", existingCfg.JiraURL)
	} else {
		fmt.Print(": ")
	}
	jiraURL, _ := reader.ReadString('\n')
	jiraURL = strings.TrimSpace(jiraURL)
	if jiraURL == "" {
		jiraURL = existingCfg.JiraURL
	}
	jiraURL = strings.TrimSuffix(jiraURL, "/")

	// Prompt for session token
	fmt.Print(ui.Label.Render("tenant.session.token: "))
	sessionToken, _ := reader.ReadString('\n')
	sessionToken = strings.TrimSpace(sessionToken)
	if sessionToken == "" {
		sessionToken = existingCfg.SessionToken
	}

	// Prompt for XSRF token
	fmt.Print(ui.Label.Render("atlassian.xsrf.token: "))
	xsrfToken, _ := reader.ReadString('\n')
	xsrfToken = strings.TrimSpace(xsrfToken)
	if xsrfToken == "" {
		xsrfToken = existingCfg.XSRFToken
	}

	// Prompt for JSESSIONID
	fmt.Print(ui.Label.Render("JSESSIONID: "))
	jsessionid, _ := reader.ReadString('\n')
	jsessionid = strings.TrimSpace(jsessionid)
	if jsessionid == "" {
		jsessionid = existingCfg.JSESSIONID
	}

	// Validate inputs
	if jiraURL == "" {
		fmt.Println(ui.Error("Jira URL is required"))
		return
	}
	if sessionToken == "" {
		fmt.Println(ui.Error("Session token is required"))
		return
	}

	// Create config
	cfg := timesheet.Config{
		JiraURL:      jiraURL,
		SessionToken: sessionToken,
		XSRFToken:    xsrfToken,
		JSESSIONID:   jsessionid,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	// Test connection
	fmt.Println(ui.Info("Testing connection..."))
	client := timesheet.NewClient(cfg)
	if err := client.TestConnection(); err != nil {
		fmt.Println(ui.Warning("Connection test failed: " + err.Error()))
		fmt.Println(ui.Muted.Render("Configuration will be saved anyway - tokens may need to be refreshed"))
	} else {
		fmt.Println(ui.SuccessMsg("Connection successful!"))
	}

	// Save config
	if err := saveTimesheetConfig(cfg); err != nil {
		fmt.Println(ui.Error("Failed to save config: " + err.Error()))
		return
	}

	fmt.Println(ui.SuccessMsg("Timesheet configuration saved!"))
	fmt.Println(ui.Muted.Render("Config file: " + config.GetConfigPath("timesheet.json")))
}

func loadTimesheetConfig() timesheet.Config {
	path := config.GetConfigPath("timesheet.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return timesheet.Config{}
	}

	var cfg timesheet.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return timesheet.Config{}
	}
	return cfg
}

func saveTimesheetConfig(cfg timesheet.Config) error {
	path := config.GetConfigPath("timesheet.json")

	// Ensure directory exists
	if err := os.MkdirAll(config.GetConfigDir(), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Write with restrictive permissions
	return os.WriteFile(path, data, 0600)
}
