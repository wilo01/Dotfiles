package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
)

// expandPath expands a leading ~ to the user's home directory
func expandPath(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return path
}

// discoverRepos lists candidate repo names: the directories under worktree_root
func discoverRepos(worktreeRoot string) ([]string, error) {
	entries, err := os.ReadDir(expandPath(worktreeRoot))
	if err != nil {
		return nil, fmt.Errorf("failed to list worktree root %s: %w", worktreeRoot, err)
	}
	var repos []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			repos = append(repos, e.Name())
		}
	}
	sort.Strings(repos)
	return repos, nil
}

// repoSource resolves the main checkout for a repo name. Resolution order:
// explicit config override, git common dir derived from any existing worktree
// under <root>/<repo>/, then the legacy ~/tds-branch-opener/branches/<repo>.
func repoSource(cfg *config.Config, repoName string) (string, error) {
	if override, ok := cfg.Agent.Repos[repoName]; ok && override.Source != "" {
		source := expandPath(override.Source)
		if isGitRepo(source) {
			return source, nil
		}
		return "", fmt.Errorf("configured source for %s is not a git repo: %s", repoName, source)
	}

	repoDir := filepath.Join(expandPath(cfg.Agent.WorktreeRoot), repoName)
	if entries, err := os.ReadDir(repoDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			candidate := filepath.Join(repoDir, e.Name())
			if source := sourceFromWorktree(candidate); source != "" {
				return source, nil
			}
		}
	}

	legacy := expandPath(filepath.Join("~/tds-branch-opener/branches", repoName))
	if isGitRepo(legacy) {
		return legacy, nil
	}

	return "", fmt.Errorf("cannot locate main checkout for %s: no existing worktree under %s and no agent.repos.%s.source in config",
		repoName, repoDir, repoName)
}

// sourceFromWorktree derives the main checkout path from a linked worktree
// via its git common dir ("" if the dir is not a linked worktree).
func sourceFromWorktree(worktreePath string) string {
	out, err := exec.Command("git", "-C", worktreePath, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return ""
	}
	commonDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(worktreePath, commonDir)
	}
	source := filepath.Dir(commonDir)
	if filepath.Clean(source) == filepath.Clean(worktreePath) {
		return "" // not a linked worktree; common dir is its own .git
	}
	return source
}

func isGitRepo(path string) bool {
	return exec.Command("git", "-C", path, "rev-parse", "--git-dir").Run() == nil
}

// findBranch looks for a branch containing the ticket key. Local branches win
// over remote ones; remote names are returned without the remotes/origin/ prefix.
func findBranch(source, key string) (string, bool) {
	out, err := exec.Command("git", "-C", source, "branch", "-a", "--list", "*"+key+"*").Output()
	if err != nil {
		return "", false
	}
	var remote string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		// "* " marks the current branch, "+ " one checked out in another worktree
		name := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "*+ "))
		if name == "" || strings.Contains(name, "->") {
			continue
		}
		if rest, found := strings.CutPrefix(name, "remotes/"); found {
			if _, branch, ok := strings.Cut(rest, "/"); ok && remote == "" {
				remote = branch
			}
			continue
		}
		return name, true // first local match wins
	}
	if remote != "" {
		return remote, true
	}
	return "", false
}

// resumeScan returns the repos (under worktree_root) that already have a
// branch containing the ticket key.
func resumeScan(cfg *config.Config, repos []string, key string) []string {
	var matches []string
	for _, repo := range repos {
		source, err := repoSource(cfg, repo)
		if err != nil {
			continue
		}
		if _, found := findBranch(source, key); found {
			matches = append(matches, repo)
		}
	}
	return matches
}

// worktreePathForBranch returns the existing worktree path that has the branch
// checked out ("" if none), parsed from git worktree list --porcelain.
func worktreePathForBranch(source, branch string) string {
	out, err := exec.Command("git", "-C", source, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return ""
	}
	var current string
	for _, line := range strings.Split(string(out), "\n") {
		if rest, found := strings.CutPrefix(line, "worktree "); found {
			current = rest
		}
		if rest, found := strings.CutPrefix(line, "branch "); found {
			if strings.TrimPrefix(rest, "refs/heads/") == branch {
				return current
			}
		}
	}
	return ""
}

// defaultBase returns the repo's default branch from origin/HEAD (e.g. master)
func defaultBase(source string) string {
	out, err := exec.Command("git", "-C", source, "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err != nil {
		return "master"
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/remotes/origin/")
}

// fetchOrigin updates remote-tracking refs (with prune) so base/branch
// resolution and fast-forwarding see the latest remote state. Best-effort.
func fetchOrigin(source string) error {
	if out, err := exec.Command("git", "-C", source, "fetch", "origin", "--prune").CombinedOutput(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// fastForwardToUpstream fast-forwards the worktree's branch to its upstream when
// it is strictly behind. It never rewrites history: if the branch is ahead or has
// diverged it warns and leaves the worktree untouched. Best-effort (prints its
// own status).
func fastForwardToUpstream(worktreePath string) {
	upstream, err := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}").Output()
	if err != nil {
		return // no upstream configured (e.g. a brand-new local branch); nothing to do
	}
	up := strings.TrimSpace(string(upstream))

	out, err := exec.Command("git", "-C", worktreePath, "rev-list", "--left-right", "--count", "@{u}...HEAD").Output()
	if err != nil {
		return
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) != 2 {
		return
	}
	behind, ahead := fields[0], fields[1]
	if behind == "0" {
		return // up to date, or ahead-only
	}
	if ahead != "0" {
		fmt.Printf("warning: branch diverged from %s (%s behind, %s ahead) — leaving worktree as-is; rebase/merge manually\n", up, behind, ahead)
		return
	}
	if out, err := exec.Command("git", "-C", worktreePath, "merge", "--ff-only", "@{u}").CombinedOutput(); err != nil {
		fmt.Printf("warning: fast-forward to %s failed: %s\n", up, strings.TrimSpace(string(out)))
		return
	}
	fmt.Printf("fast-forwarded %s commit(s) from %s\n", behind, up)
}

// addWorktree creates a worktree at path. When createBranch is true a new
// branch is created off base; otherwise the existing branch is checked out.
// Afterwards wt's post-start hooks run in the new worktree (best-effort) so
// per-repo setup like dependency install still happens.
func addWorktree(source, path, branch, base string, createBranch bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create worktree parent dir: %w", err)
	}

	args := []string{"-C", source, "worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, path, base)
	} else {
		args = append(args, path, branch)
	}
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add failed: %s", strings.TrimSpace(string(out)))
	}

	if wtBin, err := exec.LookPath("wt"); err == nil {
		if out, err := exec.Command(wtBin, "hook", "post-start", "-C", path).CombinedOutput(); err != nil {
			fmt.Printf("warning: wt post-start hooks failed: %s\n", strings.TrimSpace(string(out)))
		}
	}
	return nil
}

var slugInvalidChars = regexp.MustCompile(`[^a-z0-9]+`)

// branchNameFor builds feature/<KEY>-<slug> or bugfix/<KEY>-<slug> per the
// git-workflow.md naming convention.
func branchNameFor(key, summary, issueType string) string {
	prefix := "feature"
	if strings.EqualFold(issueType, "Bug") || strings.EqualFold(issueType, "Tweak") {
		prefix = "bugfix"
	}
	slug := strings.Trim(slugInvalidChars.ReplaceAllString(strings.ToLower(summary), "-"), "-")
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	if slug == "" {
		return fmt.Sprintf("%s/%s", prefix, key)
	}
	return fmt.Sprintf("%s/%s-%s", prefix, key, slug)
}
