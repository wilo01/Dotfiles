package agent

import (
	"errors"
	"fmt"
	"testing"

	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/task"
)

// storeWith builds a registry holding a task per key, each with one repo so it
// counts as restorable.
func storeWith(t *testing.T, keys ...string) *task.Store {
	t.Helper()
	store := newStore(t)
	for _, key := range keys {
		entry, err := task.New(key, "summary of "+key, "/tasks/"+key)
		if err != nil {
			t.Fatal(err)
		}
		entry.AddRepo(task.Repo{Name: "tds-suite", Worktree: "/tasks/" + key + "/tds-suite"})
		store.Upsert(entry)
	}
	return store
}

func ticket(key string, opts ...func(*internalJira.Ticket)) internalJira.Ticket {
	tk := internalJira.Ticket{Key: key, Summary: "summary of " + key, Status: "In Progress"}
	for _, o := range opts {
		o(&tk)
	}
	return tk
}

func subtask(tk *internalJira.Ticket) { tk.IsSubtask = true }

func byKey(t *testing.T, items []bulkTicket, key string) bulkTicket {
	t.Helper()
	for _, item := range items {
		if item.Ticket.Key == key {
			return item
		}
	}
	t.Fatalf("no planned item for %s", key)
	return bulkTicket{}
}

func TestPlanBulkSplitsRestoresFromNewTickets(t *testing.T) {
	store := storeWith(t, "SUITE-1")
	tickets := []internalJira.Ticket{ticket("SUITE-1"), ticket("SUITE-2")}

	items := planBulk(tickets, store, startOptions{})

	if got := byKey(t, items, "SUITE-1"); !got.HasTask || got.action() != "restore" {
		t.Errorf("SUITE-1 should restore, got %q", got.action())
	}
	if got := byKey(t, items, "SUITE-2"); got.HasTask || got.action() != "new — pick repos" {
		t.Errorf("SUITE-2 should be new, got %q", got.action())
	}

	restores, creates := countActions(items)
	if restores != 1 || creates != 1 {
		t.Errorf("counts = %d restores / %d creates, want 1/1", restores, creates)
	}
}

func TestPlanBulkSkipsSubtasksByDefault(t *testing.T) {
	tickets := []internalJira.Ticket{ticket("SUITE-1"), ticket("SUITE-2", subtask)}

	items := planBulk(tickets, newStore(t), startOptions{})
	if got := byKey(t, items, "SUITE-2").Skip; got != "subtask" {
		t.Errorf("subtask not skipped: %q", got)
	}

	items = planBulk(tickets, newStore(t), startOptions{IncludeSubtasks: true})
	if got := byKey(t, items, "SUITE-2").Skip; got != "" {
		t.Errorf("--include-subtasks ignored: %q", got)
	}
}

// A task whose repos were all removed cannot be restored, and must say so
// rather than failing inside ensureSession.
func TestPlanBulkSkipsTasksWithNoRepos(t *testing.T) {
	store := newStore(t)
	empty, _ := task.New("SUITE-1", "x", "/tasks/SUITE-1")
	store.Upsert(empty)

	items := planBulk([]internalJira.Ticket{ticket("SUITE-1")}, store, startOptions{})

	if got := byKey(t, items, "SUITE-1").Skip; got != "task has no repos" {
		t.Errorf("skip = %q", got)
	}
	if restores, creates := countActions(items); restores != 0 || creates != 0 {
		t.Errorf("a repo-less task was counted: %d/%d", restores, creates)
	}
}

// --max guards the interactive creations; restoring is prompt-free, so a reboot
// recovery is never truncated by it.
func TestApplyBulkLimitCapsCreationsOnly(t *testing.T) {
	store := storeWith(t, "SUITE-1", "SUITE-2", "SUITE-3")
	tickets := []internalJira.Ticket{
		ticket("SUITE-1"), ticket("SUITE-2"), ticket("SUITE-3"),
		ticket("NEW-1"), ticket("NEW-2"), ticket("NEW-3"),
	}

	items := planBulk(tickets, store, startOptions{Max: 2})

	restores, creates := countActions(items)
	if restores != 3 {
		t.Errorf("restores capped by --max: got %d, want 3", restores)
	}
	if creates != 2 {
		t.Errorf("creates = %d, want 2", creates)
	}
	if got := byKey(t, items, "NEW-3").Skip; got != "over --max limit" {
		t.Errorf("third new ticket not capped: %q", got)
	}
}

func TestApplyBulkLimitZeroMeansUnlimited(t *testing.T) {
	tickets := []internalJira.Ticket{ticket("NEW-1"), ticket("NEW-2"), ticket("NEW-3")}

	items := planBulk(tickets, newStore(t), startOptions{Max: 0})

	if _, creates := countActions(items); creates != 3 {
		t.Errorf("creates = %d, want all 3", creates)
	}
}

func TestReposCellShowsRestoredRepos(t *testing.T) {
	store := storeWith(t, "SUITE-1")
	items := planBulk([]internalJira.Ticket{ticket("SUITE-1"), ticket("SUITE-2")}, store, startOptions{})

	if got := byKey(t, items, "SUITE-1").reposCell(); got != "tds-suite" {
		t.Errorf("restored repos = %q", got)
	}
	if got := byKey(t, items, "SUITE-2").reposCell(); got != "-" {
		t.Errorf("a new ticket has no repos yet, got %q", got)
	}
}

func TestBulkActionLabels(t *testing.T) {
	cases := []struct {
		item bulkTicket
		want string
	}{
		{bulkTicket{Skip: "subtask"}, "skip: subtask"},
		{bulkTicket{HasTask: true}, "restore"},
		{bulkTicket{}, "new — pick repos"},
	}
	for _, c := range cases {
		if got := c.item.action(); got != c.want {
			t.Errorf("action() = %q, want %q", got, c.want)
		}
	}
}

// A cancelled picker must be distinguishable from a real failure, so one esc
// skips a ticket instead of abandoning the rest of the run.
func TestCancelledIsASentinelNotAString(t *testing.T) {
	if !errors.Is(errCancelled, errCancelled) {
		t.Fatal("errCancelled is not comparable with errors.Is")
	}
	wrapped := fmt.Errorf("picking repos for SUITE-1: %w", errCancelled)
	if !errors.Is(wrapped, errCancelled) {
		t.Error("a wrapped cancellation is no longer recognised")
	}
	if errors.Is(errors.New("cancelled"), errCancelled) {
		t.Error("a lookalike error must not be treated as a cancellation")
	}
}
