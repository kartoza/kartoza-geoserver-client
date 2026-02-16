package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) GetAvailableFeatureTypes(workspace, datastore string) ([]string, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes?list=available", workspace, datastore), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get available feature types: %s", string(body))
	}

	var result struct {
		List struct {
			String []string `json:"string"`
		} `json:"list"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode available feature types: %w", err)
	}

	return result.List.String, nil
}

// CreateSQLViewLayer creates a SQL View layer in GeoServer
func (c *Client) CreateSQLViewLayer(workspace, dataStore string, config SQLViewConfig) error {
	// Build the virtualTable (SQL View) configuration
	virtualTable := map[string]interface{}{
		"name": config.Name,
		"sql":  config.SQL,
	}

	// Add key column if specified
	if config.KeyColumn != "" {
		virtualTable["keyColumn"] = config.KeyColumn
	}

	// Add escape SQL setting
	virtualTable["escapeSql"] = config.EscapeSql

	// Add geometry configuration
	if config.GeometryColumn != "" {
		geometry := map[string]interface{}{
			"name": config.GeometryColumn,
			"type": config.GeometryType,
			"srid": config.GeometrySRID,
		}
		virtualTable["geometry"] = geometry
	}

	// Add parameters if any
	if len(config.Parameters) > 0 {
		params := make([]map[string]interface{}, len(config.Parameters))
		for i, p := range config.Parameters {
			param := map[string]interface{}{
				"name": p.Name,
			}
			if p.DefaultValue != "" {
				param["defaultValue"] = p.DefaultValue
			}
			if p.RegexpValidator != "" {
				param["regexpValidator"] = p.RegexpValidator
			}
			params[i] = param
		}
		virtualTable["parameter"] = params
	}

	// Build the feature type with metadata containing virtualTable
	title := config.Title
	if title == "" {
		title = config.Name
	}

	body := map[string]interface{}{
		"featureType": map[string]interface{}{
			"name":       config.Name,
			"nativeName": config.Name,
			"title":      title,
			"abstract":   config.Abstract,
			"enabled":    true,
			"advertised": true,
			"metadata": map[string]interface{}{
				"entry": []map[string]interface{}{
					{
						"@key":         "JDBC_VIRTUAL_TABLE",
						"virtualTable": virtualTable,
					},
				},
			},
		},
	}

	resp, err := c.doJSONRequest("POST", fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes", workspace, dataStore), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create SQL view layer: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) UpdateSQLViewLayer(workspace, dataStore string, config SQLViewConfig) error {
	// Build the virtualTable (SQL View) configuration
	virtualTable := map[string]interface{}{
		"name": config.Name,
		"sql":  config.SQL,
	}

	if config.KeyColumn != "" {
		virtualTable["keyColumn"] = config.KeyColumn
	}

	virtualTable["escapeSql"] = config.EscapeSql

	if config.GeometryColumn != "" {
		geometry := map[string]interface{}{
			"name": config.GeometryColumn,
			"type": config.GeometryType,
			"srid": config.GeometrySRID,
		}
		virtualTable["geometry"] = geometry
	}

	if len(config.Parameters) > 0 {
		params := make([]map[string]interface{}, len(config.Parameters))
		for i, p := range config.Parameters {
			param := map[string]interface{}{
				"name": p.Name,
			}
			if p.DefaultValue != "" {
				param["defaultValue"] = p.DefaultValue
			}
			if p.RegexpValidator != "" {
				param["regexpValidator"] = p.RegexpValidator
			}
			params[i] = param
		}
		virtualTable["parameter"] = params
	}

	title := config.Title
	if title == "" {
		title = config.Name
	}

	body := map[string]interface{}{
		"featureType": map[string]interface{}{
			"name":       config.Name,
			"nativeName": config.Name,
			"title":      title,
			"abstract":   config.Abstract,
			"enabled":    true,
			"advertised": true,
			"metadata": map[string]interface{}{
				"entry": []map[string]interface{}{
					{
						"@key":         "JDBC_VIRTUAL_TABLE",
						"virtualTable": virtualTable,
					},
				},
			},
		},
	}

	resp, err := c.doJSONRequest("PUT", fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes/%s", workspace, dataStore, config.Name), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update SQL view layer: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) DeleteSQLViewLayer(workspace, dataStore, layerName string) error {
	// First delete the layer
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/layers/%s:%s", workspace, layerName), nil, "")
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Then delete the feature type with recurse=true to clean up
	resp, err = c.doRequest("DELETE", fmt.Sprintf("/workspaces/%s/datastores/%s/featuretypes/%s?recurse=true", workspace, dataStore, layerName), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete SQL view layer: %s", string(bodyBytes))
	}

	return nil
}

