package checker

import (
	"context"
	"net/http"
	"time"
)

// PlatformChecker checks availability on developer platforms
type PlatformChecker struct {
	name      string
	urlFormat string
}

// Common platforms
var (
	GitHubChecker = &PlatformChecker{
		name:      "github",
		urlFormat: "https://github.com/%s",
	}
	NPMChecker = &PlatformChecker{
		name:      "npm",
		urlFormat: "https://www.npmjs.com/package/%s",
	}
	PyPIChecker = &PlatformChecker{
		name:      "pypi",
		urlFormat: "https://pypi.org/project/%s/",
	}
	DockerHubChecker = &PlatformChecker{
		name:      "dockerhub",
		urlFormat: "https://hub.docker.com/v2/repositories/library/%s/",
	}
	CratesChecker = &PlatformChecker{
		name:      "crates",
		urlFormat: "https://crates.io/api/v1/crates/%s",
	}
)

// NewPlatformChecker creates a custom platform checker
func NewPlatformChecker(name, urlFormat string) *PlatformChecker {
	return &PlatformChecker{name: name, urlFormat: urlFormat}
}

// ServiceName returns the service identifier
func (c *PlatformChecker) ServiceName() string {
	return c.name
}

// Check checks if the name is available on the platform
func (c *PlatformChecker) Check(ctx context.Context, name string) Result {
	result := Result{
		Name:    name,
		Service: c.ServiceName(),
		Status:  StatusUnknown,
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects - a redirect often means the resource exists
			return http.ErrUseLastResponse
		},
	}

	url := c.formatURL(name)
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		result.Status = StatusError
		result.Error = err.Error()
		return result
	}

	req.Header.Set("User-Agent", "hlp-namefinder/1.0")

	resp, err := client.Do(req)
	if err != nil {
		result.Status = StatusError
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusMovedPermanently, http.StatusFound:
		result.Status = StatusTaken
		result.Confidence = 0.95
	case http.StatusNotFound:
		result.Status = StatusAvailable
		result.Confidence = 0.9
	case http.StatusTooManyRequests:
		result.Status = StatusRateLimited
		result.Error = "rate limited"
	case http.StatusForbidden:
		// Some platforms return 403 for rate limiting
		result.Status = StatusRateLimited
		result.Error = "forbidden/rate limited"
	default:
		result.Status = StatusUnknown
		result.Error = resp.Status
	}

	return result
}

func (c *PlatformChecker) formatURL(name string) string {
	// Simple format string replacement
	url := c.urlFormat
	for i := 0; i < len(url)-1; i++ {
		if url[i] == '%' && url[i+1] == 's' {
			url = url[:i] + name + url[i+2:]
			break
		}
	}
	return url
}

// CreatePlatformCheckers creates checkers for common platforms
func CreatePlatformCheckers() []Checker {
	return []Checker{
		GitHubChecker,
		NPMChecker,
		PyPIChecker,
		DockerHubChecker,
		CratesChecker,
	}
}
