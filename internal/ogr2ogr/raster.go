package ogr2ogr

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/kartoza/kartoza-cloudbench/internal/postgres"
)

// RasterImportOptions configures a raster import operation
type RasterImportOptions struct {
	SourceFile    string // Path to source raster file (GeoTIFF, etc.)
	TargetService string // PostgreSQL service name from pg_service.conf
	TargetSchema  string // Target schema (default: "public")
	TableName     string // Target table name (auto-detected from filename if empty)
	SRID          int    // Source SRID (0 = auto-detect)
	TileSize      string // Tile size (default: "256x256")
	Overwrite     bool   // Drop existing table
	Append        bool   // Append to existing table
	CreateIndex   bool   // Create spatial index (default: true)
	OutOfDB       bool   // Store rasters out-of-database (reference file path)
	Overview      []int  // Create overview levels (e.g., [2, 4, 8])
}

// RasterImportResult contains the result of a raster import operation
type RasterImportResult struct {
	Success   bool
	TableName string
	Errors    []string
	Warnings  []string
	Duration  float64
}

// RasterInfo contains information about a raster file
type RasterInfo struct {
	Width         int
	Height        int
	Bands         int
	DataType      string
	SRID          int
	PixelSizeX    float64
	PixelSizeY    float64
	NoDataValue   *float64
	Extent        *Extent
	Driver        string
	Compression   string
}

// CheckRaster2PgsqlAvailable checks if raster2pgsql is available
func CheckRaster2PgsqlAvailable() bool {
	_, err := exec.LookPath("raster2pgsql")
	return err == nil
}

// GetRaster2PgsqlVersion returns the raster2pgsql version
func GetRaster2PgsqlVersion() (string, error) {
	// raster2pgsql doesn't have a --version flag, but we can check PostGIS version
	cmd := exec.Command("raster2pgsql")
	output, _ := cmd.Output()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "RELEASE:") || strings.Contains(line, "raster2pgsql") {
			return strings.TrimSpace(line), nil
		}
	}
	return "available", nil
}

// GetRasterInfo gets information about a raster file using gdalinfo
func GetRasterInfo(sourcePath string) (*RasterInfo, error) {
	// Use text parsing for simplicity (works with all GDAL versions)
	return getRasterInfoText(sourcePath)
}

// getRasterInfoText parses gdalinfo text output
func getRasterInfoText(sourcePath string) (*RasterInfo, error) {
	cmd := exec.Command("gdalinfo", sourcePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get raster info: %w", err)
	}

	info := &RasterInfo{}
	lines := strings.Split(string(output), "\n")

	sizeRegex := regexp.MustCompile(`Size is (\d+),\s*(\d+)`)
	sridRegex := regexp.MustCompile(`EPSG[",:](\d+)`)
	bandRegex := regexp.MustCompile(`Band (\d+)`)
	typeRegex := regexp.MustCompile(`Type=(\w+)`)
	pixelRegex := regexp.MustCompile(`Pixel Size = \(([^,]+),\s*([^)]+)\)`)
	nodataRegex := regexp.MustCompile(`NoData Value=([-\d.e]+)`)
	driverRegex := regexp.MustCompile(`Driver:\s*(\w+)/`)

	for _, line := range lines {
		if matches := sizeRegex.FindStringSubmatch(line); matches != nil {
			info.Width, _ = strconv.Atoi(matches[1])
			info.Height, _ = strconv.Atoi(matches[2])
		}
		if matches := sridRegex.FindStringSubmatch(line); matches != nil && info.SRID == 0 {
			info.SRID, _ = strconv.Atoi(matches[1])
		}
		if matches := bandRegex.FindStringSubmatch(line); matches != nil {
			band, _ := strconv.Atoi(matches[1])
			if band > info.Bands {
				info.Bands = band
			}
		}
		if matches := typeRegex.FindStringSubmatch(line); matches != nil && info.DataType == "" {
			info.DataType = matches[1]
		}
		if matches := pixelRegex.FindStringSubmatch(line); matches != nil {
			info.PixelSizeX, _ = strconv.ParseFloat(matches[1], 64)
			info.PixelSizeY, _ = strconv.ParseFloat(matches[2], 64)
		}
		if matches := nodataRegex.FindStringSubmatch(line); matches != nil {
			val, _ := strconv.ParseFloat(matches[1], 64)
			info.NoDataValue = &val
		}
		if matches := driverRegex.FindStringSubmatch(line); matches != nil {
			info.Driver = matches[1]
		}
	}

	return info, nil
}

// ImportRaster imports a raster file to PostGIS using raster2pgsql
func ImportRaster(ctx context.Context, opts RasterImportOptions, progress ProgressCallback) (*RasterImportResult, error) {
	if !CheckRaster2PgsqlAvailable() {
		return nil, fmt.Errorf("raster2pgsql not found in PATH")
	}

	// Get PostgreSQL connection info
	services, err := postgres.ParsePGServiceFile()
	if err != nil {
		return nil, fmt.Errorf("failed to parse pg_service.conf: %w", err)
	}

	svc, err := postgres.GetServiceByName(services, opts.TargetService)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}

	result := &RasterImportResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	// Determine table name
	tableName := opts.TableName
	if tableName == "" {
		baseName := filepath.Base(opts.SourceFile)
		tableName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		tableName = sanitizeTableName(tableName)
	}
	result.TableName = tableName

	// Add schema prefix if not public
	fullTableName := tableName
	if opts.TargetSchema != "" && opts.TargetSchema != "public" {
		fullTableName = opts.TargetSchema + "." + tableName
	}

	if progress != nil {
		progress(0, "Starting raster import of "+filepath.Base(opts.SourceFile))
	}

	// Build raster2pgsql command
	args := buildRaster2PgsqlArgs(opts, fullTableName)

	// Execute raster2pgsql and pipe to psql
	r2pCmd := exec.CommandContext(ctx, "raster2pgsql", args...)

	// Build psql command
	psqlArgs := []string{}
	if svc.Host != "" {
		psqlArgs = append(psqlArgs, "-h", svc.Host)
	}
	if svc.Port != "" {
		psqlArgs = append(psqlArgs, "-p", svc.Port)
	}
	if svc.User != "" {
		psqlArgs = append(psqlArgs, "-U", svc.User)
	}
	if svc.DBName != "" {
		psqlArgs = append(psqlArgs, "-d", svc.DBName)
	}
	psqlArgs = append(psqlArgs, "-q") // Quiet mode

	psqlCmd := exec.CommandContext(ctx, "psql", psqlArgs...)

	// Set PGPASSWORD environment variable
	if svc.Password != "" {
		psqlCmd.Env = append(psqlCmd.Environ(), "PGPASSWORD="+svc.Password)
	}

	// Pipe raster2pgsql output to psql
	pipe, err := r2pCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}
	psqlCmd.Stdin = pipe

	// Capture stderr from both commands
	r2pStderr, _ := r2pCmd.StderrPipe()
	psqlStderr, _ := psqlCmd.StderrPipe()

	// Start both commands
	if err := r2pCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start raster2pgsql: %w", err)
	}

	if err := psqlCmd.Start(); err != nil {
		r2pCmd.Process.Kill()
		return nil, fmt.Errorf("failed to start psql: %w", err)
	}

	// Read stderr in background
	var wg sync.WaitGroup
	var outputLines []string
	var mu sync.Mutex

	wg.Add(2)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(r2pStderr)
		for scanner.Scan() {
			mu.Lock()
			outputLines = append(outputLines, "[raster2pgsql] "+scanner.Text())
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(psqlStderr)
		for scanner.Scan() {
			mu.Lock()
			outputLines = append(outputLines, "[psql] "+scanner.Text())
			mu.Unlock()
		}
	}()

	// Wait for raster2pgsql to finish first
	r2pErr := r2pCmd.Wait()

	// Then wait for psql
	psqlErr := psqlCmd.Wait()

	wg.Wait()

	// Parse errors
	for _, line := range outputLines {
		if strings.Contains(line, "ERROR") || strings.Contains(line, "error") {
			result.Errors = append(result.Errors, line)
		} else if strings.Contains(line, "WARNING") || strings.Contains(line, "warning") {
			result.Warnings = append(result.Warnings, line)
		}
	}

	if r2pErr != nil {
		result.Success = false
		if len(result.Errors) == 0 {
			result.Errors = append(result.Errors, "raster2pgsql failed: "+r2pErr.Error())
		}
		return result, nil
	}

	if psqlErr != nil {
		result.Success = false
		if len(result.Errors) == 0 {
			result.Errors = append(result.Errors, "psql failed: "+psqlErr.Error())
		}
		return result, nil
	}

	result.Success = true
	if progress != nil {
		progress(100, "Raster import complete: "+tableName)
	}

	return result, nil
}

// buildRaster2PgsqlArgs builds the raster2pgsql command arguments
func buildRaster2PgsqlArgs(opts RasterImportOptions, tableName string) []string {
	args := []string{}

	// Tile size
	if opts.TileSize != "" {
		args = append(args, "-t", opts.TileSize)
	} else {
		args = append(args, "-t", "256x256")
	}

	// SRID
	if opts.SRID > 0 {
		args = append(args, "-s", strconv.Itoa(opts.SRID))
	}

	// Create index
	if opts.CreateIndex {
		args = append(args, "-I")
	}

	// Create table or append
	if opts.Overwrite {
		args = append(args, "-d") // Drop and recreate
	} else if opts.Append {
		args = append(args, "-a") // Append mode
	} else {
		args = append(args, "-c") // Create mode (default)
	}

	// Out-of-DB raster
	if opts.OutOfDB {
		args = append(args, "-R")
	}

	// Register raster as filesystem
	args = append(args, "-F")

	// Create constraints
	args = append(args, "-C")

	// Overview levels
	for _, level := range opts.Overview {
		args = append(args, "-l", strconv.Itoa(level))
	}

	// Source file and target table
	args = append(args, opts.SourceFile, tableName)

	return args
}

// IsRasterFile checks if a file is a raster format
func IsRasterFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	rasterExts := map[string]bool{
		".tif":  true,
		".tiff": true,
		".img":  true,
		".jp2":  true,
		".ecw":  true,
		".sid":  true,
		".asc":  true,
		".dem":  true,
		".hgt":  true,
		".nc":   true,
		".vrt":  true,
	}
	return rasterExts[ext]
}

// GetRasterExtensions returns supported raster file extensions
func GetRasterExtensions() map[string]string {
	return map[string]string{
		".tif":  "GeoTIFF",
		".tiff": "GeoTIFF",
		".img":  "ERDAS Imagine",
		".jp2":  "JPEG2000",
		".ecw":  "ECW",
		".sid":  "MrSID",
		".asc":  "ASCII Grid",
		".dem":  "USGS DEM",
		".hgt":  "SRTM HGT",
		".nc":   "NetCDF",
		".vrt":  "GDAL VRT",
	}
}
