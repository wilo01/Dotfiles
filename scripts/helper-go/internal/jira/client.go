package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client is a JIRA REST API client
type Client struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *resty.Client
}

// Ticket represents a JIRA ticket
type Ticket struct {
	Key       string    `json:"key"`
	Summary   string    `json:"summary"`
	Status    string    `json:"status"`
	Assignee  string    `json:"assignee"`
	IssueType string    `json:"issue_type"`
	Priority  string    `json:"priority"`
	ParentKey string    `json:"parent_key,omitempty"`
	IsSubtask bool      `json:"is_subtask"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Worklog represents a JIRA worklog entry
type Worklog struct {
	ID              string        `json:"id"`
	IssueKey        string        `json:"issue_key"`
	Started         time.Time     `json:"started"`
	TimeSpent       time.Duration `json:"time_spent"`
	TimeSpentStr    string        `json:"time_spent_str"` // "2h 30m" format for comparison
	Comment         string        `json:"comment"`
	Author          string        `json:"author"`
	AuthorAccountId string        `json:"author_account_id"` // Cloud: accountId for reliable matching
}

// WorklogEntry is used to create a new worklog
type WorklogEntry struct {
	TimeSpent string    `json:"timeSpent"`
	Started   time.Time `json:"started"`
	Comment   string    `json:"comment,omitempty"`
}

// CurrentUser represents the authenticated JIRA user
type CurrentUser struct {
	AccountID    string `json:"accountId"` // Cloud
	Name         string `json:"name"`      // Server/DC username
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// IssueDetails holds summary and type for an issue (used by sync)
type IssueDetails struct {
	Summary   string
	IssueType string
	ParentKey string // For sub-tasks: parent ticket key
	IsSubtask bool   // True if this is a sub-task
}

// NewClient creates a new JIRA client
func NewClient(baseURL, email, apiToken string) *Client {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	// Set auth based on whether it's Cloud or Server
	if isCloudInstance(baseURL) {
		// Cloud: use Basic auth with email:apiToken
		auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + apiToken))
		client.SetHeader("Authorization", "Basic "+auth)
	} else if strings.HasPrefix(apiToken, "JSESSIONID=") {
		// Server/DC with session cookie (for instances with basic auth disabled)
		client.SetHeader("Cookie", apiToken)
	} else if len(apiToken) == 32 && !strings.Contains(apiToken, "-") {
		// Looks like a raw session ID (32 hex chars)
		client.SetHeader("Cookie", "JSESSIONID="+apiToken)
	} else {
		// Server/DC: use Bearer token (PAT)
		client.SetHeader("Authorization", "Bearer "+apiToken)
	}

	return &Client{
		baseURL:    baseURL,
		email:      email,
		apiToken:   apiToken,
		httpClient: client,
	}
}

// isCloudInstance checks if the URL is a JIRA Cloud instance
func isCloudInstance(baseURL string) bool {
	return strings.Contains(baseURL, ".atlassian.net")
}

// stripOrderBy removes ORDER BY clause from JQL (Cloud doesn't support it in enhanced search)
func stripOrderBy(jql string) string {
	idx := strings.Index(strings.ToUpper(jql), "ORDER BY")
	if idx != -1 {
		return strings.TrimSpace(jql[:idx])
	}
	return jql
}

// GetTicket fetches a ticket by key
func (c *Client) GetTicket(key string) (*Ticket, error) {
	var result struct {
		Key    string `json:"key"`
		Fields struct {
			Summary string `json:"summary"`
			Status  struct {
				Name string `json:"name"`
			} `json:"status"`
			Assignee *struct {
				DisplayName string `json:"displayName"`
			} `json:"assignee"`
			IssueType struct {
				Name    string `json:"name"`
				Subtask bool   `json:"subtask"`
			} `json:"issuetype"`
			Priority *struct {
				Name string `json:"name"`
			} `json:"priority"`
			Parent *struct {
				Key string `json:"key"`
			} `json:"parent"`
			Updated string `json:"updated"`
		} `json:"fields"`
	}

	resp, err := c.httpClient.R().
		SetResult(&result).
		Get("/rest/api/2/issue/" + key)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch ticket: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch ticket: %s", resp.Status())
	}

	ticket := &Ticket{
		Key:       result.Key,
		Summary:   result.Fields.Summary,
		Status:    result.Fields.Status.Name,
		IssueType: result.Fields.IssueType.Name,
		IsSubtask: result.Fields.IssueType.Subtask,
	}

	if result.Fields.Assignee != nil {
		ticket.Assignee = result.Fields.Assignee.DisplayName
	}
	if result.Fields.Priority != nil {
		ticket.Priority = result.Fields.Priority.Name
	}
	if result.Fields.Parent != nil {
		ticket.ParentKey = result.Fields.Parent.Key
	}

	// Parse updated time
	if result.Fields.Updated != "" {
		if t, err := time.Parse("2006-01-02T15:04:05.000-0700", result.Fields.Updated); err == nil {
			ticket.UpdatedAt = t
		}
	}

	return ticket, nil
}

// GetTicketDescription fetches an issue's description as plain text.
// API v2 returns wiki-markup strings; ADF objects (v3-style) are flattened
// to their text nodes, mirroring the worklog comment handling.
func (c *Client) GetTicketDescription(key string) (string, error) {
	var result struct {
		Fields struct {
			Description json.RawMessage `json:"description"`
		} `json:"fields"`
	}

	resp, err := c.httpClient.R().
		SetQueryParam("fields", "description").
		SetResult(&result).
		Get("/rest/api/2/issue/" + key)

	if err != nil {
		return "", fmt.Errorf("failed to fetch description: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("failed to fetch description: %s", resp.Status())
	}

	raw := result.Fields.Description
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	// ADF fallback: collect text nodes from paragraphs
	var adf struct {
		Content []struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &adf); err == nil {
		var parts []string
		for _, p := range adf.Content {
			var line strings.Builder
			for _, t := range p.Content {
				line.WriteString(t.Text)
			}
			if line.Len() > 0 {
				parts = append(parts, line.String())
			}
		}
		return strings.Join(parts, "\n"), nil
	}

	return "", nil
}

// LogWork logs work to a ticket
func (c *Client) LogWork(key string, entry WorklogEntry) error {
	body := map[string]interface{}{
		"timeSpent": entry.TimeSpent,
		"started":   entry.Started.Format("2006-01-02T15:04:05.000-0700"),
	}

	if entry.Comment != "" {
		if isCloudInstance(c.baseURL) {
			// Cloud uses Atlassian Document Format
			body["comment"] = map[string]interface{}{
				"type":    "doc",
				"version": 1,
				"content": []map[string]interface{}{
					{
						"type": "paragraph",
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": entry.Comment,
							},
						},
					},
				},
			}
		} else {
			// Server/DC uses plain text
			body["comment"] = entry.Comment
		}
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		Post("/rest/api/2/issue/" + key + "/worklog")

	if err != nil {
		return fmt.Errorf("failed to log work: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to log work: %s - %s", resp.Status(), resp.String())
	}

	return nil
}

// GetWorklogs fetches worklogs for a ticket
func (c *Client) GetWorklogs(key string) ([]Worklog, error) {
	var result struct {
		Worklogs []struct {
			ID        string          `json:"id"`
			Started   string          `json:"started"`
			TimeSpent string          `json:"timeSpent"`
			Comment   json.RawMessage `json:"comment"`
			Author    struct {
				DisplayName string `json:"displayName"`
				AccountId   string `json:"accountId"`
			} `json:"author"`
			TimeSpentSeconds int `json:"timeSpentSeconds"`
		} `json:"worklogs"`
	}

	resp, err := c.httpClient.R().
		SetResult(&result).
		Get("/rest/api/2/issue/" + key + "/worklog")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch worklogs: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch worklogs: %s", resp.Status())
	}

	var worklogs []Worklog
	for _, w := range result.Worklogs {
		wl := Worklog{
			ID:              w.ID,
			IssueKey:        key,
			TimeSpent:       time.Duration(w.TimeSpentSeconds) * time.Second,
			TimeSpentStr:    w.TimeSpent, // "2h 30m" format from JIRA
			Author:          w.Author.DisplayName,
			AuthorAccountId: w.Author.AccountId,
		}

		// Parse started time
		if t, err := time.Parse("2006-01-02T15:04:05.000-0700", w.Started); err == nil {
			wl.Started = t
		}

		// Parse comment (can be string or ADF)
		if len(w.Comment) > 0 {
			var commentStr string
			if err := json.Unmarshal(w.Comment, &commentStr); err == nil {
				wl.Comment = commentStr
			} else {
				// Try to extract text from ADF
				var adf struct {
					Content []struct {
						Content []struct {
							Text string `json:"text"`
						} `json:"content"`
					} `json:"content"`
				}
				if err := json.Unmarshal(w.Comment, &adf); err == nil {
					for _, p := range adf.Content {
						for _, t := range p.Content {
							if t.Text != "" {
								wl.Comment = t.Text
								break
							}
						}
					}
				}
			}
		}

		worklogs = append(worklogs, wl)
	}

	return worklogs, nil
}

// GetWorklogsByDate fetches worklogs for a ticket on a specific date (current user only)
func (c *Client) GetWorklogsByDate(key string, date time.Time) ([]Worklog, error) {
	worklogs, err := c.GetWorklogs(key)
	if err != nil {
		return nil, err
	}

	// Get current user for filtering
	currentUser, err := c.GetCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Filter by date AND author (current user only)
	targetDate := date.Format("2006-01-02")
	var filtered []Worklog
	for _, wl := range worklogs {
		if wl.Started.Format("2006-01-02") == targetDate && isCurrentUserWorklog(wl, currentUser) {
			filtered = append(filtered, wl)
		}
	}

	return filtered, nil
}

// DeleteWorklog deletes a worklog by ID
func (c *Client) DeleteWorklog(key, worklogID string) error {
	resp, err := c.httpClient.R().
		SetQueryParam("adjustEstimate", "leave").
		Delete("/rest/api/2/issue/" + key + "/worklog/" + worklogID)

	if err != nil {
		return fmt.Errorf("failed to delete worklog: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete worklog: %s - %s", resp.Status(), resp.String())
	}

	return nil
}

// ExtractProjectKey extracts the project key from an issue key (e.g., VIS-1234 -> VIS)
func ExtractProjectKey(issueKey string) string {
	parts := strings.Split(issueKey, "-")
	if len(parts) >= 1 {
		return strings.ToUpper(parts[0])
	}
	return ""
}

// SearchAssignedTickets searches for tickets assigned to the current user
func (c *Client) SearchAssignedTickets(project string) ([]Ticket, error) {
	jql := "assignee = currentUser() AND status != Done"
	if project != "" {
		jql = fmt.Sprintf("project = %s AND %s", project, jql)
	}
	jql += " ORDER BY updated DESC"

	return c.Search(jql, 50)
}

// SearchSprintTickets searches for tickets in the current sprint assigned to user
func (c *Client) SearchSprintTickets() ([]Ticket, error) {
	jql := "assignee = currentUser() AND sprint in openSprints() AND status != Done ORDER BY updated DESC"
	return c.Search(jql, 50)
}

// Search performs a JQL search
func (c *Client) Search(jql string, maxResults int) ([]Ticket, error) {
	var result struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary string `json:"summary"`
				Status  struct {
					Name string `json:"name"`
				} `json:"status"`
				Assignee *struct {
					DisplayName string `json:"displayName"`
				} `json:"assignee"`
				IssueType struct {
					Name    string `json:"name"`
					Subtask bool   `json:"subtask"`
				} `json:"issuetype"`
				Priority *struct {
					Name string `json:"name"`
				} `json:"priority"`
				Parent *struct {
					Key string `json:"key"`
				} `json:"parent"`
				Updated string `json:"updated"`
			} `json:"fields"`
		} `json:"issues"`
	}

	// Cloud uses different endpoint and doesn't support ORDER BY
	endpoint := "/rest/api/2/search"
	if isCloudInstance(c.baseURL) {
		endpoint = "/rest/api/2/search/jql"
		jql = stripOrderBy(jql)
	}

	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"jql":        jql,
			"maxResults": fmt.Sprintf("%d", maxResults),
			"fields":     "summary,status,assignee,issuetype,priority,parent,updated",
		}).
		SetResult(&result).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to search: %s", resp.Status())
	}

	var tickets []Ticket
	for _, issue := range result.Issues {
		ticket := Ticket{
			Key:       issue.Key,
			Summary:   issue.Fields.Summary,
			Status:    issue.Fields.Status.Name,
			IssueType: issue.Fields.IssueType.Name,
			IsSubtask: issue.Fields.IssueType.Subtask,
		}

		if issue.Fields.Assignee != nil {
			ticket.Assignee = issue.Fields.Assignee.DisplayName
		}
		if issue.Fields.Priority != nil {
			ticket.Priority = issue.Fields.Priority.Name
		}
		if issue.Fields.Parent != nil {
			ticket.ParentKey = issue.Fields.Parent.Key
		}
		if issue.Fields.Updated != "" {
			if t, err := time.Parse("2006-01-02T15:04:05.000-0700", issue.Fields.Updated); err == nil {
				ticket.UpdatedAt = t
			}
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

// TestConnection tests the JIRA connection
func (c *Client) TestConnection() error {
	resp, err := c.httpClient.R().
		Get("/rest/api/2/myself")

	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("authentication failed: %s", resp.Status())
	}

	return nil
}

// GetCurrentUser fetches the authenticated user's info
func (c *Client) GetCurrentUser() (*CurrentUser, error) {
	var result CurrentUser

	resp, err := c.httpClient.R().
		SetResult(&result).
		Get("/rest/api/2/myself")

	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get current user: %s", resp.Status())
	}

	return &result, nil
}

// GetIssueSummaries fetches summaries for multiple tickets in a single request
func (c *Client) GetIssueSummaries(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	// Build JQL: key in (VIS-1, VIS-2, ...)
	jql := fmt.Sprintf("key in (%s)", strings.Join(keys, ", "))

	tickets, err := c.Search(jql, len(keys))
	if err != nil {
		return nil, err
	}

	summaries := make(map[string]string)
	for _, t := range tickets {
		summaries[t.Key] = t.Summary
	}

	return summaries, nil
}

// GetIssueDetails fetches summaries and types for multiple tickets in a single request
func (c *Client) GetIssueDetails(keys []string) (map[string]IssueDetails, error) {
	if len(keys) == 0 {
		return make(map[string]IssueDetails), nil
	}

	// Build JQL: key in (VIS-1, VIS-2, ...)
	jql := fmt.Sprintf("key in (%s)", strings.Join(keys, ", "))

	tickets, err := c.Search(jql, len(keys))
	if err != nil {
		return nil, err
	}

	details := make(map[string]IssueDetails)
	for _, t := range tickets {
		details[t.Key] = IssueDetails{
			Summary:   t.Summary,
			IssueType: t.IssueType,
			ParentKey: t.ParentKey,
			IsSubtask: t.IsSubtask,
		}
	}

	return details, nil
}

// SearchIssuesWithWorklogs finds issues that were updated in date range
// Note: Uses 'updated' field which works on both Cloud and Server/DC
// (worklogAuthor and worklogDate are Cloud-only JQL fields)
func (c *Client) SearchIssuesWithWorklogs(fromDate, toDate time.Time) ([]string, error) {
	// Note: ORDER BY is stripped for Cloud in Search() method
	jql := fmt.Sprintf(
		`updated >= "%s" ORDER BY updated DESC`,
		fromDate.Format("2006-01-02"),
	)

	tickets, err := c.Search(jql, 200)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, t := range tickets {
		keys = append(keys, t.Key)
	}
	return keys, nil
}

// SearchIssuesWithUserWorklogs finds issues with current user's worklogs (Cloud only)
// Uses worklogAuthor and worklogDate JQL fields for efficient server-side filtering
func (c *Client) SearchIssuesWithUserWorklogs(fromDate, toDate time.Time) ([]string, error) {
	jql := fmt.Sprintf(
		`worklogAuthor = currentUser() AND worklogDate >= "%s" AND worklogDate <= "%s"`,
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
	)

	tickets, err := c.Search(jql, 200)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, t := range tickets {
		keys = append(keys, t.Key)
	}
	return keys, nil
}

// FetchUserWorklogs fetches all worklogs for current user in date range
// progressFn is called with (current, total, issueKey) for each issue processed
func (c *Client) FetchUserWorklogs(fromDate, toDate time.Time, progressFn func(current, total int, issueKey string)) ([]Worklog, error) {
	// Get current user
	currentUser, err := c.GetCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Search for issues - use Cloud-specific JQL if available
	var issueKeys []string
	isCloud := isCloudInstance(c.baseURL)

	if isCloud {
		// Cloud: Use worklogAuthor JQL for efficient server-side filtering
		issueKeys, err = c.SearchIssuesWithUserWorklogs(fromDate, toDate)
	} else {
		// Server/DC: Fall back to updated date search
		issueKeys, err = c.SearchIssuesWithWorklogs(fromDate, toDate)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	// Fetch worklogs for each issue and filter
	var result []Worklog
	total := len(issueKeys)
	for i, key := range issueKeys {
		if progressFn != nil {
			progressFn(i+1, total, key)
		}

		worklogs, err := c.GetWorklogs(key)
		if err != nil {
			continue // Skip issues with fetch errors
		}

		for _, wl := range worklogs {
			// Filter by author - Cloud JQL pre-filters, but still verify
			// Server/DC requires this filter since we search by 'updated' date
			if !isCurrentUserWorklog(wl, currentUser) {
				continue
			}

			// Filter by date range
			wlDate := wl.Started.Truncate(24 * time.Hour)
			fromTrunc := fromDate.Truncate(24 * time.Hour)
			toTrunc := toDate.Truncate(24 * time.Hour)

			if wlDate.Before(fromTrunc) || wlDate.After(toTrunc) {
				continue
			}

			result = append(result, wl)
		}
	}

	return result, nil
}

// isCurrentUserWorklog checks if a worklog belongs to the current user
func isCurrentUserWorklog(wl Worklog, user *CurrentUser) bool {
	if user == nil {
		return false
	}
	// For Cloud: prefer AccountId comparison (most reliable)
	if user.AccountID != "" && wl.AuthorAccountId != "" {
		return wl.AuthorAccountId == user.AccountID
	}
	// Fallback: case-insensitive display name comparison
	return strings.EqualFold(wl.Author, user.DisplayName)
}
