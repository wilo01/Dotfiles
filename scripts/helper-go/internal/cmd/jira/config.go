package jira

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure JIRA credentials",
	Long: `Configure JIRA connection settings.

This command will prompt for:
- JIRA Base URL (e.g., https://company.atlassian.net)
- Email address
- API Token

For JIRA Cloud, generate an API token at:
https://id.atlassian.com/manage-profile/security/api-tokens

For JIRA Server/DC, use a Personal Access Token.`,
	Run: runConfig,
}

func runConfig(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	// Get current values
	currentURL := viper.GetString("jira.base_url")
	currentEmail := viper.GetString("jira.email")

	// Prompt for Base URL
	fmt.Print(ui.Label.Render("JIRA Base URL"))
	if currentURL != "" {
		fmt.Printf(" [%s]: ", currentURL)
	} else {
		fmt.Print(": ")
	}
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = currentURL
	}

	// Prompt for Email
	fmt.Print(ui.Label.Render("Email"))
	if currentEmail != "" {
		fmt.Printf(" [%s]: ", currentEmail)
	} else {
		fmt.Print(": ")
	}
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		email = currentEmail
	}

	// Prompt for API Token
	fmt.Print(ui.Label.Render("API Token (input hidden): "))
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

	// Validate inputs
	if baseURL == "" {
		fmt.Println(ui.Error("Base URL is required"))
		return
	}
	if email == "" {
		fmt.Println(ui.Error("Email is required"))
		return
	}
	if token == "" {
		fmt.Println(ui.Error("API Token is required"))
		return
	}

	// Test connection
	fmt.Println(ui.Info("Testing connection..."))
	client := internalJira.NewClient(baseURL, email, token)
	if err := client.TestConnection(); err != nil {
		fmt.Println(ui.Error("Connection failed: " + err.Error()))
		return
	}
	fmt.Println(ui.SuccessMsg("Connection successful!"))

	// Save to config
	viper.Set("jira.base_url", baseURL)
	viper.Set("jira.email", email)

	// Save API token to keyring
	credMgr := config.NewCredentialManager()
	if err := credMgr.SetJiraToken(token); err != nil {
		fmt.Println(ui.Warning("Could not save token to keyring: " + err.Error()))
		fmt.Println(ui.Info("Token will be read from HLP_JIRA_TOKEN or JIRA_API_TOKEN environment variable"))
	} else {
		fmt.Println(ui.SuccessMsg("API token saved to system keyring"))
	}

	// Write config file
	if err := viper.WriteConfig(); err != nil {
		// Try to create the config file
		if err := viper.SafeWriteConfig(); err != nil {
			fmt.Println(ui.Warning("Could not save config: " + err.Error()))
		}
	}

	fmt.Println(ui.SuccessMsg("Configuration saved!"))
}
