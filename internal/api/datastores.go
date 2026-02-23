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

func (c *Client) GetDataStores(workspace string) ([]models.DataStore, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/datastores", workspace), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get datastores: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Fix GeoServer's empty string response quirk
	body = fixEmptyGeoServerResponse(body)

	var result struct {
		DataStores struct {
			DataStore []models.DataStore `json:"dataStore"`
		} `json:"dataStores"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode datastores: %w", err)
	}

	// Fetch enabled status for each store (list endpoint doesn't include it)
	for i := range result.DataStores.DataStore {
		result.DataStores.DataStore[i].Workspace = workspace
		// Get individual store to fetch enabled status
		storeResp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/datastores/%s", workspace, result.DataStores.DataStore[i].Name), nil, "")
		if err == nil {
			defer storeResp.Body.Close()
			if storeResp.StatusCode == http.StatusOK {
				var storeDetail struct {
					DataStore struct {
						Enabled bool `json:"enabled"`
					} `json:"dataStore"`
				}
				if json.NewDecoder(storeResp.Body).Decode(&storeDetail) == nil {
					result.DataStores.DataStore[i].Enabled = storeDetail.DataStore.Enabled
				}
			}
		}
	}

	return result.DataStores.DataStore, nil
}

// GetCoverageStores fetches all coverage stores for a workspace
func (c *Client) DeleteDataStore(workspace, name string, recurse bool) error {
	path := fmt.Sprintf("/workspaces/%s/datastores/%s", workspace, name)
	if recurse {
		path += "?recurse=true"
	}

	resp, err := c.doRequest("DELETE", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete datastore: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) CreateDataStore(workspace string, name string, storeType models.DataStoreType, params map[string]string) error {
	// Build connection parameters in the format GeoServer expects
	// GeoServer JSON format uses "entry" array with "@key" and "$" fields
	entries := make([]map[string]string, 0)

	// Add dbtype for database stores
	if storeType == models.DataStoreTypePostGIS || storeType == models.DataStoreTypeGeoPackage {
		entries = append(entries, map[string]string{
			"@key": "dbtype",
			"$":    storeType.DBType(),
		})
	}

	// Add all provided parameters
	for key, value := range params {
		if key != "name" && value != "" { // Skip name as it's set separately
			entries = append(entries, map[string]string{
				"@key": key,
				"$":    value,
			})
		}
	}

	body := map[string]interface{}{
		"dataStore": map[string]interface{}{
			"name":    name,
			"enabled": true,
			"connectionParameters": map[string]interface{}{
				"entry": entries,
			},
		},
	}

	resp, err := c.doJSONRequest("POST", fmt.Sprintf("/workspaces/%s/datastores", workspace), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create datastore: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) UpdateDataStore(workspace, oldName, newName string) error {
	body := map[string]interface{}{
		"dataStore": map[string]string{
			"name": newName,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s/datastores/%s", workspace, oldName), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update datastore: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) GetDataStoreConfig(workspace, storeName string) (*models.DataStoreConfig, error) {
	config := &models.DataStoreConfig{
		Name:      storeName,
		Workspace: workspace,
	}

	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/datastores/%s", workspace, storeName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get data store: %s", string(bodyBytes))
	}

	var storeResult struct {
		DataStore struct {
			Name        string `json:"name"`
			Enabled     bool   `json:"enabled"`
			Description string `json:"description"`
		} `json:"dataStore"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&storeResult); err != nil {
		return nil, fmt.Errorf("failed to decode data store: %w", err)
	}

	config.Enabled = storeResult.DataStore.Enabled
	config.Description = storeResult.DataStore.Description

	return config, nil
}

func (c *Client) UpdateDataStoreConfig(workspace string, config models.DataStoreConfig) error {
	body := map[string]interface{}{
		"dataStore": map[string]interface{}{
			"name":        config.Name,
			"enabled":     config.Enabled,
			"description": config.Description,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s/datastores/%s", workspace, config.Name), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update data store: %s", string(bodyBytes))
	}

	return nil
}
func (c *Client) GetDataStoreDetails(workspace, storeName string) (*DataStoreDetails, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/datastores/%s", workspace, storeName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get data store details: %s", string(bodyBytes))
	}

	var storeResult struct {
		DataStore struct {
			Name                 string `json:"name"`
			Description          string `json:"description"`
			Type                 string `json:"type"`
			Enabled              bool   `json:"enabled"`
			ConnectionParameters struct {
				Entry []struct {
					Key   string `json:"@key"`
					Value string `json:"$"`
				} `json:"entry"`
			} `json:"connectionParameters"`
		} `json:"dataStore"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&storeResult); err != nil {
		return nil, fmt.Errorf("failed to decode data store details: %w", err)
	}

	details := &DataStoreDetails{
		Name:                 storeResult.DataStore.Name,
		Description:          storeResult.DataStore.Description,
		Type:                 storeResult.DataStore.Type,
		Enabled:              storeResult.DataStore.Enabled,
		ConnectionParameters: make(map[string]string),
	}

	// Convert entry array to map
	for _, entry := range storeResult.DataStore.ConnectionParameters.Entry {
		details.ConnectionParameters[entry.Key] = entry.Value
	}

	return details, nil
}

func (c *Client) CreateDataStoreFromDetails(workspace string, details *DataStoreDetails) error {
	// Convert connection parameters map back to entry array format
	entries := make([]map[string]string, 0, len(details.ConnectionParameters))
	for key, value := range details.ConnectionParameters {
		entries = append(entries, map[string]string{
			"@key": key,
			"$":    value,
		})
	}

	body := map[string]interface{}{
		"dataStore": map[string]interface{}{
			"name":        details.Name,
			"description": details.Description,
			"type":        details.Type,
			"enabled":     details.Enabled,
			"connectionParameters": map[string]interface{}{
				"entry": entries,
			},
		},
	}

	resp, err := c.doJSONRequest("POST", fmt.Sprintf("/workspaces/%s/datastores", workspace), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create data store: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) GetLayersForDataStore(workspace, storeName string) ([]string, error) {
	featureTypes, err := c.GetFeatureTypes(workspace, storeName)
	if err != nil {
		return nil, err
	}

	var layerNames []string
	for _, ft := range featureTypes {
		layerNames = append(layerNames, fmt.Sprintf("%s:%s", workspace, ft.Name))
	}
	return layerNames, nil
}

func (c *Client) DeleteDataStoreWithCleanup(workspace, name string, recurse bool) error {
	// If recursing, first clean up GWC caches for all associated layers
	if recurse {
		layerNames, err := c.GetLayersForDataStore(workspace, name)
		if err != nil {
			// Log but don't fail - the layers might not exist or be accessible
			fmt.Printf("Warning: could not get layers for cleanup: %v\n", err)
		} else {
			for _, layerName := range layerNames {
				// Truncate and delete GWC entries - ignore errors as cache might not exist
				_ = c.TruncateAndDeleteGWCLayer(layerName)
			}
		}
	}

	// Now delete the data store
	return c.DeleteDataStore(workspace, name, recurse)
}

func (c *Client) DownloadDataStore(workspace, name string) ([]byte, error) {
	path := fmt.Sprintf("/workspaces/%s/datastores/%s", workspace, name)
	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("data store not found: %s/%s", workspace, name)
	}

	return io.ReadAll(resp.Body)
}
