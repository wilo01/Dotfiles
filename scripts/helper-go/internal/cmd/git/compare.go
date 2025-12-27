package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	compareNoGraph    bool
	compareMaxCommits int
)

// CompareResult holds the comparison data between commit and branch
type CompareResult struct {
	Commit       string
	Branch       string
	CommitShort  string
	BranchHead   string
	Ahead        int
	Behind       int
	MergeBase    string
	CommitsAhead []CommitInfo
	BehindList   []CommitInfo
}

// CommitInfo represents a single commit
type CommitInfo struct {
	Hash    string
	Message string
}

var compareCmd = &cobra.Command{
	Use:   "compare <branch> <commit>",
	Short: "Compare a commit against a branch",
	Long: `Compare a commit against a branch to see if it's ahead or behind.

Shows a visual graph of the divergence and lists the commits that differ.

Examples:
  hlp git compare main abc123
  hlp git compare develop HEAD~5
  hlp git compare origin/main feature-branch`,
	Args: cobra.ExactArgs(2),
	RunE: runCompare,
}

func init() {
	compareCmd.Flags().BoolVarP(&compareNoGraph, "no-graph", "g", false, "hide visual graph, show only summary")
	compareCmd.Flags().IntVarP(&compareMaxCommits, "commits", "c", 5, "max commits to show in list")
}

func runCompare(cmd *cobra.Command, args []string) error {
	branch := args[0]
	commit := args[1]

	// Validate refs exist
	if err := validateRef(commit); err != nil {
		return fmt.Errorf("invalid commit reference '%s': %w", commit, err)
	}
	if err := validateRef(branch); err != nil {
		return fmt.Errorf("invalid branch reference '%s': %w", branch, err)
	}

	result, err := compareRefs(commit, branch)
	if err != nil {
		return err
	}

	printCompareResult(result)
	return nil
}

func validateRef(ref string) error {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("reference not found")
	}
	return nil
}

func compareRefs(commit, branch string) (*CompareResult, error) {
	result := &CompareResult{
		Commit: commit,
		Branch: branch,
	}

	// Get short commit hash
	if out, err := gitOutput("rev-parse", "--short", commit); err == nil {
		result.CommitShort = strings.TrimSpace(out)
	}

	// Get branch HEAD short hash
	if out, err := gitOutput("rev-parse", "--short", branch); err == nil {
		result.BranchHead = strings.TrimSpace(out)
	}

	// Count commits ahead (on commit, not on branch)
	if out, err := gitOutput("rev-list", "--count", branch+".."+commit); err == nil {
		result.Ahead, _ = strconv.Atoi(strings.TrimSpace(out))
	}

	// Count commits behind (on branch, not on commit)
	if out, err := gitOutput("rev-list", "--count", commit+".."+branch); err == nil {
		result.Behind, _ = strconv.Atoi(strings.TrimSpace(out))
	}

	// Get merge base
	if out, err := gitOutput("merge-base", commit, branch); err == nil {
		base := strings.TrimSpace(out)
		if len(base) >= 7 {
			result.MergeBase = base[:7]
		} else {
			result.MergeBase = base
		}
	}

	// Get commits ahead
	if out, err := gitOutput("log", "--oneline", branch+".."+commit); err == nil && out != "" {
		result.CommitsAhead = parseCommitLog(out)
	}

	// Get commits behind
	if out, err := gitOutput("log", "--oneline", commit+".."+branch); err == nil && out != "" {
		result.BehindList = parseCommitLog(out)
	}

	return result, nil
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parseCommitLog(output string) []CommitInfo {
	var commits []CommitInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) >= 2 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Message: parts[1],
			})
		} else if len(parts) == 1 {
			commits = append(commits, CommitInfo{Hash: parts[0]})
		}
	}
	return commits
}

func printCompareResult(r *CompareResult) {
	fmt.Println()
	fmt.Printf("%s %s %s %s\n\n",
		ui.Title.Render("Comparing:"),
		ui.Value.Render(r.CommitShort),
		ui.Muted.Render("→"),
		ui.Value.Render(r.Branch))

	// Visual graph
	if !compareNoGraph {
		printGraph(r)
		fmt.Println()
	}

	// Summary
	fmt.Println(ui.Header("Summary"))
	fmt.Println()

	if r.Ahead > 0 {
		fmt.Printf("  %s %s is %s ahead of %s\n",
			ui.Success.Render("✓"),
			ui.Value.Render(r.CommitShort),
			ui.Success.Render(fmt.Sprintf("%d commit(s)", r.Ahead)),
			ui.Muted.Render(r.Branch))
	} else {
		fmt.Printf("  %s %s is %s ahead of %s\n",
			ui.Muted.Render("○"),
			ui.Muted.Render(r.CommitShort),
			ui.Muted.Render("0 commits"),
			ui.Muted.Render(r.Branch))
	}

	if r.Behind > 0 {
		fmt.Printf("  %s %s is %s behind %s\n",
			ui.ErrorText.Render("✗"),
			ui.Value.Render(r.CommitShort),
			ui.ErrorText.Render(fmt.Sprintf("%d commit(s)", r.Behind)),
			ui.Muted.Render(r.Branch))
	} else {
		fmt.Printf("  %s %s is %s behind %s\n",
			ui.Muted.Render("○"),
			ui.Muted.Render(r.CommitShort),
			ui.Muted.Render("0 commits"),
			ui.Muted.Render(r.Branch))
	}

	// Commits ahead
	if len(r.CommitsAhead) > 0 {
		fmt.Println()
		fmt.Printf("%s (on %s, not on %s):\n",
			ui.Title.Render("Commits ahead"),
			ui.Value.Render(r.CommitShort),
			ui.Muted.Render(r.Branch))
		printCommitList(r.CommitsAhead, compareMaxCommits)
	}

	// Commits behind
	if len(r.BehindList) > 0 {
		fmt.Println()
		fmt.Printf("%s (on %s, not on %s):\n",
			ui.Title.Render("Commits behind"),
			ui.Muted.Render(r.Branch),
			ui.Value.Render(r.CommitShort))
		printCommitList(r.BehindList, compareMaxCommits)
	}

	fmt.Println()
}

func printCommitList(commits []CommitInfo, max int) {
	for i, c := range commits {
		if i >= max {
			remaining := len(commits) - max
			fmt.Printf("  %s\n", ui.Muted.Render(fmt.Sprintf("... and %d more", remaining)))
			break
		}
		fmt.Printf("  %s %s %s\n",
			ui.Muted.Render("•"),
			ui.Secondary.Render(c.Hash),
			c.Message)
	}
}
