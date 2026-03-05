package dev

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dariuszw/hlp/internal/commits"
)

func TestExtractJiraTicket(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{"standard branch", "VIS-1234-some-feature", "VIS-1234"},
		{"lowercase", "vis-5678-fix-bug", "VIS-5678"},
		{"mixed case", "Vis-9999-mixed", "VIS-9999"},
		{"no match", "feature-branch-no-ticket", ""},
		{"just numbers", "1234", ""},
		{"empty string", "", ""},
		{"ticket only", "VIS-100", "VIS-100"},
		{"multi-project prefix", "TDT-42-something", "TDT-42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJiraTicket(tt.branch)
			if got != tt.want {
				t.Errorf("extractJiraTicket(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestBuildCommitMessage(t *testing.T) {
	entry := &commits.CommitEntry{
		Ticket:     "VIS-6457",
		BranchName: "VIS-6457-some-fix",
		Logs:       []string{"Fixed the bug.", "Updated tests."},
	}

	tests := []struct {
		name        string
		ticket      string
		version     string
		newBranch   string
		entry       *commits.CommitEntry
		prNumber    int
		targetBranch string
		wantContains []string
	}{
		{
			name:        "with entry and version",
			ticket:      "VIS-6457",
			version:     "13.2AV",
			newBranch:   "VIS-6457-some-fix-13.2av",
			entry:       entry,
			prNumber:    1063,
			targetBranch: "maintenance/13.2AV",
			wantContains: []string{
				"VIS-6457 [13.2AV]",
				"VIS-6457-some-fix-13.2av",
				"- Fixed the bug.",
				"- Updated tests.",
				"Cherry-pick of #1063 onto maintenance/13.2AV",
			},
		},
		{
			name:        "nil entry falls back",
			ticket:      "VIS-1234",
			version:     "13.1AV",
			newBranch:   "VIS-1234-fix-13.1av",
			entry:       nil,
			prNumber:    999,
			targetBranch: "maintenance/13.1AV",
			wantContains: []string{
				"Cherry-pick PR #999 onto maintenance/13.1AV",
			},
		},
		{
			name:        "empty version falls back",
			ticket:      "VIS-1234",
			version:     "",
			newBranch:   "VIS-1234-fix",
			entry:       entry,
			prNumber:    500,
			targetBranch: "some-branch",
			wantContains: []string{
				"Cherry-pick PR #500 onto some-branch",
			},
		},
		{
			name:        "entry with no logs",
			ticket:      "VIS-6457",
			version:     "13.2AV",
			newBranch:   "VIS-6457-fix-13.2av",
			entry:       &commits.CommitEntry{Ticket: "VIS-6457"},
			prNumber:    1063,
			targetBranch: "maintenance/13.2AV",
			wantContains: []string{
				"VIS-6457 [13.2AV]",
				"Cherry-pick of #1063",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCommitMessage(tt.ticket, tt.version, tt.newBranch, tt.entry, tt.prNumber, tt.targetBranch)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildCommitMessage() missing %q\n\ngot:\n%s", want, got)
				}
			}
		})
	}
}

func TestBuildPRBody(t *testing.T) {
	entry := &commits.CommitEntry{
		Ticket: "VIS-7077",
		Logs:   []string{"Added email tracking.", "Log SMTP exceptions."},
	}

	tests := []struct {
		name         string
		newBranch    string
		entry        *commits.CommitEntry
		prNumber     int
		targetBranch string
		method       string
		wantContains []string
	}{
		{
			name:         "with entry and logs",
			newBranch:    "VIS-7077-fix-13.2av",
			entry:        entry,
			prNumber:     1039,
			targetBranch: "maintenance/13.2AV",
			method:       "cherry-pick",
			wantContains: []string{
				"VIS-7077-fix-13.2av",
				"- Added email tracking.",
				"- Log SMTP exceptions.",
				"- Cherry-pick of #1039 onto maintenance/13.2AV",
			},
		},
		{
			name:         "nil entry falls back to method",
			newBranch:    "VIS-1234-fix",
			entry:        nil,
			prNumber:     100,
			targetBranch: "maintenance/13.1AV",
			method:       "patch",
			wantContains: []string{
				"Cherry-pick of #100 onto maintenance/13.1AV",
				"Method: patch",
			},
		},
		{
			name:         "entry with no logs falls back",
			newBranch:    "VIS-1234-fix",
			entry:        &commits.CommitEntry{Ticket: "VIS-1234"},
			prNumber:     200,
			targetBranch: "maintenance/13.2AV",
			method:       "cherry-pick",
			wantContains: []string{
				"Cherry-pick of #200 onto maintenance/13.2AV",
				"Method: cherry-pick",
			},
		},
		{
			name:         "manual method shown when conflicts resolved interactively",
			newBranch:    "VIS-9999-fix-13.2av",
			entry:        nil,
			prNumber:     1100,
			targetBranch: "maintenance/13.2AV",
			method:       "manual",
			wantContains: []string{
				"Cherry-pick of #1100 onto maintenance/13.2AV",
				"Method: manual",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPRBody(tt.newBranch, tt.entry, tt.prNumber, tt.targetBranch, tt.method)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildPRBody() missing %q\n\ngot:\n%s", want, got)
				}
			}
		})
	}
}

func TestParseBranches(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single branch", "maintenance/13.1AV", []string{"maintenance/13.1AV"}},
		{"multiple branches", "maintenance/13.1AV,maintenance/12.1AV", []string{"maintenance/13.1AV", "maintenance/12.1AV"}},
		{"with whitespace", " maintenance/13.1AV , maintenance/12.1AV ", []string{"maintenance/13.1AV", "maintenance/12.1AV"}},
		{"trailing comma", "maintenance/13.1AV,", []string{"maintenance/13.1AV"}},
		{"leading comma", ",maintenance/13.1AV", []string{"maintenance/13.1AV"}},
		{"empty string", "", nil},
		{"only commas", ",,,", nil},
		{"spaces only", "  ,  ,  ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBranches(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseBranches(%q) = %v (len %d), want %v (len %d)", tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseBranches(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"tilde expansion", "~/Documents/Commits.md", filepath.Join(home, "Documents", "Commits.md")},
		{"absolute path unchanged", "/tmp/Commits.md", "/tmp/Commits.md"},
		{"relative path cleaned", "foo/../bar/file.md", "bar/file.md"},
		{"tilde traversal cleaned", "~/../../etc/passwd", filepath.Clean(filepath.Join(home, "../../etc/passwd"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.path)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

