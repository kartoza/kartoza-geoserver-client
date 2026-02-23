package webserver

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

// S3PreviewMetadata represents metadata for S3 layer preview
type S3PreviewMetadata struct {
	Format        string      `json:"format"`                  // "cog", "copc", "geoparquet", "geojson", "geotiff", "qgisproject", "parquet"
	PreviewType   string      `json:"previewType"`             // "raster", "pointcloud", "vector", "qgisproject", "table"
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

// RasterInfo holds extracted information about a raster file
type RasterInfo struct {
	Bounds    *S3Bounds
	CRS       string
	BandCount int
}

// ParquetInfo holds extracted information about a parquet/geoparquet file
type ParquetInfo struct {
	Bounds       *S3Bounds
	CRS          string
	FeatureCount int64
	FieldNames   []string
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
