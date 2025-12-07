package checker

import (
	"context"
	"net"
	"strings"
	"time"
)

// DomainChecker checks domain availability via DNS
type DomainChecker struct {
	tld string
}

// NewDomainChecker creates a domain checker for a specific TLD
func NewDomainChecker(tld string) *DomainChecker {
	tld = strings.TrimPrefix(tld, ".")
	return &DomainChecker{tld: tld}
}

// ServiceName returns the service identifier
func (c *DomainChecker) ServiceName() string {
	return "domain." + c.tld
}

// Check checks if the domain is available
func (c *DomainChecker) Check(ctx context.Context, name string) Result {
	domain := name + "." + c.tld

	result := Result{
		Name:    name,
		Service: c.ServiceName(),
		Status:  StatusUnknown,
	}

	// Use DNS lookup to check if domain exists
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to resolve the domain
	_, err := resolver.LookupHost(ctx, domain)

	if err != nil {
		// Check if it's a "not found" error (domain likely available)
		if dnsErr, ok := err.(*net.DNSError); ok {
			if dnsErr.IsNotFound || strings.Contains(dnsErr.Error(), "no such host") {
				result.Status = StatusAvailable
				result.Confidence = 0.7 // DNS check is not 100% reliable
				return result
			}
			if dnsErr.IsTimeout {
				result.Status = StatusError
				result.Error = "timeout"
				return result
			}
		}
		// Other errors might mean the domain exists but has no records
		result.Status = StatusUnknown
		result.Error = err.Error()
		return result
	}

	// Domain resolves - it's taken
	result.Status = StatusTaken
	result.Confidence = 0.95
	return result
}

// CommonTLDs returns a list of common TLDs to check
func CommonTLDs() []string {
	return []string{
		"com", "net", "org", "io", "app", "dev",
		"co", "me", "ai", "so", "sh",
		"pl", "eu", "de", "uk",
	}
}

// CreateDomainCheckers creates checkers for multiple TLDs
func CreateDomainCheckers(tlds []string) []Checker {
	checkers := make([]Checker, len(tlds))
	for i, tld := range tlds {
		checkers[i] = NewDomainChecker(tld)
	}
	return checkers
}
