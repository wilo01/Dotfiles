package format

import "testing"

func TestDeriveCherryBranch(t *testing.T) {
	tests := []struct {
		name         string
		sourceBranch string
		targetBranch string
		expected     string
		expectErr    bool
	}{
		{
			name:         "replace version suffix",
			sourceBranch: "vis-6893---kiosk-manual-pin-login-screen-stuck-on-ipad-13.2av",
			targetBranch: "maintenance/13.1AV",
			expected:     "vis-6893---kiosk-manual-pin-login-screen-stuck-on-ipad-13.1av",
		},
		{
			name:         "append when no version suffix",
			sourceBranch: "VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default",
			targetBranch: "maintenance/13.1AV",
			expected:     "VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default-13.1av",
		},
		{
			name:         "major version only replace",
			sourceBranch: "vis-1234-fix-bug-13av",
			targetBranch: "maintenance/12AV",
			expected:     "vis-1234-fix-bug-12av",
		},
		{
			name:         "major version only append",
			sourceBranch: "vis-1234-fix-bug",
			targetBranch: "maintenance/12AV",
			expected:     "vis-1234-fix-bug-12av",
		},
		{
			name:         "uppercase source version gets lowercased",
			sourceBranch: "VIS-100-feature-13.2AV",
			targetBranch: "maintenance/12.1AV",
			expected:     "VIS-100-feature-12.1av",
		},
		{
			name:         "target version always lowercased in output",
			sourceBranch: "vis-999-thing",
			targetBranch: "maintenance/13.0AV",
			expected:     "vis-999-thing-13.0av",
		},
		{
			name:         "invalid target format",
			sourceBranch: "vis-1234-fix-13.2av",
			targetBranch: "develop",
			expectErr:    true,
		},
		{
			name:         "invalid target no version",
			sourceBranch: "vis-1234-fix",
			targetBranch: "release/2.0",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DeriveCherryBranch(tt.sourceBranch, tt.targetBranch)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got result %q", result)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("DeriveCherryBranch(%q, %q) = %q, want %q",
					tt.sourceBranch, tt.targetBranch, result, tt.expected)
			}
		})
	}
}

func TestExtractMaintenanceVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"maintenance/13.1AV", "13.1AV"},
		{"maintenance/12AV", "12AV"},
		{"maintenance/13.0av", "13.0av"},
		{"develop", ""},
		{"release/2.0", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExtractMaintenanceVersion(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractMaintenanceVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
