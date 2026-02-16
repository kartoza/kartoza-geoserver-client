package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

func (c *Client) GetCoverageStores(workspace string) ([]models.CoverageStore, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/coveragestores", workspace), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get coverage stores: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Fix GeoServer's empty string response quirk
	body = fixEmptyGeoServerResponse(body)

	var result struct {
		CoverageStores struct {
			CoverageStore []models.CoverageStore `json:"coverageStore"`
		} `json:"coverageStores"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode coverage stores: %w", err)
	}

	// Fetch enabled status for each store (list endpoint doesn't include it)
	for i := range result.CoverageStores.CoverageStore {
		result.CoverageStores.CoverageStore[i].Workspace = workspace
		// Get individual store to fetch enabled status
		storeResp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/coveragestores/%s", workspace, result.CoverageStores.CoverageStore[i].Name), nil, "")
		if err == nil {
			defer storeResp.Body.Close()
			if storeResp.StatusCode == http.StatusOK {
				var storeDetail struct {
					CoverageStore struct {
						Enabled bool `json:"enabled"`
					} `json:"coverageStore"`
				}
				if json.NewDecoder(storeResp.Body).Decode(&storeDetail) == nil {
					result.CoverageStores.CoverageStore[i].Enabled = storeDetail.CoverageStore.Enabled
				}
			}
		}
	}

	return result.CoverageStores.CoverageStore, nil
}

func (c *Client) DeleteCoverageStore(workspace, name string, recurse bool) error {
	path := fmt.Sprintf("/workspaces/%s/coveragestores/%s", workspace, name)
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
		return fmt.Errorf("failed to delete coverage store: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) CreateCoverageStore(workspace string, name string, storeType models.CoverageStoreType, url string) error {
	body := map[string]interface{}{
		"coverageStore": map[string]interface{}{
			"name":    name,
			"type":    storeType.Type(),
			"enabled": true,
			"url":     url,
		},
	}

	resp, err := c.doJSONRequest("POST", fmt.Sprintf("/workspaces/%s/coveragestores", workspace), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create coverage store: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) UpdateCoverageStore(workspace, oldName, newName string) error {
	body := map[string]interface{}{
		"coverageStore": map[string]string{
			"name": newName,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s/coveragestores/%s", workspace, oldName), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update coverage store: %s", string(bodyBytes))
	}

	return nil
}

// PublishCoverage publishes a coverage from a coverage store
func (c *Client) PublishCoverage(workspace, coverageStore, coverageName string) error {
	// First, get the available coverages in the store to find the native name
	// GeoServer needs us to specify the coverage details
	body := map[string]interface{}{
		"coverage": map[string]interface{}{
			"name":       coverageName,
			"nativeName": coverageName,
			"title":      coverageName,
			"enabled":    true,
			"advertised": true,
		},
	}

	resp, err := c.doJSONRequest("POST", fmt.Sprintf("/workspaces/%s/coveragestores/%s/coverages", workspace, coverageStore), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to publish coverage: %s", string(bodyBytes))
	}

	return nil
}

// PublishFeatureType publishes a feature type from a data store
func (c *Client) GetCoverageStoreDetails(workspace, storeName string) (*CoverageStoreDetails, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/coveragestores/%s", workspace, storeName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get coverage store details: %s", string(bodyBytes))
	}

	var storeResult struct {
		CoverageStore struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
			Enabled     bool   `json:"enabled"`
			URL         string `json:"url"`
		} `json:"coverageStore"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&storeResult); err != nil {
		return nil, fmt.Errorf("failed to decode coverage store details: %w", err)
	}

	return &CoverageStoreDetails{
		Name:        storeResult.CoverageStore.Name,
		Description: storeResult.CoverageStore.Description,
		Type:        storeResult.CoverageStore.Type,
		Enabled:     storeResult.CoverageStore.Enabled,
		URL:         storeResult.CoverageStore.URL,
	}, nil
}

func (c *Client) CreateCoverageStoreFromDetails(workspace string, details *CoverageStoreDetails) error {
	body := map[string]interface{}{
		"coverageStore": map[string]interface{}{
			"name":        details.Name,
			"description": details.Description,
			"type":        details.Type,
			"enabled":     details.Enabled,
			"url":         details.URL,
		},
	}

	resp, err := c.doJSONRequest("POST", fmt.Sprintf("/workspaces/%s/coveragestores", workspace), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create coverage store: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) GetCoverageStoreConfig(workspace, storeName string) (*models.CoverageStoreConfig, error) {
	config := &models.CoverageStoreConfig{
		Name:      storeName,
		Workspace: workspace,
	}

	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/coveragestores/%s", workspace, storeName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get coverage store: %s", string(bodyBytes))
	}

	var storeResult struct {
		CoverageStore struct {
			Name        string `json:"name"`
			Enabled     bool   `json:"enabled"`
			Description string `json:"description"`
		} `json:"coverageStore"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&storeResult); err != nil {
		return nil, fmt.Errorf("failed to decode coverage store: %w", err)
	}

	config.Enabled = storeResult.CoverageStore.Enabled
	config.Description = storeResult.CoverageStore.Description

	return config, nil
}

func (c *Client) UpdateCoverageStoreConfig(workspace string, config models.CoverageStoreConfig) error {
	body := map[string]interface{}{
		"coverageStore": map[string]interface{}{
			"name":        config.Name,
			"enabled":     config.Enabled,
			"description": config.Description,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s/coveragestores/%s", workspace, config.Name), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update coverage store: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) GetLayersForCoverageStore(workspace, storeName string) ([]string, error) {
	coverages, err := c.GetCoverages(workspace, storeName)
	if err != nil {
		return nil, err
	}

	var layerNames []string
	for _, cov := range coverages {
		layerNames = append(layerNames, fmt.Sprintf("%s:%s", workspace, cov.Name))
	}
	return layerNames, nil
}

func (c *Client) DeleteCoverageStoreWithCleanup(workspace, name string, recurse bool) error {
	// If recursing, first clean up GWC caches for all associated layers
	if recurse {
		layerNames, err := c.GetLayersForCoverageStore(workspace, name)
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

	// Now delete the coverage store
	return c.DeleteCoverageStore(workspace, name, recurse)
}

func (c *Client) DownloadCoverageStore(workspace, name string) ([]byte, error) {
	path := fmt.Sprintf("/workspaces/%s/coveragestores/%s", workspace, name)
	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coverage store not found: %s/%s", workspace, name)
	}

	return io.ReadAll(resp.Body)
}

