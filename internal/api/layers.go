// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

func (c *Client) GetLayers(workspace string) ([]models.Layer, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/layers", workspace), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get layers: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Fix GeoServer's empty string response quirk
	body = fixEmptyGeoServerResponse(body)

	var result struct {
		Layers struct {
			Layer []models.Layer `json:"layer"`
		} `json:"layers"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode layers: %w", err)
	}

	for i := range result.Layers.Layer {
		result.Layers.Layer[i].Workspace = workspace
	}

	return result.Layers.Layer, nil
}

func (c *Client) DeleteLayer(workspace, name string) error {
	path := fmt.Sprintf("/workspaces/%s/layers/%s", workspace, name)

	resp, err := c.doRequest("DELETE", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete layer: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) GetFeatureTypes(workspace, datastore string) ([]models.FeatureType, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes", workspace, datastore), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get feature types: %s", string(body))
	}

	var result struct {
		FeatureTypes struct {
			FeatureType []models.FeatureType `json:"featureType"`
		} `json:"featureTypes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode feature types: %w", err)
	}

	for i := range result.FeatureTypes.FeatureType {
		result.FeatureTypes.FeatureType[i].Workspace = workspace
		result.FeatureTypes.FeatureType[i].Store = datastore
	}

	return result.FeatureTypes.FeatureType, nil
}

// GetAvailableFeatureTypes fetches unpublished feature types for a datastore
func (c *Client) GetCoverages(workspace, coveragestore string) ([]models.Coverage, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/coveragestores/%s/coverages", workspace, coveragestore), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get coverages: %s", string(body))
	}

	var result struct {
		Coverages struct {
			Coverage []models.Coverage `json:"coverage"`
		} `json:"coverages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode coverages: %w", err)
	}

	for i := range result.Coverages.Coverage {
		result.Coverages.Coverage[i].Workspace = workspace
		result.Coverages.Coverage[i].Store = coveragestore
	}

	return result.Coverages.Coverage, nil
}

func (c *Client) PublishFeatureType(workspace, dataStore, featureTypeName string) error {
	body := map[string]interface{}{
		"featureType": map[string]interface{}{
			"name":       featureTypeName,
			"nativeName": featureTypeName,
			"title":      featureTypeName,
			"enabled":    true,
			"advertised": true,
		},
	}

	resp, err := c.doJSONRequest("POST", fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes", workspace, dataStore), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to publish feature type: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) EnableLayer(workspace, layerName string, enabled bool) error {
	body := map[string]interface{}{
		"layer": map[string]interface{}{
			"enabled": enabled,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/layers/%s:%s", workspace, layerName), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update layer: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) SetLayerAdvertised(workspace, layerName string, advertised bool) error {
	body := map[string]interface{}{
		"layer": map[string]interface{}{
			"advertised": advertised,
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/layers/%s:%s", workspace, layerName), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update layer: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) GetLayerConfig(workspace, layerName string) (*models.LayerConfig, error) {
	config := &models.LayerConfig{
		Name:      layerName,
		Workspace: workspace,
		// Defaults - will be overwritten by resource values
		Enabled:    true,
		Advertised: true,
	}

	// Get layer info to determine resource type and href
	resp, err := c.doRequest("GET", fmt.Sprintf("/layers/%s:%s", workspace, layerName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get layer: %s", string(bodyBytes))
	}

	var layerResult struct {
		Layer struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			Queryable *bool  `json:"queryable"`
			Resource  struct {
				Class string `json:"@class"`
				Name  string `json:"name"`
				Href  string `json:"href"`
			} `json:"resource"`
			DefaultStyle struct {
				Name string `json:"name"`
			} `json:"defaultStyle"`
		} `json:"layer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&layerResult); err != nil {
		return nil, fmt.Errorf("failed to decode layer: %w", err)
	}

	if layerResult.Layer.Queryable != nil {
		config.Queryable = *layerResult.Layer.Queryable
	}
	config.DefaultStyle = layerResult.Layer.DefaultStyle.Name

	// Determine store type from resource class
	isFeatureType := layerResult.Layer.Resource.Class == "featureType"
	if isFeatureType {
		config.StoreType = "datastore"
	} else {
		config.StoreType = "coveragestore"
	}

	// Parse the resource href to extract store name
	// Href format: http://host/rest/workspaces/{ws}/datastores/{store}/featuretypes/{name}.json
	// or: http://host/rest/workspaces/{ws}/coveragestores/{store}/coverages/{name}.json
	href := layerResult.Layer.Resource.Href
	config.Store = extractStoreFromHref(href, isFeatureType)

	// Now fetch the resource (featuretype or coverage) to get enabled/advertised
	// These are on the resource, not on the layer endpoint
	if isFeatureType && config.Store != "" {
		resourceResp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes/%s", workspace, config.Store, layerName), nil, "")
		if err == nil {
			defer resourceResp.Body.Close()
			if resourceResp.StatusCode == http.StatusOK {
				var ftResult struct {
					FeatureType struct {
						Enabled    *bool `json:"enabled"`
						Advertised *bool `json:"advertised"`
					} `json:"featureType"`
				}
				if json.NewDecoder(resourceResp.Body).Decode(&ftResult) == nil {
					if ftResult.FeatureType.Enabled != nil {
						config.Enabled = *ftResult.FeatureType.Enabled
					}
					if ftResult.FeatureType.Advertised != nil {
						config.Advertised = *ftResult.FeatureType.Advertised
					}
				}
			}
		}
	} else if !isFeatureType && config.Store != "" {
		resourceResp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/coveragestores/%s/coverages/%s", workspace, config.Store, layerName), nil, "")
		if err == nil {
			defer resourceResp.Body.Close()
			if resourceResp.StatusCode == http.StatusOK {
				var covResult struct {
					Coverage struct {
						Enabled    *bool `json:"enabled"`
						Advertised *bool `json:"advertised"`
					} `json:"coverage"`
				}
				if json.NewDecoder(resourceResp.Body).Decode(&covResult) == nil {
					if covResult.Coverage.Enabled != nil {
						config.Enabled = *covResult.Coverage.Enabled
					}
					if covResult.Coverage.Advertised != nil {
						config.Advertised = *covResult.Coverage.Advertised
					}
				}
			}
		}
	}

	return config, nil
}

// extractStoreFromHref extracts the store name from a resource href
func extractStoreFromHref(href string, isFeatureType bool) string {
	// Href format for featuretype: .../datastores/{store}/featuretypes/...
	// Href format for coverage: .../coveragestores/{store}/coverages/...
	var marker string
	if isFeatureType {
		marker = "/datastores/"
	} else {
		marker = "/coveragestores/"
	}

	idx := strings.Index(href, marker)
	if idx == -1 {
		return ""
	}

	// Get everything after the marker
	remaining := href[idx+len(marker):]

	// Find the next slash
	slashIdx := strings.Index(remaining, "/")
	if slashIdx == -1 {
		return remaining
	}

	return remaining[:slashIdx]
}

// UpdateLayerConfig updates layer configuration
func (c *Client) UpdateLayerConfig(workspace string, config models.LayerConfig) error {
	// Determine if this is a feature type or coverage based on store type
	isFeatureType := config.StoreType == "datastore"

	// Update the resource (where enabled/advertised are stored)
	if isFeatureType && config.Store != "" {
		body := map[string]interface{}{
			"featureType": map[string]interface{}{
				"enabled":    config.Enabled,
				"advertised": config.Advertised,
			},
		}

		resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes/%s", workspace, config.Store, config.Name), body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to update feature type: %s", string(bodyBytes))
		}
	} else if !isFeatureType && config.Store != "" {
		body := map[string]interface{}{
			"coverage": map[string]interface{}{
				"enabled":    config.Enabled,
				"advertised": config.Advertised,
			},
		}

		resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s/coveragestores/%s/coverages/%s", workspace, config.Store, config.Name), body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to update coverage: %s", string(bodyBytes))
		}
	} else {
		return fmt.Errorf("cannot update layer: store name is required")
	}

	// Update queryable on the layer endpoint (only for vector layers)
	if isFeatureType {
		layerBody := map[string]interface{}{
			"layer": map[string]interface{}{
				"queryable": config.Queryable,
			},
		}

		resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/layers/%s:%s", workspace, config.Name), layerBody)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Non-OK status is acceptable here as queryable might not be supported
	}

	return nil
}

func (c *Client) GetLayerStyles(workspace, layerName string) (*LayerStyles, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/layers/%s:%s", workspace, layerName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get layer styles: %s", string(bodyBytes))
	}

	var result struct {
		Layer struct {
			DefaultStyle struct {
				Name string `json:"name"`
			} `json:"defaultStyle"`
			Styles struct {
				Style []struct {
					Name string `json:"name"`
				} `json:"style"`
			} `json:"styles"`
		} `json:"layer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode layer styles: %w", err)
	}

	styles := &LayerStyles{
		DefaultStyle: result.Layer.DefaultStyle.Name,
	}

	for _, s := range result.Layer.Styles.Style {
		styles.AdditionalStyles = append(styles.AdditionalStyles, s.Name)
	}

	return styles, nil
}

func (c *Client) UpdateLayerStyles(workspace, layerName, defaultStyle string, additionalStyles []string) error {
	// Build the styles array
	stylesArray := make([]map[string]string, 0)
	for _, styleName := range additionalStyles {
		stylesArray = append(stylesArray, map[string]string{"name": styleName})
	}

	body := map[string]interface{}{
		"layer": map[string]interface{}{
			"defaultStyle": map[string]string{
				"name": defaultStyle,
			},
		},
	}

	// Only include styles if there are additional styles
	if len(stylesArray) > 0 {
		body["layer"].(map[string]interface{})["styles"] = map[string]interface{}{
			"style": stylesArray,
		}
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/layers/%s:%s", workspace, layerName), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update layer styles: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) GetLayerMetadata(workspace, layerName string) (*models.LayerMetadata, error) {
	metadata := &models.LayerMetadata{
		Name:      layerName,
		Workspace: workspace,
	}

	// Get layer info to determine resource type
	resp, err := c.doRequest("GET", fmt.Sprintf("/layers/%s:%s", workspace, layerName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get layer: %s", string(bodyBytes))
	}

	var layerResult struct {
		Layer struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			Queryable *bool  `json:"queryable"`
			Resource  struct {
				Class string `json:"@class"`
				Name  string `json:"name"`
				Href  string `json:"href"`
			} `json:"resource"`
			DefaultStyle struct {
				Name string `json:"name"`
			} `json:"defaultStyle"`
		} `json:"layer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&layerResult); err != nil {
		return nil, fmt.Errorf("failed to decode layer: %w", err)
	}

	if layerResult.Layer.Queryable != nil {
		metadata.Queryable = *layerResult.Layer.Queryable
	}
	metadata.DefaultStyle = layerResult.Layer.DefaultStyle.Name

	// Extract store name and type from href (more reliable than Resource.Class)
	storeName, isFeatureType := c.extractStoreNameFromHref(layerResult.Layer.Resource.Href)
	metadata.Store = storeName

	// Fall back to Resource.Class if href extraction didn't work
	if storeName == "" {
		isFeatureType = strings.Contains(layerResult.Layer.Resource.Class, "FeatureType")
		fmt.Printf("[API] GetLayerMetadata: WARNING - could not extract store from href, falling back to Resource.Class\n")
	}

	fmt.Printf("[API] GetLayerMetadata: layer=%s, resource.href=%s, extracted store=%s, isFeatureType=%v\n",
		layerName, layerResult.Layer.Resource.Href, storeName, isFeatureType)
	if isFeatureType {
		metadata.StoreType = "datastore"
	} else {
		metadata.StoreType = "coveragestore"
	}

	// Get resource details (featuretype or coverage)
	var resourcePath string
	if isFeatureType {
		resourcePath = fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes/%s", workspace, storeName, layerName)
	} else {
		resourcePath = fmt.Sprintf("/workspaces/%s/coveragestores/%s/coverages/%s", workspace, storeName, layerName)
	}

	resourceResp, err := c.doRequest("GET", resourcePath, nil, "")
	if err != nil {
		return metadata, nil // Return what we have
	}
	defer resourceResp.Body.Close()

	if resourceResp.StatusCode != http.StatusOK {
		return metadata, nil // Return what we have
	}

	if isFeatureType {
		var ftResult struct {
			FeatureType struct {
				Name       string `json:"name"`
				NativeName string `json:"nativeName"`
				Title      string `json:"title"`
				Abstract   string `json:"abstract"`
				Keywords   struct {
					String []string `json:"string"`
				} `json:"keywords"`
				NativeCRS         string `json:"nativeCRS"`
				SRS               string `json:"srs"`
				Enabled           bool   `json:"enabled"`
				Advertised        bool   `json:"advertised"`
				NativeBoundingBox struct {
					MinX float64 `json:"minx"`
					MinY float64 `json:"miny"`
					MaxX float64 `json:"maxx"`
					MaxY float64 `json:"maxy"`
					CRS  string  `json:"crs"`
				} `json:"nativeBoundingBox"`
				LatLonBoundingBox struct {
					MinX float64 `json:"minx"`
					MinY float64 `json:"miny"`
					MaxX float64 `json:"maxx"`
					MaxY float64 `json:"maxy"`
					CRS  string  `json:"crs"`
				} `json:"latLonBoundingBox"`
				MaxFeatures   int `json:"maxFeatures"`
				NumDecimals   int `json:"numDecimals"`
				MetadataLinks struct {
					MetadataLink []struct {
						Type         string `json:"type"`
						MetadataType string `json:"metadataType"`
						Content      string `json:"content"`
					} `json:"metadataLink"`
				} `json:"metadataLinks"`
			} `json:"featureType"`
		}

		if err := json.NewDecoder(resourceResp.Body).Decode(&ftResult); err == nil {
			ft := ftResult.FeatureType
			metadata.NativeName = ft.NativeName
			metadata.Title = ft.Title
			metadata.Abstract = ft.Abstract
			metadata.Keywords = ft.Keywords.String
			metadata.NativeCRS = ft.NativeCRS
			metadata.SRS = ft.SRS
			metadata.Enabled = ft.Enabled
			metadata.Advertised = ft.Advertised
			metadata.MaxFeatures = ft.MaxFeatures
			metadata.NumDecimals = ft.NumDecimals
			metadata.NativeBoundingBox = &models.BoundingBox{
				MinX: ft.NativeBoundingBox.MinX,
				MinY: ft.NativeBoundingBox.MinY,
				MaxX: ft.NativeBoundingBox.MaxX,
				MaxY: ft.NativeBoundingBox.MaxY,
				CRS:  ft.NativeBoundingBox.CRS,
			}
			metadata.LatLonBoundingBox = &models.BoundingBox{
				MinX: ft.LatLonBoundingBox.MinX,
				MinY: ft.LatLonBoundingBox.MinY,
				MaxX: ft.LatLonBoundingBox.MaxX,
				MaxY: ft.LatLonBoundingBox.MaxY,
				CRS:  ft.LatLonBoundingBox.CRS,
			}
			for _, ml := range ft.MetadataLinks.MetadataLink {
				metadata.MetadataLinks = append(metadata.MetadataLinks, models.MetadataLink{
					Type:         ml.Type,
					MetadataType: ml.MetadataType,
					Content:      ml.Content,
				})
			}
		}
	} else {
		var covResult struct {
			Coverage struct {
				Name       string `json:"name"`
				NativeName string `json:"nativeName"`
				Title      string `json:"title"`
				Abstract   string `json:"abstract"`
				Keywords   struct {
					String []string `json:"string"`
				} `json:"keywords"`
				NativeCRS         string `json:"nativeCRS"`
				SRS               string `json:"srs"`
				Enabled           bool   `json:"enabled"`
				Advertised        bool   `json:"advertised"`
				NativeBoundingBox struct {
					MinX float64 `json:"minx"`
					MinY float64 `json:"miny"`
					MaxX float64 `json:"maxx"`
					MaxY float64 `json:"maxy"`
					CRS  string  `json:"crs"`
				} `json:"nativeBoundingBox"`
				LatLonBoundingBox struct {
					MinX float64 `json:"minx"`
					MinY float64 `json:"miny"`
					MaxX float64 `json:"maxx"`
					MaxY float64 `json:"maxy"`
					CRS  string  `json:"crs"`
				} `json:"latLonBoundingBox"`
			} `json:"coverage"`
		}

		if err := json.NewDecoder(resourceResp.Body).Decode(&covResult); err == nil {
			cov := covResult.Coverage
			metadata.NativeName = cov.NativeName
			metadata.Title = cov.Title
			metadata.Abstract = cov.Abstract
			metadata.Keywords = cov.Keywords.String
			metadata.NativeCRS = cov.NativeCRS
			metadata.SRS = cov.SRS
			metadata.Enabled = cov.Enabled
			metadata.Advertised = cov.Advertised
			metadata.NativeBoundingBox = &models.BoundingBox{
				MinX: cov.NativeBoundingBox.MinX,
				MinY: cov.NativeBoundingBox.MinY,
				MaxX: cov.NativeBoundingBox.MaxX,
				MaxY: cov.NativeBoundingBox.MaxY,
				CRS:  cov.NativeBoundingBox.CRS,
			}
			metadata.LatLonBoundingBox = &models.BoundingBox{
				MinX: cov.LatLonBoundingBox.MinX,
				MinY: cov.LatLonBoundingBox.MinY,
				MaxX: cov.LatLonBoundingBox.MaxX,
				MaxY: cov.LatLonBoundingBox.MaxY,
				CRS:  cov.LatLonBoundingBox.CRS,
			}
		}
	}

	return metadata, nil
}

// GetFeatureCount returns the feature count for a vector layer via WFS
func (c *Client) GetFeatureCount(workspace, layerName string) (int64, error) {
	// Construct WFS URL for GetFeature with resultType=hits
	wfsURL := fmt.Sprintf("%s/%s/wfs?SERVICE=WFS&VERSION=2.0.0&REQUEST=GetFeature&TYPENAMES=%s:%s&resultType=hits",
		c.baseURL, workspace, workspace, layerName)

	req, err := http.NewRequest("GET", wfsURL, nil)
	if err != nil {
		return -1, fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return -1, fmt.Errorf("WFS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("WFS request returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("failed to read response: %w", err)
	}

	bodyStr := string(body)

	// Parse numberMatched or numberOfFeatures from the response
	// WFS 2.0 uses numberMatched, WFS 1.x uses numberOfFeatures
	var count int64 = -1

	// Try numberMatched first (WFS 2.0)
	if idx := strings.Index(bodyStr, "numberMatched=\""); idx != -1 {
		start := idx + len("numberMatched=\"")
		end := strings.Index(bodyStr[start:], "\"")
		if end != -1 {
			if n, err := fmt.Sscanf(bodyStr[start:start+end], "%d", &count); err == nil && n == 1 {
				return count, nil
			}
		}
	}

	// Try numberOfFeatures (WFS 1.x)
	if idx := strings.Index(bodyStr, "numberOfFeatures=\""); idx != -1 {
		start := idx + len("numberOfFeatures=\"")
		end := strings.Index(bodyStr[start:], "\"")
		if end != -1 {
			if n, err := fmt.Sscanf(bodyStr[start:start+end], "%d", &count); err == nil && n == 1 {
				return count, nil
			}
		}
	}

	return -1, fmt.Errorf("could not parse feature count from WFS response")
}

func (c *Client) UpdateLayerMetadata(workspace string, metadata *models.LayerMetadata) error {
	isFeatureType := metadata.StoreType == "datastore"

	// Validate store name is present
	if metadata.Store == "" {
		return fmt.Errorf("store name is empty for layer %s (storeType: %s)", metadata.Name, metadata.StoreType)
	}

	fmt.Printf("[API] UpdateLayerMetadata: workspace=%s, layer=%s, store=%s, storeType=%s\n",
		workspace, metadata.Name, metadata.Store, metadata.StoreType)

	// Build update body for the resource
	var resourcePath string
	var resourceBody map[string]interface{}

	if isFeatureType {
		resourcePath = fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes/%s", workspace, metadata.Store, metadata.Name)
		updateFields := map[string]interface{}{
			"enabled":    metadata.Enabled,
			"advertised": metadata.Advertised,
		}
		if metadata.Title != "" {
			updateFields["title"] = metadata.Title
		}
		if metadata.Abstract != "" {
			updateFields["abstract"] = metadata.Abstract
		}
		if len(metadata.Keywords) > 0 {
			updateFields["keywords"] = map[string]interface{}{
				"string": metadata.Keywords,
			}
		}
		if metadata.SRS != "" {
			updateFields["srs"] = metadata.SRS
		}
		resourceBody = map[string]interface{}{
			"featureType": updateFields,
		}
	} else {
		resourcePath = fmt.Sprintf("/workspaces/%s/coveragestores/%s/coverages/%s", workspace, metadata.Store, metadata.Name)
		updateFields := map[string]interface{}{
			"enabled":    metadata.Enabled,
			"advertised": metadata.Advertised,
		}
		if metadata.Title != "" {
			updateFields["title"] = metadata.Title
		}
		if metadata.Abstract != "" {
			updateFields["abstract"] = metadata.Abstract
		}
		if len(metadata.Keywords) > 0 {
			updateFields["keywords"] = map[string]interface{}{
				"string": metadata.Keywords,
			}
		}
		if metadata.SRS != "" {
			updateFields["srs"] = metadata.SRS
		}
		resourceBody = map[string]interface{}{
			"coverage": updateFields,
		}
	}

	fmt.Printf("[API] UpdateLayerMetadata: PUT %s\n", resourcePath)

	resp, err := c.doJSONRequest("PUT", resourcePath, resourceBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("[API] UpdateLayerMetadata failed: status=%d, body=%s\n", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("failed to update layer metadata: %s", string(bodyBytes))
	}

	// Update queryable on layer endpoint (only for vector)
	if isFeatureType {
		layerBody := map[string]interface{}{
			"layer": map[string]interface{}{
				"queryable": metadata.Queryable,
			},
		}

		layerResp, err := c.doJSONRequest("PUT", fmt.Sprintf("/layers/%s:%s", workspace, metadata.Name), layerBody)
		if err != nil {
			return nil // Non-critical
		}
		defer layerResp.Body.Close()
	}

	return nil
}

// extractStoreNameFromHref extracts the store name from a GeoServer resource href
// href format: .../workspaces/{ws}/datastores/{store}/featuretypes/{name}.json
// or: .../workspaces/{ws}/coveragestores/{store}/coverages/{name}.json
func (c *Client) extractStoreNameFromHref(href string) (string, bool) {
	// Try datastores first
	if parts := strings.Split(href, "/datastores/"); len(parts) >= 2 {
		storePart := parts[1]
		storeNameParts := strings.Split(storePart, "/")
		if len(storeNameParts) >= 1 && storeNameParts[0] != "" {
			return storeNameParts[0], true // isFeatureType = true
		}
	}

	// Try coveragestores
	if parts := strings.Split(href, "/coveragestores/"); len(parts) >= 2 {
		storePart := parts[1]
		storeNameParts := strings.Split(storePart, "/")
		if len(storeNameParts) >= 1 && storeNameParts[0] != "" {
			return storeNameParts[0], false // isFeatureType = false
		}
	}

	return "", false
}

// ============================================================================
// GeoWebCache (GWC) API Methods
// ============================================================================

func (c *Client) DeleteLayerWithCleanup(workspace, name string) error {
	// Clean up GWC cache first
	layerName := fmt.Sprintf("%s:%s", workspace, name)
	_ = c.TruncateAndDeleteGWCLayer(layerName)

	// Now delete the layer
	return c.DeleteLayer(workspace, name)
}

func (c *Client) DownloadLayer(workspace, name string) ([]byte, error) {
	var path string
	if workspace == "" {
		path = fmt.Sprintf("/layers/%s", name)
	} else {
		path = fmt.Sprintf("/workspaces/%s/layers/%s", workspace, name)
	}

	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("layer not found: %s", name)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) DownloadFeatureType(workspace, store, name string) ([]byte, error) {
	path := fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes/%s", workspace, store, name)
	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feature type not found: %s/%s/%s", workspace, store, name)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) DownloadCoverage(workspace, store, name string) ([]byte, error) {
	path := fmt.Sprintf("/workspaces/%s/coveragestores/%s/coverages/%s", workspace, store, name)
	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coverage not found: %s/%s/%s", workspace, store, name)
	}

	return io.ReadAll(resp.Body)
}
