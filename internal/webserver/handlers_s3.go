// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/cloudnative"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/s3client"
)

// S3ConnectionResponse represents an S3 connection in API responses
type S3ConnectionResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	Region    string `json:"region,omitempty"`
	UseSSL    bool   `json:"useSSL"`
	PathStyle bool   `json:"pathStyle"`
	IsActive  bool   `json:"isActive"`
}

// S3ConnectionRequest represents an S3 connection create/update request
type S3ConnectionRequest struct {
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region,omitempty"`
	UseSSL    bool   `json:"useSSL"`
	PathStyle bool   `json:"pathStyle"`
}

// S3BucketResponse represents a bucket in API responses
type S3BucketResponse struct {
	Name         string `json:"name"`
	CreationDate string `json:"creationDate"`
}

// S3ObjectResponse represents an object in API responses
type S3ObjectResponse struct {
	Key             string `json:"key"`
	Size            int64  `json:"size"`
	LastModified    string `json:"lastModified"`
	ContentType     string `json:"contentType,omitempty"`
	IsFolder        bool   `json:"isFolder"`
	CloudNativeType string `json:"cloudNativeType,omitempty"` // "cog", "copc", "geoparquet", or ""
}

// S3TestConnectionResponse represents the response from testing an S3 connection
type S3TestConnectionResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	BucketCount int    `json:"bucketCount,omitempty"`
}

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

// handleS3Buckets handles bucket operations
func (s *Server) handleS3Buckets(w http.ResponseWriter, r *http.Request, connID string, pathParts []string) {
	client := s.getS3Client(connID)
	if client == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	// /api/s3/connections/{connId}/buckets
	if len(pathParts) == 0 || pathParts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			s.listS3Buckets(w, r, client)
		case http.MethodPost:
			s.createS3Bucket(w, r, client)
		case http.MethodOptions:
			s.handleCORS(w)
		default:
			s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	bucketName := pathParts[0]

	// /api/s3/connections/{connId}/buckets/{bucket}/objects
	if len(pathParts) >= 2 && pathParts[1] == "objects" {
		s.handleS3Objects(w, r, client, bucketName)
		return
	}

	// /api/s3/connections/{connId}/buckets/{bucket}/presign
	if len(pathParts) >= 2 && pathParts[1] == "presign" {
		s.handleS3Presign(w, r, client, bucketName)
		return
	}

	// /api/s3/connections/{connId}/buckets/{bucket}
	switch r.Method {
	case http.MethodDelete:
		s.deleteS3Bucket(w, r, client, bucketName)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listS3Buckets lists all buckets
func (s *Server) listS3Buckets(w http.ResponseWriter, r *http.Request, client *s3client.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]S3BucketResponse, len(buckets))
	for i, b := range buckets {
		response[i] = S3BucketResponse{
			Name:         b.Name,
			CreationDate: b.CreationDate.Format(time.RFC3339),
		}
	}

	s.jsonResponse(w, response)
}

// createS3Bucket creates a new bucket
func (s *Server) createS3Bucket(w http.ResponseWriter, r *http.Request, client *s3client.Client) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.jsonError(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.CreateBucket(ctx, req.Name); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	s.jsonResponse(w, S3BucketResponse{
		Name:         req.Name,
		CreationDate: time.Now().Format(time.RFC3339),
	})
}

// deleteS3Bucket deletes a bucket
func (s *Server) deleteS3Bucket(w http.ResponseWriter, r *http.Request, client *s3client.Client, bucketName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.DeleteBucket(ctx, bucketName); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleS3Objects handles object operations
func (s *Server) handleS3Objects(w http.ResponseWriter, r *http.Request, client *s3client.Client, bucketName string) {
	switch r.Method {
	case http.MethodGet:
		s.listS3Objects(w, r, client, bucketName)
	case http.MethodPost:
		s.uploadS3Object(w, r, client, bucketName)
	case http.MethodDelete:
		s.deleteS3Object(w, r, client, bucketName)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listS3Objects lists objects in a bucket
func (s *Server) listS3Objects(w http.ResponseWriter, r *http.Request, client *s3client.Client, bucketName string) {
	prefix := r.URL.Query().Get("prefix")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	objects, err := client.ListObjects(ctx, bucketName, prefix)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]S3ObjectResponse, len(objects))
	for i, obj := range objects {
		// Detect cloud-native format
		cnFormat := s3client.DetectCloudNativeFormat(obj.Key)
		cnType := ""
		if cnFormat != s3client.FormatUnknown {
			cnType = string(cnFormat)
		}

		response[i] = S3ObjectResponse{
			Key:             obj.Key,
			Size:            obj.Size,
			LastModified:    obj.LastModified.Format(time.RFC3339),
			ContentType:     obj.ContentType,
			IsFolder:        obj.IsDir,
			CloudNativeType: cnType,
		}
	}

	s.jsonResponse(w, response)
}

// uploadS3Object handles file upload to S3
func (s *Server) uploadS3Object(w http.ResponseWriter, r *http.Request, client *s3client.Client, bucketName string) {
	// Parse multipart form
	if err := r.ParseMultipartForm(500 << 20); err != nil { // 500MB max
		s.jsonError(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.jsonError(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get options from form
	targetKey := r.FormValue("key")
	if targetKey == "" {
		targetKey = header.Filename
	}

	convertToCloudNative := r.FormValue("convert") == "true"
	createSubfolder := r.FormValue("subfolder") == "true" // For GeoPackage layer extraction
	prefix := r.FormValue("prefix")                       // Current folder prefix

	// Debug logging
	log.Printf("[S3 Upload] File: %s, Convert: %v, Subfolder: %v, Prefix: %s",
		header.Filename, convertToCloudNative, createSubfolder, prefix)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// If conversion requested, save to temp file first
	if convertToCloudNative {
		ext := strings.ToLower(filepath.Ext(header.Filename))
		log.Printf("[S3 Upload] Extension detected: %s", ext)

		// Special handling for GeoPackage - extract all layers as GeoParquet/Parquet
		if ext == ".gpkg" {
			log.Printf("[S3 Upload] Processing GeoPackage file")
			// Save to temp file
			tempDir := os.TempDir()
			tempPath := filepath.Join(tempDir, header.Filename)
			log.Printf("[S3 Upload] Saving to temp path: %s", tempPath)
			tempFile, err := os.Create(tempPath)
			if err != nil {
				log.Printf("[S3 Upload] ERROR: Failed to create temp file: %v", err)
				s.jsonError(w, "Failed to create temp file", http.StatusInternalServerError)
				return
			}
			defer os.Remove(tempPath)

			if _, err := io.Copy(tempFile, file); err != nil {
				tempFile.Close()
				log.Printf("[S3 Upload] ERROR: Failed to save temp file: %v", err)
				s.jsonError(w, "Failed to save temp file", http.StatusInternalServerError)
				return
			}
			tempFile.Close()
			log.Printf("[S3 Upload] Temp file saved successfully")

			// Get detailed info about layers in the GeoPackage (including geometry detection)
			log.Printf("[S3 Upload] Getting layer info from GeoPackage")
			layerInfos, err := cloudnative.GetGeoPackageLayerInfo(ctx, tempPath)
			if err != nil {
				log.Printf("[S3 Upload] ERROR: Failed to get layer info: %v", err)
				s.jsonError(w, "Failed to get layer info: "+err.Error(), http.StatusInternalServerError)
				return
			}

			log.Printf("[S3 Upload] Found %d layers in GeoPackage", len(layerInfos))
			for i, layer := range layerInfos {
				log.Printf("[S3 Upload]   Layer %d: %s (geom: %s, hasGeometry: %v)",
					i, layer.Name, layer.GeometryType, layer.HasGeometry)
			}

			if len(layerInfos) == 0 {
				log.Printf("[S3 Upload] ERROR: No layers found in GeoPackage")
				s.jsonError(w, "No layers found in GeoPackage", http.StatusBadRequest)
				return
			}

			// Determine output prefix
			baseName := strings.TrimSuffix(header.Filename, ext)
			var outputPrefix string
			if createSubfolder {
				// Create subfolder named after the gpkg
				if prefix != "" {
					outputPrefix = prefix + baseName + "/"
				} else {
					outputPrefix = baseName + "/"
				}
			} else {
				// Place files directly in current folder
				outputPrefix = prefix
			}

			// Convert each layer - spatial layers to GeoParquet, non-spatial to Parquet
			opts := cloudnative.DefaultConversionOptions()
			var uploadedFiles []map[string]interface{}

			log.Printf("[S3 Upload] Starting conversion of %d layers", len(layerInfos))

			for _, layerInfo := range layerInfos {
				var fileExt string
				var contentType string
				var convErr error
				var outputPath string

				if layerInfo.HasGeometry {
					// Spatial layer -> GeoParquet (.geoparquet)
					fileExt = ".geoparquet"
					contentType = "application/vnd.apache.parquet"
					outputPath = filepath.Join(tempDir, layerInfo.Name+fileExt)
					log.Printf("[S3 Upload] Converting layer '%s' to GeoParquet: %s", layerInfo.Name, outputPath)
					convErr = cloudnative.ConvertGeoPackageLayerToGeoParquet(ctx, tempPath, layerInfo.Name, outputPath, opts, nil)
				} else {
					// Non-spatial table -> Parquet (.parquet)
					fileExt = ".parquet"
					contentType = "application/vnd.apache.parquet"
					outputPath = filepath.Join(tempDir, layerInfo.Name+fileExt)
					log.Printf("[S3 Upload] Converting layer '%s' to Parquet: %s", layerInfo.Name, outputPath)
					convErr = cloudnative.ConvertGeoPackageLayerToParquet(ctx, tempPath, layerInfo.Name, outputPath, opts, nil)
				}

				if convErr != nil {
					// Skip layers that fail conversion
					log.Printf("[S3 Upload] ERROR: Failed to convert layer '%s': %v", layerInfo.Name, convErr)
					continue
				}
				log.Printf("[S3 Upload] Successfully converted layer '%s'", layerInfo.Name)

				// Upload converted file
				convertedFile, err := os.Open(outputPath)
				if err != nil {
					log.Printf("[S3 Upload] ERROR: Failed to open converted file '%s': %v", outputPath, err)
					continue
				}

				stat, _ := convertedFile.Stat()
				targetKey := outputPrefix + layerInfo.Name + fileExt
				log.Printf("[S3 Upload] Uploading '%s' (%d bytes) to S3 key: %s", layerInfo.Name, stat.Size(), targetKey)

				putOpts := s3client.PutOptions{
					ContentType: contentType,
				}

				if err := client.PutObject(ctx, bucketName, targetKey, convertedFile, stat.Size(), putOpts); err != nil {
					convertedFile.Close()
					log.Printf("[S3 Upload] ERROR: Failed to upload '%s': %v", targetKey, err)
					continue
				}
				convertedFile.Close()
				os.Remove(outputPath) // Clean up immediately after upload

				log.Printf("[S3 Upload] Successfully uploaded '%s'", targetKey)

				uploadedFiles = append(uploadedFiles, map[string]interface{}{
					"layer":       layerInfo.Name,
					"key":         targetKey,
					"size":        stat.Size(),
					"hasGeometry": layerInfo.HasGeometry,
					"format":      fileExt[1:], // Remove leading dot
				})
			}

			log.Printf("[S3 Upload] Completed: %d layers converted and uploaded", len(uploadedFiles))

			s.jsonResponse(w, map[string]interface{}{
				"success":         true,
				"converted":       true,
				"format":          "geoparquet/parquet",
				"gpkgExtracted":   true,
				"layerCount":      len(uploadedFiles),
				"files":           uploadedFiles,
				"createSubfolder": createSubfolder,
			})
			return
		}

		// Standard conversion for other file types
		convType, shouldConvert := cloudnative.DetectRecommendedConversion(header.Filename)
		if shouldConvert && convType != cloudnative.ConversionNone {
			// Save to temp file
			tempDir := os.TempDir()
			tempPath := filepath.Join(tempDir, header.Filename)
			tempFile, err := os.Create(tempPath)
			if err != nil {
				s.jsonError(w, "Failed to create temp file", http.StatusInternalServerError)
				return
			}
			defer os.Remove(tempPath)

			if _, err := io.Copy(tempFile, file); err != nil {
				tempFile.Close()
				s.jsonError(w, "Failed to save temp file", http.StatusInternalServerError)
				return
			}
			tempFile.Close()

			// Convert
			outputPath := cloudnative.GenerateOutputPath(tempPath, convType)
			defer os.Remove(outputPath)

			opts := cloudnative.DefaultConversionOptions()
			var convErr error

			switch convType {
			case cloudnative.ConversionCOG:
				convErr = cloudnative.ConvertToCOG(ctx, tempPath, outputPath, opts, nil)
			case cloudnative.ConversionCOPC:
				convErr = cloudnative.ConvertToCOPC(ctx, tempPath, outputPath, opts, nil)
			case cloudnative.ConversionGeoParquet:
				convErr = cloudnative.ConvertToGeoParquet(ctx, tempPath, outputPath, opts, nil)
			}

			if convErr != nil {
				s.jsonError(w, "Conversion failed: "+convErr.Error(), http.StatusInternalServerError)
				return
			}

			// Upload converted file
			convertedFile, err := os.Open(outputPath)
			if err != nil {
				s.jsonError(w, "Failed to open converted file", http.StatusInternalServerError)
				return
			}
			defer convertedFile.Close()

			stat, _ := convertedFile.Stat()
			targetKey = cloudnative.GenerateOutputPath(targetKey, convType)
			if prefix != "" {
				targetKey = prefix + filepath.Base(targetKey)
			} else {
				targetKey = filepath.Base(targetKey)
			}

			putOpts := s3client.PutOptions{
				ContentType: getContentType(targetKey),
			}

			if err := client.PutObject(ctx, bucketName, targetKey, convertedFile, stat.Size(), putOpts); err != nil {
				s.jsonError(w, "Failed to upload: "+err.Error(), http.StatusInternalServerError)
				return
			}

			s.jsonResponse(w, map[string]interface{}{
				"success":   true,
				"key":       targetKey,
				"size":      stat.Size(),
				"converted": true,
				"format":    string(convType),
			})
			return
		}
	}

	// Direct upload without conversion
	if prefix != "" {
		targetKey = prefix + filepath.Base(targetKey)
	}

	putOpts := s3client.PutOptions{
		ContentType: header.Header.Get("Content-Type"),
	}
	if putOpts.ContentType == "" {
		putOpts.ContentType = getContentType(targetKey)
	}

	if err := client.PutObject(ctx, bucketName, targetKey, file, header.Size, putOpts); err != nil {
		s.jsonError(w, "Failed to upload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"success":   true,
		"key":       targetKey,
		"size":      header.Size,
		"converted": false,
	})
}

// deleteS3Object deletes an object or recursively deletes a folder and its contents
func (s *Server) deleteS3Object(w http.ResponseWriter, r *http.Request, client *s3client.Client, bucketName string) {
	key := r.URL.Query().Get("key")
	if key == "" {
		s.jsonError(w, "Object key is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Check if this is a folder (ends with /) - if so, delete recursively
	if strings.HasSuffix(key, "/") {
		// List all objects with this prefix
		objects, err := client.ListObjectsRecursive(ctx, bucketName, key)
		if err != nil {
			s.jsonError(w, "Failed to list folder contents: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Collect all keys to delete (including the folder marker itself)
		var keysToDelete []string
		for _, obj := range objects {
			keysToDelete = append(keysToDelete, obj.Key)
		}
		// Also add the folder marker itself
		keysToDelete = append(keysToDelete, key)

		if len(keysToDelete) > 0 {
			if err := client.DeleteObjects(ctx, bucketName, keysToDelete); err != nil {
				s.jsonError(w, "Failed to delete folder contents: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		// Single object delete
		if err := client.DeleteObject(ctx, bucketName, key); err != nil {
			s.jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleS3Presign generates a presigned URL
func (s *Server) handleS3Presign(w http.ResponseWriter, r *http.Request, client *s3client.Client, bucketName string) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		s.jsonError(w, "Object key is required", http.StatusBadRequest)
		return
	}

	// Default 1 hour expiry
	expires := 1 * time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url, err := client.GetPresignedURL(ctx, bucketName, key, expires)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]string{
		"url":     url,
		"expires": time.Now().Add(expires).Format(time.RFC3339),
	})
}

// S3PreviewMetadata represents metadata for S3 layer preview
type S3PreviewMetadata struct {
	Format        string      `json:"format"`      // "cog", "copc", "geoparquet", "geojson", "geotiff", "qgisproject", "parquet"
	PreviewType   string      `json:"previewType"` // "raster", "pointcloud", "vector", "qgisproject", "table"
	Bounds        *S3Bounds   `json:"bounds,omitempty"`
	CRS           string      `json:"crs,omitempty"`
	Size          int64       `json:"size"`
	Key           string      `json:"key"`
	ProxyURL      string      `json:"proxyUrl"`                // URL to proxy through backend
	GeoJSONURL    string      `json:"geojsonUrl,omitempty"`    // URL to get GeoParquet converted to GeoJSON
	AttributesURL string      `json:"attributesUrl,omitempty"` // URL to get attributes as JSON table
	BandCount     int         `json:"bandCount,omitempty"`     // Number of bands (1 = potential DEM)
	FeatureCount  int64       `json:"featureCount,omitempty"`  // Number of features/rows
	FieldNames    []string    `json:"fieldNames,omitempty"`    // Column names for table view
	Metadata      interface{} `json:"metadata,omitempty"`      // Format-specific metadata
}

// S3Bounds represents geographic bounds
type S3Bounds struct {
	MinX float64 `json:"minX"`
	MinY float64 `json:"minY"`
	MaxX float64 `json:"maxX"`
	MaxY float64 `json:"maxY"`
}

// handleS3Preview handles preview requests for S3 objects
// Pattern: /api/s3/preview/{connectionId}/{bucket}?key=object/path
func (s *Server) handleS3Preview(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}

	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/s3/preview/{connectionId}/{bucket}
	path := strings.TrimPrefix(r.URL.Path, "/api/s3/preview/")
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get object metadata
	objInfo, err := client.GetObjectInfo(ctx, bucket, objectKey)
	if err != nil {
		s.jsonError(w, "Object not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Build proxy URL (data served through our backend)
	proxyURL := "/api/s3/proxy/" + connID + "/" + bucket + "?key=" + objectKey

	// Generate presigned URL for GDAL bounds extraction only (not returned to client)
	presignedURL, _ := client.GetPresignedURL(ctx, bucket, objectKey, 1*time.Hour)

	// Detect format
	cnFormat := s3client.DetectCloudNativeFormat(objectKey)
	format := string(cnFormat)
	if format == "" || format == "unknown" {
		ext := strings.ToLower(filepath.Ext(objectKey))
		switch ext {
		case ".tif", ".tiff":
			format = "geotiff"
		case ".geojson", ".json":
			format = "geojson"
		case ".gpkg":
			format = "geopackage"
		case ".las", ".laz":
			format = "pointcloud"
		case ".qgs", ".qgz":
			format = "qgisproject"
		case ".parquet":
			format = "parquet"
		case ".geoparquet":
			format = "geoparquet"
		default:
			format = "unknown"
		}
	}

	// Determine preview type
	previewType := "unknown"
	switch format {
	case "cog", "geotiff":
		previewType = "raster"
	case "copc", "pointcloud":
		previewType = "pointcloud"
	case "geoparquet", "geojson", "geopackage":
		previewType = "vector"
	case "parquet":
		previewType = "table"
	case "qgisproject":
		previewType = "qgisproject"
	}

	metadata := S3PreviewMetadata{
		Format:      format,
		PreviewType: previewType,
		Size:        objInfo.Size,
		Key:         objectKey,
		ProxyURL:    proxyURL,
	}

	// For COG/GeoTIFF files, try to extract bounds and band count using GDAL
	if format == "cog" || format == "geotiff" {
		rasterInfo := s.extractRasterInfo(ctx, presignedURL)
		if rasterInfo != nil {
			metadata.Bounds = rasterInfo.Bounds
			metadata.CRS = rasterInfo.CRS
			metadata.BandCount = rasterInfo.BandCount
		}
	}

	// For GeoJSON, try to extract bounds from the file
	if format == "geojson" {
		bounds := s.extractGeoJSONBounds(ctx, client, bucket, objectKey)
		if bounds != nil {
			metadata.Bounds = bounds
			metadata.CRS = "EPSG:4326"
		}
	}

	// For GeoParquet/Parquet, add conversion endpoints and extract info
	if format == "geoparquet" || format == "parquet" {
		// URL for GeoJSON conversion (for map preview)
		if format == "geoparquet" {
			metadata.GeoJSONURL = "/api/s3/geojson/" + connID + "/" + bucket + "?key=" + objectKey
		}
		// URL for attributes table (for table view)
		metadata.AttributesURL = "/api/s3/attributes/" + connID + "/" + bucket + "?key=" + objectKey

		// Extract parquet info (bounds, fields, feature count) by downloading file first
		tempFile, err := s.downloadS3ToTemp(ctx, client, bucket, objectKey)
		if err == nil {
			defer os.Remove(tempFile)
			parquetInfo := s.extractParquetInfo(ctx, tempFile)
			if parquetInfo != nil {
				metadata.Bounds = parquetInfo.Bounds
				metadata.CRS = parquetInfo.CRS
				metadata.FeatureCount = parquetInfo.FeatureCount
				metadata.FieldNames = parquetInfo.FieldNames
			}
		}
	}

	s.jsonResponse(w, metadata)
}

// RasterInfo holds extracted information about a raster file
type RasterInfo struct {
	Bounds    *S3Bounds
	CRS       string
	BandCount int
}

// extractRasterInfo extracts bounds and band count from a COG/GeoTIFF using GDAL
func (s *Server) extractRasterInfo(ctx context.Context, url string) *RasterInfo {
	// Use gdalinfo with /vsicurl/ to read from remote URL
	cmd := exec.CommandContext(ctx, "gdalinfo", "-json", "/vsicurl/"+url)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	// Parse JSON output
	var info struct {
		WGS84Extent struct {
			Coordinates [][][]float64 `json:"coordinates"`
		} `json:"wgs84Extent"`
		CoordinateSystem struct {
			WKT string `json:"wkt"`
		} `json:"coordinateSystem"`
		Bands []struct {
			Band int `json:"band"`
		} `json:"bands"`
	}
	if err := json.Unmarshal(output, &info); err != nil {
		return nil
	}

	result := &RasterInfo{
		BandCount: len(info.Bands),
	}

	// Extract bounds from wgs84Extent polygon
	if len(info.WGS84Extent.Coordinates) > 0 && len(info.WGS84Extent.Coordinates[0]) >= 4 {
		coords := info.WGS84Extent.Coordinates[0]
		minX, minY := coords[0][0], coords[0][1]
		maxX, maxY := coords[0][0], coords[0][1]
		for _, c := range coords {
			if c[0] < minX {
				minX = c[0]
			}
			if c[0] > maxX {
				maxX = c[0]
			}
			if c[1] < minY {
				minY = c[1]
			}
			if c[1] > maxY {
				maxY = c[1]
			}
		}
		result.Bounds = &S3Bounds{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
		result.CRS = "EPSG:4326"
	}

	return result
}

// extractGeoJSONBounds extracts bounds from a GeoJSON file
func (s *Server) extractGeoJSONBounds(ctx context.Context, client *s3client.Client, bucket, key string) *S3Bounds {
	// Read first 100KB to check for bbox property
	data, err := client.GetObjectRange(ctx, bucket, key, 0, 102400)
	if err != nil {
		return nil
	}

	// Try to parse as GeoJSON FeatureCollection with bbox
	var fc struct {
		BBox []float64 `json:"bbox"`
	}
	if err := json.Unmarshal(data, &fc); err == nil && len(fc.BBox) >= 4 {
		return &S3Bounds{
			MinX: fc.BBox[0],
			MinY: fc.BBox[1],
			MaxX: fc.BBox[2],
			MaxY: fc.BBox[3],
		}
	}

	return nil
}

// getContentType returns MIME type based on file extension
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".tif", ".tiff":
		return "image/tiff"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".json", ".geojson":
		return "application/json"
	case ".parquet", ".geoparquet":
		return "application/vnd.apache.parquet"
	case ".gpkg":
		return "application/geopackage+sqlite3"
	case ".laz", ".las":
		return "application/octet-stream"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// handleS3Proxy proxies S3 object data through the backend
// Pattern: /api/s3/proxy/{connectionId}/{bucket}?key=object/path
func (s *Server) handleS3Proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}

	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/s3/proxy/{connectionId}/{bucket}
	path := strings.TrimPrefix(r.URL.Path, "/api/s3/proxy/")
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get object from S3
	obj, err := client.GetObject(ctx, bucket, objectKey)
	if err != nil {
		s.jsonError(w, "Failed to get object: "+err.Error(), http.StatusNotFound)
		return
	}
	defer obj.Close()

	// Set content type based on file extension
	contentType := getContentType(objectKey)
	w.Header().Set("Content-Type", contentType)

	// Set CORS headers to allow browser access
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges")

	// Enable range requests for large files (important for COG streaming)
	w.Header().Set("Accept-Ranges", "bytes")

	// Handle range requests
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		// For range requests, we need to get object info first
		objInfo, err := client.GetObjectInfo(ctx, bucket, objectKey)
		if err != nil {
			s.jsonError(w, "Failed to get object info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse range header (format: bytes=start-end)
		var start, end int64
		_, err = parseRangeHeader(rangeHeader, objInfo.Size, &start, &end)
		if err != nil {
			http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}

		// Get range from S3
		data, err := client.GetObjectRange(ctx, bucket, objectKey, start, end-start+1)
		if err != nil {
			s.jsonError(w, "Failed to get object range: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Range", "bytes "+
			formatInt64(start)+"-"+formatInt64(end)+"/"+formatInt64(objInfo.Size))
		w.Header().Set("Content-Length", formatInt64(int64(len(data))))
		w.WriteHeader(http.StatusPartialContent)
		w.Write(data)
		return
	}

	// Full object download
	_, err = io.Copy(w, obj)
	if err != nil {
		// Connection may have been closed by client, don't error
		return
	}
}

// parseRangeHeader parses HTTP Range header
func parseRangeHeader(rangeHeader string, fileSize int64, start, end *int64) (bool, error) {
	// Format: bytes=start-end or bytes=start- or bytes=-suffix
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return false, nil
	}

	rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeSpec, "-")
	if len(parts) != 2 {
		return false, nil
	}

	if parts[0] == "" {
		// Suffix range: -500 means last 500 bytes
		suffix, err := parseInt64(parts[1])
		if err != nil {
			return false, err
		}
		*start = fileSize - suffix
		*end = fileSize - 1
	} else if parts[1] == "" {
		// Open range: 500- means from 500 to end
		var err error
		*start, err = parseInt64(parts[0])
		if err != nil {
			return false, err
		}
		*end = fileSize - 1
	} else {
		// Full range: 500-999
		var err error
		*start, err = parseInt64(parts[0])
		if err != nil {
			return false, err
		}
		*end, err = parseInt64(parts[1])
		if err != nil {
			return false, err
		}
	}

	// Validate range
	if *start < 0 {
		*start = 0
	}
	if *end >= fileSize {
		*end = fileSize - 1
	}
	if *start > *end {
		return false, nil
	}

	return true, nil
}

func parseInt64(s string) (int64, error) {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, io.EOF
		}
		n = n*10 + int64(c-'0')
	}
	return n, nil
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte(n%10) + '0'
		n /= 10
	}
	return string(buf[i:])
}

// ParquetInfo holds extracted information about a parquet/geoparquet file
type ParquetInfo struct {
	Bounds       *S3Bounds
	CRS          string
	FeatureCount int64
	FieldNames   []string
}

// extractParquetInfo extracts metadata from a Parquet/GeoParquet file using ogrinfo
// filePath should be a local file path (already downloaded from S3)
func (s *Server) extractParquetInfo(ctx context.Context, filePath string) *ParquetInfo {
	// Use ogrinfo to read from local file
	cmd := exec.CommandContext(ctx, "ogrinfo", "-so", "-al", filePath)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("extractParquetInfo: ogrinfo failed: %v", err)
		return nil
	}

	info := &ParquetInfo{}
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse feature count
		if strings.HasPrefix(line, "Feature Count:") {
			countStr := strings.TrimSpace(strings.TrimPrefix(line, "Feature Count:"))
			if count, err := parseInt64(countStr); err == nil {
				info.FeatureCount = count
			}
		}

		// Parse extent (bounds)
		if strings.HasPrefix(line, "Extent:") {
			// Format: Extent: (minX, minY) - (maxX, maxY)
			info.Bounds = s.parseOGRExtent(line)
		}

		// Parse geometry column (presence indicates spatial data)
		if strings.HasPrefix(line, "Geometry Column =") {
			info.CRS = "EPSG:4326" // Assume WGS84 for GeoParquet
		}

		// Parse field names - they appear after "Geometry:" line
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "INFO") &&
			!strings.HasPrefix(line, "Layer name") && !strings.HasPrefix(line, "Feature Count") &&
			!strings.HasPrefix(line, "Extent") && !strings.HasPrefix(line, "Geometry") &&
			!strings.HasPrefix(line, "FID Column") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fieldName := strings.TrimSpace(parts[0])
				if fieldName != "" && !strings.Contains(fieldName, " ") {
					info.FieldNames = append(info.FieldNames, fieldName)
				}
			}
		}
	}

	return info
}

// parseOGRExtent parses extent from ogrinfo output
func (s *Server) parseOGRExtent(line string) *S3Bounds {
	// Format: Extent: (minX, minY) - (maxX, maxY)
	line = strings.TrimPrefix(line, "Extent:")
	line = strings.TrimSpace(line)
	line = strings.ReplaceAll(line, "(", "")
	line = strings.ReplaceAll(line, ")", "")
	parts := strings.Split(line, " - ")
	if len(parts) != 2 {
		return nil
	}

	minParts := strings.Split(strings.TrimSpace(parts[0]), ", ")
	maxParts := strings.Split(strings.TrimSpace(parts[1]), ", ")
	if len(minParts) != 2 || len(maxParts) != 2 {
		return nil
	}

	var minX, minY, maxX, maxY float64
	if _, err := fmt.Sscanf(minParts[0], "%f", &minX); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(minParts[1], "%f", &minY); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(maxParts[0], "%f", &maxX); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(maxParts[1], "%f", &maxY); err != nil {
		return nil
	}

	return &S3Bounds{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}

// handleS3GeoJSON converts GeoParquet to GeoJSON and returns it
// Pattern: /api/s3/geojson/{connectionId}/{bucket}?key=object/path&limit=1000
func (s *Server) handleS3GeoJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}

	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/s3/geojson/{connectionId}/{bucket}
	path := strings.TrimPrefix(r.URL.Path, "/api/s3/geojson/")
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

	// Get optional limit (default 1000 features for map preview)
	limitStr := r.URL.Query().Get("limit")
	limit := 1000
	if limitStr != "" {
		if l, err := parseInt64(limitStr); err == nil && l > 0 {
			limit = int(l)
		}
	}

	client := s.getS3Client(connID)
	if client == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Download file from S3 to temp file (proxy through our API)
	tempFile, err := s.downloadS3ToTemp(ctx, client, bucket, objectKey)
	if err != nil {
		s.jsonError(w, "Failed to download file from S3: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile)

	// Use ogr2ogr to convert GeoParquet to GeoJSON with limit
	// -unsetFid avoids FID field mapping issues with Arrow/Parquet
	args := []string{
		"-f", "GeoJSON",
		"-t_srs", "EPSG:4326", // Ensure output is in WGS84
		"-unsetFid",   // Avoid FID field mapping issues
		"/vsistdout/", // Output to stdout
		tempFile,
		"-limit", fmt.Sprintf("%d", limit),
	}

	cmd := exec.CommandContext(ctx, "ogr2ogr", args...)
	output, err := cmd.Output()
	if err != nil {
		// Check if it's a non-spatial parquet file
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("handleS3GeoJSON: ogr2ogr failed: %s", string(exitErr.Stderr))
		}
		s.jsonError(w, "Failed to convert to GeoJSON - file may not contain geometry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers for GeoJSON response
	w.Header().Set("Content-Type", "application/geo+json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(output)
}

// AttributeTableResponse represents the response for attribute table data
type AttributeTableResponse struct {
	Fields  []string                 `json:"fields"`
	Rows    []map[string]interface{} `json:"rows"`
	Total   int64                    `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
	HasMore bool                     `json:"hasMore"`
}

// handleS3Attributes returns attribute data from a Parquet/GeoParquet file as JSON
// Pattern: /api/s3/attributes/{connectionId}/{bucket}?key=object/path&limit=100&offset=0
func (s *Server) handleS3Attributes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}

	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/s3/attributes/{connectionId}/{bucket}
	path := strings.TrimPrefix(r.URL.Path, "/api/s3/attributes/")
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

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := parseInt64(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = int(l)
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := parseInt64(offsetStr); err == nil && o >= 0 {
			offset = int(o)
		}
	}

	client := s.getS3Client(connID)
	if client == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Download file from S3 to temp file (proxy through our API)
	tempFile, err := s.downloadS3ToTemp(ctx, client, bucket, objectKey)
	if err != nil {
		s.jsonError(w, "Failed to download file from S3: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile)

	// Use ogr2ogr to convert to GeoJSON (which includes all attributes)
	// Then parse the JSON and extract just the properties
	// -unsetFid avoids FID field mapping issues with Arrow/Parquet
	args := []string{
		"-f", "GeoJSON",
		"-unsetFid", // Avoid FID field mapping issues
		"/vsistdout/",
		tempFile,
		"-limit", fmt.Sprintf("%d", limit+offset+1), // Get one extra to check hasMore
	}

	cmd := exec.CommandContext(ctx, "ogr2ogr", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("handleS3Attributes: ogr2ogr failed: %s", string(exitErr.Stderr))
		}
		s.jsonError(w, "Failed to read attributes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse GeoJSON
	var geojson struct {
		Type     string `json:"type"`
		Features []struct {
			Properties map[string]interface{} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(output, &geojson); err != nil {
		s.jsonError(w, "Failed to parse GeoJSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract fields from first feature
	var fields []string
	if len(geojson.Features) > 0 {
		for k := range geojson.Features[0].Properties {
			fields = append(fields, k)
		}
	}

	// Apply offset and limit
	start := offset
	if start > len(geojson.Features) {
		start = len(geojson.Features)
	}
	end := start + limit
	hasMore := false
	if end >= len(geojson.Features) {
		end = len(geojson.Features)
	} else {
		hasMore = true
	}

	// Extract rows (properties only)
	rows := make([]map[string]interface{}, 0, end-start)
	for i := start; i < end; i++ {
		rows = append(rows, geojson.Features[i].Properties)
	}

	// Get total count using ogrinfo on the temp file
	totalCount := int64(len(geojson.Features))
	infoCmd := exec.CommandContext(ctx, "ogrinfo", "-so", "-al", tempFile)
	if infoOutput, err := infoCmd.Output(); err == nil {
		for _, line := range strings.Split(string(infoOutput), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Feature Count:") {
				countStr := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "Feature Count:"))
				if count, err := parseInt64(countStr); err == nil {
					totalCount = count
					hasMore = int64(offset+limit) < totalCount
				}
				break
			}
		}
	}

	response := AttributeTableResponse{
		Fields:  fields,
		Rows:    rows,
		Total:   totalCount,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}

	s.jsonResponse(w, response)
}

// downloadS3ToTemp downloads an S3 object to a temporary file and returns the file path
// The caller is responsible for removing the temp file when done
func (s *Server) downloadS3ToTemp(ctx context.Context, client *s3client.Client, bucket, key string) (string, error) {
	// Determine file extension from key
	ext := filepath.Ext(key)
	if ext == "" {
		ext = ".parquet"
	}

	// Create temp file with appropriate extension
	tempFile, err := os.CreateTemp("", "s3preview-*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Get object from S3
	obj, err := client.GetObject(ctx, bucket, key)
	if err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer obj.Close()

	// Copy to temp file
	_, err = io.Copy(tempFile, obj)
	tempFile.Close()
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to download object: %w", err)
	}

	return tempPath, nil
}

// DuckDBQueryRequest represents a DuckDB query request
type DuckDBQueryRequest struct {
	SQL    string `json:"sql"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// DuckDBQueryResponse represents the response from a DuckDB query
type DuckDBQueryResponse struct {
	Columns        []string                 `json:"columns"`
	ColumnTypes    []string                 `json:"columnTypes,omitempty"`
	Rows           []map[string]interface{} `json:"rows"`
	RowCount       int                      `json:"rowCount"`
	TotalCount     int64                    `json:"totalCount,omitempty"`
	HasMore        bool                     `json:"hasMore"`
	GeometryColumn string                   `json:"geometryColumn,omitempty"`
	SQL            string                   `json:"sql,omitempty"`
	Error          string                   `json:"error,omitempty"`
}

// DuckDBTableInfoResponse represents metadata about a Parquet file
type DuckDBTableInfoResponse struct {
	Columns        []DuckDBColumnInfoResponse `json:"columns"`
	RowCount       int64                      `json:"rowCount"`
	GeometryColumn string                     `json:"geometryColumn,omitempty"`
	BBox           []float64                  `json:"bbox,omitempty"`
	SampleQueries  []string                   `json:"sampleQueries,omitempty"`
}

// DuckDBColumnInfoResponse describes a column in a Parquet file
type DuckDBColumnInfoResponse struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

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
