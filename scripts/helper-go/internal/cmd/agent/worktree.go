package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	if direct := filepath.Join(expandPath(cfg.Agent.ReposRoot), repoName); isGitRepo(direct) {
		return direct, nil
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
