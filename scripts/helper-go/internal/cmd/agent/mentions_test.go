package agent

import (
	"testing"

	internalJira "github.com/dariuszw/hlp/internal/jira"
)

func allRepos() []SourceRepo {
	names := []string{"tds-suite", "tds-hexer", "tds-kiosk-chrome-app", "tds-api-server", "tds-student", "smtp-node"}
	repos := make([]SourceRepo, 0, len(names))
	for _, n := range names {
		repos = append(repos, SourceRepo{Name: n, Path: "/branches/" + n})
	}
	return repos
}

func mentionsFor(summary, description string) map[string]bool {
	tk := &internalJira.Ticket{Key: "SUITE-9250", Summary: summary}
	return mentionedRepos(tk, description, allRepos())
}

func TestMentionsFullRepoName(t *testing.T) {
	got := mentionsFor("fix thing", "the change is in tds-kiosk-chrome-app and nowhere else")

	if !got["tds-kiosk-chrome-app"] {
		t.Error("full repo name not matched")
	}
	if got["tds-hexer"] {
		t.Error("unrelated repo matched")
	}
}

// Prose rarely writes the tds- prefix, so the suffix counts too.
func TestMentionsBarePrefixlessAlias(t *testing.T) {
	got := mentionsFor("routing broken", "the hexer layer drops the header before the kiosk sees it")

	if !got["tds-hexer"] {
		t.Error("bare 'hexer' not matched")
	}
	if !got["tds-kiosk-chrome-app"] {
		t.Error("'kiosk' should match tds-kiosk-chrome-app")
	}
}

// The whole reason for stripping the key: every SUITE- ticket would otherwise
// mark tds-suite.
func TestTicketKeyDoesNotCountAsAMention(t *testing.T) {
	got := mentionsFor("something unrelated", "see SUITE-9250 for context")

	if got["tds-suite"] {
		t.Error("the ticket key marked tds-suite")
	}
}

// "suite" is the product name and appears on nearly every ticket, so it is a
// stop word: only the full repo name marks tds-suite.
func TestSuiteAliasIsAStopWord(t *testing.T) {
	if mentionsFor("Photos not showing in TDS suite", "")["tds-suite"] {
		t.Error("the word 'suite' alone marked tds-suite")
	}
	if !mentionsFor("fix it", "the bug is in tds-suite source/ui")["tds-suite"] {
		t.Error("the full repo name should still match")
	}
}

func TestMentionsAreWordBounded(t *testing.T) {
	if mentionsFor("rapid prototyping", "")["tds-api-server"] {
		t.Error("'api' matched inside 'rapid'")
	}
}

func TestMentionsAreCaseInsensitive(t *testing.T) {
	if !mentionsFor("HEXER routing", "")["tds-hexer"] {
		t.Error("uppercase mention missed")
	}
}

func TestEmptyDescriptionStillScansTheSummary(t *testing.T) {
	got := mentionsFor("TDS Kiosk App issues: BNZ UAT", "")

	if !got["tds-kiosk-chrome-app"] {
		t.Error("summary-only mention missed")
	}
}

func TestNoMentionsIsEmptyNotNil(t *testing.T) {
	got := mentionsFor("nothing relevant here", "still nothing")

	for name, marked := range got {
		if marked {
			t.Errorf("%s marked with no mention", name)
		}
	}
}
