// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

func (c *Client) GetLayerGroups(workspace string) ([]models.LayerGroup, error) {
	var path string
	if workspace == "" {
		path = "/layergroups"
	} else {
		path = fmt.Sprintf("/workspaces/%s/layergroups", workspace)
	}

	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get layer groups: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Fix GeoServer's empty string response quirk
	body = fixEmptyGeoServerResponse(body)

	var result struct {
		LayerGroups struct {
			LayerGroup []models.LayerGroup `json:"layerGroup"`
		} `json:"layerGroups"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode layer groups: %w", err)
	}

	for i := range result.LayerGroups.LayerGroup {
		result.LayerGroups.LayerGroup[i].Workspace = workspace
	}

	return result.LayerGroups.LayerGroup, nil
}

func (c *Client) CreateLayerGroup(workspace string, config models.LayerGroupCreate) error {
	path := fmt.Sprintf("/workspaces/%s/layergroups", workspace)

	// Build a map of layer styles for quick lookup
	layerStyleMap := make(map[string]string)
	for _, ls := range config.LayerStyles {
		if ls.StyleName != "" {
			layerStyleMap[ls.LayerName] = ls.StyleName
		}
	}

	// Build the publishables array with layer references
	publishables := make([]map[string]interface{}, len(config.Layers))
	styles := make([]map[string]interface{}, len(config.Layers))

	for i, layerName := range config.Layers {
		// Layer names should be in workspace:layer format
		publishables[i] = map[string]interface{}{
			"@type": "layer",
			"name":  layerName,
		}
		// Use assigned style if available, otherwise empty (default)
		if styleName, ok := layerStyleMap[layerName]; ok && styleName != "" {
			styles[i] = map[string]interface{}{
				"name": styleName,
			}
		} else {
			styles[i] = map[string]interface{}{}
		}
	}

	mode := config.Mode
	if mode == "" {
		mode = "SINGLE"
	}

	body := map[string]interface{}{
		"layerGroup": map[string]interface{}{
			"name": config.Name,
			"mode": mode,
			"publishables": map[string]interface{}{
				"published": publishables,
			},
			"styles": map[string]interface{}{
				"style": styles,
			},
		},
	}

	if config.Title != "" {
		body["layerGroup"].(map[string]interface{})["title"] = config.Title
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal layer group: %w", err)
	}

	resp, err := c.doRequest("POST", path, bytes.NewReader(jsonBody), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create layer group (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) GetLayerGroup(workspace, name string) (*models.LayerGroupDetails, error) {
	var path string
	if workspace == "" {
		path = fmt.Sprintf("/layergroups/%s", name)
	} else {
		path = fmt.Sprintf("/workspaces/%s/layergroups/%s", workspace, name)
	}

	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get layer group: %s", string(bodyBytes))
	}

	var result struct {
		LayerGroup struct {
			Name      string `json:"name"`
			Mode      string `json:"mode"`
			Title     string `json:"title"`
			Abstract  string `json:"abstractTxt"`
			Workspace struct {
				Name string `json:"name"`
			} `json:"workspace"`
			Publishables struct {
				Published json.RawMessage `json:"published"`
			} `json:"publishables"`
			Styles struct {
				Style json.RawMessage `json:"style"`
			} `json:"styles"`
			Bounds struct {
				MinX float64 `json:"minx"`
				MinY float64 `json:"miny"`
				MaxX float64 `json:"maxx"`
				MaxY float64 `json:"maxy"`
				CRS  string  `json:"crs"`
			} `json:"bounds"`
		} `json:"layerGroup"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode layer group: %w", err)
	}

	details := &models.LayerGroupDetails{
		Name:       result.LayerGroup.Name,
		Mode:       result.LayerGroup.Mode,
		Title:      result.LayerGroup.Title,
		Abstract:   result.LayerGroup.Abstract,
		Workspace:  result.LayerGroup.Workspace.Name,
		Enabled:    true, // Layer groups are enabled by default
		Advertised: true,
	}

	// Parse publishables - can be single object or array
	type publishedItem struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	}
	var publishedItems []publishedItem
	if len(result.LayerGroup.Publishables.Published) > 0 {
		// Try array first
		if err := json.Unmarshal(result.LayerGroup.Publishables.Published, &publishedItems); err != nil {
			// Try single object
			var single publishedItem
			if err := json.Unmarshal(result.LayerGroup.Publishables.Published, &single); err == nil {
				publishedItems = []publishedItem{single}
			}
		}
	}

	// Parse styles - can be single string, single object, array of strings, or array of objects
	type styleItem struct {
		Name string `json:"name"`
	}
	var styleNames []string
	if len(result.LayerGroup.Styles.Style) > 0 {
		// Try array of objects first
		var styleObjects []styleItem
		if err := json.Unmarshal(result.LayerGroup.Styles.Style, &styleObjects); err == nil {
			for _, s := range styleObjects {
				styleNames = append(styleNames, s.Name)
			}
		} else {
			// Try array of strings
			var styleStrings []string
			if err := json.Unmarshal(result.LayerGroup.Styles.Style, &styleStrings); err == nil {
				styleNames = styleStrings
			} else {
				// Try single object
				var singleObj styleItem
				if err := json.Unmarshal(result.LayerGroup.Styles.Style, &singleObj); err == nil {
					styleNames = []string{singleObj.Name}
				} else {
					// Try single string
					var singleStr string
					if err := json.Unmarshal(result.LayerGroup.Styles.Style, &singleStr); err == nil {
						styleNames = []string{singleStr}
					}
				}
			}
		}
	}

	// Build layers list
	for i, pub := range publishedItems {
		item := models.LayerGroupItem{
			Type: pub.Type,
			Name: pub.Name,
		}
		// Match with style if available
		if i < len(styleNames) {
			item.StyleName = styleNames[i]
		}
		details.Layers = append(details.Layers, item)
	}

	// Parse bounds
	if result.LayerGroup.Bounds.CRS != "" {
		details.Bounds = &models.Bounds{
			MinX: result.LayerGroup.Bounds.MinX,
			MinY: result.LayerGroup.Bounds.MinY,
			MaxX: result.LayerGroup.Bounds.MaxX,
			MaxY: result.LayerGroup.Bounds.MaxY,
			CRS:  result.LayerGroup.Bounds.CRS,
		}
	}

	return details, nil
}

func (c *Client) UpdateLayerGroup(workspace, name string, update models.LayerGroupUpdate) error {
	var path string
	if workspace == "" {
		path = fmt.Sprintf("/layergroups/%s", name)
	} else {
		path = fmt.Sprintf("/workspaces/%s/layergroups/%s", workspace, name)
	}

	// Build a map of layer styles for quick lookup
	layerStyleMap := make(map[string]string)
	for _, ls := range update.LayerStyles {
		if ls.StyleName != "" {
			layerStyleMap[ls.LayerName] = ls.StyleName
		}
	}

	// Build the publishables array with layer references
	publishables := make([]map[string]interface{}, len(update.Layers))
	styles := make([]map[string]interface{}, len(update.Layers))

	for i, layerName := range update.Layers {
		publishables[i] = map[string]interface{}{
			"@type": "layer",
			"name":  layerName,
		}
		// Use assigned style if available, otherwise empty (default)
		if styleName, ok := layerStyleMap[layerName]; ok && styleName != "" {
			styles[i] = map[string]interface{}{
				"name": styleName,
			}
		} else {
			styles[i] = map[string]interface{}{}
		}
	}

	body := map[string]interface{}{
		"layerGroup": map[string]interface{}{
			"mode": update.Mode,
		},
	}

	lg := body["layerGroup"].(map[string]interface{})

	if update.Title != "" {
		lg["title"] = update.Title
	}

	if len(update.Layers) > 0 {
		lg["publishables"] = map[string]interface{}{
			"published": publishables,
		}
		lg["styles"] = map[string]interface{}{
			"style": styles,
		}
	}

	resp, err := c.doJSONRequest("PUT", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update layer group: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) DeleteLayerGroup(workspace, name string) error {
	var path string
	if workspace == "" {
		path = fmt.Sprintf("/layergroups/%s", name)
	} else {
		path = fmt.Sprintf("/workspaces/%s/layergroups/%s", workspace, name)
	}

	resp, err := c.doRequest("DELETE", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete layer group: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) DownloadLayerGroup(workspace, name string) ([]byte, error) {
	var path string
	if workspace == "" {
		path = fmt.Sprintf("/layergroups/%s", name)
	} else {
		path = fmt.Sprintf("/workspaces/%s/layergroups/%s", workspace, name)
	}

	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("layer group not found: %s", name)
	}

	return io.ReadAll(resp.Body)
}
