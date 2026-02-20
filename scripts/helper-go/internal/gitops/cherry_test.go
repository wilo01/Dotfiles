package gitops

import (
	"encoding/json"
	"testing"
)

func TestPRInfoParsing(t *testing.T) {
	raw := `{
		"number": 1036,
		"title": "VIS-6893 Kiosk manual pin login screen stuck on iPad",
		"headRefName": "vis-6893---kiosk-manual-pin-login-screen-stuck-on-ipad-13.2av",
		"baseRefName": "maintenance/13.2AV",
		"state": "MERGED",
		"url": "https://github.com/org/repo/pull/1036"
	}`

	var info PRInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if info.Number != 1036 {
		t.Errorf("Number = %d, want 1036", info.Number)
	}
	if info.HeadBranch != "vis-6893---kiosk-manual-pin-login-screen-stuck-on-ipad-13.2av" {
		t.Errorf("HeadBranch = %q", info.HeadBranch)
	}
	if info.BaseBranch != "maintenance/13.2AV" {
		t.Errorf("BaseBranch = %q", info.BaseBranch)
	}
	if info.State != "MERGED" {
		t.Errorf("State = %q, want MERGED", info.State)
	}
	if info.URL != "https://github.com/org/repo/pull/1036" {
		t.Errorf("URL = %q", info.URL)
	}
}

func TestMergeCommitParsing(t *testing.T) {
	raw := `{"mergeCommit":{"oid":"abc1234def5678"}}`

	var result struct {
		MergeCommit struct {
			OID string `json:"oid"`
		} `json:"mergeCommit"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if result.MergeCommit.OID != "abc1234def5678" {
		t.Errorf("OID = %q, want abc1234def5678", result.MergeCommit.OID)
	}
}

func TestMergeCommitEmpty(t *testing.T) {
	raw := `{"mergeCommit":null}`

	var result struct {
		MergeCommit *struct {
			OID string `json:"oid"`
		} `json:"mergeCommit"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if result.MergeCommit != nil {
		t.Errorf("expected nil mergeCommit for unmerged PR")
	}
}
