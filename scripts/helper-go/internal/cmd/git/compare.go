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
	compareFormat     string // "table" or "simple"
	compareTarget     string // --target/-t flag
	compareBase       string // --base/-b flag
)

// CompareResult holds the comparison data between commit and branch
type CompareResult struct {
	Commit        string
	Branch        string
	CommitShort   string
	BranchHead    string
	Ahead         int
	Behind        int
	MergeBase     string
	CommitsAhead  []CommitInfo
	BehindList    []CommitInfo
	// Enhanced versions with metadata
	AheadEnhanced  []EnhancedCommitInfo
	BehindEnhanced []EnhancedCommitInfo
}

// CommitInfo represents a single commit (basic)
type CommitInfo struct {
	Hash    string
	Message string
}

// EnhancedCommitInfo represents a commit with full metadata
type EnhancedCommitInfo struct {
	Hash         string
	ShortHash    string
	Message      string
	Author       string
	Date         string // formatted as "16 Jan 12:42"
	BranchLabels []string
	IsTarget     bool // marks the commit being compared
}

var compareCmd = &cobra.Command{
	Use:   "compare [<commit> <base>]",
	Short: "Compare a commit against a base branch",
	Long: `Compare a commit against a base branch to see if it's ahead or behind.

Shows a visual table (like VSCode Git Graph) of the divergence with dates and authors.

Arguments:
  <commit>  The commit/branch to compare (what you're checking)
  <base>    The base branch to compare against (reference point)

Flags (alternative to positional args):
  --target/-t  The commit/branch to compare
  --base/-b    The base branch to compare against

Examples:
  # Using positional arguments:
  hlp git compare a593eaff master        # is commit in sync with master?
  hlp git compare HEAD~5 develop         # how far behind is HEAD~5 from develop?
  hlp git compare feature-branch origin/main

  # Using named flags:
  hlp git compare --target a593eaff --base master
  hlp git compare -t HEAD~5 -b develop`,
	Args: cobra.MaximumNArgs(2),
	RunE: runCompare,
}

func init() {
	compareCmd.Flags().BoolVarP(&compareNoGraph, "no-graph", "g", false, "hide visual graph, show only summary")
	compareCmd.Flags().IntVarP(&compareMaxCommits, "commits", "c", 5, "max commits to show in list")
	compareCmd.Flags().StringVarP(&compareFormat, "format", "f", "table", "output format: table or simple")
	compareCmd.Flags().StringVarP(&compareTarget, "target", "t", "", "commit/branch to compare (alternative to positional arg)")
	compareCmd.Flags().StringVarP(&compareBase, "base", "b", "", "base branch to compare against (alternative to positional arg)")
}

func runCompare(cmd *cobra.Command, args []string) error {
	var commit, branch string

	// Flags take precedence over positional args
	if compareTarget != "" || compareBase != "" {
		if compareTarget == "" || compareBase == "" {
			return fmt.Errorf("both --target and --base must be specified together")
		}
		commit = compareTarget
		branch = compareBase
	} else if len(args) == 2 {
		commit = args[0] // what you're checking
		branch = args[1] // base/reference point
	} else {
		return fmt.Errorf("usage: hlp git compare <commit> <base> or --target <commit> --base <base>")
	}

	// Validate refs exist
	if err := validateRef(commit); err != nil {
		return fmt.Errorf("invalid commit reference '%s': %w", commit, err)
	}
	if err := validateRef(branch); err != nil {
		return fmt.Errorf("invalid base reference '%s': %w", branch, err)
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

	// Get commits ahead (basic)
	if out, err := gitOutput("log", "--oneline", branch+".."+commit); err == nil && out != "" {
		result.CommitsAhead = parseCommitLog(out)
	}

	// Get commits behind (basic)
	if out, err := gitOutput("log", "--oneline", commit+".."+branch); err == nil && out != "" {
		result.BehindList = parseCommitLog(out)
	}

	// Get enhanced commit logs with author and date
	logFormat := "%H|%h|%s|%an|%ad"
	dateFormat := "%d %b %H:%M"

	// Commits ahead (on commit, not on branch) - enhanced
	if out, err := gitOutput("log", "--format="+logFormat, "--date=format:"+dateFormat, branch+".."+commit); err == nil && out != "" {
		result.AheadEnhanced = parseEnhancedCommitLog(out)
	}

	// Commits behind (on branch, not on commit) - enhanced
	if out, err := gitOutput("log", "--format="+logFormat, "--date=format:"+dateFormat, commit+".."+branch); err == nil && out != "" {
		result.BehindEnhanced = parseEnhancedCommitLog(out)
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

// parseEnhancedCommitLog parses git log output with format: hash|shortHash|message|author|date
func parseEnhancedCommitLog(output string) []EnhancedCommitInfo {
	var commits []EnhancedCommitInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			commits = append(commits, EnhancedCommitInfo{
				Hash:      parts[0],
				ShortHash: parts[1],
				Message:   parts[2],
				Author:    parts[3],
				Date:      parts[4],
			})
		} else if len(parts) >= 2 {
			// Fallback for basic format
			commits = append(commits, EnhancedCommitInfo{
				Hash:      parts[0],
				ShortHash: parts[0][:7],
				Message:   parts[1],
			})
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

	// Visual graph/table
	if !compareNoGraph {
		if compareFormat == "table" && (len(r.AheadEnhanced) > 0 || len(r.BehindEnhanced) > 0) {
			printTable(r, compareMaxCommits)
			fmt.Println()
		} else {
			printGraph(r)
			fmt.Println()
		}
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
