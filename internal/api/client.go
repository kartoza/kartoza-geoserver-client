package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/config"
)

// fixEmptyGeoServerResponse handles GeoServer's quirk of returning "" instead of {} or []
// for empty collections. This converts patterns like {"dataStores": ""} to {"dataStores": {}}
func fixEmptyGeoServerResponse(body []byte) []byte {
	// Pattern matches: "someKey": "" where the value is an empty string
	// and converts it to "someKey": {} to allow proper unmarshaling
	re := regexp.MustCompile(`("(?:dataStores|coverageStores|wmsStores|wmtsStores|styles|layers|layerGroups|featureTypes|coverages|workspaces)")\s*:\s*""`)
	return re.ReplaceAll(body, []byte(`$1: {}`))
}

// Client is a GeoServer REST API client
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new GeoServer API client from a Connection
func NewClient(conn *config.Connection) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(conn.URL, "/"),
		username: conn.Username,
		password: conn.Password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientDirect creates a new GeoServer API client with direct parameters
func NewClientDirect(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BaseURL returns the base URL of the GeoServer
func (c *Client) BaseURL() string {
	return c.baseURL
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	url := c.baseURL + "/rest" + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// doJSONRequest performs a JSON request
func (c *Client) doJSONRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	return c.doRequest(method, path, bodyReader, "application/json")
}

// TestConnection tests if the connection is valid
func (c *Client) TestConnection() error {
	resp, err := c.doRequest("GET", "/about/version", nil, "")
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetServerStatus fetches the runtime status of the GeoServer instance
func (c *Client) GetServerStatus() (*ServerStatus, error) {
	startTime := time.Now()
	status := &ServerStatus{Online: false}

	// First check if server is reachable
	resp, err := c.doRequest("GET", "/about/status", nil, "")
	if err != nil {
		status.Error = fmt.Sprintf("Connection failed: %v", err)
		return status, nil
	}
	defer resp.Body.Close()

	status.ResponseTimeMs = time.Since(startTime).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		status.Error = fmt.Sprintf("Status check failed: %s", string(body))
		return status, nil
	}

	status.Online = true

	// Parse status response for memory info
	var statusResp struct {
		About struct {
			Status []struct {
				Name  string `json:"@name"`
				Value string `json:"value"`
			} `json:"status"`
		} `json:"about"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err == nil {
		for _, s := range statusResp.About.Status {
			switch s.Name {
			case "GEOSERVER_DATA_DIR":
				// data dir info
			}
		}
	}

	// Get memory info from system status endpoint
	memResp, err := c.doRequest("GET", "/about/system-status", nil, "")
	if err == nil {
		defer memResp.Body.Close()
		if memResp.StatusCode == http.StatusOK {
			var sysStatus struct {
				Metrics struct {
					Metric []struct {
						Name      string `json:"@name"`
						Available bool   `json:"available"`
						Value     string `json:"value"`
						Unit      string `json:"unit,omitempty"`
					} `json:"metric"`
				} `json:"metrics"`
			}
			if err := json.NewDecoder(memResp.Body).Decode(&sysStatus); err == nil {
				for _, m := range sysStatus.Metrics.Metric {
					if !m.Available {
						continue
					}
					switch m.Name {
					case "MEMORY_USED":
						fmt.Sscanf(m.Value, "%d", &status.MemoryUsed)
					case "MEMORY_FREE":
						fmt.Sscanf(m.Value, "%d", &status.MemoryFree)
					case "MEMORY_TOTAL":
						fmt.Sscanf(m.Value, "%d", &status.MemoryTotal)
					case "CPU_LOAD":
						fmt.Sscanf(m.Value, "%f", &status.CPULoad)
					}
				}
				if status.MemoryTotal > 0 {
					status.MemoryUsedPct = float64(status.MemoryUsed) / float64(status.MemoryTotal) * 100
				}
			}
		}
	}

	// Get version info
	if info, err := c.GetServerInfo(); err == nil {
		status.GeoServerVersion = info.GeoServerVersion
	}

	return status, nil
}

// GetServerCounts fetches counts of various resources
func (c *Client) GetServerCounts() (*ServerStatus, error) {
	status := &ServerStatus{Online: true}

	// Count workspaces
	if workspaces, err := c.GetWorkspaces(); err == nil {
		status.WorkspaceCount = len(workspaces)

		// Count layers, stores across all workspaces
		for _, ws := range workspaces {
			if layers, err := c.GetLayers(ws.Name); err == nil {
				status.LayerCount += len(layers)
			}
			if stores, err := c.GetDataStores(ws.Name); err == nil {
				status.DataStoreCount += len(stores)
			}
			if coverages, err := c.GetCoverageStores(ws.Name); err == nil {
				status.CoverageCount += len(coverages)
			}
		}
	}

	// Count global styles
	if styles, err := c.GetStyles(""); err == nil {
		status.StyleCount = len(styles)
	}

	return status, nil
}

// GetFullServerStatus gets both runtime status and resource counts
func (c *Client) GetFullServerStatus() (*ServerStatus, error) {
	status, err := c.GetServerStatus()
	if err != nil || !status.Online {
		return status, err
	}

	// Get resource counts
	counts, _ := c.GetServerCounts()
	if counts != nil {
		status.WorkspaceCount = counts.WorkspaceCount
		status.LayerCount = counts.LayerCount
		status.DataStoreCount = counts.DataStoreCount
		status.CoverageCount = counts.CoverageCount
		status.StyleCount = counts.StyleCount
	}

	return status, nil
}

// GetServerInfo fetches information about the GeoServer instance
func (c *Client) GetServerInfo() (*ServerInfo, error) {
	resp, err := c.doRequest("GET", "/about/version", nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get server info: %s", string(body))
	}

	var response struct {
		About struct {
			Resource []struct {
				Name           string      `json:"@name"`
				Version        interface{} `json:"Version"`
				BuildTimestamp string      `json:"Build-Timestamp"`
				GitRevision    string      `json:"Git-Revision"`
			} `json:"resource"`
		} `json:"about"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode server info: %w", err)
	}

	info := &ServerInfo{}
	for _, r := range response.About.Resource {
		switch r.Name {
		case "GeoServer":
			info.GeoServerVersion = fmt.Sprintf("%v", r.Version)
			info.GeoServerBuild = r.BuildTimestamp
			info.GeoServerRevision = r.GitRevision
		case "GeoTools":
			info.GeoToolsVersion = fmt.Sprintf("%v", r.Version)
		case "GeoWebCache":
			info.GeoWebCacheVersion = fmt.Sprintf("%v", r.Version)
		}
	}

	return info, nil
}

// ReloadConfiguration reloads the GeoServer configuration
func (c *Client) ReloadConfiguration() error {
	resp, err := c.doRequest("POST", "/reload", nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to reload configuration: %s", string(body))
	}

	return nil
}

// GetServerVersion returns just the GeoServer version string
func (c *Client) GetServerVersion() (string, error) {
	info, err := c.GetServerInfo()
	if err != nil {
		return "", err
	}
	return info.GeoServerVersion, nil
}
