package jira

import "testing"

func TestIsCurrentUserWorklog(t *testing.T) {
	tests := []struct {
		name     string
		worklog  Worklog
		user     *CurrentUser
		expected bool
	}{
		// Nil user guard (prevents panic)
		{
			name:     "nil user returns false",
			worklog:  Worklog{Author: "John Doe"},
			user:     nil,
			expected: false,
		},

		// Cloud: AccountId comparison (preferred method)
		{
			name:     "cloud match by accountId",
			worklog:  Worklog{AuthorAccountId: "abc123", Author: "John Doe"},
			user:     &CurrentUser{AccountID: "abc123", DisplayName: "John Doe"},
			expected: true,
		},
		{
			name:     "cloud no match accountId",
			worklog:  Worklog{AuthorAccountId: "abc123", Author: "John Doe"},
			user:     &CurrentUser{AccountID: "xyz789", DisplayName: "John Doe"},
			expected: false,
		},
		{
			name:     "cloud accountId match ignores displayName mismatch",
			worklog:  Worklog{AuthorAccountId: "abc123", Author: "Johnny D"},
			user:     &CurrentUser{AccountID: "abc123", DisplayName: "John Doe"},
			expected: true,
		},

		// Server/DC: DisplayName fallback (case-insensitive)
		{
			name:     "server match displayName exact",
			worklog:  Worklog{Author: "John Doe"},
			user:     &CurrentUser{DisplayName: "John Doe"},
			expected: true,
		},
		{
			name:     "server match displayName case insensitive",
			worklog:  Worklog{Author: "John Doe"},
			user:     &CurrentUser{DisplayName: "john doe"},
			expected: true,
		},
		{
			name:     "server match displayName uppercase",
			worklog:  Worklog{Author: "john doe"},
			user:     &CurrentUser{DisplayName: "JOHN DOE"},
			expected: true,
		},
		{
			name:     "server no match displayName",
			worklog:  Worklog{Author: "John Doe"},
			user:     &CurrentUser{DisplayName: "Jane Smith"},
			expected: false,
		},

		// Fallback when accountId is empty on both sides
		{
			name:     "fallback to displayName when accountId empty",
			worklog:  Worklog{AuthorAccountId: "", Author: "John Doe"},
			user:     &CurrentUser{AccountID: "", DisplayName: "John Doe"},
			expected: true,
		},
		{
			name:     "fallback to displayName when only worklog accountId empty",
			worklog:  Worklog{AuthorAccountId: "", Author: "John Doe"},
			user:     &CurrentUser{AccountID: "abc123", DisplayName: "John Doe"},
			expected: true,
		},
		{
			name:     "fallback to displayName when only user accountId empty",
			worklog:  Worklog{AuthorAccountId: "abc123", Author: "John Doe"},
			user:     &CurrentUser{AccountID: "", DisplayName: "John Doe"},
			expected: true,
		},

		// Edge cases
		{
			name:     "empty author and displayName",
			worklog:  Worklog{Author: ""},
			user:     &CurrentUser{DisplayName: ""},
			expected: true, // Empty strings are equal
		},
		{
			name:     "empty author vs non-empty displayName",
			worklog:  Worklog{Author: ""},
			user:     &CurrentUser{DisplayName: "John Doe"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCurrentUserWorklog(tt.worklog, tt.user)
			if result != tt.expected {
				t.Errorf("isCurrentUserWorklog(%+v, %+v) = %v, want %v",
					tt.worklog, tt.user, result, tt.expected)
			}
		})
	}
}

func TestIsCloudInstance(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://company.atlassian.net", true},
		{"https://my-project.atlassian.net/jira", true},
		{"https://jira.company.com", false},
		{"https://localhost:8080", false},
		{"http://10.0.0.1:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isCloudInstance(tt.url)
			if result != tt.expected {
				t.Errorf("isCloudInstance(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestStripOrderBy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes ORDER BY",
			input:    "project = TEST ORDER BY updated DESC",
			expected: "project = TEST",
		},
		{
			name:     "removes order by lowercase",
			input:    "assignee = currentUser() order by created",
			expected: "assignee = currentUser()",
		},
		{
			name:     "no ORDER BY clause",
			input:    "project = TEST AND status = Open",
			expected: "project = TEST AND status = Open",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripOrderBy(tt.input)
			if result != tt.expected {
				t.Errorf("stripOrderBy(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractProjectKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VIS-1234", "VIS"},
		{"tdt-99", "TDT"},
		{"PROJ-1", "PROJ"},
		{"ABC", "ABC"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExtractProjectKey(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractProjectKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
