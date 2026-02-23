// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"

	"github.com/kartoza/kartoza-cloudbench/internal/api"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/models"
	"github.com/kartoza/kartoza-cloudbench/internal/postgres"
)

// LinkedStore tracks the relationship between a PostgreSQL connection and a GeoServer PostGIS store
type LinkedStore struct {
	GeoServerConnectionID string `json:"geoserver_connection_id"`
	Workspace             string `json:"workspace"`
	StoreName             string `json:"store_name"`
	PGServiceName         string `json:"pg_service_name"`
	SchemaName            string `json:"schema_name"`
	CreatedAt             string `json:"created_at,omitempty"`
}

// PostGISStoreParams contains the parameters for creating a PostGIS data store
type PostGISStoreParams struct {
	Name        string
	Description string
	Host        string
	Port        string
	Database    string
	Schema      string
	User        string
	Password    string
	SSLMode     string
	// Advanced options
	MinConnections      int
	MaxConnections      int
	ConnectionTimeout   int
	ValidateConnections bool
	FetchSize           int
	ExposeKeys          bool
	LooseBBox           bool
	PreparedStatements  bool
}

// DefaultPostGISStoreParams returns sensible defaults for PostGIS store parameters
func DefaultPostGISStoreParams() PostGISStoreParams {
	return PostGISStoreParams{
		Schema:              "public",
		MinConnections:      1,
		MaxConnections:      10,
		ConnectionTimeout:   20,
		ValidateConnections: true,
		FetchSize:           1000,
		ExposeKeys:          true,
		LooseBBox:           true,
		PreparedStatements:  true,
	}
}

// CreateStoreFromPGService creates a GeoServer PostGIS data store from a PostgreSQL service
func CreateStoreFromPGService(
	gsClient *api.Client,
	workspace string,
	storeName string,
	pgServiceName string,
	schema string,
) error {
	// Get the PostgreSQL service details
	services, err := postgres.ParsePGServiceFile()
	if err != nil {
		return fmt.Errorf("failed to parse pg_service.conf: %w", err)
	}

	svc, err := postgres.GetServiceByName(services, pgServiceName)
	if err != nil {
		return fmt.Errorf("PostgreSQL service not found: %w", err)
	}

	// Build store parameters
	params := DefaultPostGISStoreParams()
	params.Name = storeName
	params.Host = svc.Host
	params.Port = svc.Port
	params.Database = svc.DBName
	params.User = svc.User
	params.Password = svc.Password
	params.SSLMode = svc.SSLMode

	if schema != "" {
		params.Schema = schema
	}

	// Create the data store in GeoServer
	return createPostGISDataStore(gsClient, workspace, params)
}

// createPostGISDataStore creates a PostGIS data store in GeoServer
func createPostGISDataStore(client *api.Client, workspace string, params PostGISStoreParams) error {
	// Prepare SSL mode for JDBC
	sslMode := params.SSLMode
	if sslMode == "" {
		sslMode = "ALLOW"
	} else {
		// Map pg_service.conf sslmode to JDBC ssl modes
		switch sslMode {
		case "disable":
			sslMode = "DISABLE"
		case "allow":
			sslMode = "ALLOW"
		case "prefer":
			sslMode = "PREFER"
		case "require":
			sslMode = "REQUIRE"
		case "verify-ca", "verify-full":
			sslMode = "VERIFY"
		default:
			sslMode = "ALLOW"
		}
	}

	// Prepare port
	port := params.Port
	if port == "" {
		port = "5432"
	}

	// Build connection parameters map
	connParams := map[string]string{
		"host":                 params.Host,
		"port":                 port,
		"database":             params.Database,
		"schema":               params.Schema,
		"user":                 params.User,
		"passwd":               params.Password,
		"SSL mode":             sslMode,
		"min connections":      fmt.Sprintf("%d", params.MinConnections),
		"max connections":      fmt.Sprintf("%d", params.MaxConnections),
		"Connection timeout":   fmt.Sprintf("%d", params.ConnectionTimeout),
		"validate connections": fmt.Sprintf("%t", params.ValidateConnections),
		"fetch size":           fmt.Sprintf("%d", params.FetchSize),
		"Expose primary keys":  fmt.Sprintf("%t", params.ExposeKeys),
		"Loose bbox":           fmt.Sprintf("%t", params.LooseBBox),
		"preparedStatements":   fmt.Sprintf("%t", params.PreparedStatements),
	}

	// Use the API client to create the store using PostGIS type
	return client.CreateDataStore(workspace, params.Name, models.DataStoreTypePostGIS, connParams)
}

// PublishTableAsLayer publishes a PostgreSQL table as a GeoServer layer
func PublishTableAsLayer(
	gsClient *api.Client,
	workspace string,
	storeName string,
	tableName string,
) error {
	// Use the API client to publish the feature type
	// The table name becomes the feature type name in GeoServer
	return gsClient.PublishFeatureType(workspace, storeName, tableName)
}

// BridgeOptions contains options for creating a PostgreSQL to GeoServer bridge
type BridgeOptions struct {
	PGServiceName         string   // PostgreSQL service name
	GeoServerConnectionID string   // GeoServer connection ID
	Workspace             string   // Target workspace
	StoreName             string   // Store name to create
	Schema                string   // PostgreSQL schema (default: public)
	Tables                []string // Specific tables to publish (empty = all)
	PublishLayers         bool     // Whether to publish tables as layers
}

// CreateBridge creates a complete bridge from PostgreSQL to GeoServer
func CreateBridge(
	cfg *config.Config,
	clients map[string]*api.Client,
	opts BridgeOptions,
) (*LinkedStore, error) {
	// Get the GeoServer client
	gsClient, ok := clients[opts.GeoServerConnectionID]
	if !ok {
		return nil, fmt.Errorf("GeoServer connection not found")
	}

	// Create the PostGIS data store
	if err := CreateStoreFromPGService(
		gsClient,
		opts.Workspace,
		opts.StoreName,
		opts.PGServiceName,
		opts.Schema,
	); err != nil {
		return nil, fmt.Errorf("failed to create data store: %w", err)
	}

	// Optionally publish tables as layers
	if opts.PublishLayers && len(opts.Tables) > 0 {
		for _, table := range opts.Tables {
			if err := PublishTableAsLayer(gsClient, opts.Workspace, opts.StoreName, table); err != nil {
				// Log but continue - don't fail the whole operation
				fmt.Printf("Warning: failed to publish table %s: %v\n", table, err)
			}
		}
	}

	// Create linked store record
	link := &LinkedStore{
		GeoServerConnectionID: opts.GeoServerConnectionID,
		Workspace:             opts.Workspace,
		StoreName:             opts.StoreName,
		PGServiceName:         opts.PGServiceName,
		SchemaName:            opts.Schema,
	}

	return link, nil
}

// GetAvailableTables returns tables available for publishing from a PostgreSQL service
func GetAvailableTables(pgServiceName string) ([]string, error) {
	services, err := postgres.ParsePGServiceFile()
	if err != nil {
		return nil, err
	}

	svc, err := postgres.GetServiceByName(services, pgServiceName)
	if err != nil {
		return nil, err
	}

	db, err := svc.Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query for tables with geometry columns (PostGIS spatial tables)
	query := `
		SELECT DISTINCT f_table_name
		FROM geometry_columns
		WHERE f_table_schema = 'public'
		ORDER BY f_table_name
	`

	rows, err := db.Query(query)
	if err != nil {
		// If geometry_columns doesn't exist, fall back to all tables
		query = `
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`
		rows, err = db.Query(query)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	return tables, rows.Err()
}
