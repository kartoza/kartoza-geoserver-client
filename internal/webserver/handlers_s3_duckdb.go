package webserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/cloudnative"
)

// handleS3DuckDBQuery handles DuckDB SQL queries against Parquet/GeoParquet files
// Pattern: POST /api/s3/duckdb/{connectionId}/{bucket}?key=object/path
func (s *Server) handleS3DuckDBQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}

	// Parse path: /api/s3/duckdb/{connectionId}/{bucket}
	path := strings.TrimPrefix(r.URL.Path, "/api/s3/duckdb/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		s.jsonError(w, "Connection ID and bucket required", http.StatusBadRequest)
		return
	}

	connID := parts[0]
	bucket := parts[1]
	objectKey := r.URL.Query().Get("key")
	if objectKey == "" {
		s.jsonError(w, "Object key required", http.StatusBadRequest)
		return
	}

	client := s.getS3Client(connID)
	if client == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	// Check if DuckDB is available
	available, version := cloudnative.CheckDuckDBAvailable()
	if !available {
		s.jsonError(w, "DuckDB is not available on the server", http.StatusServiceUnavailable)
		return
	}
	log.Printf("DuckDB available: %s", version)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Download file from S3 to temp
	tempFile, err := s.downloadS3ToTemp(ctx, client, bucket, objectKey)
	if err != nil {
		s.jsonError(w, "Failed to download file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile)

	switch r.Method {
	case http.MethodGet:
		// GET returns table metadata and sample queries
		s.handleDuckDBTableInfo(w, r, ctx, tempFile, objectKey)
	case http.MethodPost:
		// POST executes a query
		s.handleDuckDBExecuteQuery(w, r, ctx, tempFile, objectKey)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleDuckDBTableInfo returns metadata about a Parquet file
func (s *Server) handleDuckDBTableInfo(w http.ResponseWriter, r *http.Request, ctx context.Context, tempFile, objectKey string) {
	info, err := cloudnative.GetParquetTableInfo(ctx, tempFile)
	if err != nil {
		s.jsonError(w, "Failed to get table info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	response := DuckDBTableInfoResponse{
		Columns:        make([]DuckDBColumnInfoResponse, len(info.Columns)),
		RowCount:       info.RowCount,
		GeometryColumn: info.GeometryColumn,
		BBox:           info.BBox,
		SampleQueries:  cloudnative.GetSampleQueries("'"+tempFile+"'", info),
	}

	for i, col := range info.Columns {
		response.Columns[i] = DuckDBColumnInfoResponse{
			Name: col.Name,
			Type: col.Type,
		}
	}

	// Replace temp file path with generic table reference in sample queries
	for i, q := range response.SampleQueries {
		response.SampleQueries[i] = strings.ReplaceAll(q, "'"+tempFile+"'", "'data'")
	}

	s.jsonResponse(w, response)
}

// handleDuckDBExecuteQuery executes a DuckDB query against a Parquet file
func (s *Server) handleDuckDBExecuteQuery(w http.ResponseWriter, r *http.Request, ctx context.Context, tempFile, objectKey string) {
	var req DuckDBQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SQL == "" {
		s.jsonError(w, "SQL query is required", http.StatusBadRequest)
		return
	}

	// Validate SQL (prevent dangerous operations)
	if err := cloudnative.ValidateSQL(req.SQL); err != nil {
		s.jsonResponse(w, DuckDBQueryResponse{
			Error: err.Error(),
		})
		return
	}

	// Replace 'data' table reference with actual file path
	sql := strings.ReplaceAll(req.SQL, "'data'", "'"+tempFile+"'")
	sql = strings.ReplaceAll(sql, "\"data\"", "'"+tempFile+"'")
	sql = strings.ReplaceAll(sql, " data ", " '"+tempFile+"' ")
	sql = strings.ReplaceAll(sql, " data;", " '"+tempFile+"';")
	sql = strings.ReplaceAll(sql, "(data)", "('"+tempFile+"')")
	// Handle "FROM data" pattern
	sql = strings.ReplaceAll(sql, "FROM data", "FROM '"+tempFile+"'")
	sql = strings.ReplaceAll(sql, "from data", "from '"+tempFile+"'")

	opts := cloudnative.DuckDBQueryOptions{
		SQL:    sql,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	result, err := cloudnative.QueryParquetFile(ctx, tempFile, opts)
	if err != nil {
		s.jsonResponse(w, DuckDBQueryResponse{
			Error: err.Error(),
			SQL:   req.SQL,
		})
		return
	}

	response := DuckDBQueryResponse{
		Columns:        result.Columns,
		Rows:           result.Rows,
		RowCount:       result.RowCount,
		HasMore:        result.HasMore,
		GeometryColumn: result.GeometryColumn,
		SQL:            req.SQL,
	}

	s.jsonResponse(w, response)
}

// handleS3DuckDBGeoJSON executes a DuckDB query and returns results as GeoJSON
// Pattern: POST /api/s3/duckdb/geojson/{connectionId}/{bucket}?key=object/path
func (s *Server) handleS3DuckDBGeoJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}

	if r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/s3/duckdb/geojson/{connectionId}/{bucket}
	path := strings.TrimPrefix(r.URL.Path, "/api/s3/duckdb/geojson/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		s.jsonError(w, "Connection ID and bucket required", http.StatusBadRequest)
		return
	}

	connID := parts[0]
	bucket := parts[1]
	objectKey := r.URL.Query().Get("key")
	if objectKey == "" {
		s.jsonError(w, "Object key required", http.StatusBadRequest)
		return
	}

	var req DuckDBQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	client := s.getS3Client(connID)
	if client == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Download file from S3 to temp
	tempFile, err := s.downloadS3ToTemp(ctx, client, bucket, objectKey)
	if err != nil {
		s.jsonError(w, "Failed to download file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile)

	// Validate SQL
	if req.SQL != "" {
		if err := cloudnative.ValidateSQL(req.SQL); err != nil {
			s.jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Replace 'data' table reference with actual file path
	sql := req.SQL
	if sql != "" {
		sql = strings.ReplaceAll(sql, "'data'", "'"+tempFile+"'")
		sql = strings.ReplaceAll(sql, "\"data\"", "'"+tempFile+"'")
		sql = strings.ReplaceAll(sql, " data ", " '"+tempFile+"' ")
		sql = strings.ReplaceAll(sql, " data;", " '"+tempFile+"';")
		sql = strings.ReplaceAll(sql, "(data)", "('"+tempFile+"')")
		sql = strings.ReplaceAll(sql, "FROM data", "FROM '"+tempFile+"'")
		sql = strings.ReplaceAll(sql, "from data", "from '"+tempFile+"'")
	}

	opts := cloudnative.DuckDBQueryOptions{
		SQL:    sql,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	geojson, err := cloudnative.QueryParquetFileAsGeoJSON(ctx, tempFile, opts)
	if err != nil {
		s.jsonError(w, "Failed to generate GeoJSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/geo+json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(geojson)
}
