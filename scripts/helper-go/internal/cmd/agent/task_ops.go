package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/hexer"
	"github.com/dariuszw/hlp/internal/task"
	"github.com/dariuszw/hlp/internal/ui"
)

// tasksPath is where the task registry lives, alongside the rest of the hlp
// config.
func tasksPath() string {
	return filepath.Join(config.GetConfigDir(), "tasks.json")
}

func loadStore() (*task.Store, error) {
	return task.Load(tasksPath())
}

// taskRootFor is the umbrella directory for a ticket. Repos always nest inside
// it, even when only one is enlisted, so adding a second repo later never has
// to move the first one.
func taskRootFor(cfg *config.Config, key string) string {
	return filepath.Join(expandPath(cfg.Agent.TasksRoot), strings.ToUpper(key))
}

// applyOptions describes the target state for a task.
type applyOptions struct {
	Key     string
	Summary string
	// Selected is the full set of repos the task should end up with, in the
	// order they were chosen; the first becomes the primary.
	Selected []string
	Base     string
	// Bases overrides Base for individual repos, so one ticket can sit on
	// master in one repo and a maintenance line in another.
	Bases        map[string]string
	Hexer        bool
	HexerModules []string
	DryRun       bool
	Force        bool
}

// applyTask reconciles a task to the requested repo set: it creates missing
// worktrees, removes dropped ones, and provisions or tears down the hexer env.
// It saves the registry only after the filesystem work succeeds.
func applyTask(cfg *config.Config, store *task.Store, opts applyOptions) (task.Task, error) {
	key := strings.ToUpper(opts.Key)

	current, existed := store.Get(key)
	if !existed {
		created, err := task.New(key, opts.Summary, taskRootFor(cfg, key))
		if err != nil {
			return task.Task{}, err
		}
		current = created
	}
	if opts.Summary != "" {
		current.Summary = opts.Summary
	}

	removed := slices.DeleteFunc(current.RepoNames(), func(name string) bool {
		return slices.Contains(opts.Selected, name)
	})

	plan, err := planRepos(cfg, current, opts)
	if err != nil {
		return task.Task{}, err
	}

	printTaskPlan(cfg, current, plan, removed, opts)
	if opts.DryRun {
		fmt.Println(ui.Info("dry-run: nothing was created or removed"))
		return current, nil
	}

	if len(removed) > 0 {
		if err := removeRepos(&current, removed, opts.Force); err != nil {
			return task.Task{}, err
		}
	}
	if err := createRepos(cfg, &current, plan); err != nil {
		return task.Task{}, err
	}

	// Selection order is the user's choice of primary, so re-order the stored
	// repos to match rather than leaving them in creation order.
	current.Repos = reorderRepos(current, opts.Selected)

	if err := applyHexer(cfg, store, &current, opts); err != nil {
		return task.Task{}, err
	}

	store.Upsert(current)
	if err := store.Save(); err != nil {
		return task.Task{}, err
	}
	return current, nil
}

// repoPlan is one repo that needs a worktree created.
type repoPlan struct {
	Name     string
	Source   string
	Worktree string
	Base     string
	Exists   bool
}

// planRepos resolves the source checkout and base branch for every selected
// repo before anything is created, so a bad repo name fails the whole run
// rather than leaving a half-built task.
func planRepos(cfg *config.Config, current task.Task, opts applyOptions) ([]repoPlan, error) {
	sources, err := discoverSourceRepos(cfg.Agent.ReposRoot)
	if err != nil {
		return nil, err
	}
	byName := map[string]string{}
	for _, s := range sources {
		byName[s.Name] = s.Path
	}

	var plan []repoPlan
	for _, name := range opts.Selected {
		source, ok := byName[name]
		if !ok {
			if override, has := cfg.Agent.Repos[name]; has && override.Source != "" {
				source = expandPath(override.Source)
			} else {
				return nil, fmt.Errorf("no repo named %q under %s (clone it first, or set agent.repos.%s.source)",
					name, expandPath(cfg.Agent.ReposRoot), name)
			}
		}

		base := opts.Bases[name]
		if base == "" {
			base = opts.Base
		}
		if base == "" {
			if override, has := cfg.Agent.Repos[name]; has && override.Base != "" {
				base = override.Base
			} else {
				resolved, err := resolveDefaultBase(source)
				if err != nil {
					return nil, err
				}
				base = resolved
			}
		}

		worktree := filepath.Join(current.TaskRoot, name)
		if existing, ok := current.Repo(name); ok && existing.Worktree != "" {
			worktree = existing.Worktree
			if existing.Base != base {
				return nil, fmt.Errorf("%s is already checked out at %s in this task; remove it with `hlp agent edit %s` before re-adding it on %s",
					name, existing.Base, current.JiraKey, base)
			}
		}

		plan = append(plan, repoPlan{
			Name:     name,
			Source:   source,
			Worktree: worktree,
			Base:     base,
			Exists:   dirExists(worktree),
		})
	}
	return plan, nil
}

// createRepos cuts a detached worktree for every planned repo that lacks one,
// rolling back the ones it created if a later repo fails.
func createRepos(cfg *config.Config, t *task.Task, plan []repoPlan) error {
	if err := os.MkdirAll(t.TaskRoot, 0o755); err != nil {
		return fmt.Errorf("failed to create task root: %w", err)
	}
	seedTaskRoot(cfg.Agent.ReposRoot, t.TaskRoot)

	var created []repoPlan
	rollback := func() {
		for _, p := range created {
			removeWorktree(p.Source, p.Worktree, true)
		}
	}

	for _, p := range plan {
		entry := task.Repo{Name: p.Name, Source: p.Source, Worktree: p.Worktree, Base: p.Base}
		if p.Exists {
			t.AddRepo(entry)
			continue
		}

		stop := ui.Spinner(fmt.Sprintf("Refreshing %s (%s)...", p.Name, p.Base))
		warnings, refreshErr := refreshSource(p.Source, p.Base)
		stop()
		for _, w := range warnings {
			fmt.Println(ui.Warning(w))
		}
		if refreshErr != nil {
			rollback()
			return refreshErr
		}

		if err := addDetachedWorktree(p.Source, p.Worktree, p.Base); err != nil {
			rollback()
			return err
		}
		seedWorktree(p.Source, p.Worktree)
		created = append(created, p)
		t.AddRepo(entry)
		fmt.Println(ui.SuccessMsg(fmt.Sprintf("%s worktree created at origin/%s", p.Name, p.Base)))
	}
	return nil
}

// removeRepos drops repos from the task, refusing to discard work unless forced.
func removeRepos(t *task.Task, names []string, force bool) error {
	var risky []string
	for _, name := range names {
		r, ok := t.Repo(name)
		if !ok {
			continue
		}
		if worktreeDirty(r.Worktree) {
			risky = append(risky, name+" (uncommitted changes)")
		} else if worktreeUnpushed(r.Worktree) {
			risky = append(risky, name+" (unpushed commits)")
		}
	}
	if len(risky) > 0 && !force {
		return fmt.Errorf("refusing to remove %s — pass --force to discard", strings.Join(risky, ", "))
	}

	if !ui.ConfirmAction(fmt.Sprintf("Remove %s from %s? Their worktrees will be deleted.", strings.Join(names, ", "), t.JiraKey)) {
		return fmt.Errorf("cancelled")
	}

	for _, name := range names {
		r, ok := t.Repo(name)
		if !ok {
			continue
		}
		if err := removeWorktree(r.Source, r.Worktree, force); err != nil {
			return err
		}
		t.RemoveRepo(name)
		fmt.Println(ui.SuccessMsg(name + " worktree removed"))
	}
	return nil
}

// reorderRepos re-sorts the stored repos into selection order, keeping any
// entry the selection does not mention at the end.
func reorderRepos(t task.Task, order []string) []task.Repo {
	out := make([]task.Repo, 0, len(t.Repos))
	for _, name := range order {
		if r, ok := t.Repo(name); ok {
			out = append(out, r)
		}
	}
	for _, r := range t.Repos {
		if !slices.Contains(order, r.Name) {
			out = append(out, r)
		}
	}
	return out
}

// applyHexer brings the task's environment up or down to match the request.
// Tearing down happens before worktrees are touched elsewhere, because the
// container is bound to the worktree path.
func applyHexer(cfg *config.Config, store *task.Store, t *task.Task, opts applyOptions) error {
	if !opts.Hexer && !t.Hexer.Enabled {
		return nil
	}

	runner, err := hexer.New(config.GetConfigDir(), cfg.Agent.Hexer, hexerProgress)
	if err != nil {
		return err
	}

	if !opts.Hexer {
		fmt.Println(ui.Info("tearing down the hexer environment"))
		if err := runner.Down(hexerSlug(t.JiraKey), t.TaskRoot, t.Hexer.Port, false); err != nil {
			return err
		}
		t.Hexer = task.Hexer{}
		return nil
	}

	tds, ok := t.Repo(cfg.Agent.Hexer.TDSRepo)
	if !ok {
		return fmt.Errorf("a hexer env needs %s in the task (it serves that worktree)", cfg.Agent.Hexer.TDSRepo)
	}

	modules, err := cfg.Agent.Hexer.ModulesByName(opts.HexerModules)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		return fmt.Errorf("a hexer env needs at least one app selected (%s)", strings.Join(cfg.Agent.Hexer.ModuleNames(), ", "))
	}

	if t.Hexer.Port == 0 || t.Hexer.DBPort == 0 {
		usedHexer, usedDB := store.UsedPorts()
		port, err := hexer.AllocatePort(cfg.Agent.Hexer.HexerPortMin, cfg.Agent.Hexer.HexerPortMax, usedHexer)
		if err != nil {
			return err
		}
		dbPort, err := hexer.AllocatePort(cfg.Agent.Hexer.DBPortMin, cfg.Agent.Hexer.DBPortMax, usedDB)
		if err != nil {
			return err
		}
		t.Hexer.Port, t.Hexer.DBPort = port, dbPort
	}
	t.Hexer.Enabled = true
	t.Hexer.Base = tds.Base
	t.Hexer.Modules = opts.HexerModules
	t.Hexer.Host = hexer.Hostname(t.JiraKey, cfg.Agent.Hexer.HostnameSuffix)

	if err := runner.Doctor(tds.Base); err != nil {
		return err
	}
	return runner.Up(hexer.UpOptions{
		Slug:      hexerSlug(t.JiraKey),
		Branch:    tds.Base,
		Worktree:  tds.Worktree,
		TaskRoot:  t.TaskRoot,
		DBPort:    t.Hexer.DBPort,
		HexerPort: t.Hexer.Port,
		Host:      t.Hexer.Host,
		Modules:   modules,
	})
}

// hexerSlug is the container/host-safe name for a task's environment.
func hexerSlug(key string) string { return strings.ToLower(key) }

func hexerProgress(step string, pct int, msg string) {
	line := fmt.Sprintf("  %s %s", ui.ProgressBar(pct, 100, 24), step)
	if msg != "" {
		line += " — " + msg
	}
	fmt.Println(line)
}

func printTaskPlan(cfg *config.Config, t task.Task, plan []repoPlan, removed []string, opts applyOptions) {
	title := t.JiraKey
	if t.Summary != "" {
		title += " — " + t.Summary
	}
	fmt.Println(ui.Header(title))
	fmt.Println(ui.KeyValue("Task root", t.TaskRoot))
	fmt.Println(ui.KeyValue("Session", t.SessionName()+existingMarker(sessionExists(t.SessionName()))))

	for _, p := range plan {
		state := "new, detached at origin/" + p.Base
		if p.Exists {
			state = "existing"
		}
		fmt.Println(ui.KeyValue("  "+p.Name, fmt.Sprintf("%s (%s)", p.Worktree, state)))
	}
	for _, name := range removed {
		fmt.Println(ui.KeyValue("  "+name, "will be REMOVED"))
	}
	if opts.Hexer {
		fmt.Println(ui.KeyValue("Hexer", fmt.Sprintf("%s · apps: %s",
			hexer.Hostname(t.JiraKey, cfg.Agent.Hexer.HostnameSuffix), strings.Join(opts.HexerModules, ", "))))
	}
}

func dirExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
