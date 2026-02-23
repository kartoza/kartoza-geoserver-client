package iceberg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client wraps the Iceberg REST Catalog API
type Client struct {
	baseURL    string
	httpClient *http.Client
	prefix     string // Optional warehouse prefix
}

// ClientConfig holds configuration for creating an Iceberg client
type ClientConfig struct {
	BaseURL string
	Prefix  string // Optional warehouse prefix (default: empty)
	Timeout time.Duration
}

// NewClient creates a new Iceberg REST Catalog client
func NewClient(config ClientConfig) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	// Ensure baseURL doesn't end with /
	config.BaseURL = strings.TrimSuffix(config.BaseURL, "/")

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		prefix: config.Prefix,
	}, nil
}

// buildURL constructs the API URL with the given path
func (c *Client) buildURL(path string) string {
	// Iceberg REST API prefix is /v1/{prefix}/...
	if c.prefix != "" {
		return fmt.Sprintf("%s/v1/%s%s", c.baseURL, c.prefix, path)
	}
	return fmt.Sprintf("%s/v1%s", c.baseURL, path)
}

// doRequest executes an HTTP request and returns the response body
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("API error: %s (type: %s, code: %d)",
				errResp.Error.Message, errResp.Error.Type, errResp.Error.Code)
		}
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetConfig retrieves the catalog configuration
func (c *Client) GetConfig(ctx context.Context) (*CatalogConfig, error) {
	// Config endpoint is at /v1/config (no prefix)
	resp, err := c.httpClient.Get(c.baseURL + "/v1/config")
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var config CatalogConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// TestConnection tests the connection to the Iceberg catalog
func (c *Client) TestConnection(ctx context.Context) (*ConnectionTestResult, error) {
	config, err := c.GetConfig(ctx)
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %v", err),
		}, err
	}

	// Try to list namespaces
	namespaces, err := c.ListNamespaces(ctx, "")
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Connection successful but failed to list namespaces: %v", err),
		}, nil
	}

	return &ConnectionTestResult{
		Success:        true,
		Message:        fmt.Sprintf("Connected successfully. Found %d namespace(s).", len(namespaces)),
		NamespaceCount: len(namespaces),
		Defaults:       config.Defaults,
	}, nil
}

// ListNamespaces returns all namespaces, optionally filtered by parent
func (c *Client) ListNamespaces(ctx context.Context, parent string) ([]Namespace, error) {
	path := "/namespaces"
	if parent != "" {
		path = fmt.Sprintf("/namespaces?parent=%s", url.QueryEscape(parent))
	}

	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response NamespacesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse namespaces: %w", err)
	}

	namespaces := make([]Namespace, len(response.Namespaces))
	for i, ns := range response.Namespaces {
		namespaces[i] = Namespace{
			Name: strings.Join(ns, "."),
			Path: ns,
		}
	}

	return namespaces, nil
}

// GetNamespace retrieves a specific namespace
func (c *Client) GetNamespace(ctx context.Context, namespace string) (*NamespaceInfo, error) {
	path := fmt.Sprintf("/namespaces/%s", url.PathEscape(namespace))

	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var info NamespaceInfo
	if err := json.Unmarshal(respBody, &info); err != nil {
		return nil, fmt.Errorf("failed to parse namespace info: %w", err)
	}

	return &info, nil
}

// CreateNamespace creates a new namespace
func (c *Client) CreateNamespace(ctx context.Context, namespace string, properties map[string]string) error {
	parts := strings.Split(namespace, ".")
	body := CreateNamespaceRequest{
		Namespace:  parts,
		Properties: properties,
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/namespaces", body)
	return err
}

// DropNamespace deletes a namespace (must be empty)
func (c *Client) DropNamespace(ctx context.Context, namespace string) error {
	path := fmt.Sprintf("/namespaces/%s", url.PathEscape(namespace))
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// ListTables returns all tables in a namespace
func (c *Client) ListTables(ctx context.Context, namespace string) ([]TableIdentifier, error) {
	path := fmt.Sprintf("/namespaces/%s/tables", url.PathEscape(namespace))

	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response TablesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse tables: %w", err)
	}

	return response.Identifiers, nil
}

// GetTable retrieves metadata for a specific table
func (c *Client) GetTable(ctx context.Context, namespace, table string) (*TableMetadata, error) {
	path := fmt.Sprintf("/namespaces/%s/tables/%s", url.PathEscape(namespace), url.PathEscape(table))

	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var metadata TableMetadata
	if err := json.Unmarshal(respBody, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse table metadata: %w", err)
	}

	return &metadata, nil
}

// CreateTable creates a new table
func (c *Client) CreateTable(ctx context.Context, namespace string, request CreateTableRequest) (*TableMetadata, error) {
	path := fmt.Sprintf("/namespaces/%s/tables", url.PathEscape(namespace))

	respBody, err := c.doRequest(ctx, http.MethodPost, path, request)
	if err != nil {
		return nil, err
	}

	var metadata TableMetadata
	if err := json.Unmarshal(respBody, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse created table: %w", err)
	}

	return &metadata, nil
}

// DropTable deletes a table
func (c *Client) DropTable(ctx context.Context, namespace, table string, purge bool) error {
	path := fmt.Sprintf("/namespaces/%s/tables/%s", url.PathEscape(namespace), url.PathEscape(table))
	if purge {
		path += "?purgeRequested=true"
	}

	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// TableExists checks if a table exists
func (c *Client) TableExists(ctx context.Context, namespace, table string) (bool, error) {
	path := fmt.Sprintf("/namespaces/%s/tables/%s", url.PathEscape(namespace), url.PathEscape(table))

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.buildURL(path), nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

// RenameTable renames a table
func (c *Client) RenameTable(ctx context.Context, srcNamespace, srcTable, dstNamespace, dstTable string) error {
	body := RenameTableRequest{
		Source: TableIdentifier{
			Namespace: []string{srcNamespace},
			Name:      srcTable,
		},
		Destination: TableIdentifier{
			Namespace: []string{dstNamespace},
			Name:      dstTable,
		},
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/tables/rename", body)
	return err
}

// GetTableSnapshots returns the snapshot history for a table
func (c *Client) GetTableSnapshots(ctx context.Context, namespace, table string) ([]Snapshot, error) {
	metadata, err := c.GetTable(ctx, namespace, table)
	if err != nil {
		return nil, err
	}

	return metadata.Metadata.Snapshots, nil
}
