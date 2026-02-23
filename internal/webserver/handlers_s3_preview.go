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

	"github.com/kartoza/kartoza-cloudbench/internal/s3client"
)

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
		"-unsetFid",           // Avoid FID field mapping issues
		"/vsistdout/",         // Output to stdout
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
