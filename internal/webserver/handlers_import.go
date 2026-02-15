package webserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/ogr2ogr"
)

// ImportJob represents an ongoing or completed import job
type ImportJob struct {
	ID          string    `json:"id"`
	SourceFile  string    `json:"source_file"`
	TargetTable string    `json:"target_table"`
	Service     string    `json:"service"`
	Status      string    `json:"status"` // "pending", "running", "completed", "failed"
	Progress    int       `json:"progress"`
	Message     string    `json:"message"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

var (
	importJobs   = make(map[string]*ImportJob)
	importJobsMu sync.RWMutex
	jobCounter   int
)

// ImportRequest represents a request to import data
type ImportRequest struct {
	SourceFile    string `json:"source_file"`     // Path to uploaded file or local file
	TargetService string `json:"target_service"`  // PostgreSQL service name
	TargetSchema  string `json:"target_schema"`   // Target schema (default: public)
	TableName     string `json:"table_name"`      // Target table name (auto-detect if empty)
	SRID          int    `json:"srid"`            // Source SRID (0 = auto-detect)
	TargetSRID    int    `json:"target_srid"`     // Target SRID (0 = keep source)
	Overwrite     bool   `json:"overwrite"`       // Overwrite existing table
	Append        bool   `json:"append"`          // Append to existing table
	SourceLayer   string `json:"source_layer"`    // Specific layer for multi-layer sources
}

// handlePGImport handles POST /api/pg/import - start an import job
func (s *Server) handlePGImport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if ogr2ogr is available
	if !ogr2ogr.CheckAvailable() {
		http.Error(w, "ogr2ogr not found on system. Please install GDAL.", http.StatusServiceUnavailable)
		return
	}

	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SourceFile == "" {
		http.Error(w, "source_file is required", http.StatusBadRequest)
		return
	}

	if req.TargetService == "" {
		http.Error(w, "target_service is required", http.StatusBadRequest)
		return
	}

	// Check if source file exists
	if _, err := os.Stat(req.SourceFile); os.IsNotExist(err) {
		http.Error(w, "Source file not found", http.StatusBadRequest)
		return
	}

	// Create import job
	importJobsMu.Lock()
	jobCounter++
	jobID := generateImportJobID()
	job := &ImportJob{
		ID:         jobID,
		SourceFile: req.SourceFile,
		Service:    req.TargetService,
		Status:     "pending",
		Progress:   0,
		Message:    "Queued for import",
		StartedAt:  time.Now(),
	}
	importJobs[jobID] = job
	importJobsMu.Unlock()

	// Start import in background
	go s.runImportJob(job, req)

	json.NewEncoder(w).Encode(map[string]string{
		"job_id":  jobID,
		"status":  "pending",
		"message": "Import job created",
	})
}

// runImportJob executes the import job
func (s *Server) runImportJob(job *ImportJob, req ImportRequest) {
	updateJob := func(status string, progress int, message string, err string) {
		importJobsMu.Lock()
		job.Status = status
		job.Progress = progress
		job.Message = message
		job.Error = err
		if status == "completed" || status == "failed" {
			job.CompletedAt = time.Now()
		}
		importJobsMu.Unlock()
	}

	updateJob("running", 0, "Starting import...", "")

	opts := ogr2ogr.ImportOptions{
		SourceFile:    req.SourceFile,
		TargetService: req.TargetService,
		TargetSchema:  req.TargetSchema,
		TableName:     req.TableName,
		SRID:          req.SRID,
		TargetSRID:    req.TargetSRID,
		Overwrite:     req.Overwrite,
		Append:        req.Append,
		SourceLayer:   req.SourceLayer,
	}

	if opts.TargetSchema == "" {
		opts.TargetSchema = "public"
	}

	// Progress callback
	progress := func(pct int, msg string) {
		updateJob("running", pct, msg, "")
	}

	ctx := context.Background()
	result, err := ogr2ogr.Import(ctx, opts, progress)
	if err != nil {
		updateJob("failed", 0, "Import failed", err.Error())
		return
	}

	if !result.Success {
		errMsg := "Import failed"
		if len(result.Errors) > 0 {
			errMsg = strings.Join(result.Errors, "; ")
		}
		updateJob("failed", 0, errMsg, errMsg)
		return
	}

	importJobsMu.Lock()
	job.TargetTable = result.TableName
	importJobsMu.Unlock()

	updateJob("completed", 100, "Import completed: "+result.TableName, "")
}

// handlePGImportStatus handles GET /api/pg/import/{jobId} - get import job status
func (s *Server) handlePGImportStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse job ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/pg/import/")
	jobID := strings.TrimSuffix(path, "/")

	if jobID == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	importJobsMu.RLock()
	job, ok := importJobs[jobID]
	importJobsMu.RUnlock()

	if !ok {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(job)
}

// handlePGDetectLayers handles POST /api/pg/detect-layers - detect layers in a file
func (s *Server) handlePGDetectLayers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if ogr2ogr is available
	if !ogr2ogr.CheckAvailable() {
		http.Error(w, "ogr2ogr not found on system", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		FilePath string `json:"file_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		http.Error(w, "file_path is required", http.StatusBadRequest)
		return
	}

	// Check if file exists
	if _, err := os.Stat(req.FilePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusBadRequest)
		return
	}

	layers, err := ogr2ogr.DetectLayers(req.FilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(layers)
}

// handlePGImportUpload handles POST /api/pg/import/upload - upload file for import
func (s *Server) handlePGImportUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 500MB)
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check if file type is supported
	if !ogr2ogr.IsSupported(header.Filename) {
		http.Error(w, "Unsupported file format", http.StatusBadRequest)
		return
	}

	// Create temp directory for uploads
	tmpDir := filepath.Join(os.TempDir(), "cloudbench-imports")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}

	// Create temp file with original extension
	ext := filepath.Ext(header.Filename)
	baseName := strings.TrimSuffix(header.Filename, ext)
	tmpFile, err := os.CreateTemp(tmpDir, baseName+"_*"+ext)
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer tmpFile.Close()

	// Copy file content
	if _, err := io.Copy(tmpFile, file); err != nil {
		os.Remove(tmpFile.Name())
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// For shapefiles, check if we need companion files
	if ext == ".shp" {
		// Try to get companion files (.dbf, .shx, .prj)
		companions := []string{".dbf", ".shx", ".prj", ".cpg"}
		for _, compExt := range companions {
			compFile, _, err := r.FormFile(strings.TrimSuffix(header.Filename, ext) + compExt)
			if err == nil {
				compPath := strings.TrimSuffix(tmpFile.Name(), ext) + compExt
				compOut, err := os.Create(compPath)
				if err == nil {
					io.Copy(compOut, compFile)
					compOut.Close()
				}
				compFile.Close()
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"file_path": tmpFile.Name(),
		"filename":  header.Filename,
		"message":   "File uploaded successfully",
	})
}

// handleOgr2ogrStatus handles GET /api/pg/ogr2ogr/status - check ogr2ogr availability
func (s *Server) handleOgr2ogrStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	available := ogr2ogr.CheckAvailable()
	version := ""
	if available {
		version, _ = ogr2ogr.GetVersion()
	}

	rasterAvailable := ogr2ogr.CheckRaster2PgsqlAvailable()
	rasterVersion := ""
	if rasterAvailable {
		rasterVersion, _ = ogr2ogr.GetRaster2PgsqlVersion()
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"available":            available,
		"version":              version,
		"raster_available":     rasterAvailable,
		"raster_version":       rasterVersion,
		"supported_formats":    ogr2ogr.SupportedFormats(),
		"supported_extensions": ogr2ogr.GetAllSupportedExtensions(),
		"vector_extensions":    ogr2ogr.GetSupportedExtensions(),
		"raster_extensions":    ogr2ogr.GetRasterExtensions(),
	})
}

// RasterImportRequest represents a request to import raster data
type RasterImportRequest struct {
	SourceFile    string `json:"source_file"`
	TargetService string `json:"target_service"`
	TargetSchema  string `json:"target_schema"`
	TableName     string `json:"table_name"`
	SRID          int    `json:"srid"`
	TileSize      string `json:"tile_size"`
	Overwrite     bool   `json:"overwrite"`
	Append        bool   `json:"append"`
	CreateIndex   bool   `json:"create_index"`
	OutOfDB       bool   `json:"out_of_db"`
}

// handlePGRasterImport handles POST /api/pg/import/raster - start a raster import job
func (s *Server) handlePGRasterImport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if raster2pgsql is available
	if !ogr2ogr.CheckRaster2PgsqlAvailable() {
		http.Error(w, "raster2pgsql not found on system. Please install PostGIS.", http.StatusServiceUnavailable)
		return
	}

	var req RasterImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SourceFile == "" {
		http.Error(w, "source_file is required", http.StatusBadRequest)
		return
	}

	if req.TargetService == "" {
		http.Error(w, "target_service is required", http.StatusBadRequest)
		return
	}

	// Check if source file exists
	if _, err := os.Stat(req.SourceFile); os.IsNotExist(err) {
		http.Error(w, "Source file not found", http.StatusBadRequest)
		return
	}

	// Create import job
	importJobsMu.Lock()
	jobCounter++
	jobID := generateImportJobID()
	job := &ImportJob{
		ID:         jobID,
		SourceFile: req.SourceFile,
		Service:    req.TargetService,
		Status:     "pending",
		Progress:   0,
		Message:    "Queued for raster import",
		StartedAt:  time.Now(),
	}
	importJobs[jobID] = job
	importJobsMu.Unlock()

	// Start import in background
	go s.runRasterImportJob(job, req)

	json.NewEncoder(w).Encode(map[string]string{
		"job_id":  jobID,
		"status":  "pending",
		"message": "Raster import job created",
	})
}

// runRasterImportJob executes the raster import job
func (s *Server) runRasterImportJob(job *ImportJob, req RasterImportRequest) {
	updateJob := func(status string, progress int, message string, err string) {
		importJobsMu.Lock()
		job.Status = status
		job.Progress = progress
		job.Message = message
		job.Error = err
		if status == "completed" || status == "failed" {
			job.CompletedAt = time.Now()
		}
		importJobsMu.Unlock()
	}

	updateJob("running", 0, "Starting raster import...", "")

	opts := ogr2ogr.RasterImportOptions{
		SourceFile:    req.SourceFile,
		TargetService: req.TargetService,
		TargetSchema:  req.TargetSchema,
		TableName:     req.TableName,
		SRID:          req.SRID,
		TileSize:      req.TileSize,
		Overwrite:     req.Overwrite,
		Append:        req.Append,
		CreateIndex:   req.CreateIndex,
		OutOfDB:       req.OutOfDB,
	}

	if opts.TargetSchema == "" {
		opts.TargetSchema = "public"
	}

	// Progress callback
	progress := func(pct int, msg string) {
		updateJob("running", pct, msg, "")
	}

	ctx := context.Background()
	result, err := ogr2ogr.ImportRaster(ctx, opts, progress)
	if err != nil {
		updateJob("failed", 0, "Raster import failed", err.Error())
		return
	}

	if !result.Success {
		errMsg := "Raster import failed"
		if len(result.Errors) > 0 {
			errMsg = strings.Join(result.Errors, "; ")
		}
		updateJob("failed", 0, errMsg, errMsg)
		return
	}

	importJobsMu.Lock()
	job.TargetTable = result.TableName
	importJobsMu.Unlock()

	updateJob("completed", 100, "Raster import completed: "+result.TableName, "")
}

// generateImportJobID generates a unique import job ID
func generateImportJobID() string {
	return "import_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

// randomString generates a random alphanumeric string
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
