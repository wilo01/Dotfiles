package agent

import (
	"fmt"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	fanoutJQL             string
	fanoutMax             int
	fanoutDryRun          bool
	fanoutIncludeSubtasks bool
	fanoutYes             bool
)

var fanoutCmd = &cobra.Command{
	Use:   "fanout",
	Short: "Spawn agents for all assigned tickets",
	Long: `Runs the configured JQL (default: all your open tickets), shows the plan,
and spawns one agent per ticket with a short stagger between launches.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFanout()
	},
}

func init() {
	fanoutCmd.Flags().StringVar(&fanoutJQL, "jql", "", "override the configured fanout JQL")
	fanoutCmd.Flags().IntVar(&fanoutMax, "max", 0, "spawn at most N agents (default: agent.max_parallel)")
	fanoutCmd.Flags().BoolVar(&fanoutDryRun, "dry-run", false, "resolve and print the plan for every ticket without spawning")
	fanoutCmd.Flags().BoolVar(&fanoutIncludeSubtasks, "include-subtasks", false, "also spawn agents for subtasks")
	fanoutCmd.Flags().BoolVar(&fanoutYes, "yes", false, "skip the confirmation prompt")
}

func runFanout() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client, err := getJiraClient()
	if err != nil {
		return err
	}

	jql := cfg.Agent.JQL
	if fanoutJQL != "" {
		jql = fanoutJQL
	}

	tickets, err := client.Search(jql, 50)
	if err != nil {
		return err
	}

	var candidates []struct {
		Key, Summary, Type, Status string
		Skip                       string
	}
	for _, t := range tickets {
		entry := struct {
			Key, Summary, Type, Status string
			Skip                       string
		}{Key: t.Key, Summary: t.Summary, Type: t.IssueType, Status: t.Status}

		if t.IsSubtask && !fanoutIncludeSubtasks {
			entry.Skip = "subtask"
		} else if sessionExists(t.Key) && agentRunning(t.Key) {
			entry.Skip = "agent already running"
		}
		candidates = append(candidates, entry)
	}

	if len(candidates) == 0 {
		fmt.Println(ui.Info("no tickets matched: " + jql))
		return nil
	}

	var rows [][]string
	spawnCount := 0
	maxSpawn := cfg.Agent.MaxParallel
	if fanoutMax > 0 {
		maxSpawn = fanoutMax
	}
	for i := range candidates {
		c := &candidates[i]
		action := "spawn"
		if c.Skip != "" {
			action = "skip: " + c.Skip
		} else if spawnCount >= maxSpawn {
			c.Skip = "over --max limit"
			action = "skip: over limit"
		} else {
			spawnCount++
		}
		summary := c.Summary
		if len(summary) > 60 {
			summary = summary[:57] + "..."
		}
		rows = append(rows, []string{c.Key, c.Type, c.Status, summary, action})
	}

	fmt.Println(ui.Table([]string{"KEY", "TYPE", "STATUS", "SUMMARY", "ACTION"}, rows))
	fmt.Println(ui.KeyValue("JQL", jql))

	if spawnCount == 0 {
		fmt.Println(ui.Info("nothing to spawn"))
		return nil
	}
	if !fanoutDryRun && !fanoutYes {
		if !ui.ConfirmAction(fmt.Sprintf("Spawn %d agent(s)?", spawnCount)) {
			return nil
		}
	}

	var failures []string
	launched := 0
	for _, c := range candidates {
		if c.Skip != "" {
			continue
		}
		ticket, err := client.GetTicket(c.Key)
		if err != nil {
			failures = append(failures, c.Key+": "+err.Error())
			continue
		}
		if launched > 0 && !fanoutDryRun {
			time.Sleep(4 * time.Second)
		}
		err = startAgent(cfg, client, ticket, startOptions{DryRun: fanoutDryRun, Manual: false})
		if err != nil {
			failures = append(failures, c.Key+": "+err.Error())
			fmt.Println(ui.Error(c.Key + ": " + err.Error()))
			continue
		}
		launched++
		fmt.Println()
	}

	fmt.Println(ui.Header("Fanout summary"))
	launchedLabel := "Launched"
	if fanoutDryRun {
		launchedLabel = "Planned (dry-run)"
	}
	fmt.Println(ui.KeyValue(launchedLabel, fmt.Sprintf("%d", launched)))
	if len(failures) > 0 {
		fmt.Println(ui.Error("Failures:\n  " + strings.Join(failures, "\n  ")))
	}
	return nil
}
