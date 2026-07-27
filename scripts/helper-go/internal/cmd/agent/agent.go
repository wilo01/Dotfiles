package agent

import (
	"fmt"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/spf13/cobra"
)

// AgentCmd is the parent command for parallel AI agent workflows
var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Parallel AI agent workflows per Jira ticket",
	Long: `Claude Code sessions per Jira ticket, each in its own git worktree and
tmux session (named after the ticket key).

  start  opens a plain interactive claude in the ticket's worktree
  spin   spawns an autonomous agent (/agent-run + skip-permissions)
  fanout spawns autonomous agents for all assigned tickets

The target repo is resolved by analyzing the ticket: --repo flag, then a scan
for an existing branch containing the key, then a headless AI triage call,
then an interactive picker as fallback.`,
}

func init() {
	AgentCmd.AddCommand(newLaunchCmd("start", "Open an interactive Claude Code session for a ticket (worktree + tmux, no agent prompt)", true))
	AgentCmd.AddCommand(newLaunchCmd("spin", "Spawn an autonomous agent for one ticket (worktree + tmux + /agent-run)", false))
	AgentCmd.AddCommand(fanoutCmd)
	AgentCmd.AddCommand(statusCmd)
	AgentCmd.AddCommand(stopCmd)
}

func getJiraClient() (*internalJira.Client, error) {
	profile, err := config.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	if profile == nil || profile.BaseURL == "" {
		return nil, fmt.Errorf("JIRA not configured. Run: hlp jira config")
	}

	credMgr := config.NewCredentialManager()
	token, err := credMgr.GetJiraTokenForProfile(profile.Name)
	if err != nil {
		return nil, fmt.Errorf("API token not found for profile '%s'", profile.Name)
	}

	return internalJira.NewClient(profile.BaseURL, profile.Email, token), nil
}
