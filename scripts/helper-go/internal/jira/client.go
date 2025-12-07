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
	ID        string        `json:"id"`
	IssueKey  string        `json:"issue_key"`
	Started   time.Time     `json:"started"`
	TimeSpent time.Duration `json:"time_spent"`
	Comment   string        `json:"comment"`
	Author    string        `json:"author"`
}

// WorklogEntry is used to create a new worklog
type WorklogEntry struct {
	TimeSpent string    `json:"timeSpent"`
	Started   time.Time `json:"started"`
	Comment   string    `json:"comment,omitempty"`
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
	} else {
		// Server/DC: use Bearer token
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

// GetTicket fetches a ticket by key
func (c *Client) GetTicket(key string) (*Ticket, error) {
	var result struct {
		Key    string `json:"key"`
		Fields struct {
			Summary   string `json:"summary"`
			Status    struct {
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
			ID      string `json:"id"`
			Started string `json:"started"`
			Comment json.RawMessage `json:"comment"`
			Author  struct {
				DisplayName string `json:"displayName"`
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
			ID:        w.ID,
			IssueKey:  key,
			TimeSpent: time.Duration(w.TimeSpentSeconds) * time.Second,
			Author:    w.Author.DisplayName,
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

// SearchAssignedTickets searches for tickets assigned to the current user
func (c *Client) SearchAssignedTickets(project string) ([]Ticket, error) {
	jql := "assignee = currentUser() AND status != Done"
	if project != "" {
		jql = fmt.Sprintf("project = %s AND %s", project, jql)
	}
	jql += " ORDER BY updated DESC"

	return c.Search(jql, 50)
}

// Search performs a JQL search
func (c *Client) Search(jql string, maxResults int) ([]Ticket, error) {
	var result struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary   string `json:"summary"`
				Status    struct {
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
				Updated string `json:"updated"`
			} `json:"fields"`
		} `json:"issues"`
	}

	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"jql":        jql,
			"maxResults": fmt.Sprintf("%d", maxResults),
			"fields":     "summary,status,assignee,issuetype,priority,updated",
		}).
		SetResult(&result).
		Get("/rest/api/2/search")

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
