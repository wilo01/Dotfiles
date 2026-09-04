package agent

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/hexer"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/task"
	"github.com/dariuszw/hlp/internal/tui/taskpicker"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	openRelaunch bool
	openNoAttach bool
	editNoAttach bool
	editOpts     = startOptions{Manual: true}
	rmForce      bool
	rmYes        bool
	rmKeepDB     bool
)

var openCmd = &cobra.Command{
	Use:   "open [TICKET-KEY]",
	Short: "Rebuild and attach the tmux session for a ticket from tasks.json",
	Long: `Recreates the ticket's tmux session from the stored task: one window per
repo, each in that repo's worktree, then resumes its Claude session.

Without a key it lists your tasks, most recently opened first, to pick from.

Reads only local state - no Jira call, no network.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		store, err := loadStore()
		if err != nil {
			return err
		}
		t, err := selectTask(store, args, "OPEN TASK")
		if err != nil {
			return err
		}

		reportDrift(t)

		return ensureSession(cfg, store, t, launchOptions{
			Manual:   true,
			Relaunch: openRelaunch,
			NoAttach: openNoAttach,
		})
	},
}

var editCmd = &cobra.Command{
	Use:   "edit [TICKET-KEY]",
	Short: "Add or remove repos on an existing ticket (and its hexer env)",
	Long: `Opens the repo picker pre-checked with the ticket's current repos.

Checking a repo creates its worktree and appends a tmux window; unchecking one
removes both, after a confirmation. Dirty or unpushed worktrees are refused
unless --force.

Without a key it lists your tasks to pick from.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		store, err := loadStore()
		if err != nil {
			return err
		}
		existing, err := selectTask(store, args, "EDIT REPOS · pick a task")
		if err != nil {
			return err
		}
		key := existing.JiraKey

		// The summary is already stored, so editing needs no Jira round trip.
		ticket := &internalJira.Ticket{Key: key, Summary: existing.Summary}

		apply, err := runPicker(cfg, store, ticket, existing, applyOptions{
			Key:          key,
			Summary:      existing.Summary,
			Description:  existing.Description,
			Base:         editOpts.Base,
			DryRun:       editOpts.DryRun,
			Force:        editOpts.Force,
			Hexer:        existing.Hexer.Enabled,
			HexerModules: existing.Hexer.Modules,
		})
		if err != nil {
			return err
		}

		t, err := applyTask(cfg, store, apply)
		if err != nil {
			return err
		}
		if editOpts.DryRun {
			return nil
		}
		if len(t.Repos) == 0 {
			fmt.Println(ui.Warning(key + " now has no repos; its tmux session is left as-is"))
			return nil
		}
		return ensureSession(cfg, store, t, launchOptions{Manual: true, NoAttach: editNoAttach})
	},
}

var rmCmd = &cobra.Command{
	Use:     "rm [TICKET-KEY]",
	Aliases: []string{"remove"},
	Short:   "Delete a ticket's worktrees, hexer env and tmux session",
	Long: `Tears down the hexer environment, removes every worktree, deletes the task
directory and drops the task from tasks.json.

Refuses to discard uncommitted or unpushed work unless --force. Branches
created inside the worktrees are never deleted from the source repos.

Without a key it lists your tasks to pick from.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		store, err := loadStore()
		if err != nil {
			return err
		}
		t, err := selectTask(store, args, "REMOVE TASK")
		if err != nil {
			return err
		}
		return removeTask(cfg, t.JiraKey)
	},
}

func removeTask(cfg *config.Config, key string) error {
	store, err := loadStore()
	if err != nil {
		return err
	}
	t, ok := store.Get(key)
	if !ok {
		return fmt.Errorf("no task for %s", key)
	}

	fmt.Println(ui.Header("Remove " + key))
	fmt.Println(ui.KeyValue("Task root", t.TaskRoot))
	var risky []string
	for _, r := range t.Repos {
		note := ""
		switch {
		case worktreeDirty(r.Worktree):
			note, risky = " — uncommitted changes", append(risky, r.Name)
		case worktreeUnpushed(r.Worktree):
			note, risky = " — unpushed commits", append(risky, r.Name)
		}
		fmt.Println(ui.KeyValue("  "+r.Name, r.Worktree+note))
	}
	if t.Hexer.Enabled {
		fmt.Println(ui.KeyValue("Hexer", fmt.Sprintf("%s (ports %d/%d)", t.Hexer.Host, t.Hexer.Port, t.Hexer.DBPort)))
	}

	if len(risky) > 0 && !rmForce {
		return fmt.Errorf("refusing to remove %s — %s would lose work; pass --force to discard",
			key, strings.Join(risky, ", "))
	}
	if !rmYes && !ui.ConfirmAction("Delete everything above?") {
		return nil
	}

	if t.Hexer.Enabled {
		runner, err := hexer.New(config.GetConfigDir(), cfg.Agent.Hexer, hexerProgress)
		if err != nil {
			return err
		}
		// The container is bound to the worktree path, so it has to go first.
		if err := runner.Down(hexer.DownOptions{
			Slug:      hexerSlug(key),
			TaskRoot:  t.TaskRoot,
			HexerPort: t.Hexer.Port,
			DBPort:    t.Hexer.DBPort,
			KeepDB:    rmKeepDB,
		}); err != nil {
			return fmt.Errorf("hexer teardown failed (nothing was removed): %w", err)
		}
		fmt.Println(ui.SuccessMsg("hexer environment torn down"))
	}

	for _, r := range t.Repos {
		if err := removeWorktree(r.Source, r.Worktree, rmForce); err != nil {
			return err
		}
		fmt.Println(ui.SuccessMsg(r.Name + " worktree removed"))
	}

	if err := os.RemoveAll(t.TaskRoot); err != nil {
		return fmt.Errorf("failed to remove %s: %w", t.TaskRoot, err)
	}
	if sessionExists(key) {
		if err := killSession(key); err != nil {
			fmt.Println(ui.Warning(err.Error()))
		}
	}

	store.Delete(key)
	if err := store.Save(); err != nil {
		return err
	}
	fmt.Println(ui.SuccessMsg(key + " removed"))
	return nil
}

// selectTask resolves the ticket a command should act on: the key given on the
// command line, or one chosen from the list when none was.
func selectTask(store *task.Store, args []string, title string) (task.Task, error) {
	if len(args) == 1 {
		key := strings.ToUpper(args[0])
		t, ok := store.Get(key)
		if !ok {
			return task.Task{}, fmt.Errorf("no task for %s — run: hlp agent start %s", key, key)
		}
		return t, nil
	}

	// The picker needs a terminal to draw on, so from a script or a pipe the
	// key is not optional.
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return task.Task{}, fmt.Errorf("no ticket key given, and there is no terminal to pick one from — pass the key explicitly")
	}

	key, err := taskpicker.Run(title, taskItems(store))
	if err != nil {
		return task.Task{}, err
	}
	if key == "" {
		return task.Task{}, errCancelled
	}
	t, ok := store.Get(key)
	if !ok {
		return task.Task{}, fmt.Errorf("no task for %s", key)
	}
	return t, nil
}

// taskItems renders the store for the picker, newest-opened first, annotating
// each task with its live tmux state so an in-progress ticket is obvious.
func taskItems(store *task.Store) []taskpicker.Item {
	tasks := store.All()
	items := make([]taskpicker.Item, 0, len(tasks))
	for _, t := range tasks {
		session := ""
		if sessionExists(t.SessionName()) {
			session = "idle"
			if primary, ok := t.PrimaryRepo(); ok && claudeRunningInWindow(t.SessionName(), primary.Name) {
				session = "claude"
			}
		}
		items = append(items, taskpicker.Item{
			Key:        t.JiraKey,
			Summary:    t.Summary,
			Repos:      t.RepoNamesInOrder(),
			Session:    session,
			Hexer:      t.Hexer.Host,
			LastOpened: humanAge(t.LastOpenedMs),
		})
	}
	return items
}

// humanAge renders a millisecond timestamp as a short relative age.
func humanAge(ms int64) string {
	if ms == 0 {
		return ""
	}
	d := time.Since(time.UnixMilli(ms))
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// reportDrift surfaces disagreements between the registry and the disk. It never
// repairs them: a missing worktree is something to decide about, not to silently
// recreate.
func reportDrift(t task.Task) {
	drift := task.Reconcile(t)
	if len(drift) == 0 {
		return
	}
	for _, d := range drift {
		switch d.Kind {
		case task.DriftTaskRootMissing:
			fmt.Println(ui.Warning("task root is missing: " + d.Path))
		case task.DriftWorktreeMissing:
			fmt.Println(ui.Warning(fmt.Sprintf("%s worktree is missing: %s (re-run start to recreate)", d.Repo, d.Path)))
		case task.DriftSourceMissing:
			fmt.Println(ui.Warning(fmt.Sprintf("%s source checkout is missing: %s", d.Repo, d.Path)))
		case task.DriftUntracked:
			fmt.Println(ui.Warning(fmt.Sprintf("%s is on disk but not in the task: %s (add it with edit)", d.Repo, d.Path)))
		}
	}
}

func init() {
	openCmd.Flags().BoolVar(&openRelaunch, "relaunch", false, "send the claude launch even if one already runs in the session")
	openCmd.Flags().BoolVar(&openNoAttach, "no-attach", false, "rebuild the session but stay where you are")
	editCmd.Flags().BoolVar(&editNoAttach, "no-attach", false, "apply the changes but stay where you are")

	editCmd.Flags().StringVar(&editOpts.Base, "base", "", "base branch for newly added repos (default: each repo's origin/HEAD)")
	editCmd.Flags().BoolVar(&editOpts.DryRun, "dry-run", false, "print the plan without creating or removing anything")
	editCmd.Flags().BoolVar(&editOpts.Force, "force", false, "allow removing worktrees with uncommitted or unpushed work")

	rmCmd.Flags().BoolVar(&rmForce, "force", false, "discard uncommitted or unpushed work")
	rmCmd.Flags().BoolVar(&rmYes, "yes", false, "skip the confirmation prompt")
	rmCmd.Flags().BoolVar(&rmKeepDB, "keep-db", false, "keep the hexer database container")
}
