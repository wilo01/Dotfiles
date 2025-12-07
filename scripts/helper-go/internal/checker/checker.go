package checker

import (
	"context"
	"sync"
)

// Status represents the availability status
type Status string

const (
	StatusAvailable   Status = "available"
	StatusTaken       Status = "taken"
	StatusError       Status = "error"
	StatusRateLimited Status = "rate_limited"
	StatusUnknown     Status = "unknown"
)

// Result represents the result of a check
type Result struct {
	Name       string         `json:"name"`
	Service    string         `json:"service"`
	Status     Status         `json:"status"`
	Confidence float64        `json:"confidence"`
	Details    map[string]any `json:"details,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// Checker defines the interface for availability checkers
type Checker interface {
	Check(ctx context.Context, name string) Result
	ServiceName() string
}

// Executor runs multiple checkers in parallel
type Executor struct {
	workers int
}

// NewExecutor creates a new parallel executor
func NewExecutor(workers int) *Executor {
	if workers <= 0 {
		workers = 10
	}
	return &Executor{workers: workers}
}

// Execute runs checkers for multiple names in parallel
func (e *Executor) Execute(ctx context.Context, names []string, checkers []Checker) map[string]map[string]Result {
	results := make(map[string]map[string]Result)
	var mu sync.Mutex

	// Initialize result maps
	for _, name := range names {
		results[name] = make(map[string]Result)
	}

	// Create work items
	type workItem struct {
		name    string
		checker Checker
	}

	var items []workItem
	for _, name := range names {
		for _, checker := range checkers {
			items = append(items, workItem{name, checker})
		}
	}

	// Worker pool
	sem := make(chan struct{}, e.workers)
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		go func(w workItem) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result := w.checker.Check(ctx, w.name)

			mu.Lock()
			results[w.name][w.checker.ServiceName()] = result
			mu.Unlock()
		}(item)
	}

	wg.Wait()
	return results
}
