// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package webserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/integration"
)

// BridgeCreateRequest represents a request to create a PostgreSQL to GeoServer bridge
type BridgeCreateRequest struct {
	PGServiceName         string   `json:"pg_service_name"`
	GeoServerConnectionID string   `json:"geoserver_connection_id"`
	Workspace             string   `json:"workspace"`
	StoreName             string   `json:"store_name"`
	Schema                string   `json:"schema"`
	Tables                []string `json:"tables"`
	PublishLayers         bool     `json:"publish_layers"`
}

// BridgeResponse represents a bridge creation response
type BridgeResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Link    *integration.LinkedStore `json:"link,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// handleBridge handles /api/bridge routes
func (s *Server) handleBridge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/bridge")
	path = strings.Trim(path, "/")

	switch {
	case path == "" || path == "create":
		if r.Method == http.MethodPost {
			s.handleCreateBridge(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case path == "tables":
		if r.Method == http.MethodGet {
			s.handleGetAvailableTables(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleCreateBridge creates a PostgreSQL to GeoServer bridge
func (s *Server) handleCreateBridge(w http.ResponseWriter, r *http.Request) {
	var req BridgeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PGServiceName == "" {
		s.jsonError(w, "pg_service_name is required", http.StatusBadRequest)
		return
	}
	if req.GeoServerConnectionID == "" {
		s.jsonError(w, "geoserver_connection_id is required", http.StatusBadRequest)
		return
	}
	if req.Workspace == "" {
		s.jsonError(w, "workspace is required", http.StatusBadRequest)
		return
	}
	if req.StoreName == "" {
		s.jsonError(w, "store_name is required", http.StatusBadRequest)
		return
	}

	// Build options
	opts := integration.BridgeOptions{
		PGServiceName:         req.PGServiceName,
		GeoServerConnectionID: req.GeoServerConnectionID,
		Workspace:             req.Workspace,
		StoreName:             req.StoreName,
		Schema:                req.Schema,
		Tables:                req.Tables,
		PublishLayers:         req.PublishLayers,
	}

	if opts.Schema == "" {
		opts.Schema = "public"
	}

	// Get clients map
	s.clientsMu.RLock()
	clients := make(map[string]*interface{})
	for k := range s.clients {
		clients[k] = nil // Just checking if key exists
	}
	s.clientsMu.RUnlock()

	// Verify GeoServer connection exists
	client := s.getClient(req.GeoServerConnectionID)
	if client == nil {
		s.jsonError(w, "GeoServer connection not found", http.StatusNotFound)
		return
	}

	// Create bridge using integration package
	link, err := integration.CreateBridge(s.config, s.clients, opts)
	if err != nil {
		json.NewEncoder(w).Encode(BridgeResponse{
			Success: false,
			Message: "Failed to create bridge",
			Error:   err.Error(),
		})
		return
	}

	// Set creation time
	link.CreatedAt = time.Now().Format(time.RFC3339)

	json.NewEncoder(w).Encode(BridgeResponse{
		Success: true,
		Message: "Bridge created successfully",
		Link:    link,
	})
}

// handleGetAvailableTables returns tables available for publishing from a PostgreSQL service
func (s *Server) handleGetAvailableTables(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" {
		s.jsonError(w, "service query parameter is required", http.StatusBadRequest)
		return
	}

	tables, err := integration.GetAvailableTables(serviceName)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": serviceName,
		"tables":  tables,
		"count":   len(tables),
	})
}
