package agent

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/task"
	"github.com/dariuszw/hlp/internal/tui/repopicker"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

// newLaunchCmd builds a launch command (start/spin) sharing the full
// pick-repos -> worktrees -> tmux pipeline; manual controls whether claude is
// opened interactively or as an autonomous agent.
func newLaunchCmd(use, short string, manual bool) *cobra.Command {
	opts := startOptions{Manual: manual}
	cmd := &cobra.Command{
		Use:   use + " <TICKET-KEY> [REPO...]",
		Short: short,
		Args:  cobra.MinimumNArgs(1),
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
			opts.Repos = args[1:]

			ticket, err := client.GetTicket(key)
			if err != nil {
				return err
			}
			if err := validateLaunchFlags(&opts); err != nil {
				return err
			}
			opts.ContextPrompt = resolveContextPrompt(cfg, opts.ContextPrompt)
			return startAgent(cfg, ticket, opts)
		},
	}
	registerLaunchFlags(cmd, &opts)
	return cmd
}

// registerLaunchFlags declares every flag that feeds startOptions. Both the
// single-ticket commands and fanout use it so the two surfaces cannot drift.
func registerLaunchFlags(cmd *cobra.Command, opts *startOptions) {
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "target repo name (skips the picker)")
	cmd.Flags().StringVar(&opts.ContextPrompt, "context", "", contextFlagUsage)
	cmd.Flags().Lookup("context").NoOptDefVal = contextPromptFromConfig
	cmd.Flags().StringVar(&opts.PermissionMode, "permission-mode", "", permissionModeFlagUsage)
	cmd.Flags().StringVar(&opts.Base, "base", "", "base branch to cut worktrees from (default: each repo's origin/HEAD)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print the plan without creating or removing anything")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "skip the assignee check and allow discarding dirty worktrees")
	cmd.Flags().BoolVar(&opts.Relaunch, "relaunch", false, "send the claude launch even if one already runs in the session")
	cmd.Flags().BoolVar(&opts.Interactive, "interactive", false, "always open the repo picker")
	cmd.Flags().BoolVar(&opts.NoAttach, "no-attach", false, "create the session but stay where you are")
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
	// Repo is the single-repo form used by fanout and --repo; Repos is the
	// multi-repo positional form.
	Repo        string
	Repos       []string
	Base        string
	DryRun      bool
	Force       bool
	Relaunch    bool
	Interactive bool
	NoAttach    bool
	// Manual opens a plain interactive claude (no agent prompt, no
	// --dangerously-skip-permissions, no CLAUDE_AGENT_MODE).
	Manual bool
	// ContextPrompt is passed to an interactive claude as its opening prompt
	// ({{KEY}} substituted). Ignored when Manual is false.
	ContextPrompt string
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
// choose repos -> ensure worktrees -> ensure hexer -> ensure tmux -> launch claude.
func startAgent(cfg *config.Config, ticket *internalJira.Ticket, opts startOptions) error {
	if ticket.Assignee == "" && !opts.Force {
		return fmt.Errorf("%s is unassigned (use --force to start anyway)", ticket.Key)
	}

	store, err := loadStore()
	if err != nil {
		return err
	}

	apply, err := resolveSelection(cfg, store, ticket, opts)
	if err != nil {
		return err
	}
	if len(apply.Selected) == 0 {
		return fmt.Errorf("no repos selected for %s", ticket.Key)
	}

	t, err := applyTask(cfg, store, apply)
	if err != nil {
		return err
	}
	if opts.DryRun {
		return nil
	}

	return ensureSession(cfg, store, t, launchOptions{
		Manual:         opts.Manual,
		ContextPrompt:  opts.ContextPrompt,
		PermissionMode: opts.PermissionMode,
		Relaunch:       opts.Relaunch,
		NoAttach:       opts.NoAttach,
	})
}

// resolveSelection decides which repos the task should end up with: the ones
// named on the command line, or whatever the picker returns.
func resolveSelection(cfg *config.Config, store *task.Store, ticket *internalJira.Ticket, opts startOptions) (applyOptions, error) {
	existing, _ := store.Get(ticket.Key)

	apply := applyOptions{
		Key:          ticket.Key,
		Summary:      ticket.Summary,
		Base:         opts.Base,
		DryRun:       opts.DryRun,
		Force:        opts.Force,
		Hexer:        existing.Hexer.Enabled,
		HexerModules: existing.Hexer.Modules,
	}

	named := opts.Repos
	if opts.Repo != "" && !slices.Contains(named, opts.Repo) {
		named = append(named, opts.Repo)
	}
	if len(named) > 0 && !opts.Interactive {
		// Named repos add to the task rather than replacing it, so
		// `start KEY tds-hexer` on an existing ticket does not silently drop
		// the repos already enlisted.
		apply.Selected = append(existing.RepoNamesInOrder(), filterNew(named, existing)...)
		return apply, nil
	}

	return runPicker(cfg, store, ticket, existing, apply)
}

func filterNew(named []string, existing task.Task) []string {
	var out []string
	for _, name := range named {
		if !existing.HasRepo(name) && !slices.Contains(out, name) {
			out = append(out, name)
		}
	}
	return out
}

// runPicker shows the EDIT REPOS screen, looping when the user asks to clone a
// repo so the new one is immediately selectable.
func runPicker(cfg *config.Config, store *task.Store, ticket *internalJira.Ticket, existing task.Task, apply applyOptions) (applyOptions, error) {
	for {
		sources, err := discoverSourceRepos(cfg.Agent.ReposRoot)
		if err != nil {
			return applyOptions{}, err
		}
		if len(sources) == 0 {
			return applyOptions{}, fmt.Errorf("no repos found under %s", expandPath(cfg.Agent.ReposRoot))
		}

		repos := make([]repopicker.Repo, 0, len(sources))
		for _, s := range sources {
			entry := repopicker.Repo{Name: s.Name, InTask: existing.HasRepo(s.Name)}
			if r, ok := existing.Repo(s.Name); ok {
				entry.Base = r.Base
			}
			if def, err := resolveDefaultBase(s.Path); err == nil {
				entry.Branches = listBaseBranches(s.Path, def, cfg.Agent.BaseBranchPatterns)
			}
			repos = append(repos, entry)
		}

		title := "EDIT REPOS · " + ticket.Key
		if ticket.Summary != "" {
			title += " · " + ticket.Summary
		}

		result, err := repopicker.Run(repopicker.Options{
			Title:          title,
			Repos:          orderByTask(repos, existing),
			Base:           apply.Base,
			HexerAvailable: cfg.Agent.Hexer.Enabled,
			HexerEnabled:   existing.Hexer.Enabled,
			HexerModules:   cfg.Agent.Hexer.ModuleNames(),
			HexerSelected:  existing.Hexer.Modules,
		})
		if err != nil {
			return applyOptions{}, err
		}

		if result.CloneRequest != "" {
			url := cloneURL(result.CloneRequest, cfg.Agent.RepoURLTemplate)
			if !ui.ConfirmAction("Clone " + url + "?") {
				continue
			}
			cloned, err := cloneRepo(cfg.Agent.ReposRoot, url)
			if err != nil {
				fmt.Println(ui.Warning(err.Error()))
				continue
			}
			fmt.Println(ui.SuccessMsg("cloned " + cloned.Name))
			continue
		}

		if !result.Confirmed {
			return applyOptions{}, fmt.Errorf("cancelled")
		}

		apply.Selected = result.Selected
		apply.Base = result.Base
		apply.Bases = result.Bases
		apply.Hexer = result.Hexer
		apply.HexerModules = result.HexerModules
		return apply, nil
	}
}

// orderByTask puts the repos already in the task first, in task order, so the
// picker's pre-checked rows match the session's window order.
func orderByTask(repos []repopicker.Repo, existing task.Task) []repopicker.Repo {
	byName := map[string]repopicker.Repo{}
	for _, r := range repos {
		byName[r.Name] = r
	}
	out := make([]repopicker.Repo, 0, len(repos))
	for _, name := range existing.RepoNamesInOrder() {
		if r, ok := byName[name]; ok {
			out = append(out, r)
			delete(byName, name)
		}
	}
	for _, r := range repos {
		if _, ok := byName[r.Name]; ok {
			out = append(out, r)
		}
	}
	return out
}
