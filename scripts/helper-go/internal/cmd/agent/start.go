package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

// newLaunchCmd builds a launch command (start/spin) sharing the full
// resolve-repo -> branch -> worktree -> tmux pipeline; manual controls
// whether claude is opened interactively or as an autonomous agent.
func newLaunchCmd(use, short string, manual bool) *cobra.Command {
	opts := startOptions{Manual: manual}
	cmd := &cobra.Command{
		Use:   use + " <TICKET-KEY> [REPO]",
		Short: short,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			client, err := getJiraClient()
			if err != nil {
				return err
			}
			key := strings.ToUpper(args[0])
			if len(args) == 2 {
				opts.Repo = args[1]
			}
			ticket, err := client.GetTicket(key)
			if err != nil {
				return err
			}
			return startAgent(cfg, client, ticket, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "target repo name (skips AI triage)")
	cmd.Flags().StringVar(&opts.Base, "base", "", "base branch for new work (default: origin/HEAD)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "resolve repo/branch/worktree and print the plan without doing anything")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "skip the assignee check")
	cmd.Flags().BoolVar(&opts.Relaunch, "relaunch", false, "send the claude launch even if one already ran in the session")
	cmd.Flags().BoolVar(&opts.Interactive, "interactive", false, "always pick the repo interactively")
	return cmd
}

type startOptions struct {
	Repo        string
	Base        string
	DryRun      bool
	Force       bool
	Relaunch    bool
	Interactive bool
	// Manual opens a plain interactive claude (no agent prompt, no
	// --dangerously-skip-permissions, no CLAUDE_AGENT_MODE).
	Manual bool
}

// startAgent runs the idempotent start sequence for one ticket:
// resolve repo -> ensure branch -> ensure worktree -> ensure tmux session -> launch claude.
func startAgent(cfg *config.Config, client *internalJira.Client, ticket *internalJira.Ticket, opts startOptions) error {
	key := ticket.Key

	if ticket.Assignee == "" && !opts.Force {
		return fmt.Errorf("%s is unassigned (use --force to start anyway)", key)
	}
	if !tmuxAvailable() {
		return fmt.Errorf("tmux not found in PATH")
	}

	repos, err := discoverRepos(cfg.Agent.WorktreeRoot)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return fmt.Errorf("no repos found under %s", cfg.Agent.WorktreeRoot)
	}

	repo, how, err := resolveRepo(cfg, client, ticket, repos, opts)
	if err != nil {
		return err
	}
	source, err := repoSource(cfg, repo)
	if err != nil {
		return err
	}

	// Refresh remote-tracking refs so base/branch resolution and the later
	// fast-forward work from the latest remote state. Best-effort (offline is fine).
	stopFetch := ui.Spinner("Fetching latest from origin...")
	fetchErr := fetchOrigin(source)
	stopFetch()
	if fetchErr != nil {
		fmt.Println(ui.Warning("git fetch failed (using local refs): " + fetchErr.Error()))
	}

	base := opts.Base
	if base == "" {
		if override, ok := cfg.Agent.Repos[repo]; ok && override.Base != "" {
			base = override.Base
		} else {
			base = defaultBase(source)
		}
	}

	branch, branchExists := findBranch(source, key)
	if !branchExists {
		branch = branchNameFor(key, ticket.Summary, ticket.IssueType)
	}

	worktreePath := worktreePathForBranch(source, branch)
	worktreeExists := worktreePath != ""
	if !worktreeExists {
		worktreePath = filepath.Join(expandPath(cfg.Agent.WorktreeRoot), repo, key)
	}

	var launch string
	if opts.Manual {
		launch = fmt.Sprintf("JIRA_KEY=%s %s", key, stripSkipPermissions(cfg.Agent.ClaudeCmd))
	} else {
		prompt := strings.ReplaceAll(cfg.Agent.Prompt, "{{KEY}}", key)
		launch = fmt.Sprintf("CLAUDE_AGENT_MODE=1 JIRA_KEY=%s %s %q", key, cfg.Agent.ClaudeCmd, prompt)
	}

	fmt.Println(ui.Header(key + " — " + ticket.Summary))
	fmt.Println(ui.KeyValue("Repo", fmt.Sprintf("%s (%s)", repo, how)))
	fmt.Println(ui.KeyValue("Branch", branch+existingMarker(branchExists)))
	fmt.Println(ui.KeyValue("Base", base))
	fmt.Println(ui.KeyValue("Worktree", worktreePath+existingMarker(worktreeExists)))
	fmt.Println(ui.KeyValue("Session", key+existingMarker(sessionExists(key))))
	fmt.Println(ui.KeyValue("Launch", launch))

	if opts.DryRun {
		fmt.Println(ui.Info("dry-run: nothing was created"))
		return nil
	}

	if !worktreeExists {
		if err := addWorktree(source, worktreePath, branch, "origin/"+base, !branchExists); err != nil {
			return err
		}
		fmt.Println(ui.SuccessMsg("worktree created"))
	}

	// Bring the worktree's branch up to date with its remote (fast-forward only).
	// Fixes the "resumed worktree is N commits behind" case; safe for fresh worktrees.
	fastForwardToUpstream(worktreePath)

	if sessionExists(key) {
		if path, err := sessionPath(key); err == nil && filepath.Clean(path) != filepath.Clean(worktreePath) {
			fmt.Println(ui.Warning(fmt.Sprintf("reusing session %s but its path is %s, not the worktree", key, path)))
		}
	} else {
		if err := newSession(key, worktreePath); err != nil {
			return err
		}
		fmt.Println(ui.SuccessMsg("tmux session created"))
	}

	if getSessionEnv(key, agentStartedEnvVar) != "" && agentRunning(key) && !opts.Relaunch {
		fmt.Println(ui.Info("claude already running in session — skipping launch (--relaunch to override)"))
		return nil
	}

	if err := sendKeys(key, launch); err != nil {
		return err
	}
	if err := setSessionEnv(key, agentStartedEnvVar, fmt.Sprintf("%d", os.Getpid())); err != nil {
		fmt.Println(ui.Warning("failed to mark session as started: " + err.Error()))
	}
	if opts.Manual {
		fmt.Println(ui.SuccessMsg("claude launched in tmux session " + key))
	} else {
		fmt.Println(ui.SuccessMsg("agent launched in tmux session " + key))
	}
	return nil
}

// stripSkipPermissions removes --dangerously-skip-permissions from the
// configured claude command so manual sessions get normal permission prompts.
func stripSkipPermissions(claudeCmd string) string {
	var kept []string
	for tok := range strings.FieldsSeq(claudeCmd) {
		if tok == "--dangerously-skip-permissions" {
			continue
		}
		kept = append(kept, tok)
	}
	return strings.Join(kept, " ")
}

func existingMarker(exists bool) string {
	if exists {
		return " (existing)"
	}
	return " (new)"
}

// resolveRepo implements the resolution chain:
// --repo flag -> resume-scan -> AI triage -> interactive picker.
func resolveRepo(cfg *config.Config, client *internalJira.Client, ticket *internalJira.Ticket, repos []string, opts startOptions) (string, string, error) {
	if opts.Repo != "" {
		for _, r := range repos {
			if r == opts.Repo {
				return r, "specified", nil
			}
		}
		return "", "", fmt.Errorf("repo %s not found under %s (known: %s)", opts.Repo, cfg.Agent.WorktreeRoot, strings.Join(repos, ", "))
	}

	if matches := resumeScan(cfg, repos, ticket.Key); len(matches) == 1 {
		return matches[0], "resume: existing branch", nil
	} else if len(matches) > 1 {
		fmt.Println(ui.Warning(fmt.Sprintf("branch %s found in multiple repos: %s", ticket.Key, strings.Join(matches, ", "))))
		repo, err := pickRepo(matches)
		return repo, "picked (multiple resume matches)", err
	}

	if opts.Interactive {
		repo, err := pickRepo(repos)
		return repo, "picked", err
	}

	description, err := client.GetTicketDescription(ticket.Key)
	if err != nil {
		fmt.Println(ui.Warning("could not fetch description for triage: " + err.Error()))
	}

	stop := ui.Spinner("AI triage: resolving target repo...")
	verdict, err := runTriage(cfg, ticket, description, repos)
	stop()
	if err != nil {
		fmt.Println(ui.Warning("triage failed: " + err.Error()))
		repo, pickErr := pickRepo(repos)
		return repo, "picked (triage failed)", pickErr
	}

	fmt.Println(ui.KeyValue("Triage", fmt.Sprintf("%s [%s] — %s", verdict.Repo, verdict.Confidence, verdict.Reason)))
	if verdict.Confidence == "low" {
		fmt.Println(ui.Warning("low triage confidence — pick the repo"))
		repo, err := pickRepo(repos)
		return repo, "picked (low confidence)", err
	}
	return verdict.Repo, "AI triage: " + verdict.Reason, nil
}

// pickRepo lets the user choose a repo via fzf, falling back to a numbered prompt
func pickRepo(repos []string) (string, error) {
	if fzf, err := exec.LookPath("fzf"); err == nil {
		cmd := exec.Command(fzf, "--prompt", "repo> ", "--height", "40%")
		cmd.Stdin = strings.NewReader(strings.Join(repos, "\n"))
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err == nil {
			if choice := strings.TrimSpace(string(out)); choice != "" {
				return choice, nil
			}
		}
	}

	for i, repo := range repos {
		fmt.Println(ui.NumberedItem(i+1, repo))
	}
	choice, err := ui.PromptChoice("Repo number")
	if err != nil {
		return "", err
	}
	var idx int
	if _, err := fmt.Sscanf(strings.TrimSpace(choice), "%d", &idx); err != nil || idx < 1 || idx > len(repos) {
		return "", fmt.Errorf("invalid choice: %s", choice)
	}
	return repos[idx-1], nil
}
