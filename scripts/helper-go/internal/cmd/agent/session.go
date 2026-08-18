package agent

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/task"
	"github.com/dariuszw/hlp/internal/ui"
)

// launchOptions shapes the claude invocation for a session.
type launchOptions struct {
	// Manual opens a plain interactive claude rather than an autonomous
	// /agent-run agent.
	Manual         bool
	ContextPrompt  string
	PermissionMode string
	Relaunch       bool
	// NoAttach leaves the user where they are instead of switching to the
	// session. Set by --no-attach, and forced on for bulk runs.
	NoAttach bool
}

// ensureSession brings the tmux session in line with the task: one window per
// repo, each in that repo's worktree, then launches claude in the primary
// repo's window. Safe to re-run; it adds and removes windows rather than
// recreating the session.
func ensureSession(cfg *config.Config, store *task.Store, t task.Task, opts launchOptions) error {
	if !tmuxAvailable() {
		return fmt.Errorf("tmux not found in PATH")
	}
	primary, ok := t.PrimaryRepo()
	if !ok {
		return fmt.Errorf("%s has no repos — nothing to open", t.JiraKey)
	}

	session := t.SessionName()
	if !sessionExists(session) {
		if err := newSessionWithWindow(session, primary.Name, primary.Worktree); err != nil {
			return err
		}
		fmt.Println(ui.SuccessMsg("tmux session " + session + " created"))
	}

	if err := syncWindows(session, windowSpecs(t), knownRepoWindows(cfg, store)); err != nil {
		return err
	}

	if claudeRunningInWindow(session, primary.Name) && !opts.Relaunch {
		fmt.Println(ui.Info("claude already running in " + session + ":" + primary.Name + " — skipping launch (--relaunch to override)"))
		if err := touch(store, t.JiraKey); err != nil {
			return err
		}
		return attach(session, primary.Name, opts)
	}

	launch := buildLaunchCommand(cfg, t, opts)
	if err := sendKeysToWindow(session, primary.Name, launch); err != nil {
		return err
	}

	if !t.SessionStarted {
		t.SessionStarted = true
		store.Upsert(t)
	}
	if err := setSessionEnv(session, agentStartedEnvVar, fmt.Sprintf("%d", os.Getpid())); err != nil {
		fmt.Println(ui.Warning("failed to mark session as started: " + err.Error()))
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("claude launched in %s:%s", session, primary.Name)))
	if err := touch(store, t.JiraKey); err != nil {
		return err
	}
	return attach(session, primary.Name, opts)
}

// attach hands the terminal to the session, landing on the primary repo's
// window.
//
// Failing to attach is never fatal: the worktrees, the session and claude are
// all already up, so the only thing lost is the convenience, and the command
// that gets you there is printed instead.
func attach(session, window string, opts launchOptions) error {
	hint := func() { fmt.Println(ui.Info("attach with: tmux attach -t " + session)) }

	if opts.NoAttach {
		hint()
		return nil
	}
	if err := selectWindow(session, window); err != nil {
		fmt.Println(ui.Warning(err.Error()))
	}

	switch err := attachSession(session); {
	case err == nil:
		return nil
	case errors.Is(err, errNotATerminal):
		hint()
	default:
		fmt.Println(ui.Warning("could not attach: " + err.Error()))
		hint()
	}
	return nil
}

func touch(store *task.Store, key string) error {
	store.Touch(key)
	return store.Save()
}

// windowSpecs is one window per enlisted repo, in task order.
func windowSpecs(t task.Task) []windowSpec {
	specs := make([]windowSpec, 0, len(t.Repos))
	for _, r := range t.Repos {
		specs = append(specs, windowSpec{Name: r.Name, Dir: r.Worktree})
	}
	return specs
}

// knownRepoWindows names every window syncWindows may remove: any window named
// after a repo hlp knows about.
//
// It deliberately spans every discoverable repo rather than the task's current
// ones. A repo that was just dropped is no longer enlisted, so deriving this
// from the task alone would leave exactly the windows that need closing.
func knownRepoWindows(cfg *config.Config, store *task.Store) map[string]bool {
	known := map[string]bool{}
	if sources, err := discoverSourceRepos(cfg.Agent.ReposRoot); err == nil {
		for _, s := range sources {
			known[s.Name] = true
		}
	}
	for _, t := range store.All() {
		for _, r := range t.Repos {
			known[r.Name] = true
		}
	}
	return known
}

// buildLaunchCommand assembles the shell line sent to the primary window.
//
// The session UUID is minted with the task, so the first launch opens it with
// --session-id and every later one resumes it. That is what makes reopening a
// ticket continue the same conversation instead of starting a fresh one.
func buildLaunchCommand(cfg *config.Config, t task.Task, opts launchOptions) string {
	sessionFlag := "--session-id " + t.ClaudeSessionID
	if t.SessionStarted {
		sessionFlag = "--resume " + t.ClaudeSessionID
	}

	if !opts.Manual {
		prompt := strings.ReplaceAll(cfg.Agent.Prompt, "{{KEY}}", t.JiraKey)
		return fmt.Sprintf("CLAUDE_AGENT_MODE=1 JIRA_KEY=%s %s %s %q",
			t.JiraKey, cfg.Agent.ClaudeCmd, sessionFlag, prompt)
	}

	launch := fmt.Sprintf("JIRA_KEY=%s %s %s", t.JiraKey, stripSkipPermissions(cfg.Agent.ClaudeCmd), sessionFlag)
	if mode := permissionModeFor(cfg, opts); mode != "" {
		launch += " --permission-mode " + mode
	}
	if opts.ContextPrompt != "" {
		launch += fmt.Sprintf(" %q", strings.ReplaceAll(opts.ContextPrompt, "{{KEY}}", t.JiraKey))
	}
	return launch
}

// permissionModes mirrors claude's --permission-mode choices. The launch line is
// delivered by tmux send-keys into a detached session, so an invalid mode would
// fail inside a pane nobody is watching; hlp rejects it up front instead.
var permissionModes = []string{"acceptEdits", "auto", "bypassPermissions", "manual", "dontAsk", "plan"}

func permissionModeFor(cfg *config.Config, opts launchOptions) string {
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
