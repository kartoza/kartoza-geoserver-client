// Package health provides health checking functionality for hosted instances.
package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/repository"
)

// Status represents the health status of an instance.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// CheckResult represents the result of a health check.
type CheckResult struct {
	InstanceID   string
	Status       Status
	ResponseTime time.Duration
	StatusCode   int
	Error        string
	CheckedAt    time.Time
}

// Checker performs health checks on instances.
type Checker struct {
	instanceRepo *repository.InstanceRepository
	httpClient   *http.Client
	interval     time.Duration
	timeout      time.Duration
	stopCh       chan struct{}
	mu           sync.RWMutex
	results      map[string]*CheckResult
}

// Config holds health checker configuration.
type Config struct {
	Interval time.Duration
	Timeout  time.Duration
}

// NewChecker creates a new health checker.
func NewChecker(instanceRepo *repository.InstanceRepository, config Config) *Checker {
	if config.Interval == 0 {
		config.Interval = 60 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &Checker{
		instanceRepo: instanceRepo,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		interval: config.Interval,
		timeout:  config.Timeout,
		stopCh:   make(chan struct{}),
		results:  make(map[string]*CheckResult),
	}
}

// Start begins the periodic health checking.
func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Run initial check
	c.checkAllInstances(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.checkAllInstances(ctx)
		}
	}
}

// Stop stops the health checker.
func (c *Checker) Stop() {
	close(c.stopCh)
}

// CheckInstance performs a health check on a single instance.
func (c *Checker) CheckInstance(ctx context.Context, instance *models.Instance) *CheckResult {
	result := &CheckResult{
		InstanceID: instance.ID,
		Status:     StatusUnknown,
		CheckedAt:  time.Now(),
	}

	// Skip instances that aren't in a checkable state
	if instance.Status != models.InstanceStatusOnline &&
		instance.Status != models.InstanceStatusStartingUp {
		result.Status = StatusUnknown
		result.Error = fmt.Sprintf("instance status is %s", instance.Status)
		return result
	}

	if instance.URL == "" {
		result.Status = StatusUnknown
		result.Error = "instance has no URL"
		return result
	}

	// Build health check URL
	healthURL := instance.URL
	if instance.HealthEndpoint != "" {
		healthURL = instance.URL + instance.HealthEndpoint
	} else {
		// Default health endpoints by product type
		switch instance.ProductID {
		case "geoserver":
			healthURL = instance.URL + "/geoserver/web/"
		case "geonode":
			healthURL = instance.URL + "/api/v2/"
		case "postgis":
			// PostGIS doesn't have HTTP endpoint, skip
			result.Status = StatusUnknown
			result.Error = "PostGIS health check not supported via HTTP"
			return result
		}
	}

	// Perform the check
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	resp, err := c.httpClient.Do(req)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Status = StatusHealthy
	} else {
		result.Status = StatusUnhealthy
		result.Error = fmt.Sprintf("unhealthy status code: %d", resp.StatusCode)
	}

	return result
}

// GetResult returns the latest health check result for an instance.
func (c *Checker) GetResult(instanceID string) *CheckResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.results[instanceID]
}

// GetAllResults returns all health check results.
func (c *Checker) GetAllResults() map[string]*CheckResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid race conditions
	results := make(map[string]*CheckResult, len(c.results))
	for k, v := range c.results {
		results[k] = v
	}
	return results
}

// checkAllInstances checks all active instances.
func (c *Checker) checkAllInstances(ctx context.Context) {
	// Get all instances that should be checked
	instances, err := c.instanceRepo.GetActiveInstances(ctx)
	if err != nil {
		return
	}

	// Check instances concurrently
	var wg sync.WaitGroup
	resultsCh := make(chan *CheckResult, len(instances))

	for _, instance := range instances {
		wg.Add(1)
		go func(inst *models.Instance) {
			defer wg.Done()
			result := c.CheckInstance(ctx, inst)
			resultsCh <- result

			// Update instance health status in database
			healthStatus := string(result.Status)
			_ = c.instanceRepo.UpdateHealthStatus(ctx, inst.ID, healthStatus)

			// If instance was starting up and is now healthy, mark as online
			if inst.Status == models.InstanceStatusStartingUp && result.Status == StatusHealthy {
				_ = c.instanceRepo.UpdateStatus(ctx, inst.ID, models.InstanceStatusOnline)
			}
		}(instance)
	}

	// Collect results
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	c.mu.Lock()
	for result := range resultsCh {
		c.results[result.InstanceID] = result
	}
	c.mu.Unlock()
}

// HealthSummary represents aggregated health statistics.
type HealthSummary struct {
	TotalInstances   int
	HealthyInstances int
	UnhealthyInstances int
	UnknownInstances int
	LastChecked      time.Time
}

// GetSummary returns aggregated health statistics.
func (c *Checker) GetSummary() *HealthSummary {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := &HealthSummary{
		TotalInstances: len(c.results),
	}

	var latestCheck time.Time
	for _, result := range c.results {
		switch result.Status {
		case StatusHealthy:
			summary.HealthyInstances++
		case StatusUnhealthy:
			summary.UnhealthyInstances++
		default:
			summary.UnknownInstances++
		}

		if result.CheckedAt.After(latestCheck) {
			latestCheck = result.CheckedAt
		}
	}

	summary.LastChecked = latestCheck
	return summary
}
