package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

func (c *Client) doGWCRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	url := c.baseURL + "/gwc/rest" + path

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

func (c *Client) doGWCJSONRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	return c.doGWCRequest(method, path, bodyReader, "application/json")
}

func (c *Client) GetGWCLayers() ([]models.GWCLayer, error) {
	resp, err := c.doGWCRequest("GET", "/layers", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get GWC layers: %s", string(body))
	}

	// GWC returns a list of layer names
	var result []string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode GWC layers: %w", err)
	}

	// Convert to GWCLayer objects
	layers := make([]models.GWCLayer, len(result))
	for i, name := range result {
		layers[i] = models.GWCLayer{
			Name:    name,
			Enabled: true, // Assume enabled if listed
		}
	}

	return layers, nil
}

func (c *Client) GetGWCLayer(layerName string) (*models.GWCLayer, error) {
	resp, err := c.doGWCRequest("GET", fmt.Sprintf("/layers/%s.json", layerName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get GWC layer: %s", string(body))
	}

	var result struct {
		GeoServerLayer struct {
			Name       string `json:"name"`
			Enabled    bool   `json:"enabled"`
			GridSubsets []struct {
				GridSetName string `json:"gridSetName"`
			} `json:"gridSubsets"`
			MimeFormats []string `json:"mimeFormats"`
		} `json:"GeoServerLayer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode GWC layer: %w", err)
	}

	layer := &models.GWCLayer{
		Name:        result.GeoServerLayer.Name,
		Enabled:     result.GeoServerLayer.Enabled,
		MimeFormats: result.GeoServerLayer.MimeFormats,
	}

	// Extract grid set names
	for _, gs := range result.GeoServerLayer.GridSubsets {
		layer.GridSubsets = append(layer.GridSubsets, gs.GridSetName)
	}

	return layer, nil
}

func (c *Client) SeedLayer(layerName string, request models.GWCSeedRequest) error {
	// Build the seed request body in GWC format
	body := map[string]interface{}{
		"seedRequest": map[string]interface{}{
			"name":        layerName,
			"gridSetId":   request.GridSetID,
			"zoomStart":   request.ZoomStart,
			"zoomStop":    request.ZoomStop,
			"format":      request.Format,
			"type":        request.Type, // seed, reseed, or truncate
			"threadCount": request.ThreadCount,
		},
	}

	if request.Bounds != nil {
		body["seedRequest"].(map[string]interface{})["bounds"] = map[string]interface{}{
			"coords": map[string]interface{}{
				"double": []float64{
					request.Bounds.MinX,
					request.Bounds.MinY,
					request.Bounds.MaxX,
					request.Bounds.MaxY,
				},
			},
			"srs": map[string]interface{}{
				"number": request.Bounds.SRS,
			},
		}
	}

	resp, err := c.doGWCJSONRequest("POST", fmt.Sprintf("/seed/%s.json", layerName), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start seed operation: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) TruncateLayer(layerName string, gridSetID, format string, zoomStart, zoomStop int) error {
	request := models.GWCSeedRequest{
		GridSetID:   gridSetID,
		ZoomStart:   zoomStart,
		ZoomStop:    zoomStop,
		Format:      format,
		Type:        "truncate",
		ThreadCount: 1,
	}
	return c.SeedLayer(layerName, request)
}

func (c *Client) GetSeedStatus(layerName string) (*models.GWCSeedStatus, error) {
	resp, err := c.doGWCRequest("GET", fmt.Sprintf("/seed/%s.json", layerName), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get seed status: %s", string(body))
	}

	// GWC returns an array of arrays with task info
	// Format: [[tiles done, tiles total, time remaining, task id, status]]
	var result struct {
		LongArrayArray [][]int64 `json:"long-array-array"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode seed status: %w", err)
	}

	status := &models.GWCSeedStatus{
		Tasks: make([]models.GWCSeedTask, 0),
	}

	for _, taskData := range result.LongArrayArray {
		if len(taskData) >= 5 {
			task := models.GWCSeedTask{
				TilesDone:     taskData[0],
				TilesTotal:    taskData[1],
				TimeRemaining: taskData[2],
				ID:            taskData[3],
				LayerName:     layerName,
			}
			// Status is encoded as an integer
			switch taskData[4] {
			case 0:
				task.Status = "Pending"
			case 1:
				task.Status = "Running"
			case 2:
				task.Status = "Done"
			case -1:
				task.Status = "Aborted"
			default:
				task.Status = "Unknown"
			}
			status.Tasks = append(status.Tasks, task)
		}
	}

	return status, nil
}

// TerminateSeedTasks terminates running seed tasks
func (c *Client) TerminateSeedTasks(killType string) error {
	resp, err := c.doGWCRequest("POST", fmt.Sprintf("/seed?kill_all=%s", killType), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to terminate seed tasks: %s", string(body))
	}

	return nil
}

func (c *Client) TerminateLayerSeedTasks(layerName string) error {
	// GWC REST API expects kill_all as a query parameter, not JSON body
	resp, err := c.doGWCRequest("POST", fmt.Sprintf("/seed/%s?kill_all=all", layerName), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to terminate layer seed tasks: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) GetGWCGridSets() ([]models.GWCGridSet, error) {
	resp, err := c.doGWCRequest("GET", "/gridsets", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get grid sets: %s", string(body))
	}

	// GWC returns a list of grid set names
	var names []string
	if err := json.NewDecoder(resp.Body).Decode(&names); err != nil {
		return nil, fmt.Errorf("failed to decode grid sets: %w", err)
	}

	gridSets := make([]models.GWCGridSet, len(names))
	for i, name := range names {
		gridSets[i] = models.GWCGridSet{Name: name}
	}

	return gridSets, nil
}

func (c *Client) GetGWCGridSet(name string) (*models.GWCGridSet, error) {
	resp, err := c.doGWCRequest("GET", fmt.Sprintf("/gridsets/%s.json", name), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get grid set: %s", string(body))
	}

	var result struct {
		GridSet struct {
			Name        string  `json:"name"`
			SRS         struct {
				Number int `json:"number"`
			} `json:"srs"`
			TileWidth   int     `json:"tileWidth"`
			TileHeight  int     `json:"tileHeight"`
			Extent      struct {
				Coords struct {
					Double []float64 `json:"double"`
				} `json:"coords"`
			} `json:"extent"`
		} `json:"gridSet"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode grid set: %w", err)
	}

	gridSet := &models.GWCGridSet{
		Name:       result.GridSet.Name,
		SRS:        fmt.Sprintf("EPSG:%d", result.GridSet.SRS.Number),
		TileWidth:  result.GridSet.TileWidth,
		TileHeight: result.GridSet.TileHeight,
	}

	if len(result.GridSet.Extent.Coords.Double) >= 4 {
		gridSet.MinX = result.GridSet.Extent.Coords.Double[0]
		gridSet.MinY = result.GridSet.Extent.Coords.Double[1]
		gridSet.MaxX = result.GridSet.Extent.Coords.Double[2]
		gridSet.MaxY = result.GridSet.Extent.Coords.Double[3]
	}

	return gridSet, nil
}

func (c *Client) GetGWCDiskQuota() (*models.GWCDiskQuota, error) {
	resp, err := c.doGWCRequest("GET", "/diskquota.json", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get disk quota: %s", string(body))
	}

	var result struct {
		DiskQuota struct {
			Enabled          bool   `json:"enabled"`
			DiskBlockSize    int    `json:"diskBlockSize"`
			CacheCleanUpFreq int    `json:"cacheCleanUpFrequency"`
			MaxConcurrent    int    `json:"maxConcurrentCleanUps"`
			GlobalQuota      struct {
				Value string `json:"value"`
				Units string `json:"units"`
			} `json:"globalQuota"`
		} `json:"gwcQuotaConfiguration"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode disk quota: %w", err)
	}

	quota := &models.GWCDiskQuota{
		Enabled:          result.DiskQuota.Enabled,
		DiskBlockSize:    result.DiskQuota.DiskBlockSize,
		CacheCleanUpFreq: result.DiskQuota.CacheCleanUpFreq,
		MaxConcurrent:    result.DiskQuota.MaxConcurrent,
	}

	if result.DiskQuota.GlobalQuota.Value != "" {
		quota.GlobalQuota = fmt.Sprintf("%s %s", result.DiskQuota.GlobalQuota.Value, result.DiskQuota.GlobalQuota.Units)
	}

	return quota, nil
}

func (c *Client) UpdateGWCDiskQuota(quota models.GWCDiskQuota) error {
	body := map[string]interface{}{
		"gwcQuotaConfiguration": map[string]interface{}{
			"enabled":               quota.Enabled,
			"diskBlockSize":         quota.DiskBlockSize,
			"cacheCleanUpFrequency": quota.CacheCleanUpFreq,
			"maxConcurrentCleanUps": quota.MaxConcurrent,
		},
	}

	resp, err := c.doGWCJSONRequest("PUT", "/diskquota.json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update disk quota: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) MassGWCTruncate(layerNames []string) error {
	for _, name := range layerNames {
		// Get layer info to find grid sets and formats
		layer, err := c.GetGWCLayer(name)
		if err != nil {
			return fmt.Errorf("failed to get layer %s info: %w", name, err)
		}

		// Truncate each grid set and format combination
		for _, gridSet := range layer.GridSubsets {
			for _, format := range layer.MimeFormats {
				if err := c.TruncateLayer(name, gridSet, format, 0, 20); err != nil {
					return fmt.Errorf("failed to truncate %s: %w", name, err)
				}
			}
		}
	}
	return nil
}

func (c *Client) DeleteGWCLayer(layerName string) error {
	resp, err := c.doGWCRequest("DELETE", fmt.Sprintf("/layers/%s", layerName), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 200 OK or 404 Not Found are both acceptable (layer may not be cached)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete GWC layer: %s", string(body))
	}

	return nil
}

func (c *Client) TruncateAndDeleteGWCLayer(layerName string) error {
	// First try to get the layer info and truncate
	layer, err := c.GetGWCLayer(layerName)
	if err == nil && layer != nil {
		// Truncate each grid set and format combination
		for _, gridSet := range layer.GridSubsets {
			for _, format := range layer.MimeFormats {
				// Ignore truncate errors - layer might not have tiles
				_ = c.TruncateLayer(layerName, gridSet, format, 0, 20)
			}
		}
	}

	// Then delete the GWC layer entry
	return c.DeleteGWCLayer(layerName)
}

