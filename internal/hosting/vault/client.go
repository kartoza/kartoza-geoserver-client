// Package vault provides HashiCorp Vault integration for credential management.
package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides access to HashiCorp Vault for managing credentials.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Config holds Vault connection configuration.
type Config struct {
	Address string
	Token   string
}

// NewClient creates a new Vault client.
func NewClient(config Config) *Client {
	return &Client{
		baseURL: config.Address,
		token:   config.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Credentials represents stored credentials for an instance.
type Credentials struct {
	URL           string            `json:"url,omitempty"`
	AdminUsername string            `json:"admin_username,omitempty"`
	AdminPassword string            `json:"admin_password,omitempty"`
	DatabaseHost  string            `json:"database_host,omitempty"`
	DatabasePort  int               `json:"database_port,omitempty"`
	DatabaseName  string            `json:"database_name,omitempty"`
	DatabaseUser  string            `json:"database_user,omitempty"`
	DatabasePass  string            `json:"database_pass,omitempty"`
	Extra         map[string]string `json:"extra,omitempty"`
}

// VaultResponse represents a generic Vault API response.
type VaultResponse struct {
	Data struct {
		Data     map[string]interface{} `json:"data"`
		Metadata struct {
			CreatedTime  string `json:"created_time"`
			Version      int    `json:"version"`
			DeletionTime string `json:"deletion_time"`
			Destroyed    bool   `json:"destroyed"`
		} `json:"metadata"`
	} `json:"data"`
}

// ReadCredentials reads credentials from a Vault path.
func (c *Client) ReadCredentials(ctx context.Context, path string) (*Credentials, error) {
	// Use KV v2 API
	url := fmt.Sprintf("%s/v1/secret/data/%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("credentials not found at path: %s", path)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Vault returned status %d: %s", resp.StatusCode, string(body))
	}

	var vaultResp VaultResponse
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse credentials from data
	creds := &Credentials{
		Extra: make(map[string]string),
	}

	if v, ok := vaultResp.Data.Data["url"].(string); ok {
		creds.URL = v
	}
	if v, ok := vaultResp.Data.Data["admin_username"].(string); ok {
		creds.AdminUsername = v
	}
	if v, ok := vaultResp.Data.Data["admin_password"].(string); ok {
		creds.AdminPassword = v
	}
	if v, ok := vaultResp.Data.Data["database_host"].(string); ok {
		creds.DatabaseHost = v
	}
	if v, ok := vaultResp.Data.Data["database_port"].(float64); ok {
		creds.DatabasePort = int(v)
	}
	if v, ok := vaultResp.Data.Data["database_name"].(string); ok {
		creds.DatabaseName = v
	}
	if v, ok := vaultResp.Data.Data["database_user"].(string); ok {
		creds.DatabaseUser = v
	}
	if v, ok := vaultResp.Data.Data["database_pass"].(string); ok {
		creds.DatabasePass = v
	}

	// Store any extra fields
	knownFields := map[string]bool{
		"url": true, "admin_username": true, "admin_password": true,
		"database_host": true, "database_port": true, "database_name": true,
		"database_user": true, "database_pass": true,
	}
	for k, v := range vaultResp.Data.Data {
		if !knownFields[k] {
			if s, ok := v.(string); ok {
				creds.Extra[k] = s
			}
		}
	}

	return creds, nil
}

// WriteCredentials writes credentials to a Vault path.
func (c *Client) WriteCredentials(ctx context.Context, path string, creds *Credentials) error {
	// Build data map
	data := make(map[string]interface{})
	if creds.URL != "" {
		data["url"] = creds.URL
	}
	if creds.AdminUsername != "" {
		data["admin_username"] = creds.AdminUsername
	}
	if creds.AdminPassword != "" {
		data["admin_password"] = creds.AdminPassword
	}
	if creds.DatabaseHost != "" {
		data["database_host"] = creds.DatabaseHost
	}
	if creds.DatabasePort != 0 {
		data["database_port"] = creds.DatabasePort
	}
	if creds.DatabaseName != "" {
		data["database_name"] = creds.DatabaseName
	}
	if creds.DatabaseUser != "" {
		data["database_user"] = creds.DatabaseUser
	}
	if creds.DatabasePass != "" {
		data["database_pass"] = creds.DatabasePass
	}
	for k, v := range creds.Extra {
		data[k] = v
	}

	// Wrap in KV v2 format
	payload := map[string]interface{}{
		"data": data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/v1/secret/data/%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Vault returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteCredentials deletes credentials at a Vault path.
func (c *Client) DeleteCredentials(ctx context.Context, path string) error {
	url := fmt.Sprintf("%s/v1/secret/data/%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Vault returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Health checks if Vault is healthy and the token is valid.
func (c *Client) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/sys/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Health endpoint returns 200 for healthy, 429/500+ for unhealthy
	if resp.StatusCode >= 500 {
		return fmt.Errorf("Vault is unhealthy: status %d", resp.StatusCode)
	}

	return nil
}