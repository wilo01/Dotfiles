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
	fanoutSpin            bool
)

var fanoutCmd = &cobra.Command{
	Use:   "fanout",
	Short: "Open a worktree + tmux session with claude for every assigned ticket",
	Long: `Runs the configured JQL (default: all your open tickets), shows the plan,
and opens one worktree + tmux session per ticket with an interactive claude,
staggered a few seconds apart.

Only tickets whose repo can be resolved from an existing branch are opened;
the rest are listed for "hlp agent start <KEY>", which lets you pick repos
yourself. --spin spawns autonomous /agent-run agents instead.`,
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

	sources, err := discoverSourceRepos(cfg.Agent.ReposRoot)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no repos found under %s", expandPath(cfg.Agent.ReposRoot))
	}
	repos := make([]string, 0, len(sources))
	for _, s := range sources {
		repos = append(repos, s.Name)
	}

	if fanoutOpts.Repo != "" && !slices.Contains(repos, fanoutOpts.Repo) {
		return fmt.Errorf("repo %s not found under %s (known: %s)", fanoutOpts.Repo, expandPath(cfg.Agent.ReposRoot), strings.Join(repos, ", "))
	}

	stop := ui.Spinner("Resolving target repos...")
	preResolve(cfg, candidates, repos, fanoutOpts, fanoutMax)
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

	return spawnAll(cfg, candidates)
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
// table can show it.
//
// Resolution is the branch scan only: a ticket with no branch anywhere has no
// unattended answer, and picking repos for it is what `hlp agent start` is for.
func preResolve(cfg *config.Config, candidates []candidate, repos []string, opts startOptions, max int) {
	// --repo and --interactive settle every ticket the same way, so the branch
	// scan has nothing left to decide.
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
		if c.Skip == "" && c.Repo == "" && !c.NeedsPicker {
			c.Skip = "no branch yet"
		}
	}

	applyMaxLimit(cfg, candidates, max)
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

func printPlan(candidates []candidate, jql string) {
	var rows [][]string
	needsPicking := false
	for i := range candidates {
		c := &candidates[i]
		summary := c.Ticket.Summary
		if len(summary) > 50 {
			summary = summary[:47] + "..."
		}
		rows = append(rows, []string{c.Ticket.Key, c.Ticket.IssueType, c.Ticket.Status, summary, c.repoCell(), c.action()})
		if c.Skip == "no branch yet" {
			needsPicking = true
		}
	}

	fmt.Println(ui.Table([]string{"KEY", "TYPE", "STATUS", "SUMMARY", "REPO", "ACTION"}, rows))
	fmt.Println(ui.KeyValue("JQL", jql))

	for i := range candidates {
		for _, w := range candidates[i].Warnings {
			fmt.Println(ui.Warning(w))
		}
	}
	if needsPicking {
		fmt.Println(ui.Info("no branch yet for some tickets: run `hlp agent start <KEY>` to pick their repos"))
	}
}

func spawnAll(cfg *config.Config, candidates []candidate) error {
	var failures []string
	launched := 0

	for i := range candidates {
		c := &candidates[i]
		if c.Skip != "" {
			continue
		}

		opts := fanoutOpts
		opts.Manual = !fanoutSpin
		// Attaching mid-loop would hand the terminal away and strand the rest.
		opts.NoAttach = true
		opts.ContextPrompt = resolveContextPrompt(cfg, fanoutOpts.ContextPrompt)
		opts.Repo = c.Repo
		// The picker is interactive, so it runs here rather than in the
		// concurrent pre-resolve pass - and never under --dry-run, which must
		// not prompt for anything.
		opts.Interactive = c.NeedsPicker
		if c.NeedsPicker && fanoutOpts.DryRun {
			fmt.Println(ui.Header(c.Ticket.Key + " - " + c.Ticket.Summary))
			fmt.Println(ui.Info("dry-run: repos would be chosen interactively"))
			fmt.Println()
			launched++
			continue
		}

		if launched > 0 && !fanoutOpts.DryRun {
			time.Sleep(4 * time.Second)
		}
		if err := startAgent(cfg, &c.Ticket, opts); err != nil {
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
