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

	blob, count, groupCount := buildStandupBlob(entries, days, baseURL, !publishNoMerge)
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

	if publishDryRun {
		webhookDisplay := webhook
		if webhookDisplay == "" {
			webhookDisplay = "<unset>"
		}
		sheetDisplay := sheetURL
		if sheetDisplay == "" {
			sheetDisplay = "<unset>"
		}
		fmt.Printf("--- DRY RUN ---\nwebhook: %s\nsheet:   %s\nentries: %d, days: %d\n---\n%s\n", webhookDisplay, sheetDisplay, count, days, blob)
		return nil
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
		return fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(respBody))
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

// buildStandupBlob renders entries as plain text matching `hlp standup show` output,
// stripped of ANSI/hyperlinks since the destination is a spreadsheet cell.
// Returns the blob, the number of CSV rows included (entry_count), and the
// number of merged groups emitted (group_count). When merge is false,
// group_count == entry_count.
func buildStandupBlob(entries []batch.Entry, limitDays int, baseURL string, merge bool) (string, int, int) {
	sortEntriesNewestFirst(entries)
	entries = applyDayWindow(entries, limitDays)
	entries = filterStandupEntries(entries, getIgnoredTickets())

	var b strings.Builder
	if !merge {
		for i, e := range entries {
			formatEntry(&b, e, baseURL, entryFormatOpts{})
			if i < len(entries)-1 {
				b.WriteString("\n")
			}
		}
		return b.String(), len(entries), len(entries)
	}

	groups := mergeEntriesByIssueKey(entries)
	for i, g := range groups {
		formatGroup(&b, g, baseURL, entryFormatOpts{})
		if i < len(groups)-1 {
			b.WriteString("\n")
		}
	}
	return b.String(), len(entries), len(groups)
}
