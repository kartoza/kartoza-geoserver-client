// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package geonode provides a client for the GeoNode REST API v2
package geonode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/config"
)

// FlexInt is an int that can be unmarshalled from a JSON string or number
// GeoNode API sometimes returns pk as a string (e.g. "440") and sometimes as an int
type FlexInt int

func (fi *FlexInt) UnmarshalJSON(b []byte) error {
	// First try to unmarshal as int
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*fi = FlexInt(i)
		return nil
	}
	// Then try as string
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if s == "" {
			*fi = 0
			return nil
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("cannot convert %q to int: %w", s, err)
		}
		*fi = FlexInt(n)
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into FlexInt", string(b))
}

func (fi FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(fi))
}

// Int returns the underlying int value
func (fi FlexInt) Int() int {
	return int(fi)
}

// Client is a GeoNode API client
type Client struct {
	baseURL    string
	username   string
	password   string
	token      string
	httpClient *http.Client
}

// NewClient creates a new GeoNode API client from a connection config
func NewClient(conn *config.GeoNodeConnection) *Client {
	baseURL := strings.TrimSuffix(conn.URL, "/")
	return &Client{
		baseURL:  baseURL,
		username: conn.Username,
		password: conn.Password,
		token:    conn.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Resource represents a GeoNode resource (base type for all resources)
// Note: Many fields use interface{} because the GeoNode API returns inconsistent types
type Resource struct {
	PK                        FlexInt     `json:"pk"`
	UUID                      string      `json:"uuid"`
	Name                      string      `json:"name,omitempty"`
	Title                     string      `json:"title"`
	Abstract                  string      `json:"abstract,omitempty"`
	ResourceType              string      `json:"resource_type"` // dataset, map, document, geostory, dashboard
	Subtype                   string      `json:"subtype,omitempty"`
	Owner                     Owner       `json:"owner,omitempty"`
	Date                      string      `json:"date,omitempty"`
	DateType                  string      `json:"date_type,omitempty"`
	Created                   string      `json:"created,omitempty"`
	LastUpdated               string      `json:"last_updated,omitempty"`
	Featured                  bool        `json:"featured"`
	Published                 bool        `json:"is_published"`
	Approved                  bool        `json:"is_approved"`
	ThumbnailURL              string      `json:"thumbnail_url,omitempty"`
	DetailURL                 string      `json:"detail_url,omitempty"`
	EmbedURL                  string      `json:"embed_url,omitempty"`
	Keywords                  interface{} `json:"keywords,omitempty"`                    // Can be []string or []object
	Category                  interface{} `json:"category,omitempty"`                    // Can be string or object
	SpatialRepresentationType interface{} `json:"spatial_representation_type,omitempty"` // Can be string or object
	SRID                      string      `json:"srid,omitempty"`
	BBox                      interface{} `json:"bbox,omitempty"`            // Complex polygon type
	LLBBox                    interface{} `json:"ll_bbox_polygon,omitempty"` // Complex polygon type
}

// Owner represents a resource owner
type Owner struct {
	PK       FlexInt `json:"pk"`
	Username string  `json:"username"`
	Name     string  `json:"name,omitempty"`
	Avatar   string  `json:"avatar,omitempty"`
}

// Dataset represents a GeoNode dataset (layer)
type Dataset struct {
	Resource
	Alternate    string `json:"alternate,omitempty"`
	DefaultStyle *Style `json:"default_style,omitempty"`
	Workspace    string `json:"workspace,omitempty"`
	StoreName    string `json:"store,omitempty"`
	StoreType    string `json:"storeType,omitempty"`
}

// Style represents a layer style
type Style struct {
	PK       FlexInt `json:"pk"`
	Name     string  `json:"name"`
	Title    string  `json:"title,omitempty"`
	SLDTitle string  `json:"sld_title,omitempty"`
	SLDURL   string  `json:"sld_url,omitempty"`
}

// Map represents a GeoNode map
type Map struct {
	Resource
	Layers []MapLayer `json:"maplayers,omitempty"`
}

// MapLayerDataset represents a dataset reference within a map layer
type MapLayerDataset struct {
	PK        FlexInt `json:"pk"`
	Alternate string  `json:"alternate,omitempty"`
	Title     string  `json:"title,omitempty"`
}

// MapLayer represents a layer within a map
type MapLayer struct {
	PK           FlexInt          `json:"pk"`
	Name         string           `json:"name"`
	Opacity      float64          `json:"opacity"`
	Visibility   bool             `json:"visibility"`
	Order        int              `json:"order"`
	Dataset      *MapLayerDataset `json:"dataset,omitempty"`
	CurrentStyle string           `json:"current_style,omitempty"`
	ExtraParams  interface{}      `json:"extra_params,omitempty"`
}

// Document represents a GeoNode document
type Document struct {
	Resource
	Extension   string `json:"extension,omitempty"`
	DocFile     string `json:"doc_file,omitempty"`
	DocURL      string `json:"doc_url,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

// GeoStory represents a GeoNode GeoStory
type GeoStory struct {
	Resource
	Data string `json:"data,omitempty"` // GeoStory JSON data
}

// Dashboard represents a GeoNode Dashboard
type Dashboard struct {
	Resource
	Data string `json:"data,omitempty"` // Dashboard JSON data
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Total     int             `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"page_size"`
	Resources json.RawMessage `json:"resources"`
}

// ResourcesResponse represents the /api/v2/resources response
type ResourcesResponse struct {
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
	Resources []Resource `json:"resources"`
}

// DatasetsResponse represents the /api/v2/datasets response
type DatasetsResponse struct {
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Datasets []Dataset `json:"datasets"`
}

// MapsResponse represents the /api/v2/maps response
type MapsResponse struct {
	Total    int   `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Maps     []Map `json:"maps"`
}

// DocumentsResponse represents the /api/v2/documents response
type DocumentsResponse struct {
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
	Documents []Document `json:"documents"`
}

// GeoStoriesResponse represents the /api/v2/geostories response
type GeoStoriesResponse struct {
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	GeoStories []GeoStory `json:"geostories"`
}

// DashboardsResponse represents the /api/v2/dashboards response
type DashboardsResponse struct {
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Dashboards []Dashboard `json:"dashboards"`
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.baseURL + path

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Add authentication
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// TestConnection tests if the GeoNode connection is working
func (c *Client) TestConnection() error {
	resp, err := c.doRequest("GET", "/api/v2/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetResources returns all resources with optional filtering
func (c *Client) GetResources(resourceType string, page, pageSize int) (*ResourcesResponse, error) {
	params := url.Values{}
	if resourceType != "" {
		params.Set("filter{resource_type}", resourceType)
	}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v2/resources"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ResourcesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetDatasets returns all datasets (layers)
func (c *Client) GetDatasets(page, pageSize int) (*DatasetsResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v2/datasets"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result DatasetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetDataset returns a single dataset by ID
func (c *Client) GetDataset(id int) (*Dataset, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v2/datasets/%d", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Dataset Dataset `json:"dataset"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Dataset, nil
}

// GetMaps returns all maps
func (c *Client) GetMaps(page, pageSize int) (*MapsResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v2/maps"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MapsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetMap returns a single map by ID
func (c *Client) GetMap(id int) (*Map, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v2/maps/%d", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Map Map `json:"map"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Map, nil
}

// GetDocuments returns all documents
func (c *Client) GetDocuments(page, pageSize int) (*DocumentsResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v2/documents"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result DocumentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetGeoStories returns all geostories
func (c *Client) GetGeoStories(page, pageSize int) (*GeoStoriesResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v2/geostories"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GeoStoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetDashboards returns all dashboards
func (c *Client) GetDashboards(page, pageSize int) (*DashboardsResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v2/dashboards"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result DashboardsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetResource returns a single resource by ID
func (c *Client) GetResource(id int) (*Resource, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v2/resources/%d", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Resource Resource `json:"resource"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Resource, nil
}

// GetWMSURL returns the WMS URL for a dataset
func (c *Client) GetWMSURL(dataset *Dataset) string {
	// GeoNode typically uses GeoServer for WMS
	// The format is usually: {base_url}/geoserver/wms
	return c.baseURL + "/geoserver/wms"
}

// GetWMSLayerName returns the WMS layer name for a dataset
func (c *Client) GetWMSLayerName(dataset *Dataset) string {
	if dataset.Alternate != "" {
		return dataset.Alternate
	}
	if dataset.Workspace != "" && dataset.Name != "" {
		return dataset.Workspace + ":" + dataset.Name
	}
	return dataset.Name
}

// UploadResponse represents the response from GeoNode upload API
type UploadResponse struct {
	Success    bool      `json:"success"`
	Status     string    `json:"status,omitempty"`
	ID         int       `json:"id,omitempty"`
	Code       string    `json:"code,omitempty"`
	URL        string    `json:"url,omitempty"`
	BBox       []float64 `json:"bbox,omitempty"`
	CRS        string    `json:"crs,omitempty"`
	State      string    `json:"state,omitempty"`
	Progress   float64   `json:"progress,omitempty"`
	Message    string    `json:"message,omitempty"`
	Error      string    `json:"error,omitempty"`
	CreateDate string    `json:"create_date,omitempty"`
}

// UploadExecutionResponse represents execution info from imports endpoint
type UploadExecutionResponse struct {
	ID           int         `json:"id"`
	ExecID       string      `json:"exec_id"`
	Name         string      `json:"name,omitempty"`
	Status       string      `json:"status"`
	State        string      `json:"state"`
	Progress     int         `json:"progress"`
	CreateDate   string      `json:"create_date,omitempty"`
	StartDate    string      `json:"start_date,omitempty"`
	EndDate      string      `json:"end_date,omitempty"`
	OutputParams interface{} `json:"output_params,omitempty"`
	Source       string      `json:"source,omitempty"`
	Action       string      `json:"action,omitempty"`
}

// UploadOptions contains optional parameters for uploads
type UploadOptions struct {
	Title    string
	Abstract string
	Keywords []string
	Charset  string
	Async    bool // If true, returns immediately and allows status polling
}

// UploadDataset uploads a file (GPKG, Shapefile, GeoTIFF) to GeoNode
// Returns the upload response with execution ID for status tracking
func (c *Client) UploadDataset(filePath string, opts *UploadOptions) (*UploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file
	fileName := filepath.Base(filePath)
	part, err := writer.CreateFormFile("base_file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file to form: %w", err)
	}

	// Add optional fields
	if opts != nil {
		if opts.Title != "" {
			writer.WriteField("title", opts.Title)
		}
		if opts.Abstract != "" {
			writer.WriteField("abstract", opts.Abstract)
		}
		if opts.Charset != "" {
			writer.WriteField("charset", opts.Charset)
		} else {
			writer.WriteField("charset", "UTF-8")
		}
		if opts.Async {
			writer.WriteField("action", "upload")
		}
	} else {
		writer.WriteField("charset", "UTF-8")
	}

	// Action field is required
	writer.WriteField("action", "upload")

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	reqURL := c.baseURL + "/api/v2/uploads/upload"
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	// Add authentication
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Use longer timeout for uploads
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result UploadResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Try to extract basic info if structure is different
		return &UploadResponse{
			Success: true,
			Message: string(respBody),
		}, nil
	}

	result.Success = true
	return &result, nil
}

// GetUploadStatus checks the status of an upload execution
func (c *Client) GetUploadStatus(executionID int) (*UploadExecutionResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v2/resource-service/execution-status/%d", executionID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result UploadExecutionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DownloadDataset downloads a dataset in the specified format
// Supported formats: shp (shapefile), gpkg (GeoPackage), csv, json (GeoJSON), xlsx
func (c *Client) DownloadDataset(datasetPK int, alternate string, format string) ([]byte, string, error) {
	// GeoNode uses the alternate name (workspace:layername) for downloads
	// Endpoint: GET /datasets/{alternate}/dataset_download?export_format={format}
	if alternate == "" {
		return nil, "", fmt.Errorf("alternate name required for download")
	}

	// URL encode the alternate name
	path := fmt.Sprintf("/datasets/%s/dataset_download", url.PathEscape(alternate))
	params := url.Values{}
	params.Set("export_format", format)
	path += "?" + params.Encode()

	reqURL := c.baseURL + path
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Use longer timeout for downloads
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Get filename from Content-Disposition header
	filename := ""
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			filename = params["filename"]
		}
	}
	if filename == "" {
		// Generate filename from alternate and format
		filename = strings.ReplaceAll(alternate, ":", "_") + "." + format
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	return data, filename, nil
}

// DeleteDataset deletes a dataset from GeoNode
func (c *Client) DeleteDataset(datasetPK int) error {
	req, err := http.NewRequest("DELETE", c.baseURL+fmt.Sprintf("/api/v2/datasets/%d", datasetPK), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Add authentication
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
