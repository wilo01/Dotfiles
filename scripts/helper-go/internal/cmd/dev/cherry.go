package dev

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dariuszw/hlp/internal/commits"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/format"
	"github.com/dariuszw/hlp/internal/gitops"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	cherryBranches    string
	cherryDryRun      bool
	cherryOpenBrowser bool
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
  5. If both fail and stdin is a TTY, prompts for manual conflict resolution
  6. Pushes the new branch and creates a PR

When conflicts require manual resolution, the tool pauses and lists conflicted
files. Resolve them in your editor, then press Enter to continue. The tool
verifies no conflict markers remain, stages, commits, and proceeds with the PR.
Press 's' to skip all remaining branches.

Examples:
  hlp cherry 1036 --branches "maintenance/13.1AV,maintenance/12.1AV"
  hlp cherry 1036 -b "maintenance/13.1AV" --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runCherry,
}

func init() {
	cherryCmd.Flags().StringVarP(&cherryBranches, "branches", "b", "", "comma-separated target branches (required)")
	cherryCmd.Flags().BoolVar(&cherryDryRun, "dry-run", false, "test cherry-pick/patch feasibility without pushing or creating PRs")
	cherryCmd.Flags().BoolVar(&cherryOpenBrowser, "open", false, "open created PR URLs in the browser")
	cherryCmd.MarkFlagRequired("branches")
}

func extractJiraTicket(branch string) string {
	match := commits.TicketPattern.FindStringSubmatch(branch)
	if len(match) >= 2 {
		return strings.ToUpper(match[1])
	}
	return ""
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Clean(filepath.Join(home, path[2:]))
		}
	}
	return filepath.Clean(path)
}

// cherryContext groups the resolved state needed for cherry-pick processing.
type cherryContext struct {
	PRNumber         int
	PRInfo           *gitops.PRInfo
	MergeCommit      string
	OriginalBranch   string
	JiraTicket       string
	CommitEntry      *commits.CommitEntry
	Cfg              *config.Config
	Interactive      bool
	OverwriteBranches *bool // nil = not yet asked, set on first "already exists" prompt
}

// cherryResult wraps a BranchResult with a signal to skip remaining branches.
type cherryResult struct {
	gitops.BranchResult
	SkipAll bool
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
	stop := ui.Spinner("Fetching PR info...")
	prInfo, err := gitops.FetchPRInfo(prNumber)
	stop()
	if err != nil {
		return err
	}
	if prInfo.State != "MERGED" {
		return fmt.Errorf("PR #%d is not merged (state: %s)", prNumber, prInfo.State)
	}

	// Fetch merge commit
	stop = ui.Spinner("Fetching merge commit...")
	mergeCommit, err := gitops.FetchMergeCommit(prNumber)
	stop()
	if err != nil {
		return err
	}
	shortCommit := mergeCommit
	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}

	// Load config (warn on error, fall back to defaults)
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		fmt.Println(ui.Warning(fmt.Sprintf("Could not load config: %v (using defaults)", cfgErr)))
		cfg = config.Default()
	}

	// Look up commit entry from Commits.md
	jiraTicket := extractJiraTicket(prInfo.HeadBranch)
	var commitEntry *commits.CommitEntry
	if jiraTicket != "" && cfg.Dev.CommitsFile != "" {
		var lookupErr error
		commitEntry, lookupErr = commits.LookupByTicket(expandPath(cfg.Dev.CommitsFile), jiraTicket)
		if lookupErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not read Commits.md: %v", lookupErr)))
		}
	}

	// Print header
	printCherryHeader(prNumber, prInfo, shortCommit)

	ctx := cherryContext{
		PRNumber:       prNumber,
		PRInfo:         prInfo,
		MergeCommit:    mergeCommit,
		OriginalBranch: originalBranch,
		JiraTicket:     jiraTicket,
		CommitEntry:    commitEntry,
		Cfg:            cfg,
		Interactive:    term.IsTerminal(int(os.Stdin.Fd())),
	}

	// Dry-run: preflight check each branch, then exit
	if cherryDryRun {
		runDryRun(ctx, targets)
		return nil
	}

	// Process each target branch
	var results []gitops.BranchResult
	for _, target := range targets {
		cr := processCherryTarget(&ctx, target)
		results = append(results, cr.BranchResult)

		if cr.SkipAll {
			fmt.Println(ui.Info("Skipping remaining branches"))
			break
		}

		// Return to original branch between targets
		if checkoutErr := gitops.CheckoutBranch(originalBranch); checkoutErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not return to %s: %v", originalBranch, checkoutErr)))
			break
		}
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

func processCherryTarget(ctx *cherryContext, targetBranch string) cherryResult {
	result := cherryResult{
		BranchResult: gitops.BranchResult{
			TargetBranch: targetBranch,
		},
	}

	// Derive new branch name
	newBranch, err := format.DeriveCherryBranch(ctx.PRInfo.HeadBranch, targetBranch)
	if err != nil {
		result.FailureReason = err.Error()
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Branch name derivation failed", Detail: err.Error(),
		})
		printBranchResult(result.BranchResult)
		return result
	}
	result.NewBranch = newBranch

	// Fetch target branch
	stop := ui.Spinner("Fetching " + targetBranch + "...")
	fetchErr := gitops.FetchBranch(targetBranch)
	stop()
	if fetchErr != nil {
		detail := fetchErr.Error()
		if strings.Contains(detail, "couldn't find remote ref") {
			detail += "\n      Hint: push it first with: git push origin " + targetBranch
		}
		result.FailureReason = "fetch failed"
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Fetch " + targetBranch + " failed", Detail: detail,
		})
		printBranchResult(result.BranchResult)
		return result
	}

	// Create feature branch (prompt once to overwrite if already exists)
	if err := gitops.CreateBranchFrom(newBranch, targetBranch); err != nil {
		if strings.Contains(err.Error(), "already exists") && ctx.Interactive {
			overwrite := ctx.OverwriteBranches
			if overwrite == nil {
				answer := ui.ConfirmAction("    Existing branch(es) found. Delete and recreate all?")
				overwrite = &answer
				ctx.OverwriteBranches = overwrite
			}
			if *overwrite {
				if delErr := gitops.DeleteLocalBranch(newBranch); delErr != nil {
					result.FailureReason = "branch deletion failed"
					result.Steps = append(result.Steps, gitops.StepResult{
						Description: "Could not delete existing branch: " + newBranch, Detail: delErr.Error(),
					})
					printBranchResult(result.BranchResult)
					return result
				}
				if retryErr := gitops.CreateBranchFrom(newBranch, targetBranch); retryErr != nil {
					result.FailureReason = "branch recreation failed"
					result.Steps = append(result.Steps, gitops.StepResult{
						Description: "Branch recreation failed", Detail: retryErr.Error(),
					})
					printBranchResult(result.BranchResult)
					return result
				}
				result.Steps = append(result.Steps, gitops.StepResult{
					Description: "Existing branch replaced: " + newBranch, Success: true,
				})
			} else {
				result.FailureReason = "skipped — branch already exists"
				result.Steps = append(result.Steps, gitops.StepResult{
					Description: "Skipped — user declined to overwrite", Info: true,
				})
				printBranchResult(result.BranchResult)
				return result
			}
		} else {
			result.FailureReason = "branch creation failed"
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Branch creation failed", Detail: err.Error(),
			})
			printBranchResult(result.BranchResult)
			return result
		}
	} else {
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Branch created: " + newBranch, Success: true,
		})
	}

	// Try cherry-pick
	stop = ui.Spinner("Cherry-picking...")
	cherryErr := gitops.CherryPick(ctx.MergeCommit)
	stop()
	if cherryErr != nil {
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Cherry-pick failed, trying patch...", Warning: true,
		})

		if abortErr := gitops.AbortCherryPick(); abortErr != nil {
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Cherry-pick abort warning (continuing to patch)", Detail: abortErr.Error(), Warning: true,
			})
		}

		// Fallback: patch via gh pr diff | git apply --3way
		stop = ui.Spinner("Applying patch...")
		patchErr := gitops.ApplyPatchFromPR(ctx.PRNumber)
		stop()
		if patchErr != nil {
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Patch failed", Detail: patchErr.Error(),
			})

			if ctx.Interactive {
				resolved, skipAll := handleInteractiveResolve(ctx, newBranch, targetBranch, &result)
				if skipAll {
					result.SkipAll = true
					if coErr := gitops.CheckoutBranch(ctx.OriginalBranch); coErr != nil {
						result.Steps = append(result.Steps, gitops.StepResult{
							Description: fmt.Sprintf("Could not return to %s: %v", ctx.OriginalBranch, coErr), Warning: true,
						})
					}
					if dlErr := gitops.DeleteLocalBranch(newBranch); dlErr != nil {
						result.Steps = append(result.Steps, gitops.StepResult{
							Description: fmt.Sprintf("Could not delete branch %s: %v", newBranch, dlErr), Warning: true,
						})
					}
					printBranchResult(result.BranchResult)
					return result
				}
				if !resolved {
					result.FailureReason = "both cherry-pick and patch failed"
					if coErr := gitops.CheckoutBranch(ctx.OriginalBranch); coErr != nil {
						result.Steps = append(result.Steps, gitops.StepResult{
							Description: fmt.Sprintf("Could not return to %s: %v", ctx.OriginalBranch, coErr), Warning: true,
						})
					}
					if dlErr := gitops.DeleteLocalBranch(newBranch); dlErr != nil {
						result.Steps = append(result.Steps, gitops.StepResult{
							Description: fmt.Sprintf("Could not delete branch %s: %v", newBranch, dlErr), Warning: true,
						})
					}
					printBranchResult(result.BranchResult)
					return result
				}
				result.Method = "manual"
				result.Steps = append(result.Steps, gitops.StepResult{
					Description: "Conflicts resolved manually", Success: true,
				})
			} else {
				result.Steps = append(result.Steps, gitops.StepResult{
					Description: "Skipped — resolve manually", Info: true,
				})
				result.FailureReason = "both cherry-pick and patch failed"
				if coErr := gitops.CheckoutBranch(ctx.OriginalBranch); coErr != nil {
					result.Steps = append(result.Steps, gitops.StepResult{
						Description: fmt.Sprintf("Could not return to %s: %v", ctx.OriginalBranch, coErr), Warning: true,
					})
				}
				if dlErr := gitops.DeleteLocalBranch(newBranch); dlErr != nil {
					result.Steps = append(result.Steps, gitops.StepResult{
						Description: fmt.Sprintf("Could not delete branch %s: %v", newBranch, dlErr), Warning: true,
					})
				}
				printBranchResult(result.BranchResult)
				return result
			}
		} else {
			// Patch succeeded — stage and commit
			if err := gitops.StageAll(); err != nil {
				result.Steps = append(result.Steps, gitops.StepResult{
					Description: "Stage failed", Detail: err.Error(),
				})
				result.FailureReason = "staging failed after patch"
				printBranchResult(result.BranchResult)
				return result
			}
			version := format.ExtractMaintenanceVersion(targetBranch)
			commitMsg := buildCommitMessage(ctx.JiraTicket, version, newBranch, ctx.CommitEntry, ctx.PRNumber, targetBranch)
			if err := gitops.CommitWithMessage(commitMsg); err != nil {
				result.Steps = append(result.Steps, gitops.StepResult{
					Description: "Commit failed", Detail: err.Error(),
				})
				result.FailureReason = "commit failed after patch"
				printBranchResult(result.BranchResult)
				return result
			}
			result.Method = "patch"
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Patch applied with --3way", Success: true,
			})
		}
	} else {
		result.Method = "cherry-pick"
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Cherry-pick applied cleanly", Success: true,
		})
	}

	// Push
	stop = ui.Spinner("Pushing " + newBranch + "...")
	pushErr := gitops.PushBranch(newBranch)
	stop()
	if pushErr != nil {
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Push failed", Detail: pushErr.Error(),
		})
		result.FailureReason = "push failed"
		printBranchResult(result.BranchResult)
		return result
	}
	result.Steps = append(result.Steps, gitops.StepResult{
		Description: "Pushed to origin", Success: true,
	})

	// Create PR
	version := format.ExtractMaintenanceVersion(targetBranch)
	prTitle := ctx.PRInfo.Title
	if version != "" {
		prTitle += " [" + version + "]"
	}
	prBody := buildPRBody(newBranch, ctx.CommitEntry, ctx.PRNumber, targetBranch, result.Method)
	stop = ui.Spinner("Creating PR...")
	prURL, prCreateErr := gitops.CreatePR(prTitle, prBody, targetBranch, newBranch)
	stop()
	if prCreateErr != nil {
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "PR creation failed", Detail: prCreateErr.Error(),
		})
		result.FailureReason = "PR creation failed"
		printBranchResult(result.BranchResult)
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

	// Write to Commits.md
	if ctx.Cfg.Dev.CommitsFile != "" {
		sha, shaErr := gitops.GetHeadCommitSHA()
		if shaErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not get HEAD SHA for Commits.md: %v", shaErr)))
		}
		files, filesErr := gitops.FetchPRFiles(ctx.PRNumber)
		if filesErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not fetch PR files for Commits.md: %v", filesErr)))
		}
		if appendErr := commits.AppendCherryEntry(expandPath(ctx.Cfg.Dev.CommitsFile), commits.CherryEntry{
			TargetBranch: targetBranch,
			Ticket:       ctx.JiraTicket,
			Version:      version,
			BranchName:   newBranch,
			Logs:         ctx.CommitEntry.GetLogs(),
			CherryNote:   fmt.Sprintf("Cherry-pick of #%d onto %s", ctx.PRNumber, targetBranch),
			ChangedFiles: files,
			CommitSHA:    sha,
		}); appendErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not update Commits.md: %v", appendErr)))
		}
	}

	printBranchResult(result.BranchResult)
	return result
}

func buildPRBody(newBranch string, entry *commits.CommitEntry, prNumber int, targetBranch, method string) string {
	if entry != nil && len(entry.Logs) > 0 {
		var b strings.Builder
		b.WriteString(newBranch + "\n\nLogs:\n\n")
		for _, log := range entry.Logs {
			b.WriteString("- " + log + "\n")
		}
		b.WriteString(fmt.Sprintf("- Cherry-pick of #%d onto %s\n", prNumber, targetBranch))
		return b.String()
	}
	return fmt.Sprintf("Cherry-pick of #%d onto %s\n\nMethod: %s", prNumber, targetBranch, method)
}

func buildCommitMessage(jiraTicket, version, newBranch string, entry *commits.CommitEntry, prNumber int, targetBranch string) string {
	if entry != nil && version != "" {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%s [%s]\n\n%s", jiraTicket, version, newBranch))
		if len(entry.Logs) > 0 {
			b.WriteString("\n\nLogs:\n")
			for _, l := range entry.Logs {
				b.WriteString("- " + l + "\n")
			}
		}
		b.WriteString(fmt.Sprintf("\nCherry-pick of #%d onto %s", prNumber, targetBranch))
		return b.String()
	}
	return fmt.Sprintf("Cherry-pick PR #%d onto %s (via patch)", prNumber, targetBranch)
}

// handleInteractiveResolve prompts the user to resolve conflicts manually.
// Returns (resolved, skipAll). When skipAll is true, all remaining branches should be skipped.
func handleInteractiveResolve(ctx *cherryContext, newBranch, targetBranch string, result *cherryResult) (bool, bool) {
	fmt.Println()
	fmt.Println(ui.Info(fmt.Sprintf("Conflicts on %s → %s", ui.Code.Render(newBranch), ui.Value.Render(targetBranch))))

	conflicts, conflictErr := gitops.GetConflictFiles()
	if conflictErr != nil {
		fmt.Println(ui.Warning(fmt.Sprintf("Could not list conflicts: %v (check with 'git diff --name-only --diff-filter=U')", conflictErr)))
	} else if len(conflicts) > 0 {
		fmt.Println(ui.Warning(fmt.Sprintf("%d conflicted file(s):", len(conflicts))))
		for _, f := range conflicts {
			fmt.Println("    " + ui.Code.Render(f))
		}
	}
	fmt.Println()

	for {
		choice, choiceErr := ui.PromptChoice("  [Enter] continue after resolving  [s] skip all remaining: ")
		if choiceErr != nil {
			// stdin closed — safest default is skip all
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Input closed — skipping remaining branches", Warning: true,
			})
			result.FailureReason = "input closed"
			return false, true
		}

		if choice == "s" {
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "User chose to skip all remaining", Info: true,
			})
			result.FailureReason = "skipped by user"
			return false, true
		}

		// User pressed Enter — check resolution state
		hasMarkers, markerErr := gitops.HasConflictMarkers()
		if markerErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not check conflict markers: %v", markerErr)))
		} else if hasMarkers {
			fmt.Println(ui.Warning("Conflict markers still present — resolve all conflicts first"))
			continue
		}

		// Stage all and check
		if stageErr := gitops.StageAll(); stageErr != nil {
			fmt.Println(ui.Error(fmt.Sprintf("Stage failed: %v", stageErr)))
			continue
		}

		hasStagedChanges, stagedErr := gitops.HasStagedChanges()
		if stagedErr != nil {
			fmt.Println(ui.Error(fmt.Sprintf("Could not check staged changes: %v", stagedErr)))
			continue
		}
		if !hasStagedChanges {
			fmt.Println(ui.Warning("No staged changes — resolve conflicts and save files"))
			continue
		}

		// Commit
		version := format.ExtractMaintenanceVersion(targetBranch)
		commitMsg := buildCommitMessage(ctx.JiraTicket, version, newBranch, ctx.CommitEntry, ctx.PRNumber, targetBranch)
		if commitErr := gitops.CommitWithMessage(commitMsg); commitErr != nil {
			fmt.Println(ui.Error(fmt.Sprintf("Commit failed: %v", commitErr)))
			continue
		}

		return true, false
	}
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

func runDryRun(ctx cherryContext, targets []string) {
	fmt.Println(ui.Info("DRY RUN — testing cherry-pick feasibility (no push/PR)"))
	fmt.Println()

	var results []gitops.BranchResult
	for _, target := range targets {
		result := processDryRunTarget(ctx, target)
		results = append(results, result)

		// Return to original branch between targets
		if checkoutErr := gitops.CheckoutBranch(ctx.OriginalBranch); checkoutErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Could not return to %s: %v", ctx.OriginalBranch, checkoutErr)))
			break
		}
	}

	printCherrySummary(results)
}

func processDryRunTarget(ctx cherryContext, targetBranch string) gitops.BranchResult {
	result := gitops.BranchResult{
		TargetBranch: targetBranch,
	}

	newBranch, err := format.DeriveCherryBranch(ctx.PRInfo.HeadBranch, targetBranch)
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
	stop := ui.Spinner("Fetching " + targetBranch + "...")
	fetchErr := gitops.FetchBranch(targetBranch)
	stop()
	if fetchErr != nil {
		detail := fetchErr.Error()
		if strings.Contains(detail, "couldn't find remote ref") {
			detail += "\n      Hint: push it first with: git push origin " + targetBranch
		}
		result.FailureReason = "fetch failed"
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Fetch " + targetBranch + " failed", Detail: detail,
		})
		printBranchResult(result)
		return result
	}

	// Create temp branch (auto-delete existing for dry-run since it's just a test)
	if err := gitops.CreateBranchFrom(newBranch, targetBranch); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println(ui.Warning("Replacing existing local branch for dry-run: " + newBranch))
			if delErr := gitops.DeleteLocalBranch(newBranch); delErr != nil {
				result.FailureReason = "branch deletion failed"
				result.Steps = append(result.Steps, gitops.StepResult{
					Description: "Could not delete existing branch: " + newBranch, Detail: delErr.Error(),
				})
				printBranchResult(result)
				return result
			}
			if retryErr := gitops.CreateBranchFrom(newBranch, targetBranch); retryErr != nil {
				result.FailureReason = "branch creation failed"
				result.Steps = append(result.Steps, gitops.StepResult{
					Description: "Branch creation failed", Detail: retryErr.Error(),
				})
				printBranchResult(result)
				return result
			}
		} else {
			result.FailureReason = "branch creation failed"
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Branch creation failed", Detail: err.Error(),
			})
			printBranchResult(result)
			return result
		}
	}

	// Cleanup helper — always delete the temp branch
	cleanup := func() {
		if coErr := gitops.CheckoutBranch(ctx.OriginalBranch); coErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Cleanup: could not return to %s: %v", ctx.OriginalBranch, coErr)))
		}
		if dlErr := gitops.DeleteLocalBranch(newBranch); dlErr != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Cleanup: could not delete branch %s: %v", newBranch, dlErr)))
		}
	}

	// Try cherry-pick
	stop = ui.Spinner("Cherry-picking...")
	cherryErr := gitops.CherryPick(ctx.MergeCommit)
	stop()
	if cherryErr != nil {
		if abortErr := gitops.AbortCherryPick(); abortErr != nil {
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Cherry-pick abort warning", Detail: abortErr.Error(), Warning: true,
			})
		}

		// Try patch
		stop = ui.Spinner("Applying patch...")
		patchErr := gitops.ApplyPatchFromPR(ctx.PRNumber)
		stop()
		if patchErr != nil {
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Cherry-pick would fail",
			})
			result.Steps = append(result.Steps, gitops.StepResult{
				Description: "Patch would fail — manual resolution needed",
			})
			result.FailureReason = "needs manual resolution"
			cleanup()
			printBranchResult(result)
			return result
		}

		result.Method = "patch"
		result.Success = true
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Cherry-pick would fail, patch would succeed", Success: true,
		})
	} else {
		result.Method = "cherry-pick"
		result.Success = true
		result.Steps = append(result.Steps, gitops.StepResult{
			Description: "Cherry-pick would apply cleanly", Success: true,
		})
	}

	cleanup()
	printBranchResult(result)
	return result
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

	// Collect successful PR URLs
	var urls []string
	for _, r := range results {
		if r.Success && r.PRURL != "" {
			urls = append(urls, r.PRURL)
		}
	}

	// Print URL list for easy copying
	if len(urls) > 0 {
		fmt.Println()
		for _, u := range urls {
			fmt.Println(u)
		}
	}

	// Open in browser if --open flag
	if cherryOpenBrowser {
		for _, u := range urls {
			_ = exec.Command("xdg-open", u).Start()
		}
	}

	fmt.Println()
}
