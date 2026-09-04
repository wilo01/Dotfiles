package agent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/task"
	"github.com/dariuszw/hlp/internal/ui"
)

// errCancelled marks a picker the user backed out of. A bulk run treats it as
// "skip this ticket" rather than a failure, so one esc does not abandon the
// rest of the tickets.
var errCancelled = errors.New("cancelled")

// bulkTicket is one ticket from the JQL and what the run intends to do with it.
type bulkTicket struct {
	Ticket internalJira.Ticket
	// Existing is the task already recorded for this ticket, if any.
	Existing task.Task
	HasTask  bool
	Skip     string
}

func (b bulkTicket) action() string {
	switch {
	case b.Skip != "":
		return "skip: " + b.Skip
	case b.HasTask:
		return "restore"
	default:
		return "new — pick repos"
	}
}

func (b bulkTicket) reposCell() string {
	if !b.HasTask {
		return "-"
	}
	return strings.Join(b.Existing.RepoNamesInOrder(), ",")
}

// startAll is `hlp agent start` with no ticket key: it runs the configured JQL
// and brings every assigned ticket up.
//
// Restores run first and without prompting, so recovering after a reboot is a
// single uninterrupted pass; only genuinely new tickets stop to ask which repos
// they need.
func startAll(cfg *config.Config, client *internalJira.Client, opts startOptions) error {
	jql := cfg.Agent.JQL
	if opts.JQL != "" {
		jql = opts.JQL
	}

	stop := ui.Spinner("Searching Jira...")
	tickets, err := client.Search(jql, 50)
	stop()
	if err != nil {
		return err
	}
	if len(tickets) == 0 {
		fmt.Println(ui.Info("no tickets matched: " + jql))
		return nil
	}

	store, err := loadStore()
	if err != nil {
		return err
	}

	items := planBulk(tickets, store, opts)
	printBulkPlan(items, jql)

	restores, creates := countActions(items)
	if restores+creates == 0 {
		fmt.Println(ui.Info("nothing to do"))
		return nil
	}
	if opts.DryRun {
		fmt.Println(ui.Info(fmt.Sprintf("dry-run: would restore %d and create %d session(s)", restores, creates)))
		return nil
	}
	if !opts.Yes && !ui.ConfirmAction(fmt.Sprintf("Restore %d and create %d session(s)?", restores, creates)) {
		return nil
	}

	done, failures := runBulk(cfg, store, items, opts)

	fmt.Println()
	fmt.Println(ui.Header("Summary"))
	fmt.Println(ui.KeyValue("Sessions up", fmt.Sprintf("%d", done)))
	if len(failures) > 0 {
		fmt.Println(ui.Error("Failures:\n  " + strings.Join(failures, "\n  ")))
	}

	if opts.NoAttach || done == 0 {
		return nil
	}
	return openFromPicker(cfg, opts)
}

// planBulk decides what to do with each ticket before anything runs, so the
// plan can be shown and confirmed as a whole.
func planBulk(tickets []internalJira.Ticket, store *task.Store, opts startOptions) []bulkTicket {
	items := make([]bulkTicket, 0, len(tickets))
	for _, ticket := range tickets {
		item := bulkTicket{Ticket: ticket}
		item.Existing, item.HasTask = store.Get(ticket.Key)

		switch {
		case ticket.IsSubtask && !opts.IncludeSubtasks:
			item.Skip = "subtask"
		case item.HasTask && len(item.Existing.Repos) == 0:
			item.Skip = "task has no repos"
		}
		items = append(items, item)
	}

	applyBulkLimit(items, opts.Max)
	return items
}

// applyBulkLimit caps how many sessions a single run brings up. Restores are
// cheap and prompt-free, so the limit only guards the interactive creations.
func applyBulkLimit(items []bulkTicket, max int) {
	if max <= 0 {
		return
	}
	created := 0
	for i := range items {
		item := &items[i]
		if item.Skip != "" || item.HasTask {
			continue
		}
		if created >= max {
			item.Skip = "over --max limit"
			continue
		}
		created++
	}
}

func countActions(items []bulkTicket) (restores, creates int) {
	for _, item := range items {
		switch {
		case item.Skip != "":
		case item.HasTask:
			restores++
		default:
			creates++
		}
	}
	return restores, creates
}

// runBulk restores every existing task first, then walks the new tickets so the
// prompt-free work is already done by the time anything asks a question.
func runBulk(cfg *config.Config, store *task.Store, items []bulkTicket, opts startOptions) (int, []string) {
	var failures []string
	done := 0

	for _, item := range items {
		if item.Skip != "" || !item.HasTask {
			continue
		}
		fmt.Println(ui.Info("restoring " + item.Ticket.Key))
		if err := ensureSession(cfg, store, item.Existing, launchOptions{Manual: true, NoAttach: true}); err != nil {
			failures = append(failures, item.Ticket.Key+": "+err.Error())
			fmt.Println(ui.Error(item.Ticket.Key + ": " + err.Error()))
			continue
		}
		done++
	}

	for _, item := range items {
		if item.Skip != "" || item.HasTask {
			continue
		}
		// No header here: applyTask prints one for the ticket it is about to
		// build, and two in a row reads like the run stuttered.
		fmt.Println()

		ticket := item.Ticket
		single := opts
		single.NoAttach = true
		single.Repos = nil

		if err := startAgent(cfg, &ticket, single); err != nil {
			if errors.Is(err, errCancelled) {
				fmt.Println(ui.Info("skipped " + item.Ticket.Key))
				continue
			}
			failures = append(failures, item.Ticket.Key+": "+err.Error())
			fmt.Println(ui.Error(item.Ticket.Key + ": " + err.Error()))
			continue
		}
		done++
	}
	return done, failures
}

// openFromPicker ends a bulk run by letting the user jump into one of the
// sessions it just brought up.
func openFromPicker(cfg *config.Config, opts startOptions) error {
	store, err := loadStore()
	if err != nil {
		return err
	}
	t, err := selectTask(store, nil, "OPEN TASK")
	if err != nil {
		// Declining to pick one is a normal way to end the run.
		if errors.Is(err, errCancelled) {
			return nil
		}
		return err
	}
	return ensureSession(cfg, store, t, launchOptions{Manual: true, Relaunch: opts.Relaunch})
}

func printBulkPlan(items []bulkTicket, jql string) {
	var rows [][]string
	for _, item := range items {
		summary := item.Ticket.Summary
		if len(summary) > 45 {
			summary = summary[:42] + "..."
		}
		rows = append(rows, []string{
			item.Ticket.Key, item.Ticket.Status, summary, item.reposCell(), item.action(),
		})
	}
	fmt.Println(ui.Table([]string{"KEY", "STATUS", "SUMMARY", "REPOS", "ACTION"}, rows))
	fmt.Println(ui.KeyValue("JQL", jql))
}
