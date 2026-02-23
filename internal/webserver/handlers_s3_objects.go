package webserver

import (
	"context"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/cloudnative"
	"github.com/kartoza/kartoza-cloudbench/internal/s3client"
)

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
			s.handleGeoPackageUpload(w, r, ctx, client, bucketName, file, header, prefix, createSubfolder)
			return
		}

		// Standard conversion for other file types
		if s.handleCloudNativeConversion(w, r, ctx, client, bucketName, file, header, targetKey, prefix) {
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

// handleGeoPackageUpload handles GeoPackage file upload with layer extraction
func (s *Server) handleGeoPackageUpload(w http.ResponseWriter, r *http.Request, ctx context.Context, client *s3client.Client, bucketName string, file io.Reader, header *multipart.FileHeader, prefix string, createSubfolder bool) {
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
	ext := strings.ToLower(filepath.Ext(header.Filename))
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
}

// handleCloudNativeConversion handles conversion to cloud-native formats
// Returns true if the request was handled (either success or error), false if no conversion needed
func (s *Server) handleCloudNativeConversion(w http.ResponseWriter, r *http.Request, ctx context.Context, client *s3client.Client, bucketName string, file io.Reader, header *multipart.FileHeader, targetKey, prefix string) bool {
	convType, shouldConvert := cloudnative.DetectRecommendedConversion(header.Filename)
	if !shouldConvert || convType == cloudnative.ConversionNone {
		return false
	}

	// Save to temp file
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, header.Filename)
	tempFile, err := os.Create(tempPath)
	if err != nil {
		s.jsonError(w, "Failed to create temp file", http.StatusInternalServerError)
		return true
	}
	defer os.Remove(tempPath)

	if _, err := io.Copy(tempFile, file); err != nil {
		tempFile.Close()
		s.jsonError(w, "Failed to save temp file", http.StatusInternalServerError)
		return true
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
		return true
	}

	// Upload converted file
	convertedFile, err := os.Open(outputPath)
	if err != nil {
		s.jsonError(w, "Failed to open converted file", http.StatusInternalServerError)
		return true
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
		return true
	}

	s.jsonResponse(w, map[string]interface{}{
		"success":   true,
		"key":       targetKey,
		"size":      stat.Size(),
		"converted": true,
		"format":    string(convType),
	})
	return true
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
