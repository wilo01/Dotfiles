package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	internalJira "github.com/dariuszw/hlp/internal/jira"
)

const triageTimeout = 3 * time.Minute

// TriageVerdict is the strict-JSON answer expected from the headless triage call
type TriageVerdict struct {
	Repo       string `json:"repo"`
	Confidence string `json:"confidence"`
	Reason     string `json:"reason"`
}

// repoHints gives the triage model a one-line purpose per known repo.
// Repos without a hint are still offered as candidates.
var repoHints = map[string]string{
	"tds-suite":                  "core Suite app: Oracle PL/SQL packages, Liquibase changesets, ExtJS admin UI, .apex resource templates, kiosk/muster UI",
	"tds-hexer":                  "Node Express 5 API gateway replacing Apex Listener: routing, auth strategies, Memcached caching, serves .apex routes",
	"tds-suite-api":              "serverless Suite REST API (Node Lambda, 281 endpoints): visitors, persons, badges, reference data",
	"tds-api-server":             "MyVisitor REST API (Node Express) serving the visitor web app and kiosk",
	"tds-visitor-web-app":        "visitor self-service portal: pre-booking, check-in UI",
	"tds-kiosk-chrome-app":       "on-site kiosk ChromeOS app: walk-in registration, check-in/out flows",
	"tds-kiosk-chrome-extension": "kiosk Chrome extension companion",
	"tds-kiosk-iwa":              "kiosk isolated web app variant",
	"tds-mcp-server":             "TypeScript MCP server exposing TDS tools to AI agents",
	"tds-access-integrations":    "Java bridges to external access-control systems (Feenics, CCure, S2, Lenel, Genetec)",
	"tds-automation":             "deployment automation: Jenkins pipelines, CodeDeploy specs, deploy shell scripts, Ansible",
	"tds-linx":                   "Linx routing/metadata service (Node TS, MongoDB) driving Hexer ALB rules",
	"tds-hop-psv-interface":      "Dynamics <-> TDS sync interface",
	"tds-hr-interface":           "HR data interface",
	"tds-acre-qrcode-generator":  "QR code generation service (Node Express)",
	"tds-pvm-daa-v2":             "PVM DAA v2 service",
	"tds-cpp":                    "C++ components",
}

// buildTriagePrompt assembles the headless prompt: ticket fields + candidate repos
func buildTriagePrompt(ticket *internalJira.Ticket, description string, repos []string) string {
	var b strings.Builder
	b.WriteString("You are a repo-triage classifier for the TDS/acre Identity ecosystem.\n")
	b.WriteString("Given a Jira ticket, decide which single repository the implementation work belongs in.\n\n")
	b.WriteString("Candidate repositories:\n")
	for _, repo := range repos {
		if hint, ok := repoHints[repo]; ok {
			fmt.Fprintf(&b, "- %s: %s\n", repo, hint)
		} else {
			fmt.Fprintf(&b, "- %s\n", repo)
		}
	}
	b.WriteString("\nTicket:\n")
	fmt.Fprintf(&b, "Key: %s\n", ticket.Key)
	fmt.Fprintf(&b, "Type: %s\n", ticket.IssueType)
	fmt.Fprintf(&b, "Summary: %s\n", ticket.Summary)
	if description != "" {
		desc := description
		if len(desc) > 4000 {
			desc = desc[:4000] + "\n[truncated]"
		}
		fmt.Fprintf(&b, "Description:\n%s\n", desc)
	}
	b.WriteString("\nIf the ticket spans multiple repos, pick the PRIMARY repo where most work happens.\n")
	b.WriteString("Respond with ONLY a JSON object, no markdown fences, no prose:\n")
	b.WriteString(`{"repo": "<one of the candidates>", "confidence": "high|medium|low", "reason": "<one sentence>"}` + "\n")
	return b.String()
}

// triageFunc classifies a ticket into one of repos. Callers inject it so the
// bulk resolution path can be tested without shelling out to a model. Non-fatal
// problems go to warn rather than stdout, so concurrent callers can buffer them
// instead of interleaving output.
type triageFunc func(cfg *config.Config, client *internalJira.Client, ticket *internalJira.Ticket, repos []string, warn func(string)) (*TriageVerdict, error)

// triageRepo fetches the ticket description and classifies it. A missing
// description is not fatal - summary and issue type alone are usually enough to
// pick a repo - so it warns and classifies anyway.
func triageRepo(cfg *config.Config, client *internalJira.Client, ticket *internalJira.Ticket, repos []string, warn func(string)) (*TriageVerdict, error) {
	description, err := client.GetTicketDescription(ticket.Key)
	if err != nil && warn != nil {
		warn("could not fetch description for triage: " + err.Error())
	}
	return runTriage(cfg, ticket, description, repos)
}

// runTriage executes the configured headless triage command with the prompt on
// stdin and parses the JSON verdict.
func runTriage(cfg *config.Config, ticket *internalJira.Ticket, description string, repos []string) (*TriageVerdict, error) {
	triageCmd := cfg.Agent.TriageCmd
	if triageCmd == "" {
		return nil, fmt.Errorf("agent.triage_cmd not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), triageTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", triageCmd)
	cmd.Stdin = strings.NewReader(buildTriagePrompt(ticket, description, repos))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("triage command failed: %w", err)
	}

	verdict, err := parseTriageVerdict(string(out))
	if err != nil {
		return nil, err
	}

	if !slices.Contains(repos, verdict.Repo) {
		return nil, fmt.Errorf("triage picked %q which is not a candidate repo", verdict.Repo)
	}
	return verdict, nil
}

// parseTriageVerdict extracts the first JSON object from model output,
// tolerating markdown fences or surrounding prose.
func parseTriageVerdict(output string) (*TriageVerdict, error) {
	start := strings.Index(output, "{")
	end := strings.LastIndex(output, "}")
	if start == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object in triage output: %.200s", output)
	}

	var verdict TriageVerdict
	if err := json.Unmarshal([]byte(output[start:end+1]), &verdict); err != nil {
		return nil, fmt.Errorf("invalid triage JSON: %w", err)
	}
	if verdict.Repo == "" {
		return nil, fmt.Errorf("triage JSON missing repo field")
	}
	if verdict.Confidence == "" {
		verdict.Confidence = "low"
	}
	return &verdict, nil
}
