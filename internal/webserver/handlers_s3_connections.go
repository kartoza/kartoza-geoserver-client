package webserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/s3client"
)

// handleS3Connections handles GET /api/s3/connections and POST /api/s3/connections
func (s *Server) handleS3Connections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listS3Connections(w, r)
	case http.MethodPost:
		s.createS3Connection(w, r)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTestS3ConnectionDirect handles POST /api/s3/connections/test
func (s *Server) handleTestS3ConnectionDirect(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}
	if r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req S3ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Endpoint == "" || req.AccessKey == "" || req.SecretKey == "" {
		s.jsonError(w, "Endpoint, accessKey, and secretKey are required", http.StatusBadRequest)
		return
	}

	// Create a temporary client to test the connection
	client, err := s3client.NewClientDirect(req.Endpoint, req.AccessKey, req.SecretKey, req.Region, req.UseSSL, req.PathStyle)
	if err != nil {
		s.jsonResponse(w, S3TestConnectionResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.TestConnection(ctx)
	if err != nil {
		s.jsonResponse(w, S3TestConnectionResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	s.jsonResponse(w, S3TestConnectionResponse{
		Success:     result.Success,
		Message:     result.Message,
		BucketCount: result.BucketCount,
	})
}

// handleS3ConnectionByID handles requests to /api/s3/connections/{id}
func (s *Server) handleS3ConnectionByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/s3/connections/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		s.jsonError(w, "Connection ID required", http.StatusBadRequest)
		return
	}

	connID := parts[0]

	// Check if this is a test request
	if len(parts) >= 2 && parts[1] == "test" {
		if r.Method == http.MethodPost || r.Method == http.MethodGet {
			s.testS3Connection(w, r, connID)
			return
		}
	}

	// Check if this is a buckets request
	if len(parts) >= 2 && parts[1] == "buckets" {
		s.handleS3Buckets(w, r, connID, parts[2:])
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getS3Connection(w, r, connID)
	case http.MethodPut:
		s.updateS3Connection(w, r, connID)
	case http.MethodDelete:
		s.deleteS3Connection(w, r, connID)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listS3Connections returns all S3 connections
func (s *Server) listS3Connections(w http.ResponseWriter, r *http.Request) {
	connections := make([]S3ConnectionResponse, len(s.config.S3Connections))
	for i, conn := range s.config.S3Connections {
		connections[i] = S3ConnectionResponse{
			ID:        conn.ID,
			Name:      conn.Name,
			Endpoint:  conn.Endpoint,
			AccessKey: conn.AccessKey,
			Region:    conn.Region,
			UseSSL:    conn.UseSSL,
			PathStyle: conn.PathStyle,
			IsActive:  conn.IsActive,
		}
	}
	s.jsonResponse(w, connections)
}

// getS3Connection returns a single S3 connection by ID
func (s *Server) getS3Connection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetS3Connection(connID)
	if conn == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	s.jsonResponse(w, S3ConnectionResponse{
		ID:        conn.ID,
		Name:      conn.Name,
		Endpoint:  conn.Endpoint,
		AccessKey: conn.AccessKey,
		Region:    conn.Region,
		UseSSL:    conn.UseSSL,
		PathStyle: conn.PathStyle,
		IsActive:  conn.IsActive,
	})
}

// createS3Connection creates a new S3 connection
func (s *Server) createS3Connection(w http.ResponseWriter, r *http.Request) {
	var req S3ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Endpoint == "" || req.AccessKey == "" || req.SecretKey == "" {
		s.jsonError(w, "Name, endpoint, accessKey, and secretKey are required", http.StatusBadRequest)
		return
	}

	id := generateUniqueID("s3")

	conn := config.S3Connection{
		ID:        id,
		Name:      req.Name,
		Endpoint:  req.Endpoint,
		AccessKey: req.AccessKey,
		SecretKey: req.SecretKey,
		Region:    req.Region,
		UseSSL:    req.UseSSL,
		PathStyle: req.PathStyle,
	}

	s.config.AddS3Connection(conn)
	s.addS3Client(&conn)

	if err := s.saveConfig(); err != nil {
		s.jsonError(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	s.jsonResponse(w, S3ConnectionResponse{
		ID:        conn.ID,
		Name:      conn.Name,
		Endpoint:  conn.Endpoint,
		AccessKey: conn.AccessKey,
		Region:    conn.Region,
		UseSSL:    conn.UseSSL,
		PathStyle: conn.PathStyle,
		IsActive:  false,
	})
}

// updateS3Connection updates an existing S3 connection
func (s *Server) updateS3Connection(w http.ResponseWriter, r *http.Request, connID string) {
	var req S3ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn := s.config.GetS3Connection(connID)
	if conn == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.Name != "" {
		conn.Name = req.Name
	}
	if req.Endpoint != "" {
		conn.Endpoint = req.Endpoint
	}
	if req.AccessKey != "" {
		conn.AccessKey = req.AccessKey
	}
	if req.SecretKey != "" {
		conn.SecretKey = req.SecretKey
	}
	conn.Region = req.Region
	conn.UseSSL = req.UseSSL
	conn.PathStyle = req.PathStyle

	s.config.UpdateS3Connection(*conn)
	s.removeS3Client(connID)
	s.addS3Client(conn)

	if err := s.saveConfig(); err != nil {
		s.jsonError(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, S3ConnectionResponse{
		ID:        conn.ID,
		Name:      conn.Name,
		Endpoint:  conn.Endpoint,
		AccessKey: conn.AccessKey,
		Region:    conn.Region,
		UseSSL:    conn.UseSSL,
		PathStyle: conn.PathStyle,
		IsActive:  conn.IsActive,
	})
}

// deleteS3Connection deletes an S3 connection
func (s *Server) deleteS3Connection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetS3Connection(connID)
	if conn == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	s.config.RemoveS3Connection(connID)
	s.removeS3Client(connID)

	if err := s.saveConfig(); err != nil {
		s.jsonError(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// testS3Connection tests an S3 connection
func (s *Server) testS3Connection(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getS3Client(connID)
	if client == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.TestConnection(ctx)
	if err != nil {
		s.jsonResponse(w, S3TestConnectionResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	s.jsonResponse(w, S3TestConnectionResponse{
		Success:     result.Success,
		Message:     result.Message,
		BucketCount: result.BucketCount,
	})
}
