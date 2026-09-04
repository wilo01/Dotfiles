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
	Long: `Claude Code sessions per Jira ticket. One ticket is one task: N git
worktrees (one per repo you pick), one tmux session named after the ticket key
with one window per repo, one resumable Claude session, and optionally a
dedicated hexer environment.

  start  pick repos, create worktrees, open the session
         with no key: bring up every ticket agent.jql returns
  open   rebuild the session from tasks.json (no Jira call, no picker)
  edit   add or remove repos on an existing ticket
  status list every task with its worktree and session state
  stop   interrupt a running agent (--kill ends the session)
  cert   trust the tds-hexer dev certificate (one import covers every env)
  rm     delete a ticket's worktrees, hexer env and session
  spin   spawn an autonomous agent

Worktrees are created detached at origin/<default branch>, freshly fetched. No
branch is created: Claude branches when it starts work.

Tasks are recorded in tasks.json next to the rest of the hlp config.`,
}

func init() {
	AgentCmd.AddCommand(newLaunchCmd("start", "Pick repos and open an interactive Claude Code session for a ticket", true))
	AgentCmd.AddCommand(newLaunchCmd("spin", "Spawn an autonomous agent for one ticket (worktrees + tmux + /agent-run)", false))
	AgentCmd.AddCommand(openCmd)
	AgentCmd.AddCommand(editCmd)
	AgentCmd.AddCommand(statusCmd)
	AgentCmd.AddCommand(stopCmd)
	AgentCmd.AddCommand(rmCmd)
	AgentCmd.AddCommand(certCmd)
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
