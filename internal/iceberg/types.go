// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package iceberg

// ConnectionTestResult represents the result of testing an Iceberg connection
type ConnectionTestResult struct {
	Success        bool              `json:"success"`
	Message        string            `json:"message"`
	NamespaceCount int               `json:"namespaceCount,omitempty"`
	Defaults       map[string]string `json:"defaults,omitempty"`
}

// CatalogConfig represents the catalog configuration response
type CatalogConfig struct {
	Defaults  map[string]string `json:"defaults"`
	Overrides map[string]string `json:"overrides"`
}

// ErrorResponse represents an error from the REST API
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// Namespace represents an Iceberg namespace
type Namespace struct {
	Name string   `json:"name"` // Dot-separated name (e.g., "db.schema")
	Path []string `json:"path"` // Array path (e.g., ["db", "schema"])
}

// NamespacesResponse represents the response from listing namespaces
type NamespacesResponse struct {
	Namespaces [][]string `json:"namespaces"`
}

// NamespaceInfo represents detailed namespace information
type NamespaceInfo struct {
	Namespace  []string          `json:"namespace"`
	Properties map[string]string `json:"properties"`
}

// CreateNamespaceRequest represents a request to create a namespace
type CreateNamespaceRequest struct {
	Namespace  []string          `json:"namespace"`
	Properties map[string]string `json:"properties,omitempty"`
}

// TableIdentifier represents a table's full identifier
type TableIdentifier struct {
	Namespace []string `json:"namespace"`
	Name      string   `json:"name"`
}

// TablesResponse represents the response from listing tables
type TablesResponse struct {
	Identifiers []TableIdentifier `json:"identifiers"`
}

// TableMetadata represents complete table metadata
type TableMetadata struct {
	MetadataLocation string `json:"metadata-location"`
	Metadata         struct {
		FormatVersion   int                `json:"format-version"`
		TableUUID       string             `json:"table-uuid"`
		Location        string             `json:"location"`
		LastSequenceNum int64              `json:"last-sequence-number"`
		LastUpdatedMS   int64              `json:"last-updated-ms"`
		LastColumnID    int                `json:"last-column-id"`
		CurrentSchemaID int                `json:"current-schema-id"`
		Schemas         []Schema           `json:"schemas"`
		PartitionSpecs  []PartitionSpec    `json:"partition-specs"`
		SortOrders      []SortOrder        `json:"sort-orders"`
		Properties      map[string]string  `json:"properties"`
		Snapshots       []Snapshot         `json:"snapshots"`
		SnapshotLog     []SnapshotLogEntry `json:"snapshot-log"`
		CurrentSnapshot *int64             `json:"current-snapshot-id"`
	} `json:"metadata"`
}

// Schema represents an Iceberg table schema
type Schema struct {
	SchemaID           int     `json:"schema-id"`
	Type               string  `json:"type"`
	Fields             []Field `json:"fields"`
	IdentifierFieldIDs []int   `json:"identifier-field-ids,omitempty"`
}

// Field represents a field in a schema
type Field struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Type     interface{} `json:"type"` // Can be string or nested struct
	Required bool        `json:"required"`
	Doc      string      `json:"doc,omitempty"`
}

// PartitionSpec represents a partition specification
type PartitionSpec struct {
	SpecID int              `json:"spec-id"`
	Fields []PartitionField `json:"fields"`
}

// PartitionField represents a field in a partition spec
type PartitionField struct {
	SourceID  int    `json:"source-id"`
	FieldID   int    `json:"field-id"`
	Name      string `json:"name"`
	Transform string `json:"transform"`
}

// SortOrder represents a sort order
type SortOrder struct {
	OrderID int         `json:"order-id"`
	Fields  []SortField `json:"fields"`
}

// SortField represents a field in a sort order
type SortField struct {
	SourceID  int    `json:"source-id"`
	Transform string `json:"transform"`
	Direction string `json:"direction"`
	NullOrder string `json:"null-order"`
}

// Snapshot represents a table snapshot
type Snapshot struct {
	SnapshotID       int64             `json:"snapshot-id"`
	ParentSnapshotID *int64            `json:"parent-snapshot-id,omitempty"`
	SequenceNumber   int64             `json:"sequence-number"`
	TimestampMS      int64             `json:"timestamp-ms"`
	ManifestList     string            `json:"manifest-list"`
	Summary          map[string]string `json:"summary"`
	SchemaID         *int              `json:"schema-id,omitempty"`
}

// SnapshotLogEntry represents an entry in the snapshot log
type SnapshotLogEntry struct {
	TimestampMS int64 `json:"timestamp-ms"`
	SnapshotID  int64 `json:"snapshot-id"`
}

// CreateTableRequest represents a request to create a table
type CreateTableRequest struct {
	Name          string            `json:"name"`
	Location      string            `json:"location,omitempty"`
	Schema        Schema            `json:"schema"`
	PartitionSpec *PartitionSpec    `json:"partition-spec,omitempty"`
	WriteOrder    *SortOrder        `json:"write-order,omitempty"`
	StageCreate   bool              `json:"stage-create,omitempty"`
	Properties    map[string]string `json:"properties,omitempty"`
}

// RenameTableRequest represents a request to rename a table
type RenameTableRequest struct {
	Source      TableIdentifier `json:"source"`
	Destination TableIdentifier `json:"destination"`
}

// TableStats represents statistics about a table
type TableStats struct {
	RowCount          int64 `json:"rowCount"`
	FileCount         int   `json:"fileCount"`
	TotalSizeBytes    int64 `json:"totalSizeBytes"`
	DataFileSizeBytes int64 `json:"dataFileSizeBytes"`
	LastUpdated       int64 `json:"lastUpdated"`
	SnapshotCount     int   `json:"snapshotCount"`
}

// GeometryColumnInfo represents information about a geometry column
type GeometryColumnInfo struct {
	Name         string    `json:"name"`
	GeometryType string    `json:"geometryType,omitempty"` // Point, LineString, Polygon, etc.
	SRID         int       `json:"srid,omitempty"`
	BBox         []float64 `json:"bbox,omitempty"` // [minX, minY, maxX, maxY]
}

// SpatialTableInfo extends table metadata with spatial information
type SpatialTableInfo struct {
	TableIdentifier
	Schema          Schema               `json:"schema"`
	GeometryColumns []GeometryColumnInfo `json:"geometryColumns,omitempty"`
	RowCount        int64                `json:"rowCount,omitempty"`
	IsSpatial       bool                 `json:"isSpatial"`
	LastUpdated     int64                `json:"lastUpdated,omitempty"`
}
