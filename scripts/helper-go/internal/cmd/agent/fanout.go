package agent

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

// fanoutOpts carries the flags shared with start/spin; the vars below are the
// ones that only make sense for a bulk run.
var (
	fanoutOpts            = startOptions{Manual: true}
	fanoutJQL             string
	fanoutMax             int
	fanoutIncludeSubtasks bool
	fanoutYes             bool
	fanoutTriage          bool
	fanoutSpin            bool
)

var fanoutCmd = &cobra.Command{
	Use:   "fanout",
	Short: "Open a worktree + tmux session with claude for every assigned ticket",
	Long: `Runs the configured JQL (default: all your open tickets), shows the plan,
and opens one worktree + tmux session per ticket with an interactive claude,
staggered a few seconds apart.

Only tickets that already have a branch are opened by default; the rest are
listed, since guessing a repo for work that hasn't started is what --triage is
for. --spin restores the old behaviour of spawning autonomous /agent-run agents.`,
	// Without this, `--context "my prompt"` parses as a bare --context plus a
	// stray positional, silently using the configured prompt instead of yours.
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFanout()
	},
}

func init() {
	registerLaunchFlags(fanoutCmd, &fanoutOpts)
	fanoutCmd.Flags().StringVar(&fanoutJQL, "jql", "", "override the configured fanout JQL")
	fanoutCmd.Flags().IntVar(&fanoutMax, "max", 0, "spawn at most N agents (default: agent.max_parallel)")
	fanoutCmd.Flags().BoolVar(&fanoutIncludeSubtasks, "include-subtasks", false, "also spawn agents for subtasks")
	fanoutCmd.Flags().BoolVar(&fanoutYes, "yes", false, "skip the confirmation prompt")
	fanoutCmd.Flags().BoolVar(&fanoutTriage, "triage", false, "AI-triage the target repo for unstarted tickets (agent.triage_statuses)")
	fanoutCmd.Flags().BoolVar(&fanoutSpin, "spin", false, "spawn autonomous /agent-run agents instead of interactive sessions")
}

// candidate is one ticket from the JQL plus the verdict on what to do with it.
type candidate struct {
	Ticket internalJira.Ticket
	Skip   string
	Repo   string
	// How describes where Repo came from, shown as the start header's provenance.
	How string
	// NeedsPicker means the repo could not be resolved unattended and the user
	// must choose one when this ticket is spawned.
	NeedsPicker bool
	Warnings    []string
}

func (c *candidate) action() string {
	if c.Skip != "" {
		return "skip: " + c.Skip
	}
	return "spawn"
}

func (c *candidate) repoCell() string {
	switch {
	case c.Repo != "":
		return c.Repo
	case c.NeedsPicker:
		return "picker"
	default:
		return "-"
	}
}

// validateFanoutFlags defers to the shared launch rules, with --spin standing in
// for "not Manual" so the same combinations are rejected here as on spin itself.
func validateFanoutFlags(opts startOptions, spin bool) error {
	opts.Manual = !spin
	return validateLaunchFlags(&opts)
}

func runFanout() error {
	if err := validateFanoutFlags(fanoutOpts, fanoutSpin); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client, err := getJiraClient()
	if err != nil {
		return err
	}

	jql := cfg.Agent.JQL
	if fanoutJQL != "" {
		jql = fanoutJQL
	}

	tickets, err := client.Search(jql, 50)
	if err != nil {
		return err
	}
	if len(tickets) == 0 {
		fmt.Println(ui.Info("no tickets matched: " + jql))
		return nil
	}

	candidates := buildCandidates(tickets)

	repos, err := discoverRepos(cfg.Agent.WorktreeRoot)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return fmt.Errorf("no repos found under %s", cfg.Agent.WorktreeRoot)
	}

	if fanoutOpts.Repo != "" && !slices.Contains(repos, fanoutOpts.Repo) {
		return fmt.Errorf("repo %s not found under %s (known: %s)", fanoutOpts.Repo, cfg.Agent.WorktreeRoot, strings.Join(repos, ", "))
	}

	stop := ui.Spinner("Resolving target repos...")
	preResolve(cfg, client, candidates, repos, fanoutOpts, fanoutTriage, fanoutMax, triageRepo)
	stop()

	printPlan(candidates, jql)

	spawnable := 0
	for i := range candidates {
		if candidates[i].Skip == "" {
			spawnable++
		}
	}
	if spawnable == 0 {
		fmt.Println(ui.Info("nothing to spawn"))
		return nil
	}
	if !fanoutOpts.DryRun && !fanoutYes {
		if !ui.ConfirmAction(fmt.Sprintf("Spawn %d session(s)?", spawnable)) {
			return nil
		}
	}

	return spawnAll(cfg, client, candidates, repos)
}

// buildCandidates turns search results into candidates, marking the ones that
// can never be spawned regardless of how their repo resolves.
func buildCandidates(tickets []internalJira.Ticket) []candidate {
	candidates := make([]candidate, 0, len(tickets))
	for _, t := range tickets {
		c := candidate{Ticket: t}
		switch {
		case t.IsSubtask && !fanoutIncludeSubtasks:
			c.Skip = "subtask"
		case sessionExists(t.Key) && agentRunning(t.Key) && !fanoutOpts.Relaunch:
			c.Skip = "agent already running"
		}
		candidates = append(candidates, c)
	}
	return candidates
}

// preResolve fills in each candidate's repo ahead of the spawn loop, so the plan
// table can show it and so triage never runs unattended by default.
//
// It runs in three phases because the --max cut has to sit between the free and
// the paid one: scanning for an existing branch costs nothing, so it happens for
// every ticket, and only what survives that (and the limit) is worth a model
// call. Cutting earlier would let tickets that turn out to be unspawnable eat
// slots that a resumable ticket further down the list could have used.
func preResolve(cfg *config.Config, client *internalJira.Client, candidates []candidate, repos []string, opts startOptions, triage bool, max int, fn triageFunc) {
	// --repo and --interactive settle every ticket the same way, so neither the
	// branch scan nor triage has anything left to decide.
	switch {
	case opts.Repo != "":
		for i := range candidates {
			candidates[i].Repo, candidates[i].How = opts.Repo, "specified"
		}
		applyMaxLimit(cfg, candidates, max)
		return
	case opts.Interactive:
		for i := range candidates {
			if candidates[i].Skip == "" {
				candidates[i].NeedsPicker = true
			}
		}
		applyMaxLimit(cfg, candidates, max)
		return
	}

	concurrently(cfg, candidates, func(c *candidate) {
		scanForBranch(cfg, c, repos)
	})

	for i := range candidates {
		c := &candidates[i]
		if c.Skip != "" || c.Repo != "" || c.NeedsPicker {
			continue
		}
		if !triage {
			c.Skip = "needs triage"
		} else if !statusAllowsTriage(cfg, c.Ticket.Status) {
			c.Skip = "not " + strings.Join(cfg.Agent.TriageStatuses, "/")
		}
	}

	applyMaxLimit(cfg, candidates, max)

	if !triage {
		return
	}
	concurrently(cfg, candidates, func(c *candidate) {
		if c.Repo == "" && !c.NeedsPicker {
			triageCandidate(cfg, client, c, repos, fn)
		}
	})
}

// concurrently runs work over every candidate that is still in play, bounded by
// agent.max_parallel. Nothing inside work may print: the spinner owns stdout.
func concurrently(cfg *config.Config, candidates []candidate, work func(*candidate)) {
	sem := make(chan struct{}, max(cfg.Agent.MaxParallel, 1))
	var wg sync.WaitGroup

	for i := range candidates {
		if candidates[i].Skip != "" {
			continue
		}
		wg.Add(1)
		go func(c *candidate) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			work(c)
		}(&candidates[i])
	}
	wg.Wait()
}

func applyMaxLimit(cfg *config.Config, candidates []candidate, max int) {
	if max <= 0 {
		max = cfg.Agent.MaxParallel
	}
	spawnable := 0
	for i := range candidates {
		c := &candidates[i]
		if c.Skip != "" {
			continue
		}
		if spawnable >= max {
			c.Skip = "over --max limit"
			continue
		}
		spawnable++
	}
}

func scanForBranch(cfg *config.Config, c *candidate, repos []string) {
	matches := resumeScan(cfg, repos, c.Ticket.Key)
	switch len(matches) {
	case 0:
	case 1:
		c.Repo, c.How = matches[0], "resume: existing branch"
	default:
		c.Warnings = append(c.Warnings, fmt.Sprintf("branch %s found in multiple repos: %s", c.Ticket.Key, strings.Join(matches, ", ")))
		c.NeedsPicker = true
	}
}

func triageCandidate(cfg *config.Config, client *internalJira.Client, c *candidate, repos []string, fn triageFunc) {
	verdict, err := fn(cfg, client, &c.Ticket, repos, func(msg string) {
		c.Warnings = append(c.Warnings, c.Ticket.Key+": "+msg)
	})
	if err != nil {
		c.Warnings = append(c.Warnings, c.Ticket.Key+": triage failed: "+err.Error())
		c.NeedsPicker = true
		return
	}
	if verdict.Confidence == "low" {
		c.Warnings = append(c.Warnings, fmt.Sprintf("%s: low triage confidence (%s)", c.Ticket.Key, verdict.Reason))
		c.NeedsPicker = true
		return
	}
	c.Repo, c.How = verdict.Repo, "AI triage: "+verdict.Reason
}

// statusAllowsTriage reports whether a ticket is early enough in the workflow
// that guessing its repo is meaningful. Anything past these columns either
// already has a branch or is too far along to guess at.
func statusAllowsTriage(cfg *config.Config, status string) bool {
	return slices.ContainsFunc(cfg.Agent.TriageStatuses, func(s string) bool {
		return strings.EqualFold(s, status)
	})
}

func printPlan(candidates []candidate, jql string) {
	var rows [][]string
	needsTriage := false
	for i := range candidates {
		c := &candidates[i]
		summary := c.Ticket.Summary
		if len(summary) > 50 {
			summary = summary[:47] + "..."
		}
		rows = append(rows, []string{c.Ticket.Key, c.Ticket.IssueType, c.Ticket.Status, summary, c.repoCell(), c.action()})
		if c.Skip == "needs triage" {
			needsTriage = true
		}
	}

	fmt.Println(ui.Table([]string{"KEY", "TYPE", "STATUS", "SUMMARY", "REPO", "ACTION"}, rows))
	fmt.Println(ui.KeyValue("JQL", jql))

	for i := range candidates {
		for _, w := range candidates[i].Warnings {
			fmt.Println(ui.Warning(w))
		}
	}
	if needsTriage {
		fmt.Println(ui.Info("no branch yet for some tickets: run `hlp agent start <KEY>` for those, or --triage to resolve them here"))
	}
}

func spawnAll(cfg *config.Config, client *internalJira.Client, candidates []candidate, repos []string) error {
	var failures []string
	launched := 0

	for i := range candidates {
		c := &candidates[i]
		if c.Skip != "" {
			continue
		}

		opts := fanoutOpts
		opts.Manual = !fanoutSpin
		opts.ContextPrompt = resolveContextPrompt(cfg, fanoutOpts.ContextPrompt)
		opts.Repo, opts.RepoHow = c.Repo, c.How
		// The per-ticket picker below replaces the blanket one, and startAgent
		// must not re-resolve a repo the pre-pass already settled.
		opts.Interactive = false
		// The picker is interactive, so it has to run here rather than in the
		// concurrent pre-resolve pass - and never under --dry-run, which must
		// not prompt for anything.
		if c.NeedsPicker {
			if fanoutOpts.DryRun {
				fmt.Println(ui.Header(c.Ticket.Key + " - " + c.Ticket.Summary))
				fmt.Println(ui.Info("dry-run: repo would be chosen interactively"))
				fmt.Println()
				launched++
				continue
			}
			fmt.Println(ui.Header(c.Ticket.Key + " - " + c.Ticket.Summary))
			repo, err := pickRepo(repos)
			if err != nil {
				failures = append(failures, c.Ticket.Key+": "+err.Error())
				fmt.Println(ui.Error(c.Ticket.Key + ": " + err.Error()))
				continue
			}
			opts.Repo, opts.RepoHow = repo, "picked"
		}

		if launched > 0 && !fanoutOpts.DryRun {
			time.Sleep(4 * time.Second)
		}
		if err := startAgent(cfg, client, &c.Ticket, opts); err != nil {
			failures = append(failures, c.Ticket.Key+": "+err.Error())
			fmt.Println(ui.Error(c.Ticket.Key + ": " + err.Error()))
			continue
		}
		launched++
		fmt.Println()
	}

	fmt.Println(ui.Header("Fanout summary"))
	launchedLabel := "Launched"
	if fanoutOpts.DryRun {
		launchedLabel = "Planned (dry-run)"
	}
	fmt.Println(ui.KeyValue(launchedLabel, fmt.Sprintf("%d", launched)))
	if len(failures) > 0 {
		fmt.Println(ui.Error("Failures:\n  " + strings.Join(failures, "\n  ")))
	}
	return nil
}
