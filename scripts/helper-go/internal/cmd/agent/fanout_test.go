package agent

import (
	"strings"
	"testing"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
	"github.com/dariuszw/hlp/internal/task"
)

// testConfig points the repos root at an empty dir so the branch scan finds
// nothing and every candidate falls through to "no branch yet".
func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.Default()
	cfg.Agent.ReposRoot = t.TempDir()
	cfg.Agent.TasksRoot = t.TempDir()
	cfg.Agent.MaxParallel = 4
	return cfg
}

// launchTask is a minimal task whose Claude session has never been started, so
// buildLaunchCommand emits --session-id rather than --resume.
func launchTask() task.Task {
	return task.Task{JiraKey: "VIS-1", ClaudeSessionID: testSessionID}
}

const testSessionID = "11111111-2222-4333-8444-555555555555"

func candidatesFor(tickets ...internalJira.Ticket) []candidate {
	out := make([]candidate, 0, len(tickets))
	for _, t := range tickets {
		out = append(out, candidate{Ticket: t})
	}
	return out
}

func findCandidate(t *testing.T, candidates []candidate, key string) *candidate {
	t.Helper()
	for i := range candidates {
		if candidates[i].Ticket.Key == key {
			return &candidates[i]
		}
	}
	t.Fatalf("candidate %s not found", key)
	return nil
}

func TestPreResolveSkipsAlreadySkippedCandidates(t *testing.T) {
	cfg := testConfig(t)
	candidates := candidatesFor(internalJira.Ticket{Key: "VIS-1", Status: "To Do"})
	candidates[0].Skip = "over --max limit"

	preResolve(cfg, candidates, []string{"repo-a"}, startOptions{}, 0)

	if candidates[0].Repo != "" {
		t.Error("an over-limit ticket must not be resolved")
	}
	if candidates[0].Skip != "over --max limit" {
		t.Errorf("skip reason overwritten: %q", candidates[0].Skip)
	}
}

func TestApplyMaxLimit(t *testing.T) {
	cfg := testConfig(t)
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1"},
		internalJira.Ticket{Key: "VIS-2"},
		internalJira.Ticket{Key: "VIS-3"},
	)

	applyMaxLimit(cfg, candidates, 2)

	if candidates[0].Skip != "" || candidates[1].Skip != "" {
		t.Errorf("first two should spawn, got %q and %q", candidates[0].Skip, candidates[1].Skip)
	}
	if candidates[2].Skip != "over --max limit" {
		t.Errorf("third: skip = %q, want %q", candidates[2].Skip, "over --max limit")
	}
}

// A ticket that can't be spawned must not consume a --max slot, otherwise a
// resumable ticket further down the list is cut for no reason.
func TestApplyMaxLimitIgnoresSkippedCandidates(t *testing.T) {
	cfg := testConfig(t)
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1"},
		internalJira.Ticket{Key: "VIS-2"},
		internalJira.Ticket{Key: "VIS-3"},
	)
	candidates[0].Skip = "needs triage"
	candidates[1].Skip = "subtask"

	applyMaxLimit(cfg, candidates, 1)

	if candidates[2].Skip != "" {
		t.Errorf("VIS-3 should have taken the free slot, got %q", candidates[2].Skip)
	}
}

func TestBuildCandidatesSkipsSubtasks(t *testing.T) {
	candidates := buildCandidates([]internalJira.Ticket{{Key: "VIS-1", IsSubtask: true}})

	if candidates[0].Skip != "subtask" {
		t.Errorf("skip = %q, want %q", candidates[0].Skip, "subtask")
	}
}

func TestResolveContextPrompt(t *testing.T) {
	cfg := config.Default()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"unset stays empty", "", ""},
		{"sentinel resolves to config", contextPromptFromConfig, cfg.Agent.ContextPrompt},
		{"custom passes through", "where did I leave off on {{KEY}}?", "where did I leave off on {{KEY}}?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveContextPrompt(cfg, tt.input); got != tt.want {
				t.Errorf("resolveContextPrompt(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}

	if !strings.Contains(cfg.Agent.ContextPrompt, "{{KEY}}") {
		t.Error("the default context prompt must template the ticket key")
	}
}

func TestValidateFanoutFlags(t *testing.T) {
	if err := validateFanoutFlags(startOptions{ContextPrompt: "/brief {{KEY}}"}, true); err == nil {
		t.Error("--context with --spin should be rejected")
	}
	if err := validateFanoutFlags(startOptions{ContextPrompt: "/brief {{KEY}}"}, false); err != nil {
		t.Errorf("--context alone should be valid, got %v", err)
	}
	if err := validateFanoutFlags(startOptions{}, true); err != nil {
		t.Errorf("--spin alone should be valid, got %v", err)
	}
	if err := validateFanoutFlags(startOptions{PermissionMode: "autoo"}, false); err == nil {
		t.Error("a typo'd permission mode should be rejected before launch")
	}
}

func TestValidatePermissionMode(t *testing.T) {
	for _, mode := range append([]string{""}, permissionModes...) {
		if err := validatePermissionMode(mode); err != nil {
			t.Errorf("mode %q should be valid, got %v", mode, err)
		}
	}
	for _, mode := range []string{"autoo", "AUTO", "accept-edits", "yolo"} {
		if err := validatePermissionMode(mode); err == nil {
			t.Errorf("mode %q should be rejected", mode)
		}
	}
}

func TestCandidateRepoCell(t *testing.T) {
	tests := []struct {
		name string
		c    candidate
		want string
	}{
		{"resolved", candidate{Repo: "tds-suite"}, "tds-suite"},
		{"picker", candidate{NeedsPicker: true}, "picker"},
		{"unresolved", candidate{}, "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.repoCell(); got != tt.want {
				t.Errorf("repoCell() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildLaunchCommand(t *testing.T) {
	cfg := config.Default()

	t.Run("manual session is plain and interactive", func(t *testing.T) {
		got := buildLaunchCommand(cfg, launchTask(), launchOptions{Manual: true})

		if strings.Contains(got, "--dangerously-skip-permissions") {
			t.Errorf("manual session must keep permission prompts: %q", got)
		}
		if strings.Contains(got, "/agent-run") || strings.Contains(got, "CLAUDE_AGENT_MODE") {
			t.Errorf("manual session must not run autonomously: %q", got)
		}
		want := "JIRA_KEY=VIS-1 claude --session-id " + testSessionID + " --permission-mode auto"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("context prompt is templated and quoted", func(t *testing.T) {
		got := buildLaunchCommand(cfg, launchTask(), launchOptions{
			Manual:        true,
			ContextPrompt: "brief me on {{KEY}}",
		})

		want := "JIRA_KEY=VIS-1 claude --session-id " + testSessionID + ` --permission-mode auto "brief me on VIS-1"`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("autonomous session carries the agent prompt", func(t *testing.T) {
		got := buildLaunchCommand(cfg, launchTask(), launchOptions{})

		want := "CLAUDE_AGENT_MODE=1 JIRA_KEY=VIS-1 claude --dangerously-skip-permissions --session-id " +
			testSessionID + ` "/agent-run VIS-1"`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("context prompt is ignored when autonomous", func(t *testing.T) {
		got := buildLaunchCommand(cfg, launchTask(), launchOptions{ContextPrompt: "brief me on {{KEY}}"})

		if strings.Contains(got, "brief me") {
			t.Errorf("autonomous launch must not carry a briefing prompt: %q", got)
		}
	})
}

func TestBuildLaunchCommandPermissionMode(t *testing.T) {
	cfg := config.Default()

	t.Run("config default applies to interactive sessions", func(t *testing.T) {
		got := buildLaunchCommand(cfg, launchTask(), launchOptions{Manual: true})

		if !strings.Contains(got, "--permission-mode auto") {
			t.Errorf("got %q", got)
		}
	})

	t.Run("flag overrides config", func(t *testing.T) {
		got := buildLaunchCommand(cfg, launchTask(), launchOptions{Manual: true, PermissionMode: "plan"})

		if !strings.Contains(got, "--permission-mode plan") || strings.Contains(got, "auto") {
			t.Errorf("got %q", got)
		}
	})

	t.Run("empty config mode omits the flag", func(t *testing.T) {
		bare := config.Default()
		bare.Agent.PermissionMode = ""

		got := buildLaunchCommand(bare, launchTask(), launchOptions{Manual: true})
		if got != "JIRA_KEY=VIS-1 claude --session-id "+testSessionID {
			t.Errorf("got %q", got)
		}
	})

	t.Run("autonomous launch is unaffected", func(t *testing.T) {
		got := buildLaunchCommand(cfg, launchTask(), launchOptions{PermissionMode: "plan"})

		if strings.Contains(got, "--permission-mode") {
			t.Errorf("spin should keep using --dangerously-skip-permissions: %q", got)
		}
	})
}

// A task whose session has already been opened resumes it, so reopening a
// ticket continues the same conversation rather than starting a new one.
func TestBuildLaunchCommandResumesAStartedSession(t *testing.T) {
	cfg := config.Default()
	started := launchTask()
	started.SessionStarted = true

	got := buildLaunchCommand(cfg, started, launchOptions{Manual: true})

	if !strings.Contains(got, "--resume "+testSessionID) {
		t.Errorf("expected --resume, got %q", got)
	}
	if strings.Contains(got, "--session-id") {
		t.Errorf("a started session must not be re-created: %q", got)
	}
}

func TestPreResolveWithForcedRepo(t *testing.T) {
	cfg := testConfig(t)
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-2", Status: "In Progress"},
	)

	preResolve(cfg, candidates, []string{"repo-a"}, startOptions{Repo: "repo-a"}, 0)

	for i := range candidates {
		if candidates[i].Repo != "repo-a" || candidates[i].Skip != "" {
			t.Errorf("%s: repo=%q skip=%q, want repo-a and no skip",
				candidates[i].Ticket.Key, candidates[i].Repo, candidates[i].Skip)
		}
	}
}

func TestPreResolveInteractiveDefersEveryRepoToThePicker(t *testing.T) {
	cfg := testConfig(t)
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-2", Status: "Backlog"},
	)

	preResolve(cfg, candidates, []string{"repo-a"}, startOptions{Interactive: true}, 0)

	for i := range candidates {
		if !candidates[i].NeedsPicker {
			t.Errorf("%s: NeedsPicker = false, want true", candidates[i].Ticket.Key)
		}
	}
}

// --context and --permission-mode only shape an interactive claude, so an
// autonomous run must reject them instead of accepting and ignoring them.
func TestValidateLaunchFlagsRejectsInteractiveOnlyFlagsOnSpin(t *testing.T) {
	tests := []struct {
		name    string
		opts    startOptions
		wantErr bool
	}{
		{"context on spin", startOptions{ContextPrompt: "brief {{KEY}}"}, true},
		{"permission mode on spin", startOptions{PermissionMode: "auto"}, true},
		{"context on start", startOptions{Manual: true, ContextPrompt: "brief {{KEY}}"}, false},
		{"permission mode on start", startOptions{Manual: true, PermissionMode: "auto"}, false},
		{"bad mode on start", startOptions{Manual: true, PermissionMode: "autoo"}, true},
		{"bare spin", startOptions{}, false},
		{"bare start", startOptions{Manual: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.opts
			if err := validateLaunchFlags(&opts); (err != nil) != tt.wantErr {
				t.Errorf("validateLaunchFlags(%+v) error = %v, wantErr %v", tt.opts, err, tt.wantErr)
			}
		})
	}
}

// fanout --spin must reject the same combinations that spin itself does.
func TestFanoutAndSpinAgreeOnFlagValidity(t *testing.T) {
	for _, opts := range []startOptions{
		{ContextPrompt: "brief {{KEY}}"},
		{PermissionMode: "auto"},
		{PermissionMode: "autoo"},
		{},
	} {
		spinOpts := opts
		spinOpts.Manual = false
		wantErr := validateLaunchFlags(&spinOpts) != nil

		if gotErr := validateFanoutFlags(opts, true) != nil; gotErr != wantErr {
			t.Errorf("%+v: fanout --spin error = %v, spin error = %v", opts, gotErr, wantErr)
		}
	}
}
