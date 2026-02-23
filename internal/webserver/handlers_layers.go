// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package webserver

import (
	"encoding/json"
	"net/http"

	"github.com/kartoza/kartoza-cloudbench/internal/api"
	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

// LayerResponse represents a layer in API responses
type LayerResponse struct {
	Name         string `json:"name"`
	Workspace    string `json:"workspace"`
	Store        string `json:"store,omitempty"`
	StoreType    string `json:"storeType,omitempty"`
	Type         string `json:"type,omitempty"`
	Enabled      bool   `json:"enabled"`
	Advertised   bool   `json:"advertised"`
	Queryable    bool   `json:"queryable"`
	DefaultStyle string `json:"defaultStyle,omitempty"`
}

// LayerUpdateRequest represents a layer update request
type LayerUpdateRequest struct {
	Enabled    bool `json:"enabled"`
	Advertised bool `json:"advertised"`
	Queryable  bool `json:"queryable"`
}

// handleLayers handles layer related requests
// Pattern: /api/layers/{connId}/{workspace} or /api/layers/{connId}/{workspace}/{layer}
// Also handles: /api/layers/{connId}/{workspace}/{layer}/count
func (s *Server) handleLayers(w http.ResponseWriter, r *http.Request) {
	connID, workspace, layer, action := parseStorePathParams(r.URL.Path, "/api/layers")

	if connID == "" || workspace == "" {
		s.jsonError(w, "Connection ID and workspace are required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	// Handle actions on specific layers
	if layer != "" && action != "" {
		switch action {
		case "count":
			s.handleLayerFeatureCount(w, r, client, workspace, layer)
			return
		}
	}

	if layer == "" {
		// Operating on layer collection
		switch r.Method {
		case http.MethodGet:
			s.listLayers(w, r, client, workspace)
		case http.MethodOptions:
			s.handleCORS(w)
		default:
			s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		// Operating on a specific layer
		switch r.Method {
		case http.MethodGet:
			s.getLayer(w, r, client, workspace, layer)
		case http.MethodPut:
			s.updateLayer(w, r, client, workspace, layer)
		case http.MethodDelete:
			s.deleteLayer(w, r, client, workspace, layer)
		case http.MethodOptions:
			s.handleCORS(w)
		default:
			s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// listLayers returns all layers for a workspace
func (s *Server) listLayers(w http.ResponseWriter, r *http.Request, client *api.Client, workspace string) {
	layers, err := client.GetLayers(workspace)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]LayerResponse, len(layers))
	for i, layer := range layers {
		enabled := true
		if layer.Enabled != nil {
			enabled = *layer.Enabled
		}
		response[i] = LayerResponse{
			Name:      layer.Name,
			Workspace: workspace,
			Type:      layer.Type,
			Enabled:   enabled,
		}
	}
	s.jsonResponse(w, response)
}

// getLayer returns a specific layer
func (s *Server) getLayer(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, layer string) {
	config, err := client.GetLayerConfig(workspace, layer)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.jsonResponse(w, LayerResponse{
		Name:         config.Name,
		Workspace:    config.Workspace,
		Store:        config.Store,
		StoreType:    config.StoreType,
		Enabled:      config.Enabled,
		Advertised:   config.Advertised,
		Queryable:    config.Queryable,
		DefaultStyle: config.DefaultStyle,
	})
}

// updateLayer updates a layer
func (s *Server) updateLayer(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, layer string) {
	var req LayerUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current layer config to preserve store info
	currentConfig, err := client.GetLayerConfig(workspace, layer)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	config := models.LayerConfig{
		Name:       layer,
		Workspace:  workspace,
		Store:      currentConfig.Store,
		StoreType:  currentConfig.StoreType,
		Enabled:    req.Enabled,
		Advertised: req.Advertised,
		Queryable:  req.Queryable,
	}

	if err := client.UpdateLayerConfig(workspace, config); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, LayerResponse{
		Name:       config.Name,
		Workspace:  workspace,
		Store:      config.Store,
		StoreType:  config.StoreType,
		Enabled:    config.Enabled,
		Advertised: config.Advertised,
		Queryable:  config.Queryable,
	})
}

// deleteLayer deletes a layer and cleans up its GWC cache
func (s *Server) deleteLayer(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, layer string) {
	// Use cleanup method to also remove GWC cache
	if err := client.DeleteLayerWithCleanup(workspace, layer); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleFeatureTypes handles feature type related requests
// Pattern: /api/featuretypes/{connId}/{workspace}/{store}
func (s *Server) handleFeatureTypes(w http.ResponseWriter, r *http.Request) {
	connID, workspace, store, _ := parseStorePathParams(r.URL.Path, "/api/featuretypes")

	if connID == "" || workspace == "" || store == "" {
		s.jsonError(w, "Connection ID, workspace, and store are required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.listFeatureTypes(w, r, client, workspace, store)
	case http.MethodPost:
		s.publishFeatureType(w, r, client, workspace, store)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listFeatureTypes returns all feature types for a data store
func (s *Server) listFeatureTypes(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, store string) {
	featureTypes, err := client.GetFeatureTypes(workspace, store)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]map[string]string, len(featureTypes))
	for i, ft := range featureTypes {
		response[i] = map[string]string{
			"name":      ft.Name,
			"workspace": workspace,
			"store":     store,
		}
	}
	s.jsonResponse(w, response)
}

// publishFeatureType publishes a new feature type
func (s *Server) publishFeatureType(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, store string) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.jsonError(w, "Feature type name is required", http.StatusBadRequest)
		return
	}

	if err := client.PublishFeatureType(workspace, store, req.Name); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	s.jsonResponse(w, map[string]string{
		"name":      req.Name,
		"workspace": workspace,
		"store":     store,
	})
}

// handleCoverages handles coverage related requests
// Pattern: /api/coverages/{connId}/{workspace}/{store}
func (s *Server) handleCoverages(w http.ResponseWriter, r *http.Request) {
	connID, workspace, store, _ := parseStorePathParams(r.URL.Path, "/api/coverages")

	if connID == "" || workspace == "" || store == "" {
		s.jsonError(w, "Connection ID, workspace, and store are required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.listCoverages(w, r, client, workspace, store)
	case http.MethodPost:
		s.publishCoverage(w, r, client, workspace, store)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listCoverages returns all coverages for a coverage store
func (s *Server) listCoverages(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, store string) {
	coverages, err := client.GetCoverages(workspace, store)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]map[string]string, len(coverages))
	for i, cov := range coverages {
		response[i] = map[string]string{
			"name":      cov.Name,
			"workspace": workspace,
			"store":     store,
		}
	}
	s.jsonResponse(w, response)
}

// publishCoverage publishes a new coverage
func (s *Server) publishCoverage(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, store string) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.jsonError(w, "Coverage name is required", http.StatusBadRequest)
		return
	}

	if err := client.PublishCoverage(workspace, store, req.Name); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	s.jsonResponse(w, map[string]string{
		"name":      req.Name,
		"workspace": workspace,
		"store":     store,
	})
}

// LayerMetadataResponse represents comprehensive layer metadata in API responses
type LayerMetadataResponse struct {
	Name              string                 `json:"name"`
	NativeName        string                 `json:"nativeName,omitempty"`
	Workspace         string                 `json:"workspace"`
	Store             string                 `json:"store"`
	StoreType         string                 `json:"storeType"`
	Title             string                 `json:"title,omitempty"`
	Abstract          string                 `json:"abstract,omitempty"`
	Keywords          []string               `json:"keywords,omitempty"`
	NativeCRS         string                 `json:"nativeCRS,omitempty"`
	SRS               string                 `json:"srs,omitempty"`
	Enabled           bool                   `json:"enabled"`
	Advertised        bool                   `json:"advertised"`
	Queryable         bool                   `json:"queryable"`
	NativeBoundingBox *BoundingBoxResponse   `json:"nativeBoundingBox,omitempty"`
	LatLonBoundingBox *BoundingBoxResponse   `json:"latLonBoundingBox,omitempty"`
	AttributionTitle  string                 `json:"attributionTitle,omitempty"`
	AttributionHref   string                 `json:"attributionHref,omitempty"`
	AttributionLogo   string                 `json:"attributionLogo,omitempty"`
	MetadataLinks     []MetadataLinkResponse `json:"metadataLinks,omitempty"`
	DefaultStyle      string                 `json:"defaultStyle,omitempty"`
	MaxFeatures       int                    `json:"maxFeatures,omitempty"`
	NumDecimals       int                    `json:"numDecimals,omitempty"`
}

// BoundingBoxResponse represents a geographic bounding box
type BoundingBoxResponse struct {
	MinX float64 `json:"minx"`
	MinY float64 `json:"miny"`
	MaxX float64 `json:"maxx"`
	MaxY float64 `json:"maxy"`
	CRS  string  `json:"crs,omitempty"`
}

// MetadataLinkResponse represents a metadata link
type MetadataLinkResponse struct {
	Type         string `json:"type"`
	MetadataType string `json:"metadataType"`
	Content      string `json:"content"`
}

// LayerMetadataUpdateRequest represents a layer metadata update request
type LayerMetadataUpdateRequest struct {
	Title            string                 `json:"title,omitempty"`
	Abstract         string                 `json:"abstract,omitempty"`
	Keywords         []string               `json:"keywords,omitempty"`
	SRS              string                 `json:"srs,omitempty"`
	Enabled          *bool                  `json:"enabled,omitempty"`
	Advertised       *bool                  `json:"advertised,omitempty"`
	Queryable        *bool                  `json:"queryable,omitempty"`
	AttributionTitle string                 `json:"attributionTitle,omitempty"`
	AttributionHref  string                 `json:"attributionHref,omitempty"`
	MetadataLinks    []MetadataLinkResponse `json:"metadataLinks,omitempty"`
}

// handleLayerMetadata handles comprehensive layer metadata requests
// Pattern: /api/layermetadata/{connId}/{workspace}/{layer}
func (s *Server) handleLayerMetadata(w http.ResponseWriter, r *http.Request) {
	connID, workspace, layer := parsePathParams(r.URL.Path, "/api/layermetadata")

	if connID == "" || workspace == "" || layer == "" {
		s.jsonError(w, "Connection ID, workspace, and layer name are required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getLayerMetadata(w, r, client, workspace, layer)
	case http.MethodPut:
		s.updateLayerMetadata(w, r, client, workspace, layer)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getLayerMetadata returns comprehensive layer metadata
func (s *Server) getLayerMetadata(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, layer string) {
	metadata, err := client.GetLayerMetadata(workspace, layer)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	response := LayerMetadataResponse{
		Name:             metadata.Name,
		NativeName:       metadata.NativeName,
		Workspace:        metadata.Workspace,
		Store:            metadata.Store,
		StoreType:        metadata.StoreType,
		Title:            metadata.Title,
		Abstract:         metadata.Abstract,
		Keywords:         metadata.Keywords,
		NativeCRS:        metadata.NativeCRS,
		SRS:              metadata.SRS,
		Enabled:          metadata.Enabled,
		Advertised:       metadata.Advertised,
		Queryable:        metadata.Queryable,
		AttributionTitle: metadata.AttributionTitle,
		AttributionHref:  metadata.AttributionHref,
		AttributionLogo:  metadata.AttributionLogo,
		DefaultStyle:     metadata.DefaultStyle,
		MaxFeatures:      metadata.MaxFeatures,
		NumDecimals:      metadata.NumDecimals,
	}

	if metadata.NativeBoundingBox != nil {
		response.NativeBoundingBox = &BoundingBoxResponse{
			MinX: metadata.NativeBoundingBox.MinX,
			MinY: metadata.NativeBoundingBox.MinY,
			MaxX: metadata.NativeBoundingBox.MaxX,
			MaxY: metadata.NativeBoundingBox.MaxY,
			CRS:  metadata.NativeBoundingBox.CRS,
		}
	}

	if metadata.LatLonBoundingBox != nil {
		response.LatLonBoundingBox = &BoundingBoxResponse{
			MinX: metadata.LatLonBoundingBox.MinX,
			MinY: metadata.LatLonBoundingBox.MinY,
			MaxX: metadata.LatLonBoundingBox.MaxX,
			MaxY: metadata.LatLonBoundingBox.MaxY,
			CRS:  metadata.LatLonBoundingBox.CRS,
		}
	}

	for _, ml := range metadata.MetadataLinks {
		response.MetadataLinks = append(response.MetadataLinks, MetadataLinkResponse{
			Type:         ml.Type,
			MetadataType: ml.MetadataType,
			Content:      ml.Content,
		})
	}

	s.jsonResponse(w, response)
}

// updateLayerMetadata updates layer metadata
func (s *Server) updateLayerMetadata(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, layer string) {
	var req LayerMetadataUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current metadata to merge updates
	metadata, err := client.GetLayerMetadata(workspace, layer)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Update fields from request
	if req.Title != "" {
		metadata.Title = req.Title
	}
	if req.Abstract != "" {
		metadata.Abstract = req.Abstract
	}
	if req.Keywords != nil {
		metadata.Keywords = req.Keywords
	}
	if req.SRS != "" {
		metadata.SRS = req.SRS
	}
	if req.Enabled != nil {
		metadata.Enabled = *req.Enabled
	}
	if req.Advertised != nil {
		metadata.Advertised = *req.Advertised
	}
	if req.Queryable != nil {
		metadata.Queryable = *req.Queryable
	}
	if req.AttributionTitle != "" {
		metadata.AttributionTitle = req.AttributionTitle
	}
	if req.AttributionHref != "" {
		metadata.AttributionHref = req.AttributionHref
	}
	if req.MetadataLinks != nil {
		metadata.MetadataLinks = nil
		for _, ml := range req.MetadataLinks {
			metadata.MetadataLinks = append(metadata.MetadataLinks, models.MetadataLink{
				Type:         ml.Type,
				MetadataType: ml.MetadataType,
				Content:      ml.Content,
			})
		}
	}

	if err := client.UpdateLayerMetadata(workspace, metadata); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated metadata
	s.getLayerMetadata(w, r, client, workspace, layer)
}

// LayerStylesResponse represents the styles associated with a layer
type LayerStylesResponse struct {
	DefaultStyle     string   `json:"defaultStyle"`
	AdditionalStyles []string `json:"additionalStyles"`
}

// LayerStylesUpdateRequest represents a request to update layer styles
type LayerStylesUpdateRequest struct {
	DefaultStyle     string   `json:"defaultStyle"`
	AdditionalStyles []string `json:"additionalStyles"`
}

// handleLayerStyles handles layer style association requests
// Pattern: /api/layerstyles/{connId}/{workspace}/{layer}
func (s *Server) handleLayerStyles(w http.ResponseWriter, r *http.Request) {
	connID, workspace, layer := parsePathParams(r.URL.Path, "/api/layerstyles")

	if connID == "" || workspace == "" || layer == "" {
		s.jsonError(w, "Connection ID, workspace, and layer name are required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getLayerStyles(w, r, client, workspace, layer)
	case http.MethodPut:
		s.updateLayerStyles(w, r, client, workspace, layer)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getLayerStyles returns the styles associated with a layer
func (s *Server) getLayerStyles(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, layer string) {
	styles, err := client.GetLayerStyles(workspace, layer)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, LayerStylesResponse{
		DefaultStyle:     styles.DefaultStyle,
		AdditionalStyles: styles.AdditionalStyles,
	})
}

// updateLayerStyles updates the styles associated with a layer
func (s *Server) updateLayerStyles(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, layer string) {
	var req LayerStylesUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.DefaultStyle == "" {
		s.jsonError(w, "Default style is required", http.StatusBadRequest)
		return
	}

	if err := client.UpdateLayerStyles(workspace, layer, req.DefaultStyle, req.AdditionalStyles); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, LayerStylesResponse{
		DefaultStyle:     req.DefaultStyle,
		AdditionalStyles: req.AdditionalStyles,
	})
}

// handleLayerFeatureCount handles feature count requests
// Pattern: /api/layers/{connId}/{workspace}/{layer}/count
func (s *Server) handleLayerFeatureCount(w http.ResponseWriter, r *http.Request, client *api.Client, workspace, layer string) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}

	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := client.GetFeatureCount(workspace, layer)
	if err != nil {
		// Return -1 for layers where count cannot be determined (e.g., raster)
		s.jsonResponse(w, map[string]int64{"count": -1})
		return
	}

	s.jsonResponse(w, map[string]int64{"count": count})
}
