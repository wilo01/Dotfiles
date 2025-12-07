package timesheet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client handles Timesheet plugin GraphQL operations
type Client struct {
	jiraURL    string
	httpClient *resty.Client
}

// Config holds timesheet configuration
type Config struct {
	JiraURL      string `json:"jira_url"`
	SessionToken string `json:"session_token"`
	XSRFToken    string `json:"xsrf_token"`
	JSESSIONID   string `json:"jsessionid"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// NewClient creates a new Timesheet client
func NewClient(cfg Config) *Client {
	client := resty.New().
		SetBaseURL(cfg.JiraURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "*/*").
		SetHeader("apollographql-client-name", "GATEWAY").
		SetHeader("cache-control", "no-cache").
		SetHeader("pragma", "no-cache").
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second)

	// Set cookies
	if cfg.SessionToken != "" {
		client.SetCookie(&http.Cookie{Name: "tenant.session.token", Value: cfg.SessionToken})
	}
	if cfg.XSRFToken != "" {
		client.SetCookie(&http.Cookie{Name: "atlassian.xsrf.token", Value: cfg.XSRFToken})
	}
	if cfg.JSESSIONID != "" {
		client.SetCookie(&http.Cookie{Name: "JSESSIONID", Value: cfg.JSESSIONID})
	}

	return &Client{
		jiraURL:    cfg.JiraURL,
		httpClient: client,
	}
}

// WorklogEntry represents a worklog to be created
type WorklogEntry struct {
	IssueKey  string
	TimeSpent string // Jira format: "30m", "2h", "1d 4h"
	Started   time.Time
	Comment   string
}

// LogWork logs work time to a Jira issue via Timesheet plugin
func (c *Client) LogWork(entry WorklogEntry) error {
	// Format timestamp for Jira
	startedStr := entry.Started.Format("2006-01-02T15:04:05.000-0700")

	// Build the GraphQL mutation
	mutation := buildMutation(c.jiraURL, entry.IssueKey, entry.TimeSpent, startedStr, entry.Comment)

	var result struct {
		Data struct {
			InvokeExtension struct {
				Success  bool `json:"success"`
				Response struct {
					Body string `json:"body"`
				} `json:"response"`
				Errors []struct {
					Message    string `json:"message"`
					Extensions struct {
						ErrorType  string `json:"errorType"`
						StatusCode int    `json:"statusCode"`
					} `json:"extensions"`
				} `json:"errors"`
			} `json:"invokeExtension"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	resp, err := c.httpClient.R().
		SetBody(mutation).
		SetResult(&result).
		Post("/gateway/api/graphql")

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("request failed: %s", resp.Status())
	}

	// Check for GraphQL errors
	if len(result.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", result.Errors[0].Message)
	}

	// Check for extension errors
	if len(result.Data.InvokeExtension.Errors) > 0 {
		return fmt.Errorf("extension error: %s", result.Data.InvokeExtension.Errors[0].Message)
	}

	if !result.Data.InvokeExtension.Success {
		// Try to parse the response body for more details
		var bodyResp struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal([]byte(result.Data.InvokeExtension.Response.Body), &bodyResp); err == nil {
			if bodyResp.Error != "" {
				return fmt.Errorf("worklog failed: %s", bodyResp.Error)
			}
			if bodyResp.Message != "" {
				return fmt.Errorf("worklog failed: %s", bodyResp.Message)
			}
		}
		return fmt.Errorf("worklog failed: unknown error")
	}

	return nil
}

// TestConnection tests if the timesheet configuration is valid
func (c *Client) TestConnection() error {
	// Try a simple query to verify cookies are valid
	query := map[string]interface{}{
		"operationName": nil,
		"variables":     map[string]interface{}{},
		"query":         "{ __typename }",
	}

	resp, err := c.httpClient.R().
		SetBody(query).
		Post("/gateway/api/graphql")

	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized || resp.StatusCode() == http.StatusForbidden {
		return fmt.Errorf("authentication failed: session tokens may be expired")
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("connection failed: %s", resp.Status())
	}

	return nil
}

func buildMutation(siteURL, issueKey, timeSpent, started, comment string) map[string]interface{} {
	body := map[string]interface{}{
		"issueIdOrKey": issueKey,
		"timeSpent":    timeSpent,
		"started":      started,
		"newEstimate":  "0m",
		"attributes":   map[string]interface{}{},
	}

	if comment != "" {
		body["comment"] = comment
	}

	return map[string]interface{}{
		"operationName": "forge_ui_invokeExtension",
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"contextIds":  []string{},
				"extensionId": "ari:cloud:ecosystem::extension/48cddd1c-9611-49cd-a49c-9ff6f9bee138/42ccec14-73af-418b-beb9-c76305bfe500/static/timesheet-global-page",
				"payload": map[string]interface{}{
					"call": map[string]interface{}{
						"path":       "/forge/worklog",
						"method":     "POST",
						"body":       body,
						"invokeType": "ui-remote-fetch",
					},
					"context": map[string]interface{}{
						"cloudId":         "",
						"localId":         "ari:cloud:ecosystem::extension/48cddd1c-9611-49cd-a49c-9ff6f9bee138/42ccec14-73af-418b-beb9-c76305bfe500/static/timesheet-global-page",
						"environmentId":   "42ccec14-73af-418b-beb9-c76305bfe500",
						"environmentType": "PRODUCTION",
						"moduleKey":       "timesheet-global-page",
						"siteUrl":         siteURL,
						"appVersion":      "4.9.0",
						"extension": map[string]interface{}{
							"type": "jira:globalPage",
							"jira": map[string]interface{}{
								"isNewNavigation": true,
							},
						},
					},
				},
				"entryPoint": "resolver",
			},
		},
		"query": `mutation forge_ui_invokeExtension($input: InvokeExtensionInput!) {
  invokeExtension(input: $input) {
    success
    response {
      body
      __typename
    }
    contextToken {
      jwt
      expiresAt
      __typename
    }
    errors {
      message
      extensions {
        errorType
        statusCode
        ... on InvokeExtensionPayloadErrorExtension {
          fields {
            authInfoUrl
            __typename
          }
          __typename
        }
        __typename
      }
      __typename
    }
    __typename
  }
}`,
	}
}
