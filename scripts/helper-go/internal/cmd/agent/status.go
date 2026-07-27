package agent

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var ticketKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]+-\d+$`)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all ticket agents: session, worktree, branch, PR",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func runStatus() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	root := expandPath(cfg.Agent.WorktreeRoot)
	repos, err := discoverRepos(cfg.Agent.WorktreeRoot)
	if err != nil {
		return err
	}

	var rows [][]string
	for _, repo := range repos {
		worktrees, err := filepath.Glob(filepath.Join(root, repo, "*"))
		if err != nil {
			continue
		}
		for _, wtPath := range worktrees {
			key := filepath.Base(wtPath)
			if !ticketKeyPattern.MatchString(key) || !isGitRepo(wtPath) {
				continue
			}

			session := "-"
			if sessionExists(key) {
				session = "idle"
				if agentRunning(key) {
					session = "agent"
				}
			}

			branch := gitOutput(wtPath, "rev-parse", "--abbrev-ref", "HEAD")
			ahead := branchAhead(wtPath, branch)
			pr := prState(wtPath, branch)

			rows = append(rows, []string{key, repo, session, branch, ahead, pr})
		}
	}

	if len(rows) == 0 {
		fmt.Println(ui.Info("no ticket worktrees found under " + root))
		return nil
	}

	fmt.Println(ui.Table([]string{"KEY", "REPO", "SESSION", "BRANCH", "AHEAD", "PR"}, rows))
	return nil
}

func gitOutput(dir string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return "-"
	}
	return strings.TrimSpace(string(out))
}

// branchAhead reports "N ahead" vs upstream, "unpushed" when no upstream exists
func branchAhead(dir, branch string) string {
	if branch == "-" || branch == "HEAD" {
		return "-"
	}
	upstream := gitOutput(dir, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if upstream == "-" {
		return "unpushed"
	}
	counts := gitOutput(dir, "rev-list", "--count", "--left-right", branch+"..."+upstream)
	if counts == "-" {
		return "-"
	}
	parts := strings.Fields(counts)
	if len(parts) == 2 {
		return fmt.Sprintf("+%s/-%s", parts[0], parts[1])
	}
	return "-"
}

// prState queries gh for a PR on the branch; degrades to "-" without gh/auth
func prState(dir, branch string) string {
	if branch == "-" || branch == "HEAD" {
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
