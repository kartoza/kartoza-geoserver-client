package webserver

import (
	"encoding/json"
	"net/http"

	"github.com/kartoza/kartoza-geoserver-client/internal/api"
	"github.com/kartoza/kartoza-geoserver-client/internal/models"
)

// GWCLayerResponse represents a cached layer in the API response
type GWCLayerResponse struct {
	Name        string   `json:"name"`
	Enabled     bool     `json:"enabled"`
	GridSubsets []string `json:"gridSubsets"`
	MimeFormats []string `json:"mimeFormats"`
}

// GWCSeedTaskResponse represents a seed task in the API response
type GWCSeedTaskResponse struct {
	ID            int64   `json:"id"`
	TilesDone     int64   `json:"tilesDone"`
	TilesTotal    int64   `json:"tilesTotal"`
	TimeRemaining int64   `json:"timeRemaining"`
	Status        string  `json:"status"`
	LayerName     string  `json:"layerName"`
	Progress      float64 `json:"progress"` // Percentage 0-100
}

// GWCGridSetResponse represents a grid set in the API response
type GWCGridSetResponse struct {
	Name       string  `json:"name"`
	SRS        string  `json:"srs"`
	TileWidth  int     `json:"tileWidth"`
	TileHeight int     `json:"tileHeight"`
	MinX       float64 `json:"minX,omitempty"`
	MinY       float64 `json:"minY,omitempty"`
	MaxX       float64 `json:"maxX,omitempty"`
	MaxY       float64 `json:"maxY,omitempty"`
}

// handleGWCLayers handles requests to /api/gwc/layers/{connId} and /api/gwc/layers/{connId}/{layerName}
func (s *Server) handleGWCLayers(w http.ResponseWriter, r *http.Request) {
	connID, layerName, _ := parsePathParams(r.URL.Path, "/api/gwc/layers")

	if connID == "" {
		s.jsonError(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if layerName == "" {
			s.getGWCLayers(w, r, client)
		} else {
			s.getGWCLayer(w, r, client, layerName)
		}
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getGWCLayers returns all cached layers
func (s *Server) getGWCLayers(w http.ResponseWriter, r *http.Request, client *api.Client) {
	layers, err := client.GetGWCLayers()
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]GWCLayerResponse, len(layers))
	for i, layer := range layers {
		response[i] = GWCLayerResponse{
			Name:        layer.Name,
			Enabled:     layer.Enabled,
			GridSubsets: layer.GridSubsets,
			MimeFormats: layer.MimeFormats,
		}
	}

	s.jsonResponse(w, response)
}

// getGWCLayer returns details for a specific cached layer
func (s *Server) getGWCLayer(w http.ResponseWriter, r *http.Request, client *api.Client, layerName string) {
	layer, err := client.GetGWCLayer(layerName)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GWCLayerResponse{
		Name:        layer.Name,
		Enabled:     layer.Enabled,
		GridSubsets: layer.GridSubsets,
		MimeFormats: layer.MimeFormats,
	}

	s.jsonResponse(w, response)
}

// handleGWCSeed handles requests to /api/gwc/seed/{connId}/{layerName}
func (s *Server) handleGWCSeed(w http.ResponseWriter, r *http.Request) {
	connID, layerName, _ := parsePathParams(r.URL.Path, "/api/gwc/seed")

	if connID == "" {
		s.jsonError(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if layerName != "" {
			s.getSeedStatus(w, r, client, layerName)
		} else {
			s.jsonError(w, "Layer name is required", http.StatusBadRequest)
		}
	case http.MethodPost:
		if layerName != "" {
			s.startSeedOperation(w, r, client, layerName)
		} else {
			s.jsonError(w, "Layer name is required", http.StatusBadRequest)
		}
	case http.MethodDelete:
		if layerName != "" {
			s.terminateLayerSeed(w, r, client, layerName)
		} else {
			s.terminateAllSeeds(w, r, client)
		}
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getSeedStatus returns the status of seed tasks for a layer
func (s *Server) getSeedStatus(w http.ResponseWriter, r *http.Request, client *api.Client, layerName string) {
	status, err := client.GetSeedStatus(layerName)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]GWCSeedTaskResponse, len(status.Tasks))
	for i, task := range status.Tasks {
		progress := 0.0
		if task.TilesTotal > 0 {
			progress = float64(task.TilesDone) / float64(task.TilesTotal) * 100
		}
		response[i] = GWCSeedTaskResponse{
			ID:            task.ID,
			TilesDone:     task.TilesDone,
			TilesTotal:    task.TilesTotal,
			TimeRemaining: task.TimeRemaining,
			Status:        task.Status,
			LayerName:     task.LayerName,
			Progress:      progress,
		}
	}

	s.jsonResponse(w, response)
}

// SeedRequest represents a seed request from the client
type SeedRequest struct {
	GridSetID   string     `json:"gridSetId"`
	ZoomStart   int        `json:"zoomStart"`
	ZoomStop    int        `json:"zoomStop"`
	Format      string     `json:"format"`
	Type        string     `json:"type"` // seed, reseed, truncate
	ThreadCount int        `json:"threadCount"`
	Bounds      *GWCBounds `json:"bounds,omitempty"`
}

// GWCBounds represents geographic bounds
type GWCBounds struct {
	MinX float64 `json:"minX"`
	MinY float64 `json:"minY"`
	MaxX float64 `json:"maxX"`
	MaxY float64 `json:"maxY"`
	SRS  string  `json:"srs"`
}

// startSeedOperation starts a seed/reseed/truncate operation
func (s *Server) startSeedOperation(w http.ResponseWriter, r *http.Request, client *api.Client, layerName string) {
	var req SeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.GridSetID == "" {
		s.jsonError(w, "gridSetId is required", http.StatusBadRequest)
		return
	}
	if req.Format == "" {
		s.jsonError(w, "format is required", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		req.Type = "seed" // Default to seed
	}
	if req.ThreadCount <= 0 {
		req.ThreadCount = 1
	}

	// Build the seed request
	seedReq := models.GWCSeedRequest{
		GridSetID:   req.GridSetID,
		ZoomStart:   req.ZoomStart,
		ZoomStop:    req.ZoomStop,
		Format:      req.Format,
		Type:        req.Type,
		ThreadCount: req.ThreadCount,
	}

	if req.Bounds != nil {
		seedReq.Bounds = &models.GWCBounds{
			MinX: req.Bounds.MinX,
			MinY: req.Bounds.MinY,
			MaxX: req.Bounds.MaxX,
			MaxY: req.Bounds.MaxY,
			SRS:  req.Bounds.SRS,
		}
	}

	if err := client.SeedLayer(layerName, seedReq); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"success": true,
		"message": "Seed operation started",
		"layer":   layerName,
		"type":    req.Type,
	})
}

// terminateLayerSeed terminates seed tasks for a specific layer
func (s *Server) terminateLayerSeed(w http.ResponseWriter, r *http.Request, client *api.Client, layerName string) {
	if err := client.TerminateLayerSeedTasks(layerName); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"success": true,
		"message": "Seed tasks terminated",
		"layer":   layerName,
	})
}

// terminateAllSeeds terminates all running seed tasks
func (s *Server) terminateAllSeeds(w http.ResponseWriter, r *http.Request, client *api.Client) {
	killType := r.URL.Query().Get("type")
	if killType == "" {
		killType = "all"
	}

	if err := client.TerminateSeedTasks(killType); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"success":  true,
		"message":  "Seed tasks terminated",
		"killType": killType,
	})
}

// handleGWCTruncate handles requests to /api/gwc/truncate/{connId}/{layerName}
func (s *Server) handleGWCTruncate(w http.ResponseWriter, r *http.Request) {
	connID, layerName, _ := parsePathParams(r.URL.Path, "/api/gwc/truncate")

	if connID == "" {
		s.jsonError(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if layerName == "" {
		s.jsonError(w, "Layer name is required", http.StatusBadRequest)
		return
	}

	// Get optional parameters from query or body
	var req struct {
		GridSetID string `json:"gridSetId"`
		Format    string `json:"format"`
		ZoomStart int    `json:"zoomStart"`
		ZoomStop  int    `json:"zoomStop"`
	}

	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// If no specific parameters, truncate all cached tiles
	if req.GridSetID == "" || req.Format == "" {
		// Get layer info to find grid sets and formats
		layer, err := client.GetGWCLayer(layerName)
		if err != nil {
			s.jsonError(w, "Failed to get layer info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Truncate all combinations
		for _, gridSet := range layer.GridSubsets {
			for _, format := range layer.MimeFormats {
				zoomStop := 20
				if req.ZoomStop > 0 {
					zoomStop = req.ZoomStop
				}
				if err := client.TruncateLayer(layerName, gridSet, format, req.ZoomStart, zoomStop); err != nil {
					s.jsonError(w, "Failed to truncate: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	} else {
		// Truncate specific combination
		zoomStop := 20
		if req.ZoomStop > 0 {
			zoomStop = req.ZoomStop
		}
		if err := client.TruncateLayer(layerName, req.GridSetID, req.Format, req.ZoomStart, zoomStop); err != nil {
			s.jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	s.jsonResponse(w, map[string]interface{}{
		"success": true,
		"message": "Cache truncated",
		"layer":   layerName,
	})
}

// handleGWCGridSets handles requests to /api/gwc/gridsets/{connId}
func (s *Server) handleGWCGridSets(w http.ResponseWriter, r *http.Request) {
	connID, gridSetName, _ := parsePathParams(r.URL.Path, "/api/gwc/gridsets")

	if connID == "" {
		s.jsonError(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if gridSetName == "" {
		s.getGWCGridSets(w, r, client)
	} else {
		s.getGWCGridSet(w, r, client, gridSetName)
	}
}

// getGWCGridSets returns all available grid sets
func (s *Server) getGWCGridSets(w http.ResponseWriter, r *http.Request, client *api.Client) {
	gridSets, err := client.GetGWCGridSets()
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]GWCGridSetResponse, len(gridSets))
	for i, gs := range gridSets {
		response[i] = GWCGridSetResponse{
			Name:       gs.Name,
			SRS:        gs.SRS,
			TileWidth:  gs.TileWidth,
			TileHeight: gs.TileHeight,
			MinX:       gs.MinX,
			MinY:       gs.MinY,
			MaxX:       gs.MaxX,
			MaxY:       gs.MaxY,
		}
	}

	s.jsonResponse(w, response)
}

// getGWCGridSet returns details for a specific grid set
func (s *Server) getGWCGridSet(w http.ResponseWriter, r *http.Request, client *api.Client, name string) {
	gridSet, err := client.GetGWCGridSet(name)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GWCGridSetResponse{
		Name:       gridSet.Name,
		SRS:        gridSet.SRS,
		TileWidth:  gridSet.TileWidth,
		TileHeight: gridSet.TileHeight,
		MinX:       gridSet.MinX,
		MinY:       gridSet.MinY,
		MaxX:       gridSet.MaxX,
		MaxY:       gridSet.MaxY,
	}

	s.jsonResponse(w, response)
}

// handleGWCDiskQuota handles requests to /api/gwc/diskquota/{connId}
func (s *Server) handleGWCDiskQuota(w http.ResponseWriter, r *http.Request) {
	connID, _, _ := parsePathParams(r.URL.Path, "/api/gwc/diskquota")

	if connID == "" {
		s.jsonError(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getGWCDiskQuota(w, r, client)
	case http.MethodPut:
		s.updateGWCDiskQuota(w, r, client)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GWCDiskQuotaResponse represents disk quota configuration
type GWCDiskQuotaResponse struct {
	Enabled          bool   `json:"enabled"`
	DiskBlockSize    int    `json:"diskBlockSize"`
	CacheCleanUpFreq int    `json:"cacheCleanUpFrequency"`
	MaxConcurrent    int    `json:"maxConcurrentCleanUps"`
	GlobalQuota      string `json:"globalQuota"`
}

// getGWCDiskQuota returns the disk quota configuration
func (s *Server) getGWCDiskQuota(w http.ResponseWriter, r *http.Request, client *api.Client) {
	quota, err := client.GetGWCDiskQuota()
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GWCDiskQuotaResponse{
		Enabled:          quota.Enabled,
		DiskBlockSize:    quota.DiskBlockSize,
		CacheCleanUpFreq: quota.CacheCleanUpFreq,
		MaxConcurrent:    quota.MaxConcurrent,
		GlobalQuota:      quota.GlobalQuota,
	}

	s.jsonResponse(w, response)
}

// updateGWCDiskQuota updates the disk quota configuration
func (s *Server) updateGWCDiskQuota(w http.ResponseWriter, r *http.Request, client *api.Client) {
	var req models.GWCDiskQuota
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := client.UpdateGWCDiskQuota(req); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"success": true,
		"message": "Disk quota updated",
	})
}
