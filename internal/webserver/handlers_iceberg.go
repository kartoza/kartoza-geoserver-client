package webserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/iceberg"
)

// IcebergConnectionResponse represents an Iceberg connection in API responses
type IcebergConnectionResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Prefix     string `json:"prefix,omitempty"`
	S3Endpoint string `json:"s3Endpoint,omitempty"`
	AccessKey  string `json:"accessKey,omitempty"`
	Region     string `json:"region,omitempty"`
	IsActive   bool   `json:"isActive"`
}

// IcebergConnectionRequest represents an Iceberg connection create/update request
type IcebergConnectionRequest struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Prefix     string `json:"prefix,omitempty"`
	S3Endpoint string `json:"s3Endpoint,omitempty"`
	AccessKey  string `json:"accessKey,omitempty"`
	SecretKey  string `json:"secretKey,omitempty"`
	Region     string `json:"region,omitempty"`
}

// IcebergTestConnectionResponse represents the response from testing an Iceberg connection
type IcebergTestConnectionResponse struct {
	Success        bool              `json:"success"`
	Message        string            `json:"message"`
	NamespaceCount int               `json:"namespaceCount,omitempty"`
	Defaults       map[string]string `json:"defaults,omitempty"`
}

// IcebergNamespaceResponse represents a namespace in API responses
type IcebergNamespaceResponse struct {
	Name       string            `json:"name"`
	Path       []string          `json:"path"`
	Properties map[string]string `json:"properties,omitempty"`
}

// IcebergTableResponse represents a table in API responses
type IcebergTableResponse struct {
	Namespace       string   `json:"namespace"`
	Name            string   `json:"name"`
	Location        string   `json:"location,omitempty"`
	FormatVersion   int      `json:"formatVersion,omitempty"`
	RowCount        int64    `json:"rowCount,omitempty"`
	SnapshotCount   int      `json:"snapshotCount,omitempty"`
	LastUpdatedMS   int64    `json:"lastUpdatedMs,omitempty"`
	HasGeometry     bool     `json:"hasGeometry,omitempty"`
	GeometryColumns []string `json:"geometryColumns,omitempty"`
}

// IcebergSchemaResponse represents a table schema in API responses
type IcebergSchemaResponse struct {
	SchemaID int                         `json:"schemaId"`
	Type     string                      `json:"type"`
	Fields   []IcebergFieldResponse      `json:"fields"`
}

// IcebergFieldResponse represents a field in a schema
type IcebergFieldResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Doc      string `json:"doc,omitempty"`
}

// IcebergSnapshotResponse represents a snapshot in API responses
type IcebergSnapshotResponse struct {
	SnapshotID     int64             `json:"snapshotId"`
	SequenceNumber int64             `json:"sequenceNumber"`
	TimestampMS    int64             `json:"timestampMs"`
	Summary        map[string]string `json:"summary,omitempty"`
	ParentID       *int64            `json:"parentId,omitempty"`
}

// handleIcebergConnections handles GET /api/iceberg/connections and POST /api/iceberg/connections
func (s *Server) handleIcebergConnections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listIcebergConnections(w, r)
	case http.MethodPost:
		s.createIcebergConnection(w, r)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTestIcebergConnectionDirect handles POST /api/iceberg/connections/test
func (s *Server) handleTestIcebergConnectionDirect(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}
	if r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IcebergConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		s.jsonError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Create a temporary client to test the connection
	client, err := iceberg.NewClient(iceberg.ClientConfig{
		BaseURL: req.URL,
		Prefix:  req.Prefix,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		s.jsonResponse(w, IcebergTestConnectionResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := client.TestConnection(ctx)
	if err != nil {
		s.jsonResponse(w, IcebergTestConnectionResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	s.jsonResponse(w, IcebergTestConnectionResponse{
		Success:        result.Success,
		Message:        result.Message,
		NamespaceCount: result.NamespaceCount,
		Defaults:       result.Defaults,
	})
}

// handleIcebergConnectionByID handles requests to /api/iceberg/connections/{id}
func (s *Server) handleIcebergConnectionByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/iceberg/connections/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		s.jsonError(w, "Connection ID required", http.StatusBadRequest)
		return
	}

	connID := parts[0]

	// Check if this is a test request
	if len(parts) >= 2 && parts[1] == "test" {
		if r.Method == http.MethodPost || r.Method == http.MethodGet {
			s.testIcebergConnection(w, r, connID)
			return
		}
	}

	// Check if this is a namespaces request
	if len(parts) >= 2 && parts[1] == "namespaces" {
		s.handleIcebergNamespaces(w, r, connID, parts[2:])
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getIcebergConnection(w, r, connID)
	case http.MethodPut:
		s.updateIcebergConnection(w, r, connID)
	case http.MethodDelete:
		s.deleteIcebergConnection(w, r, connID)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listIcebergConnections returns all Iceberg connections
func (s *Server) listIcebergConnections(w http.ResponseWriter, r *http.Request) {
	connections := make([]IcebergConnectionResponse, len(s.config.IcebergConnections))
	for i, conn := range s.config.IcebergConnections {
		connections[i] = IcebergConnectionResponse{
			ID:         conn.ID,
			Name:       conn.Name,
			URL:        conn.URL,
			Prefix:     conn.Prefix,
			S3Endpoint: conn.S3Endpoint,
			AccessKey:  conn.AccessKey,
			Region:     conn.Region,
			IsActive:   conn.IsActive,
		}
	}
	s.jsonResponse(w, connections)
}

// getIcebergConnection returns a single Iceberg connection by ID
func (s *Server) getIcebergConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetIcebergConnection(connID)
	if conn == nil {
		s.jsonError(w, "Iceberg connection not found", http.StatusNotFound)
		return
	}

	s.jsonResponse(w, IcebergConnectionResponse{
		ID:         conn.ID,
		Name:       conn.Name,
		URL:        conn.URL,
		Prefix:     conn.Prefix,
		S3Endpoint: conn.S3Endpoint,
		AccessKey:  conn.AccessKey,
		Region:     conn.Region,
		IsActive:   conn.IsActive,
	})
}

// createIcebergConnection creates a new Iceberg connection
func (s *Server) createIcebergConnection(w http.ResponseWriter, r *http.Request) {
	var req IcebergConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.URL == "" {
		s.jsonError(w, "Name and URL are required", http.StatusBadRequest)
		return
	}

	id := generateUniqueID("iceberg")

	conn := config.IcebergCatalogConnection{
		ID:         id,
		Name:       req.Name,
		URL:        req.URL,
		Prefix:     req.Prefix,
		S3Endpoint: req.S3Endpoint,
		AccessKey:  req.AccessKey,
		SecretKey:  req.SecretKey,
		Region:     req.Region,
	}

	s.config.AddIcebergConnection(conn)
	s.addIcebergClient(&conn)

	if err := s.saveConfig(); err != nil {
		s.jsonError(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	s.jsonResponse(w, IcebergConnectionResponse{
		ID:         conn.ID,
		Name:       conn.Name,
		URL:        conn.URL,
		Prefix:     conn.Prefix,
		S3Endpoint: conn.S3Endpoint,
		AccessKey:  conn.AccessKey,
		Region:     conn.Region,
		IsActive:   false,
	})
}

// updateIcebergConnection updates an existing Iceberg connection
func (s *Server) updateIcebergConnection(w http.ResponseWriter, r *http.Request, connID string) {
	var req IcebergConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn := s.config.GetIcebergConnection(connID)
	if conn == nil {
		s.jsonError(w, "Iceberg connection not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.Name != "" {
		conn.Name = req.Name
	}
	if req.URL != "" {
		conn.URL = req.URL
	}
	conn.Prefix = req.Prefix
	conn.S3Endpoint = req.S3Endpoint
	if req.AccessKey != "" {
		conn.AccessKey = req.AccessKey
	}
	if req.SecretKey != "" {
		conn.SecretKey = req.SecretKey
	}
	conn.Region = req.Region

	s.config.UpdateIcebergConnection(*conn)
	s.removeIcebergClient(connID)
	s.addIcebergClient(conn)

	if err := s.saveConfig(); err != nil {
		s.jsonError(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, IcebergConnectionResponse{
		ID:         conn.ID,
		Name:       conn.Name,
		URL:        conn.URL,
		Prefix:     conn.Prefix,
		S3Endpoint: conn.S3Endpoint,
		AccessKey:  conn.AccessKey,
		Region:     conn.Region,
		IsActive:   conn.IsActive,
	})
}

// deleteIcebergConnection deletes an Iceberg connection
func (s *Server) deleteIcebergConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetIcebergConnection(connID)
	if conn == nil {
		s.jsonError(w, "Iceberg connection not found", http.StatusNotFound)
		return
	}

	s.config.RemoveIcebergConnection(connID)
	s.removeIcebergClient(connID)

	if err := s.saveConfig(); err != nil {
		s.jsonError(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// testIcebergConnection tests an Iceberg connection
func (s *Server) testIcebergConnection(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getIcebergClient(connID)
	if client == nil {
		s.jsonError(w, "Iceberg connection not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := client.TestConnection(ctx)
	if err != nil {
		s.jsonResponse(w, IcebergTestConnectionResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	s.jsonResponse(w, IcebergTestConnectionResponse{
		Success:        result.Success,
		Message:        result.Message,
		NamespaceCount: result.NamespaceCount,
		Defaults:       result.Defaults,
	})
}

// handleIcebergNamespaces handles namespace operations
func (s *Server) handleIcebergNamespaces(w http.ResponseWriter, r *http.Request, connID string, pathParts []string) {
	client := s.getIcebergClient(connID)
	if client == nil {
		s.jsonError(w, "Iceberg connection not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// /api/iceberg/connections/{connId}/namespaces
	if len(pathParts) == 0 || pathParts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			s.listIcebergNamespaces(w, r, client, ctx)
		case http.MethodPost:
			s.createIcebergNamespace(w, r, client, ctx)
		case http.MethodOptions:
			s.handleCORS(w)
		default:
			s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	namespace := pathParts[0]

	// /api/iceberg/connections/{connId}/namespaces/{namespace}/tables
	if len(pathParts) >= 2 && pathParts[1] == "tables" {
		s.handleIcebergTables(w, r, client, ctx, namespace, pathParts[2:])
		return
	}

	// /api/iceberg/connections/{connId}/namespaces/{namespace}
	switch r.Method {
	case http.MethodGet:
		s.getIcebergNamespace(w, r, client, ctx, namespace)
	case http.MethodDelete:
		s.deleteIcebergNamespace(w, r, client, ctx, namespace)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listIcebergNamespaces lists all namespaces
func (s *Server) listIcebergNamespaces(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context) {
	parent := r.URL.Query().Get("parent")

	namespaces, err := client.ListNamespaces(ctx, parent)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]IcebergNamespaceResponse, len(namespaces))
	for i, ns := range namespaces {
		response[i] = IcebergNamespaceResponse{
			Name: ns.Name,
			Path: ns.Path,
		}
	}

	s.jsonResponse(w, response)
}

// getIcebergNamespace gets a namespace's details
func (s *Server) getIcebergNamespace(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context, namespace string) {
	info, err := client.GetNamespace(ctx, namespace)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.jsonResponse(w, IcebergNamespaceResponse{
		Name:       strings.Join(info.Namespace, "."),
		Path:       info.Namespace,
		Properties: info.Properties,
	})
}

// createIcebergNamespace creates a new namespace
func (s *Server) createIcebergNamespace(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context) {
	var req struct {
		Name       string            `json:"name"`
		Properties map[string]string `json:"properties,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.jsonError(w, "Namespace name is required", http.StatusBadRequest)
		return
	}

	if err := client.CreateNamespace(ctx, req.Name, req.Properties); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	s.jsonResponse(w, IcebergNamespaceResponse{
		Name:       req.Name,
		Path:       strings.Split(req.Name, "."),
		Properties: req.Properties,
	})
}

// deleteIcebergNamespace deletes a namespace
func (s *Server) deleteIcebergNamespace(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context, namespace string) {
	if err := client.DropNamespace(ctx, namespace); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleIcebergTables handles table operations
func (s *Server) handleIcebergTables(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context, namespace string, pathParts []string) {
	// /api/iceberg/connections/{connId}/namespaces/{namespace}/tables
	if len(pathParts) == 0 || pathParts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			s.listIcebergTables(w, r, client, ctx, namespace)
		case http.MethodOptions:
			s.handleCORS(w)
		default:
			s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	tableName := pathParts[0]

	// /api/iceberg/connections/{connId}/namespaces/{namespace}/tables/{table}/snapshots
	if len(pathParts) >= 2 && pathParts[1] == "snapshots" {
		if r.Method == http.MethodGet {
			s.listIcebergSnapshots(w, r, client, ctx, namespace, tableName)
			return
		}
	}

	// /api/iceberg/connections/{connId}/namespaces/{namespace}/tables/{table}/schema
	if len(pathParts) >= 2 && pathParts[1] == "schema" {
		if r.Method == http.MethodGet {
			s.getIcebergTableSchema(w, r, client, ctx, namespace, tableName)
			return
		}
	}

	// /api/iceberg/connections/{connId}/namespaces/{namespace}/tables/{table}
	switch r.Method {
	case http.MethodGet:
		s.getIcebergTable(w, r, client, ctx, namespace, tableName)
	case http.MethodDelete:
		s.deleteIcebergTable(w, r, client, ctx, namespace, tableName)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listIcebergTables lists all tables in a namespace
func (s *Server) listIcebergTables(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context, namespace string) {
	tables, err := client.ListTables(ctx, namespace)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]IcebergTableResponse, len(tables))
	for i, t := range tables {
		response[i] = IcebergTableResponse{
			Namespace: strings.Join(t.Namespace, "."),
			Name:      t.Name,
		}
	}

	s.jsonResponse(w, response)
}

// getIcebergTable gets a table's metadata
func (s *Server) getIcebergTable(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context, namespace, tableName string) {
	metadata, err := client.GetTable(ctx, namespace, tableName)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check for geometry columns
	var geometryColumns []string
	hasGeometry := false
	if len(metadata.Metadata.Schemas) > 0 {
		currentSchemaID := metadata.Metadata.CurrentSchemaID
		for _, schema := range metadata.Metadata.Schemas {
			if schema.SchemaID == currentSchemaID {
				for _, field := range schema.Fields {
					fieldType := formatFieldType(field.Type)
					if strings.Contains(strings.ToLower(fieldType), "geometry") ||
						strings.Contains(strings.ToLower(field.Name), "geom") {
						geometryColumns = append(geometryColumns, field.Name)
						hasGeometry = true
					}
				}
				break
			}
		}
	}

	// Extract stats from latest snapshot
	var rowCount int64
	if metadata.Metadata.CurrentSnapshot != nil {
		for _, snap := range metadata.Metadata.Snapshots {
			if snap.SnapshotID == *metadata.Metadata.CurrentSnapshot {
				if countStr, ok := snap.Summary["total-records"]; ok {
					// Parse the row count (it's a string representation of a number)
					var count int64
					if err := json.Unmarshal([]byte(countStr), &count); err == nil {
						rowCount = count
					}
				}
				break
			}
		}
	}

	s.jsonResponse(w, IcebergTableResponse{
		Namespace:       namespace,
		Name:            tableName,
		Location:        metadata.Metadata.Location,
		FormatVersion:   metadata.Metadata.FormatVersion,
		RowCount:        rowCount,
		SnapshotCount:   len(metadata.Metadata.Snapshots),
		LastUpdatedMS:   metadata.Metadata.LastUpdatedMS,
		HasGeometry:     hasGeometry,
		GeometryColumns: geometryColumns,
	})
}

// getIcebergTableSchema gets a table's schema
func (s *Server) getIcebergTableSchema(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context, namespace, tableName string) {
	metadata, err := client.GetTable(ctx, namespace, tableName)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Find current schema
	var currentSchema *iceberg.Schema
	for i, schema := range metadata.Metadata.Schemas {
		if schema.SchemaID == metadata.Metadata.CurrentSchemaID {
			currentSchema = &metadata.Metadata.Schemas[i]
			break
		}
	}

	if currentSchema == nil && len(metadata.Metadata.Schemas) > 0 {
		currentSchema = &metadata.Metadata.Schemas[0]
	}

	if currentSchema == nil {
		s.jsonError(w, "No schema found", http.StatusNotFound)
		return
	}

	// Convert fields
	fields := make([]IcebergFieldResponse, len(currentSchema.Fields))
	for i, f := range currentSchema.Fields {
		fields[i] = IcebergFieldResponse{
			ID:       f.ID,
			Name:     f.Name,
			Type:     formatFieldType(f.Type),
			Required: f.Required,
			Doc:      f.Doc,
		}
	}

	s.jsonResponse(w, IcebergSchemaResponse{
		SchemaID: currentSchema.SchemaID,
		Type:     currentSchema.Type,
		Fields:   fields,
	})
}

// deleteIcebergTable deletes a table
func (s *Server) deleteIcebergTable(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context, namespace, tableName string) {
	purge := r.URL.Query().Get("purge") == "true"

	if err := client.DropTable(ctx, namespace, tableName, purge); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// listIcebergSnapshots lists all snapshots for a table
func (s *Server) listIcebergSnapshots(w http.ResponseWriter, r *http.Request, client *iceberg.Client, ctx context.Context, namespace, tableName string) {
	snapshots, err := client.GetTableSnapshots(ctx, namespace, tableName)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]IcebergSnapshotResponse, len(snapshots))
	for i, snap := range snapshots {
		response[i] = IcebergSnapshotResponse{
			SnapshotID:     snap.SnapshotID,
			SequenceNumber: snap.SequenceNumber,
			TimestampMS:    snap.TimestampMS,
			Summary:        snap.Summary,
			ParentID:       snap.ParentSnapshotID,
		}
	}

	s.jsonResponse(w, response)
}

// formatFieldType converts Iceberg field type to string
func formatFieldType(t interface{}) string {
	switch v := t.(type) {
	case string:
		return v
	case map[string]interface{}:
		if typeStr, ok := v["type"].(string); ok {
			if typeStr == "list" {
				if elemType, ok := v["element-type"].(string); ok {
					return "list<" + elemType + ">"
				}
			}
			if typeStr == "map" {
				keyType := "?"
				valType := "?"
				if kt, ok := v["key-type"].(string); ok {
					keyType = kt
				}
				if vt, ok := v["value-type"].(string); ok {
					valType = vt
				}
				return "map<" + keyType + "," + valType + ">"
			}
			if typeStr == "struct" {
				return "struct"
			}
			return typeStr
		}
		return "complex"
	default:
		return "unknown"
	}
}
