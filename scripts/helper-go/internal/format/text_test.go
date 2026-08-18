package format

import "testing"

func TestJiraBranch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VIS-1234 Add new feature", "git checkout -b VIS-1234-add-new-feature"},
		{"vis-5678 fix login bug", "git checkout -b VIS-5678-fix-login-bug"},
		{"TASK-99 allowed 13 things", "git checkout -b TASK-99-allowed-13-things"},
		{"TDT-123 Update version-2 config", "git checkout -b TDT-123-update-version-2-config"},
		{"  VIS-100  spaces everywhere  ", "git checkout -b VIS-100-spaces-everywhere"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := JiraBranch(tt.input)
			if result != tt.expected {
				t.Errorf("JiraBranch(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VIS-1234 work in progress", "git stash push -u -m VIS-1234-work-in-progress"},
		{"save my changes", "git stash push -u -m save-my-changes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Stash(tt.input)
			if result != tt.expected {
				t.Errorf("Stash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"VIS-1234 My Feature", "vis-1234-my-feature"},
		{"  Multiple   Spaces  ", "multiple-spaces"},
		{"Special@#$Characters", "special-characters"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Dash(tt.input)
			if result != tt.expected {
				t.Errorf("Dash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPRTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"vis-1234 add new feature", "VIS-1234 add new feature"},
		{"TDT-99 Fix bug in tdt-100 module", "TDT-99 Fix bug in TDT-100 module"},
		{"no ticket here", "no ticket here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := PRTitle(tt.input)
			if result != tt.expected {
				t.Errorf("PRTitle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VIS-1234 My Report", "vis-1234-my-report.md"},
		{"Meeting Notes", "meeting-notes.md"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Filename(tt.input)
			if result != tt.expected {
				t.Errorf("Filename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractTicket(t *testing.T) {
	tests := []struct {
		input          string
		expectedTicket string
		expectedDesc   string
	}{
		{"VIS-1234 add feature", "VIS-1234", "add feature"},
		{"vis-99-fix-bug", "VIS-99", "fix-bug"},
		{"no ticket", "", "no ticket"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ticket, desc := ExtractTicket(tt.input)
			if ticket != tt.expectedTicket {
				t.Errorf("ExtractTicket(%q) ticket = %q, want %q", tt.input, ticket, tt.expectedTicket)
			}
			if desc != tt.expectedDesc {
				t.Errorf("ExtractTicket(%q) desc = %q, want %q", tt.input, desc, tt.expectedDesc)
			}
		})
	}
}
