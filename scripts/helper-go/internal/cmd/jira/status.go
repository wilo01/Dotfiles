package jira

import (
	"fmt"
	"time"

	"github.com/dariuszw/hlp/internal/context"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/dariuszw/hlp/pkg/duration"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display current work status",
	Long: `Display information about the current work context and today's logged time.

Shows:
- Current git repository and branch
- Detected JIRA ticket
- Today's logged work (if JIRA is configured)`,
	Run: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) {
	detector := context.NewDetector()
	ctx := detector.Detect()

	fmt.Println(ui.Header("Work Status"))
	fmt.Println()

	// Context info
	if ctx.Repository != "" {
		fmt.Println(ui.KeyValue("Repository", ctx.Repository))
	}
	if ctx.Branch != "" {
		fmt.Println(ui.KeyValue("Branch", ctx.Branch))
	}
	if ctx.Ticket != "" {
		fmt.Println(ui.KeyValue("Ticket", ui.Primary.Render(ctx.Ticket)))
		if ctx.Description != "" {
			fmt.Println(ui.KeyValue("Description", ctx.Description))
		}
	} else {
		fmt.Println(ui.Muted.Render("  No ticket detected"))
	}

	fmt.Println()

	// Today's work
	client, err := getJiraClient()
	if err != nil {
		fmt.Println(ui.Muted.Render("  JIRA not configured - run 'hlp jira config'"))
		return
	}

	if ctx.Ticket == "" {
		return
	}

	fmt.Println(ui.Title.Render("Today's Work"))

	worklogs, err := client.GetWorklogs(ctx.Ticket)
	if err != nil {
		fmt.Println(ui.Warning("Could not fetch worklogs: " + err.Error()))
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	var totalToday time.Duration
	var todayLogs []struct {
		time    time.Duration
		comment string
	}

	for _, w := range worklogs {
		if w.Started.Truncate(24 * time.Hour).Equal(today) {
			totalToday += w.TimeSpent
			todayLogs = append(todayLogs, struct {
				time    time.Duration
				comment string
			}{w.TimeSpent, w.Comment})
		}
	}

	if len(todayLogs) == 0 {
		fmt.Println(ui.Muted.Render("  No work logged today"))
	} else {
		for _, log := range todayLogs {
			comment := log.comment
			if comment == "" {
				comment = "(no comment)"
			}
			fmt.Printf("  %s %s\n",
				ui.Success.Render(duration.Format(log.time)),
				ui.Muted.Render(comment))
		}
		fmt.Println()
		fmt.Println(ui.KeyValue("Total today", ui.SuccessBold.Render(duration.Format(totalToday))))
	}
}
