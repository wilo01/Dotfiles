package agent

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
)

// testConfig points the worktree root at an empty dir so resumeScan finds no
// repos and every candidate falls through to the triage decision.
func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.Default()
	cfg.Agent.WorktreeRoot = t.TempDir()
	cfg.Agent.MaxParallel = 4
	return cfg
}

// recordingTriage is a triageFunc that records the keys it was asked about and
// returns whatever the per-key table says.
type recordingTriage struct {
	mu       sync.Mutex
	calls    []string
	verdicts map[string]*TriageVerdict
	errs     map[string]error
}

func (r *recordingTriage) fn(_ *config.Config, _ *internalJira.Client, ticket *internalJira.Ticket, _ []string, warn func(string)) (*TriageVerdict, error) {
	r.mu.Lock()
	r.calls = append(r.calls, ticket.Key)
	r.mu.Unlock()

	if err, ok := r.errs[ticket.Key]; ok {
		return nil, err
	}
	if v, ok := r.verdicts[ticket.Key]; ok {
		return v, nil
	}
	warn("no verdict configured")
	return nil, fmt.Errorf("unexpected key %s", ticket.Key)
}

func (r *recordingTriage) called(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Contains(r.calls, key)
}

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

func TestPreResolveWithoutTriageNeverCallsTheModel(t *testing.T) {
	cfg := testConfig(t)
	rec := &recordingTriage{}
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-2", Status: "Backlog"},
	)

	preResolve(cfg, nil, candidates, []string{"repo-a"}, startOptions{}, false, 0, rec.fn)

	if len(rec.calls) != 0 {
		t.Fatalf("triage ran with --triage off: %v", rec.calls)
	}
	for _, key := range []string{"VIS-1", "VIS-2"} {
		if got := findCandidate(t, candidates, key).Skip; got != "needs triage" {
			t.Errorf("%s: skip = %q, want %q", key, got, "needs triage")
		}
	}
}

func TestPreResolveTriagesOnlyConfiguredStatuses(t *testing.T) {
	cfg := testConfig(t)
	rec := &recordingTriage{
		verdicts: map[string]*TriageVerdict{
			"VIS-1": {Repo: "repo-a", Confidence: "high", Reason: "backlog item"},
			"VIS-2": {Repo: "repo-a", Confidence: "high", Reason: "todo item"},
			"VIS-3": {Repo: "repo-a", Confidence: "high", Reason: "lowercase todo"},
		},
	}
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1", Status: "Backlog"},
		internalJira.Ticket{Key: "VIS-2", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-3", Status: "to do"},
		internalJira.Ticket{Key: "VIS-4", Status: "In Progress"},
		internalJira.Ticket{Key: "VIS-5", Status: "Code Review"},
	)

	preResolve(cfg, nil, candidates, []string{"repo-a"}, startOptions{}, true, 0, rec.fn)

	for _, key := range []string{"VIS-1", "VIS-2", "VIS-3"} {
		if !rec.called(key) {
			t.Errorf("%s should have been triaged", key)
		}
		c := findCandidate(t, candidates, key)
		if c.Repo != "repo-a" {
			t.Errorf("%s: repo = %q, want repo-a", key, c.Repo)
		}
		if !strings.HasPrefix(c.How, "AI triage:") {
			t.Errorf("%s: how = %q, want an AI triage provenance", key, c.How)
		}
	}

	for _, key := range []string{"VIS-4", "VIS-5"} {
		if rec.called(key) {
			t.Errorf("%s is past the triage statuses and must not be triaged", key)
		}
		if skip := findCandidate(t, candidates, key).Skip; skip == "" {
			t.Errorf("%s: expected a skip reason, got none", key)
		}
	}
}

func TestPreResolveDegradesToPicker(t *testing.T) {
	cfg := testConfig(t)
	rec := &recordingTriage{
		verdicts: map[string]*TriageVerdict{
			"VIS-1": {Repo: "repo-a", Confidence: "low", Reason: "could be either"},
			"VIS-3": {Repo: "repo-a", Confidence: "high", Reason: "clear"},
		},
		errs: map[string]error{"VIS-2": fmt.Errorf("triage command failed")},
	}
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-2", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-3", Status: "To Do"},
	)

	preResolve(cfg, nil, candidates, []string{"repo-a"}, startOptions{}, true, 0, rec.fn)

	for _, key := range []string{"VIS-1", "VIS-2"} {
		c := findCandidate(t, candidates, key)
		if !c.NeedsPicker {
			t.Errorf("%s: NeedsPicker = false, want true", key)
		}
		if c.Repo != "" {
			t.Errorf("%s: repo = %q, want empty so the picker decides", key, c.Repo)
		}
		if len(c.Warnings) == 0 {
			t.Errorf("%s: degraded silently, expected a warning", key)
		}
	}

	// One ticket's failure must not affect the others in the batch.
	if c := findCandidate(t, candidates, "VIS-3"); c.Repo != "repo-a" || c.NeedsPicker {
		t.Errorf("VIS-3 was affected by its neighbours: repo=%q picker=%v", c.Repo, c.NeedsPicker)
	}
}

func TestPreResolveSkipsAlreadySkippedCandidates(t *testing.T) {
	cfg := testConfig(t)
	rec := &recordingTriage{}
	candidates := candidatesFor(internalJira.Ticket{Key: "VIS-1", Status: "To Do"})
	candidates[0].Skip = "over --max limit"

	preResolve(cfg, nil, candidates, []string{"repo-a"}, startOptions{}, true, 0, rec.fn)

	if rec.called("VIS-1") {
		t.Error("an over-limit ticket must not cost a model call")
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

func TestPreResolveCutsBeforeTriagingButAfterScanning(t *testing.T) {
	cfg := testConfig(t)
	rec := &recordingTriage{
		verdicts: map[string]*TriageVerdict{
			"VIS-1": {Repo: "repo-a", Confidence: "high", Reason: "first"},
			"VIS-2": {Repo: "repo-a", Confidence: "high", Reason: "second"},
		},
	}
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-2", Status: "To Do"},
	)

	preResolve(cfg, nil, candidates, []string{"repo-a"}, startOptions{}, true, 1, rec.fn)

	if !rec.called("VIS-1") {
		t.Error("the in-limit ticket should have been triaged")
	}
	if rec.called("VIS-2") {
		t.Error("an over-limit ticket must not cost a model call")
	}
	if candidates[1].Skip != "over --max limit" {
		t.Errorf("VIS-2: skip = %q, want %q", candidates[1].Skip, "over --max limit")
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

func TestBuildLaunchCommandPermissionMode(t *testing.T) {
	cfg := config.Default()

	t.Run("config default applies to interactive sessions", func(t *testing.T) {
		got := buildLaunchCommand(cfg, "VIS-1", startOptions{Manual: true})

		if got != "JIRA_KEY=VIS-1 claude --permission-mode auto" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("flag overrides config", func(t *testing.T) {
		got := buildLaunchCommand(cfg, "VIS-1", startOptions{Manual: true, PermissionMode: "plan"})

		if !strings.Contains(got, "--permission-mode plan") || strings.Contains(got, "auto") {
			t.Errorf("got %q", got)
		}
	})

	t.Run("empty config mode omits the flag", func(t *testing.T) {
		bare := config.Default()
		bare.Agent.PermissionMode = ""

		if got := buildLaunchCommand(bare, "VIS-1", startOptions{Manual: true}); got != "JIRA_KEY=VIS-1 claude" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("mode comes before the briefing prompt", func(t *testing.T) {
		got := buildLaunchCommand(cfg, "VIS-1", startOptions{Manual: true, ContextPrompt: "brief {{KEY}}"})

		want := `JIRA_KEY=VIS-1 claude --permission-mode auto "brief VIS-1"`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("autonomous launch is unaffected", func(t *testing.T) {
		got := buildLaunchCommand(cfg, "VIS-1", startOptions{PermissionMode: "plan"})

		if strings.Contains(got, "--permission-mode") {
			t.Errorf("spin should keep using --dangerously-skip-permissions: %q", got)
		}
	})
}

func TestStatusAllowsTriage(t *testing.T) {
	cfg := config.Default()

	tests := map[string]bool{
		"Backlog":     true,
		"To Do":       true,
		"to do":       true,
		"BACKLOG":     true,
		"In Progress": false,
		"Code Review": false,
		"":            false,
	}
	for status, want := range tests {
		if got := statusAllowsTriage(cfg, status); got != want {
			t.Errorf("statusAllowsTriage(%q) = %v, want %v", status, got, want)
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
		got := buildLaunchCommand(cfg, "VIS-1", startOptions{Manual: true})

		if strings.Contains(got, "--dangerously-skip-permissions") {
			t.Errorf("manual session must keep permission prompts: %q", got)
		}
		if strings.Contains(got, "/agent-run") || strings.Contains(got, "CLAUDE_AGENT_MODE") {
			t.Errorf("manual session must not run autonomously: %q", got)
		}
		if got != "JIRA_KEY=VIS-1 claude --permission-mode auto" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("context prompt is templated and quoted", func(t *testing.T) {
		got := buildLaunchCommand(cfg, "VIS-1", startOptions{
			Manual:        true,
			ContextPrompt: "brief me on {{KEY}}",
		})

		if got != `JIRA_KEY=VIS-1 claude --permission-mode auto "brief me on VIS-1"` {
			t.Errorf("got %q", got)
		}
	})

	t.Run("autonomous session is unchanged", func(t *testing.T) {
		got := buildLaunchCommand(cfg, "VIS-1", startOptions{})

		want := `CLAUDE_AGENT_MODE=1 JIRA_KEY=VIS-1 claude --dangerously-skip-permissions "/agent-run VIS-1"`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("context prompt is ignored when autonomous", func(t *testing.T) {
		got := buildLaunchCommand(cfg, "VIS-1", startOptions{ContextPrompt: "brief me on {{KEY}}"})

		if strings.Contains(got, "brief me") {
			t.Errorf("autonomous launch must not carry a briefing prompt: %q", got)
		}
	})
}

func TestPreResolveWithForcedRepo(t *testing.T) {
	cfg := testConfig(t)
	rec := &recordingTriage{}
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-2", Status: "In Progress"},
	)

	preResolve(cfg, nil, candidates, []string{"repo-a"}, startOptions{Repo: "repo-a"}, true, 0, rec.fn)

	if len(rec.calls) != 0 {
		t.Errorf("--repo settles every ticket, triage should not run: %v", rec.calls)
	}
	for i := range candidates {
		if candidates[i].Repo != "repo-a" || candidates[i].Skip != "" {
			t.Errorf("%s: repo=%q skip=%q, want repo-a and no skip",
				candidates[i].Ticket.Key, candidates[i].Repo, candidates[i].Skip)
		}
	}
}

func TestPreResolveInteractiveDefersEveryRepoToThePicker(t *testing.T) {
	cfg := testConfig(t)
	rec := &recordingTriage{}
	candidates := candidatesFor(
		internalJira.Ticket{Key: "VIS-1", Status: "To Do"},
		internalJira.Ticket{Key: "VIS-2", Status: "Backlog"},
	)

	preResolve(cfg, nil, candidates, []string{"repo-a"}, startOptions{Interactive: true}, true, 0, rec.fn)

	if len(rec.calls) != 0 {
		t.Errorf("--interactive means the user picks, triage should not run: %v", rec.calls)
	}
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
