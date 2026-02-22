package webserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/geonode"
)

// handleGeoNodeConnections handles GeoNode connection CRUD operations
// GET /api/geonode/connections - List all connections
// POST /api/geonode/connections - Create a new connection
func (s *Server) handleGeoNodeConnections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listGeoNodeConnections(w, r)
	case http.MethodPost:
		s.createGeoNodeConnection(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGeoNodeConnectionByID handles single connection operations
// GET /api/geonode/connections/{id} - Get connection details
// PUT /api/geonode/connections/{id} - Update connection
// DELETE /api/geonode/connections/{id} - Delete connection
func (s *Server) handleGeoNodeConnectionByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/geonode/connections/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Connection ID required", http.StatusBadRequest)
		return
	}

	connID := parts[0]

	// Check if this is a sub-resource request
	if len(parts) > 1 {
		switch parts[1] {
		case "test":
			s.testGeoNodeConnection(w, r, connID)
		case "datasets":
			s.getGeoNodeDatasets(w, r, connID)
		case "maps":
			s.getGeoNodeMaps(w, r, connID)
		case "documents":
			s.getGeoNodeDocuments(w, r, connID)
		case "geostories":
			s.getGeoNodeGeoStories(w, r, connID)
		case "dashboards":
			s.getGeoNodeDashboards(w, r, connID)
		case "resources":
			s.getGeoNodeResources(w, r, connID)
		case "upload":
			s.handleGeoNodeUpload(w, r, connID)
		case "download":
			// download/{pk}/{alternate}?format={format}
			if len(parts) > 2 {
				s.handleGeoNodeDownload(w, r, connID, parts[2:])
			} else {
				http.Error(w, "Dataset PK required for download", http.StatusBadRequest)
			}
		case "wms":
			// Proxy WMS requests to GeoNode's GeoServer
			s.handleGeoNodeWMSProxy(w, r, connID)
		default:
			http.Error(w, "Unknown sub-resource", http.StatusNotFound)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getGeoNodeConnection(w, r, connID)
	case http.MethodPut:
		s.updateGeoNodeConnection(w, r, connID)
	case http.MethodDelete:
		s.deleteGeoNodeConnection(w, r, connID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listGeoNodeConnections returns all GeoNode connections
func (s *Server) listGeoNodeConnections(w http.ResponseWriter, r *http.Request) {
	connections := s.config.GeoNodeConnections
	if connections == nil {
		connections = []config.GeoNodeConnection{}
	}

	// Mask passwords in response
	safeConns := make([]map[string]interface{}, len(connections))
	for i, conn := range connections {
		safeConns[i] = map[string]interface{}{
			"id":        conn.ID,
			"name":      conn.Name,
			"url":       conn.URL,
			"username":  conn.Username,
			"has_token": conn.Token != "",
			"is_active": conn.IsActive,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeConns)
}

// createGeoNodeConnection creates a new GeoNode connection
func (s *Server) createGeoNodeConnection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Token    string `json:"token,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.URL == "" {
		http.Error(w, "Name and URL are required", http.StatusBadRequest)
		return
	}

	conn := config.GeoNodeConnection{
		ID:       uuid.New().String(),
		Name:     input.Name,
		URL:      strings.TrimSuffix(input.URL, "/"),
		Username: input.Username,
		Password: input.Password,
		Token:    input.Token,
		IsActive: true,
	}

	// Test connection before saving
	client := geonode.NewClient(&conn)
	if err := client.TestConnection(); err != nil {
		http.Error(w, "Connection test failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.config.AddGeoNodeConnection(conn)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store client
	s.geonodeClientsMu.Lock()
	s.geonodeClients[conn.ID] = client
	s.geonodeClientsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       conn.ID,
		"name":     conn.Name,
		"url":      conn.URL,
		"username": conn.Username,
	})
}

// getGeoNodeConnection returns a single connection
func (s *Server) getGeoNodeConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetGeoNodeConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        conn.ID,
		"name":      conn.Name,
		"url":       conn.URL,
		"username":  conn.Username,
		"has_token": conn.Token != "",
		"is_active": conn.IsActive,
	})
}

// updateGeoNodeConnection updates an existing connection
func (s *Server) updateGeoNodeConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetGeoNodeConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	var input struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Token    string `json:"token,omitempty"`
		IsActive *bool  `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name != "" {
		conn.Name = input.Name
	}
	if input.URL != "" {
		conn.URL = strings.TrimSuffix(input.URL, "/")
	}
	if input.Username != "" {
		conn.Username = input.Username
	}
	if input.Password != "" {
		conn.Password = input.Password
	}
	if input.Token != "" {
		conn.Token = input.Token
	}
	if input.IsActive != nil {
		conn.IsActive = *input.IsActive
	}

	s.config.UpdateGeoNodeConnection(*conn)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update client
	s.geonodeClientsMu.Lock()
	s.geonodeClients[conn.ID] = geonode.NewClient(conn)
	s.geonodeClientsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       conn.ID,
		"name":     conn.Name,
		"url":      conn.URL,
		"username": conn.Username,
	})
}

// deleteGeoNodeConnection removes a connection
func (s *Server) deleteGeoNodeConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetGeoNodeConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	s.config.RemoveGeoNodeConnection(connID)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove client
	s.geonodeClientsMu.Lock()
	delete(s.geonodeClients, connID)
	s.geonodeClientsMu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// testGeoNodeConnection tests a connection
func (s *Server) testGeoNodeConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetGeoNodeConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	client := geonode.NewClient(conn)
	if err := client.TestConnection(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// getGeoNodeClient returns the client for a connection
func (s *Server) getGeoNodeClient(connID string) *geonode.Client {
	s.geonodeClientsMu.RLock()
	client, ok := s.geonodeClients[connID]
	s.geonodeClientsMu.RUnlock()

	if ok {
		return client
	}

	// Create client if not exists
	conn := s.config.GetGeoNodeConnection(connID)
	if conn == nil {
		return nil
	}

	client = geonode.NewClient(conn)
	s.geonodeClientsMu.Lock()
	s.geonodeClients[connID] = client
	s.geonodeClientsMu.Unlock()

	return client
}

// getGeoNodeResources returns all resources for a connection
func (s *Server) getGeoNodeResources(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getGeoNodeClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	resourceType := r.URL.Query().Get("type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 100
	}

	resources, err := client.GetResources(resourceType, page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get resources: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

// getGeoNodeDatasets returns datasets (layers) for a connection
func (s *Server) getGeoNodeDatasets(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getGeoNodeClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 100
	}

	datasets, err := client.GetDatasets(page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get datasets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(datasets)
}

// getGeoNodeMaps returns maps for a connection
func (s *Server) getGeoNodeMaps(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getGeoNodeClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 100
	}

	maps, err := client.GetMaps(page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get maps: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(maps)
}

// getGeoNodeDocuments returns documents for a connection
func (s *Server) getGeoNodeDocuments(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getGeoNodeClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 100
	}

	documents, err := client.GetDocuments(page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get documents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// getGeoNodeGeoStories returns geostories for a connection
func (s *Server) getGeoNodeGeoStories(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getGeoNodeClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 100
	}

	geostories, err := client.GetGeoStories(page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get geostories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(geostories)
}

// getGeoNodeDashboards returns dashboards for a connection
func (s *Server) getGeoNodeDashboards(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getGeoNodeClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 100
	}

	dashboards, err := client.GetDashboards(page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get dashboards: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboards)
}

// handleGeoNodeTestConnection tests a connection without saving
func (s *Server) handleGeoNodeTestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		URL      string `json:"url"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Token    string `json:"token,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn := &config.GeoNodeConnection{
		URL:      strings.TrimSuffix(input.URL, "/"),
		Username: input.Username,
		Password: input.Password,
		Token:    input.Token,
	}

	client := geonode.NewClient(conn)
	if err := client.TestConnection(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleGeoNodeUpload handles file upload to GeoNode
// POST /api/geonode/connections/{id}/upload
func (s *Server) handleGeoNodeUpload(w http.ResponseWriter, r *http.Request, connID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client := s.getGeoNodeClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	// Parse multipart form (64MB max for larger files)
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get optional parameters
	title := r.FormValue("title")
	abstract := r.FormValue("abstract")

	// Create temporary file to save upload
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "geonode-upload-*-"+header.Filename)
	if err != nil {
		http.Error(w, "Failed to create temporary file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy uploaded file to temp file
	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Failed to save uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	// Upload to GeoNode
	opts := &geonode.UploadOptions{
		Title:    title,
		Abstract: abstract,
	}

	result, err := client.UploadDataset(tempFile.Name(), opts)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": result.Success,
		"id":      result.ID,
		"status":  result.Status,
		"state":   result.State,
		"url":     result.URL,
		"message": result.Message,
	})
}

// handleGeoNodeDownload handles file download from GeoNode
// GET /api/geonode/connections/{id}/download/{pk}/{alternate}?format={format}
func (s *Server) handleGeoNodeDownload(w http.ResponseWriter, r *http.Request, connID string, pathParts []string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client := s.getGeoNodeClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	// pathParts should contain [pk, alternate...]
	if len(pathParts) < 2 {
		http.Error(w, "Dataset PK and alternate name required", http.StatusBadRequest)
		return
	}

	pk, err := strconv.Atoi(pathParts[0])
	if err != nil {
		http.Error(w, "Invalid dataset PK", http.StatusBadRequest)
		return
	}

	// Alternate name may contain slashes, so join the rest
	alternate := strings.Join(pathParts[1:], "/")
	// URL decode the alternate name
	alternate, _ = url.PathUnescape(alternate)

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "gpkg" // Default to GeoPackage
	}

	// Download the dataset
	data, filename, err := client.DownloadDataset(pk, alternate, format)
	if err != nil {
		http.Error(w, "Download failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

// handleGeoNodeWMSProxy proxies WMS requests to GeoNode's GeoServer
// GET /api/geonode/connections/{id}/wms?SERVICE=WMS&...
// This solves CORS issues when fetching WMS tiles from external GeoServers
func (s *Server) handleGeoNodeWMSProxy(w http.ResponseWriter, r *http.Request, connID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	conn := s.config.GetGeoNodeConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	// Build the GeoServer WMS URL
	// GeoNode typically exposes GeoServer at /geoserver/ows (public OWS endpoint)
	// This endpoint often has less restrictive auth than /geoserver/wms
	geoserverURL := conn.URL + "/geoserver/ows"

	// Forward all query parameters
	targetURL, err := url.Parse(geoserverURL)
	if err != nil {
		http.Error(w, "Invalid GeoServer URL", http.StatusInternalServerError)
		return
	}
	targetURL.RawQuery = r.URL.RawQuery

	// Create request to GeoServer
	req, err := http.NewRequest(http.MethodGet, targetURL.String(), nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Forward relevant headers
	if acceptHeader := r.Header.Get("Accept"); acceptHeader != "" {
		req.Header.Set("Accept", acceptHeader)
	}

	// Set a proper User-Agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; CloudBench/1.0; WMS Proxy)")

	// Note: We intentionally do NOT send authentication for WMS proxy requests
	// Public layers should be accessible without auth, and sending invalid/partial
	// credentials can cause the server to reject the request with 401

	// Create a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch from GeoServer: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers (selectively)
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
	}

	// Add CORS and CORP headers to allow cross-origin embedding
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Stream the response body
	io.Copy(w, resp.Body)
}
