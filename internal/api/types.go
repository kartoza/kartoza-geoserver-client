// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package api

// ServerInfo contains information about the GeoServer instance
type ServerInfo struct {
	GeoServerVersion   string
	GeoServerBuild     string
	GeoServerRevision  string
	GeoToolsVersion    string
	GeoWebCacheVersion string
}

// ServerStatus contains runtime status information
type ServerStatus struct {
	Online           bool    `json:"online"`
	ResponseTimeMs   int64   `json:"responseTimeMs"`
	MemoryUsed       int64   `json:"memoryUsed"`    // bytes
	MemoryFree       int64   `json:"memoryFree"`    // bytes
	MemoryTotal      int64   `json:"memoryTotal"`   // bytes
	MemoryUsedPct    float64 `json:"memoryUsedPct"` // percentage
	CPULoad          float64 `json:"cpuLoad"`       // percentage (if available)
	WorkspaceCount   int     `json:"workspaceCount"`
	LayerCount       int     `json:"layerCount"`
	DataStoreCount   int     `json:"dataStoreCount"`
	CoverageCount    int     `json:"coverageCount"`
	StyleCount       int     `json:"styleCount"`
	Error            string  `json:"error,omitempty"`
	GeoServerVersion string  `json:"geoserverVersion,omitempty"`
}

// SQLViewParameter defines a parameter for a SQL View query
type SQLViewParameter struct {
	Name            string `json:"name"`
	DefaultValue    string `json:"defaultValue,omitempty"`
	RegexpValidator string `json:"regexpValidator,omitempty"`
}

// SQLViewConfig contains configuration for creating a SQL View layer
type SQLViewConfig struct {
	Name           string             // Layer name
	Title          string             // Human-readable title
	Abstract       string             // Description
	SQL            string             // The SQL query
	KeyColumn      string             // Primary key column (optional)
	GeometryColumn string             // Geometry column name
	GeometryType   string             // Geometry type (Point, LineString, Polygon, etc.)
	GeometrySRID   int                // SRID (e.g., 4326)
	Parameters     []SQLViewParameter // Query parameters
	EscapeSql      bool               // Whether to escape SQL
}

// LayerStyles represents the styles associated with a layer
type LayerStyles struct {
	DefaultStyle     string   `json:"defaultStyle"`
	AdditionalStyles []string `json:"styles"`
}

// DataStoreDetails contains full datastore configuration for syncing
type DataStoreDetails struct {
	Name                 string            `json:"name"`
	Description          string            `json:"description,omitempty"`
	Type                 string            `json:"type"`
	Enabled              bool              `json:"enabled"`
	ConnectionParameters map[string]string `json:"connectionParameters"`
}

// CoverageStoreDetails contains full coverage store configuration for syncing
type CoverageStoreDetails struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
	URL         string `json:"url"`
}
