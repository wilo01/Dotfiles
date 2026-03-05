package commits

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleCommitsMd = `# 2026-02-19 15:55
## master
VIS-6457

VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default

Logs:
- Tyiping reference code is uppercase now.
- Added option to set uppercase keyboard when needed.

Changed files: source/ui-kiosk/app/global/BadgeHandler.js\ source/ui-kiosk/app/plugins/Keyboard.js\
- Commit branch HASH [deaf92c98f09](VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default)

# 2026-02-20 08:16
## master
VIS-7077 (#1039)

VIS-7077-swift-chubb-emails-not-sending-consistently

Logs:
- Added v_send_success to track email sending status.
- Log SMTP exceptions with recipient address.
- Update queue only on successful email send.

Changed files: source/server/database/sql/safe/packages/ca_email_pak.sql\
- Commit branch HASH [c24463f438](VIS-7077-swift-chubb-emails-not-sending-consistently)

# 2026-02-24 08:24
## master
Merge pull request #1063 from acreidentity/VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default

VIS-6457

Changed files: source/ui-kiosk/app/global/BadgeHandler.js\
- Commit branch HASH [1a498e16](VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default-13.2av)
`

func TestLookupByTicket_Found(t *testing.T) {
	tmp := writeTemp(t, sampleCommitsMd)

	entry, err := LookupByTicket(tmp, "VIS-6457")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.Ticket != "VIS-6457" {
		t.Errorf("ticket = %q, want VIS-6457", entry.Ticket)
	}
	if entry.BranchName != "VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default" {
		t.Errorf("branch = %q", entry.BranchName)
	}
	if len(entry.Logs) != 2 {
		t.Fatalf("logs count = %d, want 2: %v", len(entry.Logs), entry.Logs)
	}
	if entry.Logs[0] != "Tyiping reference code is uppercase now." {
		t.Errorf("log[0] = %q", entry.Logs[0])
	}
}

func TestLookupByTicket_FindsMostRecent(t *testing.T) {
	tmp := writeTemp(t, sampleCommitsMd)

	entry, err := LookupByTicket(tmp, "VIS-7077")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.Ticket != "VIS-7077" {
		t.Errorf("ticket = %q, want VIS-7077", entry.Ticket)
	}
	if len(entry.Logs) != 3 {
		t.Fatalf("logs count = %d, want 3", len(entry.Logs))
	}
}

func TestLookupByTicket_NotFound(t *testing.T) {
	tmp := writeTemp(t, sampleCommitsMd)

	entry, err := LookupByTicket(tmp, "VIS-9999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry != nil {
		t.Errorf("expected nil, got %+v", entry)
	}
}

func TestLookupByTicket_FileNotFound(t *testing.T) {
	_, err := LookupByTicket("/nonexistent/file.md", "VIS-1234")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLookupByTicket_CaseInsensitive(t *testing.T) {
	tmp := writeTemp(t, sampleCommitsMd)

	entry, err := LookupByTicket(tmp, "vis-6457")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
}

func TestLookupByTicket_SkipsNonMaster(t *testing.T) {
	content := `# 2026-02-19 15:55
## maintenance/13.2AV
VIS-1234

VIS-1234-some-branch

Logs:
- This is on maintenance, not master.

Changed files: some/file.js\
- Commit branch HASH [abc123](VIS-1234-some-branch)
`
	tmp := writeTemp(t, content)

	entry, err := LookupByTicket(tmp, "VIS-1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry != nil {
		t.Errorf("expected nil for non-master entry, got %+v", entry)
	}
}

func TestLookupByTicket_PrefersLatestMasterEntry(t *testing.T) {
	content := `# 2026-02-18 10:00
## master
VIS-5000

VIS-5000-old-branch

Logs:
- Old log line.

Changed files: old/file.js\
- Commit branch HASH [aaa111](VIS-5000-old-branch)

# 2026-02-19 10:00
## master
VIS-5000

VIS-5000-new-branch

Logs:
- New log line.
- Second new log.

Changed files: new/file.js\
- Commit branch HASH [bbb222](VIS-5000-new-branch)
`
	tmp := writeTemp(t, content)

	entry, err := LookupByTicket(tmp, "VIS-5000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.BranchName != "VIS-5000-new-branch" {
		t.Errorf("branch = %q, want VIS-5000-new-branch", entry.BranchName)
	}
	if len(entry.Logs) != 2 {
		t.Fatalf("logs count = %d, want 2", len(entry.Logs))
	}
	if entry.Logs[0] != "New log line." {
		t.Errorf("log[0] = %q", entry.Logs[0])
	}
}

func TestAppendCherryEntry(t *testing.T) {
	tmp := writeTemp(t, "# existing content\n")

	err := AppendCherryEntry(tmp, CherryEntry{
		TargetBranch: "maintenance/13.2AV",
		Ticket:       "VIS-6457",
		Version:      "13.2AV",
		BranchName:   "VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default-13.2av",
		Logs:         []string{"Tyiping reference code is uppercase now.", "Added option to set uppercase keyboard when needed."},
		CherryNote:   "Cherry-pick of #1063 onto maintenance/13.2AV",
		ChangedFiles: []string{"source/ui-kiosk/app/global/BadgeHandler.js", "source/ui-kiosk/app/plugins/Keyboard.js"},
		CommitSHA:    "abc123def456",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tmp)
	content := string(data)

	assertContains(t, content, "## maintenance/13.2AV")
	assertContains(t, content, "VIS-6457 [13.2AV]")
	assertContains(t, content, "VIS-6457-alphanumeric-booking-reference-all-caps-keyboard-default-13.2av")
	assertContains(t, content, "- Tyiping reference code is uppercase now.")
	assertContains(t, content, "- Cherry-pick of #1063 onto maintenance/13.2AV")
	assertContains(t, content, `Changed files: source/ui-kiosk/app/global/BadgeHandler.js\ source/ui-kiosk/app/plugins/Keyboard.js\ `)
	assertContains(t, content, "[abc123def456]")
}

func TestAppendCherryEntry_NoLogs(t *testing.T) {
	tmp := writeTemp(t, "")

	err := AppendCherryEntry(tmp, CherryEntry{
		TargetBranch: "maintenance/13.1AV",
		Ticket:       "VIS-9999",
		Version:      "13.1AV",
		BranchName:   "VIS-9999-some-fix-13.1av",
		CherryNote:   "Cherry-pick of #100 onto maintenance/13.1AV",
		CommitSHA:    "deadbeef",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tmp)
	content := string(data)

	assertContains(t, content, "VIS-9999 [13.1AV]")
	assertContains(t, content, "- Cherry-pick of #100 onto maintenance/13.1AV")
}

func TestAppendCherryEntry_NoVersion(t *testing.T) {
	tmp := writeTemp(t, "")

	err := AppendCherryEntry(tmp, CherryEntry{
		TargetBranch: "maintenance/13.2AV",
		Ticket:       "VIS-1234",
		Version:      "",
		BranchName:   "VIS-1234-some-fix",
		CherryNote:   "Cherry-pick of #42 onto maintenance/13.2AV",
		CommitSHA:    "cafe1234",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tmp)
	content := string(data)

	// Without version, ticket line should NOT have brackets
	assertContains(t, content, "VIS-1234\n")
	if strings.Contains(content, "VIS-1234 [") {
		t.Errorf("expected no version bracket, got:\n%s", content)
	}
}

func TestAppendCherryEntry_PreservesExisting(t *testing.T) {
	existing := "# 2026-02-19 15:55\n## master\nVIS-6457\nExisting content\n"
	tmp := writeTemp(t, existing)

	err := AppendCherryEntry(tmp, CherryEntry{
		TargetBranch: "maintenance/13.2AV",
		Ticket:       "VIS-6457",
		Version:      "13.2AV",
		BranchName:   "VIS-6457-fix-13.2av",
		CherryNote:   "Cherry-pick of #1063 onto maintenance/13.2AV",
		CommitSHA:    "abc123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tmp)
	content := string(data)

	// Existing content must still be present
	assertContains(t, content, "# 2026-02-19 15:55")
	assertContains(t, content, "Existing content")
	// New entry appended
	assertContains(t, content, "## maintenance/13.2AV")
	assertContains(t, content, "VIS-6457 [13.2AV]")
}

func TestGetLogs_NilReceiver(t *testing.T) {
	var entry *CommitEntry
	logs := entry.GetLogs()
	if logs != nil {
		t.Errorf("expected nil, got %v", logs)
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "Commits.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected content to contain %q\n\ngot:\n%s", substr, s)
	}
}
