package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
			if err := validateLaunchFlags(&opts); err != nil {
				return err
			}
			opts.ContextPrompt = resolveContextPrompt(cfg, opts.ContextPrompt)
			return startAgent(cfg, client, ticket, opts)
		},
	}
	registerLaunchFlags(cmd, &opts)
	return cmd
}

// registerLaunchFlags declares every flag that feeds startOptions. Both the
// single-ticket commands and fanout use it so the two surfaces cannot drift.
func registerLaunchFlags(cmd *cobra.Command, opts *startOptions) {
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "target repo name (skips AI triage)")
	cmd.Flags().StringVar(&opts.ContextPrompt, "context", "", contextFlagUsage)
	cmd.Flags().Lookup("context").NoOptDefVal = contextPromptFromConfig
	cmd.Flags().StringVar(&opts.PermissionMode, "permission-mode", "", permissionModeFlagUsage)
	cmd.Flags().StringVar(&opts.Base, "base", "", "base branch for new work (default: origin/HEAD)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "resolve repo/branch/worktree and print the plan without doing anything")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "skip the assignee check")
	cmd.Flags().BoolVar(&opts.Relaunch, "relaunch", false, "send the claude launch even if one already ran in the session")
	cmd.Flags().BoolVar(&opts.Interactive, "interactive", false, "always pick the repo interactively")
}

// validateLaunchFlags rejects combinations that would otherwise be accepted and
// silently do nothing. Both --context and --permission-mode only shape an
// interactive claude, so passing them to an autonomous run is a mistake worth
// surfacing rather than ignoring.
func validateLaunchFlags(opts *startOptions) error {
	if !opts.Manual {
		if opts.ContextPrompt != "" {
			return fmt.Errorf("--context only applies to interactive sessions: an autonomous agent reads the ticket itself")
		}
		if opts.PermissionMode != "" {
			return fmt.Errorf("--permission-mode only applies to interactive sessions: autonomous agents run with agent.claude_cmd as configured")
		}
	}
	return validatePermissionMode(opts.PermissionMode)
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
	// ContextPrompt is passed to an interactive claude as its opening prompt
	// ({{KEY}} substituted). Ignored when Manual is false.
	ContextPrompt string
	// RepoHow describes where Repo came from, for callers that resolved it
	// ahead of resolveRepo. Empty means "specified on the command line".
	RepoHow string
	// PermissionMode overrides agent.permission_mode for this launch. Ignored
	// when Manual is false.
	PermissionMode string
}

// contextPromptFromConfig is the value --context takes when passed bare, later
// swapped for agent.context_prompt.
const contextPromptFromConfig = "@config"

// Setting NoOptDefVal makes --context valid on its own, which is also what stops
// pflag accepting the space-separated form, hence the =TMPL in the usage text.
const contextFlagUsage = "open claude with a briefing prompt; bare uses agent.context_prompt, " +
	"or --context=TMPL for your own ({{KEY}} is substituted). Requires the = form."

const permissionModeFlagUsage = "claude permission mode for interactive sessions " +
	"(default: agent.permission_mode; use 'manual' for normal prompts)"

func resolveContextPrompt(cfg *config.Config, flagValue string) string {
	if flagValue == contextPromptFromConfig {
		return cfg.Agent.ContextPrompt
	}
	return flagValue
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

	launch := buildLaunchCommand(cfg, key, opts)

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

// buildLaunchCommand assembles the shell line sent to the ticket's tmux session.
// Manual sessions get a plain claude the user drives, optionally opened on a
// briefing prompt; otherwise claude starts autonomously on the agent prompt.
func buildLaunchCommand(cfg *config.Config, key string, opts startOptions) string {
	if !opts.Manual {
		prompt := strings.ReplaceAll(cfg.Agent.Prompt, "{{KEY}}", key)
		return fmt.Sprintf("CLAUDE_AGENT_MODE=1 JIRA_KEY=%s %s %q", key, cfg.Agent.ClaudeCmd, prompt)
	}

	launch := fmt.Sprintf("JIRA_KEY=%s %s", key, stripSkipPermissions(cfg.Agent.ClaudeCmd))
	if mode := permissionMode(cfg, opts); mode != "" {
		launch += " --permission-mode " + mode
	}
	if opts.ContextPrompt != "" {
		launch += fmt.Sprintf(" %q", strings.ReplaceAll(opts.ContextPrompt, "{{KEY}}", key))
	}
	return launch
}

// permissionModes mirrors claude's --permission-mode choices. The launch line is
// delivered by tmux send-keys into a detached session, so an invalid mode would
// fail inside a pane nobody is watching; hlp rejects it up front instead.
var permissionModes = []string{"acceptEdits", "auto", "bypassPermissions", "manual", "dontAsk", "plan"}

func permissionMode(cfg *config.Config, opts startOptions) string {
	if opts.PermissionMode != "" {
		return opts.PermissionMode
	}
	return cfg.Agent.PermissionMode
}

func validatePermissionMode(mode string) error {
	if mode == "" || slices.Contains(permissionModes, mode) {
		return nil
	}
	return fmt.Errorf("unknown permission mode %q (valid: %s)", mode, strings.Join(permissionModes, ", "))
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
		how := opts.RepoHow
		if how == "" {
			how = "specified"
		}
		for _, r := range repos {
			if r == opts.Repo {
				return r, how, nil
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

	// Warnings are buffered rather than printed inline: the spinner owns the
	// current line until it stops.
	var warnings []string
	stop := ui.Spinner("AI triage: resolving target repo...")
	verdict, err := triageRepo(cfg, client, ticket, repos, func(msg string) {
		warnings = append(warnings, msg)
	})
	stop()
	for _, w := range warnings {
		fmt.Println(ui.Warning(w))
	}
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
