package dev

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dariuszw/hlp/internal/format"
	"github.com/dariuszw/hlp/internal/gitops"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	cherryBranches string
	cherryDryRun   bool
)

var cherryCmd = &cobra.Command{
	Use:   "cherry <pr-number>",
	Short: "Cherry-pick a merged PR onto maintenance branches",
	Long: `Cherry-pick a merged GitHub PR onto one or more maintenance branches.

For each target branch:
  1. Fetches the target branch from origin
  2. Creates a feature branch (version suffix replaced or appended)
  3. Attempts cherry-pick of the merge commit
  4. Falls back to patch (gh pr diff | git apply --3way) if cherry-pick fails
  5. Pushes the new branch and creates a PR

Examples:
  hlp cherry 1036 --branches "maintenance/13.1AV,maintenance/12.1AV"
  hlp cherry 1036 -b "maintenance/13.1AV" --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runCherry,
}

func init() {
	cherryCmd.Flags().StringVarP(&cherryBranches, "branches", "b", "", "comma-separated target branches (required)")
	cherryCmd.Flags().BoolVar(&cherryDryRun, "dry-run", false, "print plan without executing")
	cherryCmd.MarkFlagRequired("branches")
}

func runCherry(cmd *cobra.Command, args []string) error {
	prNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid PR number: %s", args[0])
	}

	targets := parseBranches(cherryBranches)
	if len(targets) == 0 {
		return fmt.Errorf("no target branches specified")
	}

	// Check working tree is clean
	clean, err := gitops.IsWorkingTreeClean()
	if err != nil {
		return fmt.Errorf("failed to check working tree: %w", err)
	}
	if !clean {
		fmt.Println(ui.Error("Working tree is not clean. Commit or stash changes first."))
		return fmt.Errorf("dirty working tree")
	}

	// Save current branch for restore
	originalBranch, err := gitops.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Fetch PR info
	prInfo, err := gitops.FetchPRInfo(prNumber)
	if err != nil {
		return err
	}
	if prInfo.State != "MERGED" {
		return fmt.Errorf("PR #%d is not merged (state: %s)", prNumber, prInfo.State)
	}

	// Fetch merge commit
	mergeCommit, err := gitops.FetchMergeCommit(prNumber)
	if err != nil {
		return err
	}
	shortCommit := mergeCommit
	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}

	// Print header
	printCherryHeader(prNumber, prInfo, shortCommit)

	// Dry-run: print plan and exit
	if cherryDryRun {
		printDryRunPlan(prInfo, targets, shortCommit)
		return nil
	}

	// Process each target branch
	var results []gitops.BranchResult
	for _, target := range targets {
		result := processCherryTarget(prNumber, prInfo, mergeCommit, target, originalBranch)
		results = append(results, result)

		// Return to original branch between targets
		_ = gitops.CheckoutBranch(originalBranch)
	}

	// Print summary
	printCherrySummary(results)

	return nil
}

func parseBranches(input string) []string {
	var branches []string
	for _, b := range strings.Split(input, ",") {
		b = strings.TrimSpace(b)
		if b != "" {
			branches = append(branches, b)
		}
	}
	return branches
}

func processCherryTarget(prNumber int, prInfo *gitops.PRInfo, mergeCommit, targetBranch, originalBranch string) gitops.BranchResult {
	result := gitops.BranchResult{
		TargetBranch: targetBranch,
	}

	// Derive new branch name
	newBranch, err := format.DeriveCherryBranch(prInfo.HeadBranch, targetBranch)
	if err != nil {
		result.FailureReason = err.Error()
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Branch name derivation failed", Detail: err.Error(),
		})
		printBranchResult(result)
		return result
	}
	result.NewBranch = newBranch

	// Fetch target branch
	if err := gitops.FetchBranch(targetBranch); err != nil {
		result.FailureReason = "fetch failed"
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Fetch " + targetBranch + " failed", Detail: err.Error(),
		})
		printBranchResult(result)
		return result
	}

	// Create feature branch
	if err := gitops.CreateBranchFrom(newBranch, targetBranch); err != nil {
		result.FailureReason = "branch creation failed"
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Branch creation failed", Detail: err.Error(),
		})
		printBranchResult(result)
		return result
	}
	result.Steps = append(result.Steps, gitops.StepResult{
		Description: "Branch created: " + newBranch, Success: true,
	})

	// Try cherry-pick
	if err := gitops.CherryPick(mergeCommit); err != nil {
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Cherry-pick failed, trying patch...", Warning: true,
		})
		_ = gitops.AbortCherryPick()

		// Fallback: patch via gh pr diff | git apply --3way
		if err := gitops.ApplyPatchFromPR(prNumber); err != nil {
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Cherry-pick failed",
			})
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Patch failed", Detail: err.Error(),
			})
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Skipped — resolve manually", Info: true,
			})
			result.FailureReason = "both cherry-pick and patch failed"
			// Cleanup: return to original branch and delete the failed branch
			_ = gitops.CheckoutBranch(originalBranch)
			_ = gitops.DeleteLocalBranch(newBranch)
			printBranchResult(result)
			return result
		}

		// Patch succeeded — stage and commit
		if err := gitops.StageAll(); err != nil {
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Stage failed", Detail: err.Error(),
			})
			result.FailureReason = "staging failed after patch"
			printBranchResult(result)
			return result
		}
		commitMsg := fmt.Sprintf("Cherry-pick PR #%d onto %s (via patch)", prNumber, targetBranch)
		if err := gitops.CommitWithMessage(commitMsg); err != nil {
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Commit failed", Detail: err.Error(),
			})
			result.FailureReason = "commit failed after patch"
			printBranchResult(result)
			return result
		}
		result.Method = "patch"
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Patch applied with --3way", Success: true,
		})
	} else {
		result.Method = "cherry-pick"
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Cherry-pick applied cleanly", Success: true,
		})
	}

	// Push
	if err := gitops.PushBranch(newBranch); err != nil {
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Push failed", Detail: err.Error(),
		})
		result.FailureReason = "push failed"
		printBranchResult(result)
		return result
	}
	result.Steps = append(result.Steps, gitops.StepResult{
		Description: "Pushed to origin", Success: true,
	})

	// Create PR
	version := format.ExtractMaintenanceVersion(targetBranch)
	prTitle := prInfo.Title
	if version != "" {
		prTitle += " [" + version + "]"
	}
	prBody := fmt.Sprintf("Cherry-pick of #%d onto %s\n\nMethod: %s", prNumber, targetBranch, result.Method)
	prURL, err := gitops.CreatePR(prTitle, prBody, targetBranch, newBranch)
	if err != nil {
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "PR creation failed", Detail: err.Error(),
		})
		result.FailureReason = "PR creation failed"
		printBranchResult(result)
		return result
	}

	// Extract PR number from URL (e.g., "https://github.com/org/repo/pull/1036")
	parts := strings.Split(strings.TrimRight(prURL, "/"), "/")
	if len(parts) > 0 {
		if num, convErr := strconv.Atoi(parts[len(parts)-1]); convErr == nil {
			result.PRNumber = num
		}
	}
	result.PRURL = prURL
	result.Success = true
	result.Steps = append(result.Steps, gitops.StepResult{
		Description: fmt.Sprintf("PR #%d created", result.PRNumber),
		Success:     true,
		Detail:      prURL,
	})

	printBranchResult(result)
	return result
}

// --- Display helpers ---

func printCherryHeader(prNumber int, info *gitops.PRInfo, shortCommit string) {
	fmt.Println()
	fmt.Println(ui.Header(fmt.Sprintf("Cherry PR #%d", prNumber)))
	fmt.Println()
	fmt.Printf("  %s %s %s %s %s\n",
		ui.Label.Render("Source:"),
		ui.Code.Render(info.HeadBranch),
		ui.Muted.Render("→"),
		ui.Value.Render(info.BaseBranch),
		ui.Muted.Render("(merged)"))
	fmt.Printf("  %s %s\n", ui.Label.Render("Commit:"), ui.Code.Render(shortCommit))
	fmt.Println()
}

func printBranchResult(result gitops.BranchResult) {
	fmt.Printf("  %s\n", ui.Value.Render(result.TargetBranch))
	for _, step := range result.Steps {
		var line string
		switch {
		case step.Warning:
			line = "    " + ui.Warning(step.Description)
		case step.Info:
			line = "    " + ui.Info(step.Description)
		case step.Success:
			line = "    " + ui.SuccessMsg(step.Description)
			if step.Detail != "" {
				line += " " + ui.Hyperlink(step.Detail, step.Detail)
			}
		default:
			line = "    " + ui.Error(step.Description)
			if step.Detail != "" {
				line += "\n      " + ui.Muted.Render(step.Detail)
			}
		}
		fmt.Println(line)
	}
	fmt.Println()
}

func printDryRunPlan(info *gitops.PRInfo, targets []string, shortCommit string) {
	fmt.Println(ui.Info("DRY RUN — no changes will be made"))
	fmt.Println()
	for _, target := range targets {
		newBranch, err := format.DeriveCherryBranch(info.HeadBranch, target)
		if err != nil {
			fmt.Printf("  %s %s\n", ui.Error(target+":"), err.Error())
			continue
		}
		fmt.Printf("  %s\n", ui.Value.Render(target))
		fmt.Printf("    %s Create branch: %s\n", ui.Muted.Render("•"), ui.Code.Render(newBranch))
		fmt.Printf("    %s Cherry-pick commit: %s\n", ui.Muted.Render("•"), ui.Code.Render(shortCommit))
		fmt.Printf("    %s Push and create PR targeting %s\n", ui.Muted.Render("•"), target)
		fmt.Println()
	}
}

func printCherrySummary(results []gitops.BranchResult) {
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}
	total := len(results)

	fmt.Println(ui.Divider(40))
	summary := fmt.Sprintf("Summary: %d/%d cherry-picks created", successCount, total)
	if successCount == total {
		fmt.Printf("  %s\n", ui.SuccessMsg(summary))
	} else if successCount > 0 {
		fmt.Printf("  %s\n", ui.Warning(summary))
	} else {
		fmt.Printf("  %s\n", ui.Error(summary))
	}
	fmt.Println()
}
