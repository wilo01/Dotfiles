package gitops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PRInfo holds metadata about a GitHub pull request.
type PRInfo struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	HeadBranch string `json:"headRefName"`
	BaseBranch string `json:"baseRefName"`
	State      string `json:"state"`
	URL        string `json:"url"`
}

// BranchResult holds the outcome of processing one target branch.
type BranchResult struct {
	TargetBranch  string
	NewBranch     string
	Method        string // "cherry-pick", "patch", ""
	PRNumber      int
	PRURL         string
	Success       bool
	FailureReason string
	Steps         []StepResult
}

// StepResult records each sub-step for display.
type StepResult struct {
	Description string
	Detail      string
	Success     bool
	Warning     bool
	Info        bool
}

// FetchPRInfo retrieves PR metadata using gh pr view.
func FetchPRInfo(prNumber int) (*PRInfo, error) {
	out, err := ghOutput("pr", "view", fmt.Sprintf("%d", prNumber),
		"--json", "number,title,headRefName,baseRefName,state,url")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR #%d: %w", prNumber, err)
	}

	var info PRInfo
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		return nil, fmt.Errorf("failed to parse PR info: %w", err)
	}
	return &info, nil
}

// FetchMergeCommit returns the merge commit SHA for a merged PR.
func FetchMergeCommit(prNumber int) (string, error) {
	out, err := ghOutput("pr", "view", fmt.Sprintf("%d", prNumber),
		"--json", "mergeCommit")
	if err != nil {
		return "", fmt.Errorf("failed to fetch merge commit: %w", err)
	}

	var result struct {
		MergeCommit struct {
			OID string `json:"oid"`
		} `json:"mergeCommit"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return "", fmt.Errorf("failed to parse merge commit: %w", err)
	}
	if result.MergeCommit.OID == "" {
		return "", fmt.Errorf("PR #%d has no merge commit (is it merged?)", prNumber)
	}
	return result.MergeCommit.OID, nil
}

// IsWorkingTreeClean checks for uncommitted changes.
func IsWorkingTreeClean() (bool, error) {
	out, err := gitOutput("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// GetCurrentBranch returns the current branch name.
func GetCurrentBranch() (string, error) {
	out, err := gitOutput("branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// FetchBranch runs git fetch origin <branch>.
func FetchBranch(branch string) error {
	return gitRun("fetch", "origin", "--", branch)
}

// CreateBranchFrom creates a new branch from origin/<base>.
func CreateBranchFrom(newBranch, baseBranch string) error {
	return gitRun("switch", "-c", newBranch, "origin/"+baseBranch)
}

// CherryPick attempts cherry-pick of a merge commit (using -m 1 for first parent).
func CherryPick(commitSHA string) error {
	return gitRun("cherry-pick", "-m", "1", commitSHA)
}

// AbortCherryPick aborts an in-progress cherry-pick.
func AbortCherryPick() error {
	return gitRun("cherry-pick", "--abort")
}

// ApplyPatchFromPR fetches the PR diff and applies it with --3way.
func ApplyPatchFromPR(prNumber int) error {
	diff, err := ghOutput("pr", "diff", fmt.Sprintf("%d", prNumber))
	if err != nil {
		return fmt.Errorf("failed to get PR diff: %w", err)
	}

	cmd := exec.Command("git", "apply", "--3way", "-")
	cmd.Stdin = strings.NewReader(diff)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("patch apply failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// StageAll stages all changes.
func StageAll() error {
	return gitRun("add", "-A")
}

// CommitWithMessage creates a commit with the given message.
func CommitWithMessage(message string) error {
	return gitRun("commit", "-m", message)
}

// PushBranch pushes the branch to origin.
func PushBranch(branch string) error {
	return gitRun("push", "-u", "origin", "--", branch)
}

// CreatePR creates a pull request via gh pr create and returns the PR URL.
func CreatePR(title, body, base, head string) (string, error) {
	out, err := ghOutput("pr", "create",
		"--title", title,
		"--body", body,
		"--base", base,
		"--head", head)
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// CheckoutBranch switches to an existing branch.
func CheckoutBranch(branch string) error {
	return gitRun("switch", "--", branch)
}

// DeleteLocalBranch deletes a local branch.
func DeleteLocalBranch(branch string) error {
	return gitRun("branch", "-D", "--", branch)
}

// FetchPRFiles returns the list of changed file paths for a PR.
func FetchPRFiles(prNumber int) ([]string, error) {
	out, err := ghOutput("pr", "view", fmt.Sprintf("%d", prNumber),
		"--json", "files")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR files: %w", err)
	}

	var result struct {
		Files []struct {
			Path string `json:"path"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return nil, fmt.Errorf("failed to parse PR files: %w", err)
	}

	paths := make([]string, 0, len(result.Files))
	for _, f := range result.Files {
		paths = append(paths, f.Path)
	}
	return paths, nil
}

// GetHeadCommitSHA returns the SHA of the current HEAD commit.
func GetHeadCommitSHA() (string, error) {
	out, err := gitOutput("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetConflictFiles returns the list of unmerged (conflicted) file paths.
func GetConflictFiles() ([]string, error) {
	out, err := gitOutput("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// HasConflictMarkers returns true if any tracked files still contain conflict markers.
func HasConflictMarkers() (bool, error) {
	cmd := exec.Command("git", "diff", "--check")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Non-zero exit means conflict markers found (exit code 2)
			return len(output) > 0, nil
		}
		// Real error (git not found, permissions, etc.)
		return false, fmt.Errorf("git diff --check: %w", err)
	}
	return false, nil
}

// HasStagedChanges returns true if the index has staged changes.
func HasStagedChanges() (bool, error) {
	out, err := gitOutput("diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// --- internal helpers ---

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

func gitRun(args ...string) error {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(output)))
	}
	return nil
}

func ghOutput(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = strings.TrimSpace(stdout.String())
		}
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("gh %s: %s", args[0], errMsg)
	}
	return stdout.String(), nil
}
