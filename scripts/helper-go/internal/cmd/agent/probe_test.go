package agent

import (
	"os"
	"testing"
)

// TestProbeRealRepo is a manual check against a real clone, skipped unless
// HLP_PROBE_REPO points at one. It reads only.
func TestProbeRealRepo(t *testing.T) {
	source := os.Getenv("HLP_PROBE_REPO")
	if source == "" {
		t.Skip("set HLP_PROBE_REPO to a real clone to run this")
	}

	def, err := resolveDefaultBase(source)
	if err != nil {
		t.Fatalf("resolveDefaultBase: %v", err)
	}
	branches := listBaseBranches(source, def, []string{"maintenance/*", "release/*"})
	t.Logf("default=%s, %d bases offered:", def, len(branches))
	for _, b := range branches {
		t.Logf("  %s", b)
	}
}
