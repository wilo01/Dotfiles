package standup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/batch"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	publishDays    int
	publishDryRun  bool
	publishWebhook string
	publishNoMerge bool
)

func init() {
	publishCmd.Flags().IntVar(&publishDays, "days", 0, "Limit to last N days (0=smart default: Mon=3, else=1)")
	publishCmd.Flags().BoolVar(&publishDryRun, "dry-run", false, "Print payload without posting")
	publishCmd.Flags().StringVar(&publishWebhook, "webhook", "", "Override webhook URL (defaults to HLP_STANDUP_WEBHOOK_URL env)")
	publishCmd.Flags().BoolVar(&publishNoMerge, "no-merge", false, "Disable merging duplicate JIRA tickets into one group")
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish standup blob to n8n webhook -> Google Sheet",
	Long: `Send today's standup entries (formatted same as 'hlp standup show') to the
configured n8n webhook, which appends them to the Daily TDS Google Sheet.

The webhook URL is taken from --webhook flag or HLP_STANDUP_WEBHOOK_URL env var.

Examples:
  hlp standup publish                       # Smart-default range (Mon=3 days, else=1)
  hlp standup publish --days 3              # Last 3 days (Friday + weekend + today)
  hlp standup publish --dry-run             # See what would be sent`,
	RunE: runPublish,
}

func runPublish(cmd *cobra.Command, args []string) error {
	if err := doPublish(); err != nil {
		fmt.Fprintf(os.Stderr, "publish failed: %v\n", err)
		return err
	}
	return nil
}

func doPublish() error {
	cfg, _ := config.Load()

	webhook := publishWebhook
	if webhook == "" {
		webhook = os.Getenv("HLP_STANDUP_WEBHOOK_URL")
	}
	if webhook == "" {
		webhook = cfg.Preferences.StandupWebhookURL
	}
	if webhook == "" && !publishDryRun {
		return fmt.Errorf("no webhook URL: set --webhook, HLP_STANDUP_WEBHOOK_URL, or preferences.standup_webhook_url in ~/.config/hlp/config.yaml")
	}

	workflowEditor := cfg.Preferences.StandupWorkflowURL

	sheetURL := cfg.Sheets.DailyTabURL
	if sheetURL == "" && !publishDryRun {
		return fmt.Errorf("no sheet URL: set google_sheets.daily_tab_url in ~/.config/hlp/config.yaml")
	}

	profile, _ := config.GetActiveProfile()
	csvPath := batch.DefaultCSVPathForProfile(profile)
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return fmt.Errorf("worklog CSV not found: %s", csvPath)
	}

	entries, err := batch.ParseCSV(csvPath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no entries in worklog CSV")
	}

	days := smartDefaultDays(publishDays)

	var baseURL string
	if profile != nil {
		baseURL = profile.BaseURL
	}

	merge := !publishNoMerge

	if publishDryRun {
		webhookDisplay := webhook
		if webhookDisplay == "" {
			webhookDisplay = "<unset>"
		}
		sheetDisplay := sheetURL
		if sheetDisplay == "" {
			sheetDisplay = "<unset>"
		}
		workflowDisplay := workflowEditor
		if workflowDisplay == "" {
			workflowDisplay = "<unset>"
		}

		// Count entries to show in header
		entriesCopy := make([]batch.Entry, len(entries))
		copy(entriesCopy, entries)
		sortEntriesNewestFirst(entriesCopy)
		filtered := filterStandupEntries(applyDayWindow(entriesCopy, days), getIgnoredTickets())
		count := len(filtered)
		if count == 0 {
			return fmt.Errorf("no entries in last %d day(s)", days)
		}

		fmt.Printf("%s\n%s\n%s\n%s\n",
			ui.Title.Render("--- DRY RUN ---"),
			ui.KeyValue("webhook", webhookDisplay),
			ui.KeyValue("sheet", sheetDisplay),
			ui.KeyValue("workflow", workflowDisplay))
		if workflowEditor == "" {
			fmt.Println(ui.Warning("workflow URL not configured - add preferences.standup_workflow_url to config to see editor link on errors"))
		}
		fmt.Println(ui.Muted.Render("───"))

		// Render entries with same colors as hlp standup show
		if err := renderPlainEntries(entries, days, baseURL, merge); err != nil {
			return err
		}

		fmt.Printf("\n%s\n%s\n",
			ui.KeyValuePadded("entries", fmt.Sprintf("%d", count), 5),
			ui.KeyValuePadded("days", fmt.Sprintf("%d", days), 5))
		// Test webhook and show response
		if webhook != "" {
			testPayload := map[string]any{
				"publish_date": time.Now().Format("2006-01-02"),
				"blob":         "[Dry-run test - no data pushed]",
				"entry_count":  count,
				"group_count":  0,
				"days":         days,
			}
			testBody, _ := json.Marshal(testPayload)
			req, _ := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(testBody))
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(ui.Warning("webhook unreachable: " + err.Error()))
			} else {
				defer resp.Body.Close()
				respBody, _ := io.ReadAll(resp.Body)
				if resp.StatusCode >= 300 {
					fmt.Println(ui.Warning(fmt.Sprintf("webhook would FAIL (HTTP %d):", resp.StatusCode)))
					fmt.Println(colorizeJSONResponse(respBody))
				} else {
					fmt.Println(ui.Success.Render("webhook responded OK"))
				}
			}
		}

		return nil
	}

	blob, count, groupCount := buildStandupBlob(entries, days, baseURL, merge, entryFormatOpts{})
	if count == 0 {
		return fmt.Errorf("no entries in last %d day(s)", days)
	}

	payload := map[string]any{
		"publish_date": time.Now().Format("2006-01-02"),
		"blob":         blob,
		"entry_count":  count,
		"group_count":  groupCount,
		"days":         days,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post webhook: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		var prettyErr string
		var pp json.RawMessage
		if json.Unmarshal(respBody, &pp) == nil {
			indented, _ := json.MarshalIndent(pp, "", "  ")
			prettyErr = colorizeJSON(string(indented))
		} else {
			prettyErr = string(respBody)
		}
		errMsg := fmt.Sprintf("webhook returned %d:\n%s", resp.StatusCode, prettyErr)
		if workflowEditor != "" {
			errMsg += fmt.Sprintf("\n  n8n editor: %s", workflowEditor)
		}
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			errMsg += "\n  n8n auth may have expired - open the n8n URL and re-login"
		}
		return fmt.Errorf("%s", errMsg)
	}

	printPublishResult(count, days, webhook, sheetURL, respBody)
	return nil
}

func printPublishResult(count, days int, webhook, sheetURL string, respBody []byte) {
	var resp struct {
		OK            bool   `json:"ok"`
		PublishDate   string `json:"publish_date"`
		BlobChars     int    `json:"blob_chars"`
		InsertedAtRow string `json:"inserted_at_row"`
	}
	_ = json.Unmarshal(respBody, &resp)

	icon := ui.Success.Render("✓")
	headline := fmt.Sprintf("%s Published %s entries  %s",
		icon,
		ui.Value.Render(fmt.Sprintf("%d", count)),
		ui.Muted.Render(fmt.Sprintf("(last %d day%s)", days, plural(days))),
	)
	fmt.Println(headline)

	tabLabel := ui.Primary.Render("Daily schedule 2026")
	if sheetURL != "" {
		tabLabel = ui.Hyperlink(tabLabel, sheetURL)
	}

	if resp.InsertedAtRow != "" {
		fmt.Printf("  %s row %s of %s · %s chars · %s\n",
			ui.Muted.Render("→"),
			ui.Value.Render(resp.InsertedAtRow),
			tabLabel,
			ui.Value.Render(fmt.Sprintf("%d", resp.BlobChars)),
			ui.Muted.Render(resp.PublishDate),
		)
	}
	if sheetURL != "" {
		fmt.Println("  " + ui.Link.Render(sheetURL))
	}
	if host := webhookHost(webhook); host != "" {
		fmt.Println("  " + ui.Muted.Render("via "+host))
	}
}

// webhookHost extracts only the host (and scheme) from a webhook URL so the
// path (which can contain token-like segments on services like n8n) isn't
// printed to terminal logs.
func webhookHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "" {
		return u.Scheme + "://" + u.Host
	}
	return u.Host
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// colorizeJSONResponse unmarshals raw JSON bytes, re-indents, and colorizes for terminal display.
func colorizeJSONResponse(body []byte) string {
	var pp json.RawMessage
	if json.Unmarshal(body, &pp) != nil {
		return string(body)
	}
	indented, err := json.MarshalIndent(pp, "", "  ")
	if err != nil {
		return string(body)
	}
	return colorizeJSON(string(indented))
}

// colorizeJSON applies ANSI color codes to indented JSON text for terminal display.
// Keys are cyan, string values are green, numbers are bold cyan, booleans/null are yellow,
// and structural characters ({ } [ ] ,) are gray.
func colorizeJSON(text string) string {
	lines := strings.Split(text, "\n")
	var colored []string
	for _, line := range lines {
		// Find the key: pattern
		colonIdx := strings.Index(line, ":")
		if colonIdx >= 0 {
			keyPart := line[:colonIdx]
			valPart := strings.TrimSpace(line[colonIdx+1:])
			// Color the key (everything before the colon)
			keyPart = ui.Primary.Render(keyPart)
			// Color the colon
			coloredLine := keyPart + " " + ui.Muted.Render(":")
			if valPart != "" {
				coloredLine += " " + colorizeJSONValue(valPart)
			}
			colored = append(colored, coloredLine)
		} else {
			// Line with just structural chars — pass through as-is
			colored = append(colored, line)
		}
	}
	return strings.Join(colored, "\n")
}

func colorizeJSONValue(val string) string {
	if strings.HasPrefix(val, "\"") {
		return ui.Success.Render(val)
	}
	if val == "true" || val == "false" || val == "null" {
		return ui.Highlight.Render(val)
	}
	// Number — includes potential trailing comma
	if len(val) > 0 && (val[0] >= '0' && val[0] <= '9' || val[0] == '-') {
		return ui.Value.Render(val)
	}
	return val
}

// buildStandupBlob renders entries as plain text matching `hlp standup show` output,
// stripped of ANSI/hyperlinks since the destination is a spreadsheet cell.
// Returns the blob, the number of CSV rows included (entry_count), and the
// number of merged groups emitted (group_count). When merge is false,
// group_count == entry_count.
func buildStandupBlob(entries []batch.Entry, limitDays int, baseURL string, merge bool, opts ...entryFormatOpts) (string, int, int) {
	sortEntriesNewestFirst(entries)
	entries = applyDayWindow(entries, limitDays)
	entries = filterStandupEntries(entries, getIgnoredTickets())

	o := entryFormatOpts{}
	if len(opts) > 0 {
		o = opts[0]
	}

	var b strings.Builder
	if !merge {
		for i, e := range entries {
			formatEntry(&b, e, baseURL, o)
			if i < len(entries)-1 {
				b.WriteString("\n")
			}
		}
		return b.String(), len(entries), len(entries)
	}

	groups := mergeEntriesByIssueKey(entries)
	for i, g := range groups {
		formatGroup(&b, g, baseURL, o)
		if i < len(groups)-1 {
			b.WriteString("\n")
		}
	}
	return b.String(), len(entries), len(groups)
}
