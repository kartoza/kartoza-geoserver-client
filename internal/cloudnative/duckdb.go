package cloudnative

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// DuckDBQueryResult represents the result of a DuckDB query
type DuckDBQueryResult struct {
	Columns      []string                 `json:"columns"`
	ColumnTypes  []string                 `json:"columnTypes"`
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"rowCount"`
	TotalCount   int64                    `json:"totalCount,omitempty"`
	HasMore      bool                     `json:"hasMore"`
	GeometryColumn string                 `json:"geometryColumn,omitempty"`
	SQL          string                   `json:"sql,omitempty"`
}

// DuckDBQueryOptions configures a DuckDB query
type DuckDBQueryOptions struct {
	SQL            string `json:"sql"`
	Limit          int    `json:"limit,omitempty"`
	Offset         int    `json:"offset,omitempty"`
	IncludeGeoJSON bool   `json:"includeGeoJson,omitempty"`
}

// DuckDBTableInfo contains metadata about a Parquet file
type DuckDBTableInfo struct {
	Columns        []DuckDBColumnInfo `json:"columns"`
	RowCount       int64              `json:"rowCount"`
	GeometryColumn string             `json:"geometryColumn,omitempty"`
	BBox           []float64          `json:"bbox,omitempty"`
}

// DuckDBColumnInfo describes a column in a Parquet file
type DuckDBColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// CheckDuckDBAvailable checks if DuckDB CLI is available
func CheckDuckDBAvailable() (bool, string) {
	cmd := exec.Command("duckdb", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false, ""
	}
	return true, strings.TrimSpace(string(output))
}

// GetParquetTableInfo returns metadata about a Parquet/GeoParquet file
func GetParquetTableInfo(ctx context.Context, filePath string) (*DuckDBTableInfo, error) {
	// First, get column information
	columnsSQL := fmt.Sprintf(`
		INSTALL spatial;
		LOAD spatial;
		SELECT column_name, data_type
		FROM (DESCRIBE SELECT * FROM '%s');
	`, escapeFilePath(filePath))

	columnsResult, err := executeDuckDBQuery(ctx, columnsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get column info: %w", err)
	}

	info := &DuckDBTableInfo{
		Columns: make([]DuckDBColumnInfo, 0),
	}

	// Parse column information
	var geometryColumn string
	for _, row := range columnsResult {
		colName := getStringValue(row, "column_name")
		colType := getStringValue(row, "data_type")

		info.Columns = append(info.Columns, DuckDBColumnInfo{
			Name: colName,
			Type: colType,
		})

		// Detect geometry column (common names and types)
		lowerName := strings.ToLower(colName)
		lowerType := strings.ToLower(colType)
		if geometryColumn == "" && (lowerName == "geometry" || lowerName == "geom" || lowerName == "wkb_geometry" ||
			strings.Contains(lowerType, "geometry") || strings.Contains(lowerType, "blob")) {
			geometryColumn = colName
		}
	}
	info.GeometryColumn = geometryColumn

	// Get row count
	countSQL := fmt.Sprintf(`SELECT COUNT(*) as cnt FROM '%s';`, escapeFilePath(filePath))
	countResult, err := executeDuckDBQuery(ctx, countSQL)
	if err == nil && len(countResult) > 0 {
		if cnt, ok := countResult[0]["cnt"]; ok {
			info.RowCount = toInt64(cnt)
		}
	}

	// Try to get bounding box if geometry column exists
	if geometryColumn != "" {
		bboxSQL := fmt.Sprintf(`
			INSTALL spatial;
			LOAD spatial;
			SELECT
				MIN(ST_XMin(ST_GeomFromWKB(%s))) as minx,
				MIN(ST_YMin(ST_GeomFromWKB(%s))) as miny,
				MAX(ST_XMax(ST_GeomFromWKB(%s))) as maxx,
				MAX(ST_YMax(ST_GeomFromWKB(%s))) as maxy
			FROM '%s'
			WHERE %s IS NOT NULL;
		`, geometryColumn, geometryColumn, geometryColumn, geometryColumn, escapeFilePath(filePath), geometryColumn)

		bboxResult, err := executeDuckDBQuery(ctx, bboxSQL)
		if err == nil && len(bboxResult) > 0 {
			minx := toFloat64(bboxResult[0]["minx"])
			miny := toFloat64(bboxResult[0]["miny"])
			maxx := toFloat64(bboxResult[0]["maxx"])
			maxy := toFloat64(bboxResult[0]["maxy"])
			if minx != 0 || miny != 0 || maxx != 0 || maxy != 0 {
				info.BBox = []float64{minx, miny, maxx, maxy}
			}
		}
	}

	return info, nil
}

// QueryParquetFile executes a SQL query against a Parquet/GeoParquet file
func QueryParquetFile(ctx context.Context, filePath string, opts DuckDBQueryOptions) (*DuckDBQueryResult, error) {
	// Apply defaults
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}

	// Build the query with limit/offset
	sql := opts.SQL
	if sql == "" {
		sql = fmt.Sprintf("SELECT * FROM '%s'", escapeFilePath(filePath))
	}

	// Wrap the query to add limit/offset
	wrappedSQL := fmt.Sprintf(`
		INSTALL spatial;
		LOAD spatial;
		WITH user_query AS (%s)
		SELECT * FROM user_query
		LIMIT %d OFFSET %d;
	`, sql, limit+1, opts.Offset) // Fetch one extra to check hasMore

	rows, err := executeDuckDBQuery(ctx, wrappedSQL)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Check if there are more results
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	// Build result
	result := &DuckDBQueryResult{
		Rows:     rows,
		RowCount: len(rows),
		HasMore:  hasMore,
		SQL:      opts.SQL,
	}

	// Extract column names from first row
	if len(rows) > 0 {
		for col := range rows[0] {
			result.Columns = append(result.Columns, col)
		}
	}

	// Detect geometry column
	for _, col := range result.Columns {
		lowerCol := strings.ToLower(col)
		if lowerCol == "geometry" || lowerCol == "geom" || lowerCol == "wkb_geometry" {
			result.GeometryColumn = col
			break
		}
	}

	return result, nil
}

// QueryParquetFileAsGeoJSON executes a query and returns results as GeoJSON
func QueryParquetFileAsGeoJSON(ctx context.Context, filePath string, opts DuckDBQueryOptions) ([]byte, error) {
	// Get table info to find geometry column
	info, err := GetParquetTableInfo(ctx, filePath)
	if err != nil {
		return nil, err
	}

	if info.GeometryColumn == "" {
		return nil, fmt.Errorf("no geometry column found in parquet file")
	}

	// Apply defaults
	limit := opts.Limit
	if limit <= 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	// Build column list excluding geometry
	var nonGeomCols []string
	for _, col := range info.Columns {
		if col.Name != info.GeometryColumn {
			nonGeomCols = append(nonGeomCols, fmt.Sprintf("'%s', \"%s\"", col.Name, col.Name))
		}
	}
	propertiesExpr := "json_object(" + strings.Join(nonGeomCols, ", ") + ")"

	// Build GeoJSON query using DuckDB's spatial extension
	sql := opts.SQL
	if sql == "" {
		sql = fmt.Sprintf("SELECT * FROM '%s'", escapeFilePath(filePath))
	}

	geojsonSQL := fmt.Sprintf(`
		INSTALL spatial;
		LOAD spatial;
		WITH user_query AS (%s),
		limited AS (SELECT * FROM user_query LIMIT %d OFFSET %d)
		SELECT json_group_array(
			json_object(
				'type', 'Feature',
				'geometry', ST_AsGeoJSON(ST_GeomFromWKB(%s))::JSON,
				'properties', %s
			)
		) as features
		FROM limited
		WHERE %s IS NOT NULL;
	`, sql, limit, opts.Offset, info.GeometryColumn, propertiesExpr, info.GeometryColumn)

	rows, err := executeDuckDBQuery(ctx, geojsonSQL)
	if err != nil {
		return nil, fmt.Errorf("GeoJSON query failed: %w", err)
	}

	// Build FeatureCollection
	var features json.RawMessage = []byte("[]")
	if len(rows) > 0 {
		if f, ok := rows[0]["features"]; ok {
			if fStr, ok := f.(string); ok {
				features = json.RawMessage(fStr)
			}
		}
	}

	fc := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	}

	return json.Marshal(fc)
}

// QueryS3ParquetFile queries a Parquet file directly from S3 using DuckDB's httpfs
func QueryS3ParquetFile(ctx context.Context, s3URL, accessKey, secretKey, region string, opts DuckDBQueryOptions) (*DuckDBQueryResult, error) {
	// Apply defaults
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}

	// Build S3 configuration
	s3Config := fmt.Sprintf(`
		INSTALL httpfs;
		LOAD httpfs;
		INSTALL spatial;
		LOAD spatial;
		SET s3_access_key_id='%s';
		SET s3_secret_access_key='%s';
		SET s3_region='%s';
	`, accessKey, secretKey, region)

	sql := opts.SQL
	if sql == "" {
		sql = fmt.Sprintf("SELECT * FROM '%s'", s3URL)
	}

	fullSQL := fmt.Sprintf(`%s
		WITH user_query AS (%s)
		SELECT * FROM user_query
		LIMIT %d OFFSET %d;
	`, s3Config, sql, limit+1, opts.Offset)

	rows, err := executeDuckDBQuery(ctx, fullSQL)
	if err != nil {
		return nil, fmt.Errorf("S3 query failed: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	result := &DuckDBQueryResult{
		Rows:     rows,
		RowCount: len(rows),
		HasMore:  hasMore,
		SQL:      opts.SQL,
	}

	if len(rows) > 0 {
		for col := range rows[0] {
			result.Columns = append(result.Columns, col)
		}
	}

	return result, nil
}

// executeDuckDBQuery executes a SQL query using the DuckDB CLI
func executeDuckDBQuery(ctx context.Context, sql string) ([]map[string]interface{}, error) {
	// Create a temporary file for the query
	tmpFile, err := os.CreateTemp("", "duckdb-query-*.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write SQL to temp file
	if _, err := tmpFile.WriteString(sql); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write SQL: %w", err)
	}
	tmpFile.Close()

	// Execute DuckDB with JSON output
	cmd := exec.CommandContext(ctx, "duckdb", "-json", "-cmd", ".read "+tmpFile.Name())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("duckdb failed: %w, stderr: %s", err, stderr.String())
	}

	// Parse JSON output
	output := strings.TrimSpace(stdout.String())
	if output == "" || output == "[]" {
		return []map[string]interface{}{}, nil
	}

	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w, output: %s", err, output)
	}

	return rows, nil
}

// escapeFilePath escapes a file path for use in DuckDB SQL
func escapeFilePath(path string) string {
	// Escape single quotes by doubling them
	return strings.ReplaceAll(path, "'", "''")
}

// getStringValue safely extracts a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// toInt64 converts an interface to int64
func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

// toFloat64 converts an interface to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

// CreateS3URL builds an S3 URL for DuckDB's httpfs extension
func CreateS3URL(endpoint, bucket, key string, useSSL bool) string {
	// DuckDB expects s3:// URLs
	// For custom endpoints (MinIO), we need to use the endpoint URL format
	protocol := "http"
	if useSSL {
		protocol = "https"
	}

	// Remove protocol from endpoint if present
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	// For MinIO/custom S3, use path-style URLs
	return fmt.Sprintf("%s://%s/%s/%s", protocol, endpoint, bucket, key)
}

// GetSampleQueries returns sample queries for a GeoParquet file
func GetSampleQueries(tableName string, info *DuckDBTableInfo) []string {
	queries := []string{
		fmt.Sprintf("SELECT * FROM '%s' LIMIT 10", tableName),
		fmt.Sprintf("SELECT COUNT(*) as count FROM '%s'", tableName),
	}

	if info.GeometryColumn != "" {
		queries = append(queries,
			fmt.Sprintf("SELECT *, ST_AsText(ST_GeomFromWKB(%s)) as geom_text FROM '%s' LIMIT 10", info.GeometryColumn, tableName),
			fmt.Sprintf("SELECT ST_GeometryType(ST_GeomFromWKB(%s)) as geom_type, COUNT(*) FROM '%s' GROUP BY 1", info.GeometryColumn, tableName),
		)
	}

	// Add queries based on column types
	for _, col := range info.Columns {
		if col.Name == info.GeometryColumn {
			continue
		}
		lowerType := strings.ToLower(col.Type)
		if strings.Contains(lowerType, "varchar") || strings.Contains(lowerType, "string") {
			queries = append(queries,
				fmt.Sprintf("SELECT DISTINCT \"%s\" FROM '%s' LIMIT 20", col.Name, tableName),
			)
			break
		}
	}

	return queries
}

// ValidateSQL performs basic SQL validation (prevents dangerous operations)
func ValidateSQL(sql string) error {
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))

	// Block dangerous statements
	dangerousKeywords := []string{
		"DROP", "DELETE", "TRUNCATE", "INSERT", "UPDATE",
		"CREATE", "ALTER", "GRANT", "REVOKE",
		"COPY", "EXPORT", "ATTACH",
	}

	for _, kw := range dangerousKeywords {
		if strings.HasPrefix(upperSQL, kw) || strings.Contains(upperSQL, " "+kw+" ") {
			return fmt.Errorf("query contains disallowed keyword: %s", kw)
		}
	}

	// Must be a SELECT or WITH statement
	if !strings.HasPrefix(upperSQL, "SELECT") && !strings.HasPrefix(upperSQL, "WITH") {
		return fmt.Errorf("only SELECT and WITH statements are allowed")
	}

	return nil
}

// DownloadAndQueryParquetFile downloads an S3 object to temp and queries it
// This is used when direct S3 access isn't configured
func DownloadAndQueryParquetFile(ctx context.Context, localPath string, opts DuckDBQueryOptions) (*DuckDBQueryResult, error) {
	// Validate SQL
	if opts.SQL != "" {
		if err := ValidateSQL(opts.SQL); err != nil {
			return nil, err
		}
	}

	// Check file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", localPath)
	}

	return QueryParquetFile(ctx, localPath, opts)
}

// GetParquetFileExtent returns the geographic extent of a GeoParquet file
func GetParquetFileExtent(ctx context.Context, filePath string) ([]float64, error) {
	info, err := GetParquetTableInfo(ctx, filePath)
	if err != nil {
		return nil, err
	}

	if len(info.BBox) == 4 {
		return info.BBox, nil
	}

	return nil, fmt.Errorf("no extent available")
}

// GetAbsPath converts a path to absolute
func GetAbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}
