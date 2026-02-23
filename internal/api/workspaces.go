// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

// GetWorkspaces fetches all workspaces
func (c *Client) GetWorkspaces() ([]models.Workspace, error) {
	resp, err := c.doRequest("GET", "/workspaces", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get workspaces: %s", string(body))
	}

	// Read body to handle empty workspaces case
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// GeoServer returns {"workspaces": ""} when there are no workspaces
	// Check for this case first
	var emptyCheck struct {
		Workspaces interface{} `json:"workspaces"`
	}
	if err := json.Unmarshal(body, &emptyCheck); err != nil {
		return nil, fmt.Errorf("failed to decode workspaces: %w", err)
	}

	// If workspaces is a string (empty), return empty slice
	if _, ok := emptyCheck.Workspaces.(string); ok {
		return []models.Workspace{}, nil
	}

	// Otherwise parse normally
	var result struct {
		Workspaces struct {
			Workspace []models.Workspace `json:"workspace"`
		} `json:"workspaces"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode workspaces: %w", err)
	}

	return result.Workspaces.Workspace, nil
}

// CreateWorkspace creates a new workspace with just a name
func (c *Client) CreateWorkspace(name string) error {
	return c.CreateWorkspaceWithConfig(models.WorkspaceConfig{Name: name})
}

// CreateWorkspaceWithConfig creates a new workspace with full configuration
func (c *Client) CreateWorkspaceWithConfig(config models.WorkspaceConfig) error {
	body := map[string]interface{}{
		"workspace": map[string]interface{}{
			"name": config.Name,
		},
	}

	resp, err := c.doJSONRequest("POST", "/workspaces", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create workspace: %s", string(bodyBytes))
	}

	// If we have isolated set to true, set it via settings
	if config.Isolated {
		if err := c.UpdateWorkspaceSettings(config.Name, true); err != nil {
			return fmt.Errorf("failed to set isolated workspace: %w", err)
		}
	}

	return nil
}

// DeleteWorkspace deletes a workspace
func (c *Client) DeleteWorkspace(name string, recurse bool) error {
	path := fmt.Sprintf("/workspaces/%s?recurse=%t", name, recurse)
	resp, err := c.doRequest("DELETE", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete workspace: %s", string(body))
	}

	return nil
}

// UpdateWorkspace updates a workspace name
func (c *Client) UpdateWorkspace(oldName, newName string) error {
	body := map[string]interface{}{
		"workspace": map[string]interface{}{
			"name": newName,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s", oldName), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update workspace: %s", string(bodyBytes))
	}

	return nil
}

// GetWorkspaceConfig retrieves the full configuration for a workspace
func (c *Client) GetWorkspaceConfig(name string) (*models.WorkspaceConfig, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s", name), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get workspace config: %s", string(bodyBytes))
	}

	var result struct {
		Workspace struct {
			Name     string `json:"name"`
			Isolated bool   `json:"isolated"`
		} `json:"workspace"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode workspace config: %w", err)
	}

	return &models.WorkspaceConfig{
		Name:     result.Workspace.Name,
		Isolated: result.Workspace.Isolated,
	}, nil
}

// GetDefaultWorkspace retrieves the default workspace
func (c *Client) GetDefaultWorkspace() (string, error) {
	resp, err := c.doRequest("GET", "/workspaces/default", nil, "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get default workspace")
	}

	var result struct {
		Workspace struct {
			Name string `json:"name"`
		} `json:"workspace"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode default workspace: %w", err)
	}

	return result.Workspace.Name, nil
}

// SetDefaultWorkspace sets the default workspace
func (c *Client) SetDefaultWorkspace(name string) error {
	body := map[string]interface{}{
		"workspace": map[string]interface{}{
			"name": name,
		},
	}

	resp, err := c.doJSONRequest("PUT", "/workspaces/default", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set default workspace: %s", string(bodyBytes))
	}

	return nil
}

// isWorkspaceServiceEnabled checks if a service is enabled for a workspace
func (c *Client) isWorkspaceServiceEnabled(workspace, service string) bool {
	resp, err := c.doRequest("GET", fmt.Sprintf("/services/%s/workspaces/%s/settings", service, workspace), nil, "")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// UpdateWorkspaceSettings updates workspace isolation settings
func (c *Client) UpdateWorkspaceSettings(workspace string, enabled bool) error {
	body := map[string]interface{}{
		"workspace": map[string]interface{}{
			"isolated": enabled,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s", workspace), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update workspace settings: %s", string(bodyBytes))
	}

	return nil
}

// EnableWorkspaceService enables or disables a service for a workspace
func (c *Client) EnableWorkspaceService(workspace, service string, enabled bool) error {
	body := map[string]interface{}{
		service: map[string]interface{}{
			"enabled": enabled,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/services/%s/workspaces/%s/settings", service, workspace), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to enable workspace service: %s", string(bodyBytes))
	}

	return nil
}

// UpdateWorkspaceWithConfig updates a workspace with full configuration
func (c *Client) UpdateWorkspaceWithConfig(oldName string, config models.WorkspaceConfig) error {
	// First update the name if it changed
	if oldName != config.Name {
		if err := c.UpdateWorkspace(oldName, config.Name); err != nil {
			return err
		}
	}

	// Update isolation settings
	if err := c.UpdateWorkspaceSettings(config.Name, config.Isolated); err != nil {
		return err
	}

	return nil
}

// DownloadWorkspace downloads a workspace configuration as JSON
func (c *Client) DownloadWorkspace(name string) ([]byte, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s", name), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download workspace")
	}

	return io.ReadAll(resp.Body)
}
