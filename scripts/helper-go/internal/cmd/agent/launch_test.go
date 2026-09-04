package agent

import (
	"strings"
	"testing"

	"github.com/dariuszw/hlp/internal/config"
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

// Reopening a ticket starts a fresh conversation. Resuming by UUID aborts the
// launch outright once that conversation is gone, leaving the window at a bare
// shell, so no session flag is passed after the first launch.
func TestBuildLaunchCommandStartsFreshOnAStartedSession(t *testing.T) {
	cfg := config.Default()
	started := launchTask()
	started.SessionStarted = true

	got := buildLaunchCommand(cfg, started, launchOptions{Manual: true})

	if strings.Contains(got, "--resume") {
		t.Errorf("a started session must not be resumed: %q", got)
	}
	if strings.Contains(got, "--session-id") {
		t.Errorf("a started session must not be re-created: %q", got)
	}
	if strings.Contains(got, "  ") {
		t.Errorf("dropping the session flag left a double space: %q", got)
	}
	if want := "JIRA_KEY=VIS-1 claude"; !strings.HasPrefix(got, want) {
		t.Errorf("got %q, want prefix %q", got, want)
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

// A restored tmux session brings the hexer window back as an idle shell, so
// only the login shell must read as "nothing running" — anything else is a
// provisioning run that a second `start` would stack on top of.
func TestHexerProvisioningIgnoresIdleShells(t *testing.T) {
	cases := []struct {
		name string
		cmds []string
		want bool
	}{
		{"restored idle shell", []string{"zsh"}, false},
		{"no window at all", nil, false},
		{"provisioning under bash", []string{"bash"}, true},
		{"waiting on docker", []string{"docker"}, true},
		{"liquibase java step", []string{"java"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := paneCommandsIndicateWork(tc.cmds); got != tc.want {
				t.Errorf("paneCommandsIndicateWork(%v) = %v, want %v", tc.cmds, got, tc.want)
			}
		})
	}
}
