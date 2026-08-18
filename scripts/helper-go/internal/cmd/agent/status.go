package agent

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/dariuszw/hlp/internal/task"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var statusDrift bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show every ticket task: repos, worktree state, session, PR",
	Long: `Lists the tasks in tasks.json, most recently opened first, with one row per
repo. BRANCH shows "detached@<sha>" until Claude creates a branch.

Pass --drift to also check each task against the filesystem.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func runStatus() error {
	store, err := loadStore()
	if err != nil {
		return err
	}

	tasks := store.All()
	if len(tasks) == 0 {
		fmt.Println(ui.Info("no tasks yet — run: hlp agent start <TICKET-KEY>"))
		return nil
	}

	var rows [][]string
	for _, t := range tasks {
		session := sessionState(t)
		env := hexerCell(t)

		if len(t.Repos) == 0 {
			rows = append(rows, []string{t.JiraKey, "-", "-", "-", session, env, "-"})
			continue
		}
		for i, r := range t.Repos {
			key := t.JiraKey
			if i > 0 {
				key = "" // one key per task, repos listed beneath it
			}
			branch := worktreeHead(r.Worktree)
			rows = append(rows, []string{
				key, r.Name, branch, branchAhead(r.Worktree, branch), session, env, prState(r.Worktree, branch),
			})
		}
	}

	fmt.Println(ui.Table([]string{"KEY", "REPO", "BRANCH", "AHEAD", "SESSION", "HEXER", "PR"}, rows))

	if statusDrift {
		for _, t := range tasks {
			if drift := task.Reconcile(t); len(drift) > 0 {
				fmt.Println()
				fmt.Println(ui.Header(t.JiraKey + " drift"))
				reportDrift(t)
			}
		}
	}
	return nil
}

// sessionState describes the tmux session: absent, idle, or running claude.
func sessionState(t task.Task) string {
	name := t.SessionName()
	if !sessionExists(name) {
		return "-"
	}
	primary, ok := t.PrimaryRepo()
	if ok && claudeRunningInWindow(name, primary.Name) {
		return "claude"
	}
	return "idle"
}

func hexerCell(t task.Task) string {
	if !t.Hexer.Enabled {
		return "-"
	}
	return fmt.Sprintf(":%d", t.Hexer.Port)
}

func gitOutput(dir string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return "-"
	}
	return strings.TrimSpace(string(out))
}

// branchAhead reports "+ahead/-behind" vs upstream. A detached worktree has no
// upstream and nothing to compare, which is the normal state until Claude
// creates a branch.
func branchAhead(dir, branch string) string {
	if branch == "-" || branch == "?" || strings.HasPrefix(branch, "detached") {
		return "-"
	}
	upstream := gitOutput(dir, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if upstream == "-" {
		return "unpushed"
	}
	counts := gitOutput(dir, "rev-list", "--count", "--left-right", branch+"..."+upstream)
	parts := strings.Fields(counts)
	if len(parts) != 2 {
		return "-"
	}
	return fmt.Sprintf("+%s/-%s", parts[0], parts[1])
}

// prState queries gh for a PR on the branch; degrades to "-" without gh/auth
func prState(dir, branch string) string {
	if branch == "-" || branch == "?" || strings.HasPrefix(branch, "detached") {
		return "-"
	}
	gh, err := exec.LookPath("gh")
	if err != nil {
		return "-"
	}
	cmd := exec.Command(gh, "pr", "list", "--head", branch, "--state", "all",
		"--json", "number,isDraft,state", "--limit", "1")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "-"
	}
	var prs []struct {
		Number  int    `json:"number"`
		IsDraft bool   `json:"isDraft"`
		State   string `json:"state"`
	}
	if json.Unmarshal(out, &prs) != nil || len(prs) == 0 {
		return "none"
	}
	state := strings.ToLower(prs[0].State)
	if prs[0].IsDraft {
		state = "draft"
	}
	return fmt.Sprintf("#%d %s", prs[0].Number, state)
}

func init() {
	statusCmd.Flags().BoolVar(&statusDrift, "drift", false, "also check each task against the filesystem")
}
