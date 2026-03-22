// Package deploy provides deployment functionality for hosted instances.
package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// JenkinsClient provides access to Jenkins API for deployment jobs.
type JenkinsClient struct {
	baseURL    string
	username   string
	token      string
	httpClient *http.Client
}

// JenkinsConfig holds Jenkins connection configuration.
type JenkinsConfig struct {
	URL      string
	Username string
	Token    string
}

// NewJenkinsClient creates a new Jenkins client.
func NewJenkinsClient(config JenkinsConfig) *JenkinsClient {
	return &JenkinsClient{
		baseURL:  config.URL,
		username: config.Username,
		token:    config.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TriggerBuildParams represents parameters for triggering a Jenkins build.
type TriggerBuildParams struct {
	JobName    string
	Parameters map[string]string
}

// BuildInfo represents information about a Jenkins build.
type BuildInfo struct {
	Number    int    `json:"number"`
	URL       string `json:"url"`
	Building  bool   `json:"building"`
	Result    string `json:"result"`
	Duration  int64  `json:"duration"`
	Timestamp int64  `json:"timestamp"`
}

// QueueItem represents a Jenkins queue item.
type QueueItem struct {
	ID         int    `json:"id"`
	URL        string `json:"url"`
	Executable *struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	} `json:"executable"`
	Blocked bool   `json:"blocked"`
	Why     string `json:"why"`
}

// TriggerBuild triggers a Jenkins build and returns the queue item ID.
func (c *JenkinsClient) TriggerBuild(ctx context.Context, params TriggerBuildParams) (int, error) {
	// Build URL with parameters
	buildURL := fmt.Sprintf("%s/job/%s/buildWithParameters", c.baseURL, url.PathEscape(params.JobName))

	// Encode parameters
	formData := url.Values{}
	for k, v := range params.Parameters {
		formData.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", buildURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("Jenkins returned status %d: %s", resp.StatusCode, string(body))
	}

	// Get queue item from Location header
	location := resp.Header.Get("Location")
	if location == "" {
		return 0, fmt.Errorf("no Location header in response")
	}

	// Parse queue ID from location
	var queueID int
	_, err = fmt.Sscanf(location, "%s/queue/item/%d/", new(string), &queueID)
	if err != nil {
		// Try alternative format
		_, err = fmt.Sscanf(location, c.baseURL+"/queue/item/%d/", &queueID)
		if err != nil {
			return 0, fmt.Errorf("failed to parse queue ID from location: %s", location)
		}
	}

	return queueID, nil
}

// GetQueueItem gets information about a queue item.
func (c *JenkinsClient) GetQueueItem(ctx context.Context, queueID int) (*QueueItem, error) {
	queueURL := fmt.Sprintf("%s/queue/item/%d/api/json", c.baseURL, queueID)

	req, err := http.NewRequestWithContext(ctx, "GET", queueURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jenkins returned status %d", resp.StatusCode)
	}

	var item QueueItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &item, nil
}

// GetBuildInfo gets information about a specific build.
func (c *JenkinsClient) GetBuildInfo(ctx context.Context, jobName string, buildNumber int) (*BuildInfo, error) {
	buildURL := fmt.Sprintf("%s/job/%s/%d/api/json", c.baseURL, url.PathEscape(jobName), buildNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", buildURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jenkins returned status %d", resp.StatusCode)
	}

	var info BuildInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &info, nil
}

// WaitForBuild waits for a queued build to start and returns the build number.
func (c *JenkinsClient) WaitForBuild(ctx context.Context, queueID int, timeout time.Duration) (int, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		item, err := c.GetQueueItem(ctx, queueID)
		if err != nil {
			return 0, err
		}

		if item.Executable != nil {
			return item.Executable.Number, nil
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(2 * time.Second):
			continue
		}
	}

	return 0, fmt.Errorf("timeout waiting for build to start")
}

// WaitForBuildCompletion waits for a build to complete.
func (c *JenkinsClient) WaitForBuildCompletion(ctx context.Context, jobName string, buildNumber int, timeout time.Duration) (*BuildInfo, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		info, err := c.GetBuildInfo(ctx, jobName, buildNumber)
		if err != nil {
			return nil, err
		}

		if !info.Building {
			return info, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			continue
		}
	}

	return nil, fmt.Errorf("timeout waiting for build completion")
}
