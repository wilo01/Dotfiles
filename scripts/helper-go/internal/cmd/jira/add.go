package jira

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var addSync bool
var addQuiet bool
var addDryRun bool

var addCmd = &cobra.Command{
	Use:   "add [ticket-key...]",
	Short: "Add JIRA ticket(s) to worklogs.csv as draft",
	Long: `Add one or more JIRA tickets to worklogs.csv with DRAFT status.

The tickets will be prepared for time logging but not processed by --batch
until you fill in the TimeSpent and change the status from DRAFT to empty.

Use --sync to automatically fetch tickets from your current sprint.

Examples:
  hlp jira add VIS-1234              # Add single ticket
  hlp jira add VIS-1234 VIS-5678     # Add multiple tickets
  hlp jira add --sync                # Sync tickets from current sprint
  hlp jira add --sync VIS-extra      # Sync sprint + add extra ticket`,
	Run: runAdd,
}

func init() {
	addCmd.Flags().BoolVarP(&addSync, "sync", "s", false, "Sync tickets from current sprint")
	addCmd.Flags().BoolVarP(&addQuiet, "quiet", "q", false, "Suppress output unless errors (for cron)")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "Preview changes without modifying files")
}

// ticketKeyPattern validates JIRA ticket format (PROJECT-NUMBER)
var ticketKeyPattern = regexp.MustCompile(`^[A-Z]+-\d+$`)

func runAdd(cmd *cobra.Command, args []string) {
	// Require either --sync or ticket keys
	if !addSync && len(args) == 0 {
		fmt.Println(ui.Error("Provide ticket keys or use --sync to fetch from sprint"))
		fmt.Println(ui.Muted.Render("Examples:"))
		fmt.Println(ui.Muted.Render("  hlp jira add VIS-1234"))
		fmt.Println(ui.Muted.Render("  hlp jira add --sync"))
		return
	}

	// Show dry-run header
	if addDryRun {
		fmt.Println(ui.Warning("[DRY-RUN] Preview mode - no changes will be made"))
		fmt.Println()
	}

	// Get current profile for CSV path
	profile, _ := config.GetActiveProfile()
	csvPath := batch.DefaultCSVPathForProfile(profile)

	// Get JIRA client once for all tickets
	client, err := getJiraClient()
	if err != nil {
		fmt.Println(ui.Error(err.Error()))
		return
	}

	// Parse existing entries for duplicate checking
	todayStr := time.Now().Format("02.01.2006")
	entries, _ := batch.ParseCSV(csvPath)

	// Build list of tickets to process
	var ticketKeys []string

	// If --sync, fetch sprint tickets first
	if addSync {
		if !addQuiet {
			fmt.Println("Fetching tickets from current sprint...")
		}
		sprintTickets, err := client.SearchSprintTickets()
		if err != nil {
			fmt.Println(ui.Error("Failed to fetch sprint tickets: " + err.Error()))
			return
		}
		if !addQuiet {
			fmt.Printf("Found %s tickets in sprint\n", ui.Success.Render(fmt.Sprintf("%d", len(sprintTickets))))
			fmt.Println()
		}

		for _, t := range sprintTickets {
			ticketKeys = append(ticketKeys, t.Key)
		}
	}

	// Add manual ticket keys
	for _, arg := range args {
		ticketKeys = append(ticketKeys, strings.ToUpper(strings.TrimSpace(arg)))
	}

	// Track results
	added := 0
	updated := 0
	skipped := 0

	// Process each ticket
	for _, ticketKey := range ticketKeys {
		// Validate ticket key format
		if !ticketKeyPattern.MatchString(ticketKey) {
			if !addQuiet {
				fmt.Printf("%s %s - invalid format (expected PROJECT-123)\n",
					ui.ErrorText.Render("✗"),
					ui.Primary.Render(ticketKey))
			}
			skipped++
			continue
		}

		// Check if ticket exists for today
		var existingEntry *batch.Entry
		for i, e := range entries {
			if strings.EqualFold(e.IssueKey, ticketKey) {
				entryDate := strings.Split(e.Date, " ")[0]
				if entryDate == todayStr {
					existingEntry = &entries[i]
					break
				}
			}
		}

		// Handle existing entry
		if existingEntry != nil {
			// If already has time logged (DONE/SYNC/UPDATED), skip
			if existingEntry.Status == batch.StatusDone ||
				existingEntry.Status == batch.StatusSync ||
				existingEntry.Status == batch.StatusUpdated ||
				existingEntry.TimeSpent != "" {
				if !addQuiet {
					fmt.Printf("%s %s - already logged today (%s %s)\n",
						ui.Muted.Render("⊘"),
						ui.Muted.Render(ticketKey),
						ui.Muted.Render(existingEntry.TimeSpent),
						ui.Muted.Render(existingEntry.Status))
				}
				skipped++
				continue
			}

			// If DRAFT, update description
			if existingEntry.Status == batch.StatusDraft {
				ticket, err := client.GetTicket(ticketKey)
				if err != nil {
					if !addQuiet {
						fmt.Printf("%s %s - %s\n",
							ui.ErrorText.Render("✗"),
							ui.Primary.Render(ticketKey),
							ui.Muted.Render(err.Error()))
					}
					skipped++
					continue
				}

				// Update description (skip in dry-run)
				if !addDryRun {
					err = batch.UpdateEntryDescription(csvPath, existingEntry.RowNumber, ticket.Summary)
					if err != nil {
						if !addQuiet {
							fmt.Printf("%s %s - %s\n",
								ui.ErrorText.Render("✗"),
								ui.Primary.Render(ticketKey),
								ui.Muted.Render(err.Error()))
						}
						skipped++
						continue
					}
				}

				if !addQuiet {
					if addDryRun {
						fmt.Printf("%s %s %s - %s %s\n",
							ui.Primary.Render("↻"),
							ui.Primary.Render(ticketKey),
							ui.Muted.Render("["+ticket.IssueType+"]"),
							ui.Muted.Render(truncateString(ticket.Summary, 40)),
							ui.Muted.Render("(would update)"))
					} else {
						fmt.Printf("%s %s %s - %s %s\n",
							ui.Primary.Render("↻"),
							ui.Primary.Render(ticketKey),
							ui.Muted.Render("["+ticket.IssueType+"]"),
							ui.Muted.Render(truncateString(ticket.Summary, 40)),
							ui.Muted.Render("(updated)"))
					}
				}
				updated++

				// TODO: Consider batching CSV updates instead of re-reading after each ticket
				// Re-read entries for next iteration (skip in dry-run)
				if !addDryRun {
					entries, _ = batch.ParseCSV(csvPath)
				}
				continue
			}
		}

		// Fetch ticket info for new entry
		ticket, err := client.GetTicket(ticketKey)
		if err != nil {
			if !addQuiet {
				fmt.Printf("%s %s - %s\n",
					ui.ErrorText.Render("✗"),
					ui.Primary.Render(ticketKey),
					ui.Muted.Render(err.Error()))
			}
			skipped++
			continue
		}

		// Determine issue key, subtask key, type, and description
		// For sub-tasks: use parent key/type and combined description
		issueKey := ticket.Key
		subtaskKey := ""
		issueType := ticket.IssueType
		description := ticket.Summary

		if ticket.IsSubtask && ticket.ParentKey != "" {
			// TODO: Log error when parent ticket fetch fails instead of silently continuing
			parentTicket, err := client.GetTicket(ticket.ParentKey)
			if err == nil {
				subtaskKey = ticket.Key
				issueKey = parentTicket.Key
				issueType = parentTicket.IssueType
				description = parentTicket.Summary + " > " + ticket.Summary
			}
		}

		// Prepend to CSV with DRAFT status (skip in dry-run)
		if !addDryRun {
			currentDateTime := time.Now().Format("02.01.2006 15:04")
			err = batch.PrependEntryWithStatus(
				csvPath,
				issueKey,
				subtaskKey,
				issueType,
				description,
				"",              // timeSpent - empty for draft
				currentDateTime,
				"",              // comment
				"",              // subtaskLogInd - defaults to N
				batch.StatusDraft,
			)
			if err != nil {
				if !addQuiet {
					fmt.Printf("%s %s - %s\n",
						ui.ErrorText.Render("✗"),
						ui.Primary.Render(ticketKey),
						ui.Muted.Render(err.Error()))
				}
				skipped++
				continue
			}
		}

		// Success
		if !addQuiet {
			displayKey := issueKey
			if subtaskKey != "" {
				displayKey = issueKey + " > " + subtaskKey
			}
			if addDryRun {
				fmt.Printf("%s %s %s - %s %s\n",
					ui.Success.Render("✓"),
					ui.Primary.Render(displayKey),
					ui.Muted.Render("["+issueType+"]"),
					ui.Muted.Render(truncateString(description, 40)),
					ui.Muted.Render("(would add)"))
			} else {
				fmt.Printf("%s %s %s - %s\n",
					ui.Success.Render("✓"),
					ui.Primary.Render(displayKey),
					ui.Muted.Render("["+issueType+"]"),
					ui.Muted.Render(truncateString(description, 40)))
			}
		}
		added++

		// Re-read entries for next iteration (skip in dry-run)
		if !addDryRun {
			entries, _ = batch.ParseCSV(csvPath)
		}
	}

	// Summary (skip if quiet)
	if addQuiet {
		return
	}

	fmt.Println()
	total := added + updated
	if total > 0 {
		summaryParts := []string{}
		if added > 0 {
			if addDryRun {
				summaryParts = append(summaryParts, fmt.Sprintf("%s would be added", ui.Success.Render(fmt.Sprintf("%d", added))))
			} else {
				summaryParts = append(summaryParts, fmt.Sprintf("%s added", ui.Success.Render(fmt.Sprintf("%d", added))))
			}
		}
		if updated > 0 {
			if addDryRun {
				summaryParts = append(summaryParts, fmt.Sprintf("%s would be updated", ui.Primary.Render(fmt.Sprintf("%d", updated))))
			} else {
				summaryParts = append(summaryParts, fmt.Sprintf("%s updated", ui.Primary.Render(fmt.Sprintf("%d", updated))))
			}
		}
		fmt.Println(strings.Join(summaryParts, ", "))

		if addDryRun {
			fmt.Println()
			fmt.Println(ui.Warning("No changes made (dry-run mode)"))
		} else {
			fmt.Println()
			fmt.Println(ui.Muted.Render("Next steps:"))
			fmt.Printf("  1. Edit %s\n", ui.Primary.Render(csvPath))
			fmt.Printf("  2. Fill in TimeSpent (e.g., 2h, 30m)\n")
			fmt.Printf("  3. Change status from DRAFT to empty\n")
			fmt.Printf("  4. Run: %s\n", ui.Primary.Render("hlp jira log --batch"))
		}
	}
	if skipped > 0 {
		fmt.Printf("Skipped %s ticket(s)\n", ui.WarningText.Render(fmt.Sprintf("%d", skipped)))
	}
	if total == 0 && skipped == 0 {
		fmt.Println(ui.Info("No tickets to add"))
	}
}

// RunAutoSync performs sprint sync, returns error if failed.
// If quiet=true, suppresses output except errors.
func RunAutoSync(quiet bool) error {
	// Get current profile for CSV path
	profile, _ := config.GetActiveProfile()
	csvPath := batch.DefaultCSVPathForProfile(profile)

	// Get JIRA client
	client, err := getJiraClient()
	if err != nil {
		return err
	}

	// Parse existing entries for duplicate checking
	todayStr := time.Now().Format("02.01.2006")
	entries, _ := batch.ParseCSV(csvPath)

	// Fetch sprint tickets
	if !quiet {
		fmt.Println("Fetching tickets from current sprint...")
	}
	sprintTickets, err := client.SearchSprintTickets()
	if err != nil {
		return fmt.Errorf("failed to fetch sprint tickets: %w", err)
	}

	if !quiet {
		fmt.Printf("Found %d tickets in sprint\n", len(sprintTickets))
	}

	// Track results
	added := 0
	updated := 0
	skipped := 0

	// Process each ticket
	for _, sprintTicket := range sprintTickets {
		ticketKey := sprintTicket.Key

		// Check if ticket exists for today
		var existingEntry *batch.Entry
		for i, e := range entries {
			if strings.EqualFold(e.IssueKey, ticketKey) {
				entryDate := strings.Split(e.Date, " ")[0]
				if entryDate == todayStr {
					existingEntry = &entries[i]
					break
				}
			}
		}

		// Handle existing entry
		if existingEntry != nil {
			// If already has time logged, skip
			if existingEntry.Status == batch.StatusDone ||
				existingEntry.Status == batch.StatusSync ||
				existingEntry.Status == batch.StatusUpdated ||
				existingEntry.TimeSpent != "" {
				skipped++
				continue
			}

			// If DRAFT, update description
			if existingEntry.Status == batch.StatusDraft {
				ticket, err := client.GetTicket(ticketKey)
				if err != nil {
					skipped++
					continue
				}

				err = batch.UpdateEntryDescription(csvPath, existingEntry.RowNumber, ticket.Summary)
				if err != nil {
					skipped++
					continue
				}

				if !quiet {
					fmt.Printf("%s %s %s - %s %s\n",
						ui.Primary.Render("↻"),
						ui.Primary.Render(ticketKey),
						ui.Muted.Render("["+ticket.IssueType+"]"),
						ui.Muted.Render(truncateString(ticket.Summary, 40)),
						ui.Muted.Render("(updated)"))
				}
				updated++
				entries, _ = batch.ParseCSV(csvPath)
				continue
			}
		}

		// Fetch ticket info for new entry
		ticket, err := client.GetTicket(ticketKey)
		if err != nil {
			skipped++
			continue
		}

		// Prepare entry
		currentDateTime := time.Now().Format("02.01.2006 15:04")

		// Prepend to CSV with DRAFT status
		err = batch.PrependEntryWithStatus(
			csvPath,
			ticketKey,
			"",              // subtaskKey - empty for now
			ticket.IssueType,
			ticket.Summary,
			"",              // timeSpent - empty for draft
			currentDateTime,
			"",              // comment
			"",              // subtaskLogInd - defaults to N
			batch.StatusDraft,
		)
		if err != nil {
			skipped++
			continue
		}

		if !quiet {
			fmt.Printf("%s %s %s - %s\n",
				ui.Success.Render("✓"),
				ui.Primary.Render(ticketKey),
				ui.Muted.Render("["+ticket.IssueType+"]"),
				ui.Muted.Render(truncateString(ticket.Summary, 40)))
		}
		added++
		entries, _ = batch.ParseCSV(csvPath)
	}

	// Summary (only if not quiet and changes made)
	if !quiet && (added > 0 || updated > 0) {
		fmt.Printf("\nAuto-sync: %d added, %d updated, %d skipped\n", added, updated, skipped)
	}

	// Update last sync time
	config.UpdateLastSyncTime()

	return nil
}

// IsQuietMode checks if --quiet flag is set on the command or its parents
func IsQuietMode(cmd *cobra.Command) bool {
	quietFlag := cmd.Flag("quiet")
	if quietFlag != nil && quietFlag.Changed {
		return true
	}
	return addQuiet
}
