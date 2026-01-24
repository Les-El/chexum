package checkpoint

import (
	"context"
	"sync"
)

// IssueCollector collects issues from multiple engines.
type IssueCollector struct {
	mu     sync.Mutex
	issues []Issue
}

// NewIssueCollector creates a new collector.
func NewIssueCollector() *IssueCollector {
	return &IssueCollector{
		issues: make([]Issue, 0),
	}
}

// Collect adds issues to the collector.
func (c *IssueCollector) Collect(issues []Issue) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.issues = append(c.issues, issues...)
}

// Issues returns all collected issues.
func (c *IssueCollector) Issues() []Issue {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Issue(nil), c.issues...)
}

// Runner coordinates multiple analysis engines.
type Runner struct {
	engines   []AnalysisEngine
	collector *IssueCollector
}

// NewRunner creates a new runner.
func NewRunner(engines []AnalysisEngine) *Runner {
	return &Runner{
		engines:   engines,
		collector: NewIssueCollector(),
	}
}

// Run executes all registered engines.
func (r *Runner) Run(ctx context.Context, path string) error {
	for _, engine := range r.engines {
		issues, err := engine.Analyze(ctx, path)
		if err != nil {
			return err
		}
		r.collector.Collect(issues)
	}
	return nil
}

// GetIssues returns the findings from the last run.
func (r *Runner) GetIssues() []Issue {
	return r.collector.Issues()
}
