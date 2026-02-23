// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package webserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
	"github.com/kartoza/kartoza-cloudbench/internal/postgres"
)

// SearchResult represents a single search result
type SearchResult struct {
	Type         string   `json:"type"` // workspace, datastore, coveragestore, layer, style, layergroup, pgservice, pgschema, pgtable, pgview, pgcolumn, pgfunction, s3connection, s3bucket, s3object, qgisproject, geonodeconnection, geonodedataset, geonodemap, geonodedocument
	Name         string   `json:"name"`
	Workspace    string   `json:"workspace,omitempty"`
	StoreName    string   `json:"storeName,omitempty"`
	StoreType    string   `json:"storeType,omitempty"`
	ConnectionID string   `json:"connectionId"`
	ServerName   string   `json:"serverName"`
	Tags         []string `json:"tags"` // Additional tags/badges
	Description  string   `json:"description,omitempty"`
	Icon         string   `json:"icon"` // Nerd font icon codepoint
	// PostgreSQL-specific fields
	ServiceName string `json:"serviceName,omitempty"` // PostgreSQL service name
	SchemaName  string `json:"schemaName,omitempty"`  // PostgreSQL schema name
	TableName   string `json:"tableName,omitempty"`   // PostgreSQL table/view name
	DataType    string `json:"dataType,omitempty"`    // Column data type or function return type
	// S3-specific fields
	S3ConnectionID string `json:"s3ConnectionId,omitempty"` // S3 connection ID
	S3Bucket       string `json:"s3Bucket,omitempty"`       // S3 bucket name
	S3Key          string `json:"s3Key,omitempty"`          // S3 object key
	// QGIS-specific fields
	QGISProjectID   string `json:"qgisProjectId,omitempty"`   // QGIS project ID
	QGISProjectPath string `json:"qgisProjectPath,omitempty"` // QGIS project file path
	// GeoNode-specific fields
	GeoNodeConnectionID string `json:"geonodeConnectionId,omitempty"` // GeoNode connection ID
	GeoNodeResourcePK   int    `json:"geonodeResourcePk,omitempty"`   // GeoNode resource primary key
	GeoNodeAlternate    string `json:"geonodeAlternate,omitempty"`    // GeoNode dataset alternate (workspace:layer)
	GeoNodeURL          string `json:"geonodeUrl,omitempty"`          // GeoNode base URL
}

// SearchResponse represents the search API response
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}

// handleSearch handles the universal search endpoint
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if query == "" {
		s.jsonResponse(w, SearchResponse{
			Query:   "",
			Results: []SearchResult{},
			Total:   0,
		})
		return
	}

	// Optional connection filter
	connFilter := r.URL.Query().Get("connection")

	var results []SearchResult

	// Search across all connections
	for _, conn := range s.config.Connections {
		if connFilter != "" && conn.ID != connFilter {
			continue
		}

		client := s.getClient(conn.ID)
		if client == nil {
			continue
		}

		// Search workspaces
		workspaces, err := client.GetWorkspaces()
		if err == nil {
			for _, ws := range workspaces {
				if matchesQuery(ws.Name, query) {
					results = append(results, SearchResult{
						Type:         "workspace",
						Name:         ws.Name,
						ConnectionID: conn.ID,
						ServerName:   conn.Name,
						Tags:         []string{"Workspace"},
						Icon:         "\uf07b", // fa-folder
					})
				}

				// Search data stores in this workspace
				dataStores, err := client.GetDataStores(ws.Name)
				if err == nil {
					for _, ds := range dataStores {
						if matchesQuery(ds.Name, query) {
							tags := []string{"Data Store"}
							if ds.Type != "" {
								tags = append(tags, ds.Type)
							}
							results = append(results, SearchResult{
								Type:         "datastore",
								Name:         ds.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         tags,
								Icon:         "\uf1c0", // fa-database
							})
						}
					}
				}

				// Search coverage stores in this workspace
				coverageStores, err := client.GetCoverageStores(ws.Name)
				if err == nil {
					for _, cs := range coverageStores {
						if matchesQuery(cs.Name, query) {
							tags := []string{"Coverage Store"}
							if cs.Type != "" {
								tags = append(tags, cs.Type)
							}
							results = append(results, SearchResult{
								Type:         "coveragestore",
								Name:         cs.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         tags,
								Icon:         "\uf03e", // fa-image
							})
						}
					}
				}

				// Search layers in this workspace
				layers, err := client.GetLayers(ws.Name)
				if err == nil {
					for _, layer := range layers {
						if matchesQuery(layer.Name, query) {
							tags := []string{"Layer"}
							if layer.Type != "" {
								tags = append(tags, layer.Type)
							}
							results = append(results, SearchResult{
								Type:         "layer",
								Name:         layer.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         tags,
								Icon:         "\uf5fd", // fa-layer-group
							})
						}
					}
				}

				// Search styles in this workspace
				styles, err := client.GetStyles(ws.Name)
				if err == nil {
					for _, style := range styles {
						if matchesQuery(style.Name, query) {
							tags := []string{"Style"}
							if style.Format != "" {
								tags = append(tags, strings.ToUpper(style.Format))
							}
							results = append(results, SearchResult{
								Type:         "style",
								Name:         style.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         tags,
								Icon:         "\uf53f", // fa-palette
							})
						}
					}
				}

				// Search layer groups in this workspace
				layerGroups, err := client.GetLayerGroups(ws.Name)
				if err == nil {
					for _, lg := range layerGroups {
						if matchesQuery(lg.Name, query) {
							results = append(results, SearchResult{
								Type:         "layergroup",
								Name:         lg.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         []string{"Layer Group"},
								Icon:         "\uf02d", // fa-book
							})
						}
					}
				}
			}
		}

		// Also search global styles (no workspace)
		globalStyles, err := client.GetStyles("")
		if err == nil {
			for _, style := range globalStyles {
				if matchesQuery(style.Name, query) {
					tags := []string{"Style", "Global"}
					if style.Format != "" {
						tags = append(tags, strings.ToUpper(style.Format))
					}
					results = append(results, SearchResult{
						Type:         "style",
						Name:         style.Name,
						ConnectionID: conn.ID,
						ServerName:   conn.Name,
						Tags:         tags,
						Icon:         "\uf53f", // fa-palette
					})
				}
			}
		}

		// Search global layer groups
		globalLayerGroups, err := client.GetLayerGroups("")
		if err == nil {
			for _, lg := range globalLayerGroups {
				if matchesQuery(lg.Name, query) {
					results = append(results, SearchResult{
						Type:         "layergroup",
						Name:         lg.Name,
						ConnectionID: conn.ID,
						ServerName:   conn.Name,
						Tags:         []string{"Layer Group", "Global"},
						Icon:         "\uf02d", // fa-book
					})
				}
			}
		}
	}

	// Search PostgreSQL services
	if postgres.PGServiceFileExists() {
		pgResults := s.searchPostgreSQLEntities(query)
		results = append(results, pgResults...)
	}

	// Search S3 connections
	s3Results := s.searchS3Entities(query)
	results = append(results, s3Results...)

	// Search QGIS projects
	qgisResults := s.searchQGISProjects(query)
	results = append(results, qgisResults...)

	// Search GeoNode connections and resources
	geonodeResults := s.searchGeoNodeEntities(query)
	results = append(results, geonodeResults...)

	// Limit results
	maxResults := 50
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	s.jsonResponse(w, SearchResponse{
		Query:   query,
		Results: results,
		Total:   len(results),
	})
}

// matchesQuery checks if a name matches the search query
func matchesQuery(name, query string) bool {
	return strings.Contains(strings.ToLower(name), query)
}

// getTypeIcon returns the appropriate Nerd Font icon for a node type
func getTypeIcon(nodeType models.NodeType) string {
	return nodeType.Icon()
}

// handleSearchSuggestions returns popular/recent search suggestions
func (s *Server) handleSearchSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return common resource type suggestions
	suggestions := []string{
		"layers",
		"styles",
		"workspaces",
		"raster",
		"vector",
		"shapefile",
		"geotiff",
		"geopackage",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"suggestions": suggestions,
	})
}

// searchPostgreSQLEntities searches across all PostgreSQL services for entities matching the query
func (s *Server) searchPostgreSQLEntities(query string) []SearchResult {
	var results []SearchResult

	services, err := postgres.ParsePGServiceFile()
	if err != nil {
		return results
	}

	for _, svc := range services {
		// Skip hidden services
		if svc.Hidden {
			continue
		}

		// Search service name
		if matchesQuery(svc.Name, query) {
			results = append(results, SearchResult{
				Type:        "pgservice",
				Name:        svc.Name,
				ServerName:  svc.Name,
				ServiceName: svc.Name,
				Tags:        []string{"PostgreSQL", "Service"},
				Icon:        "\ue76e", // postgresql icon
				Description: svc.DBName,
			})
		}

		// Connect to the database to search entities
		db, err := svc.Connect()
		if err != nil {
			continue
		}

		// Search schemas, tables, views, columns, and functions
		results = append(results, s.searchPGSchemas(db, svc.Name, query)...)
		results = append(results, s.searchPGTables(db, svc.Name, query)...)
		results = append(results, s.searchPGViews(db, svc.Name, query)...)
		results = append(results, s.searchPGColumns(db, svc.Name, query)...)
		results = append(results, s.searchPGFunctions(db, svc.Name, query)...)

		db.Close()
	}

	return results
}

// searchPGSchemas searches for PostgreSQL schemas matching the query
func (s *Server) searchPGSchemas(db *sql.DB, serviceName, query string) []SearchResult {
	var results []SearchResult

	rows, err := db.Query(`
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT LIKE 'pg_%'
		  AND schema_name != 'information_schema'
		  AND LOWER(schema_name) LIKE '%' || $1 || '%'
		ORDER BY schema_name
		LIMIT 20
	`, query)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			continue
		}
		results = append(results, SearchResult{
			Type:        "pgschema",
			Name:        schemaName,
			ServerName:  serviceName,
			ServiceName: serviceName,
			SchemaName:  schemaName,
			Tags:        []string{"PostgreSQL", "Schema"},
			Icon:        "\uf07b", // folder icon
		})
	}

	return results
}

// searchPGTables searches for PostgreSQL tables matching the query
func (s *Server) searchPGTables(db *sql.DB, serviceName, query string) []SearchResult {
	var results []SearchResult

	rows, err := db.Query(`
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_type = 'BASE TABLE'
		  AND table_schema NOT LIKE 'pg_%'
		  AND table_schema != 'information_schema'
		  AND LOWER(table_name) LIKE '%' || $1 || '%'
		ORDER BY table_schema, table_name
		LIMIT 30
	`, query)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, tableName string
		if err := rows.Scan(&schemaName, &tableName); err != nil {
			continue
		}
		results = append(results, SearchResult{
			Type:        "pgtable",
			Name:        tableName,
			ServerName:  serviceName,
			ServiceName: serviceName,
			SchemaName:  schemaName,
			Tags:        []string{"PostgreSQL", "Table", schemaName},
			Icon:        "\uf0ce", // table icon
			Description: schemaName + "." + tableName,
		})
	}

	return results
}

// searchPGViews searches for PostgreSQL views matching the query
func (s *Server) searchPGViews(db *sql.DB, serviceName, query string) []SearchResult {
	var results []SearchResult

	rows, err := db.Query(`
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_type = 'VIEW'
		  AND table_schema NOT LIKE 'pg_%'
		  AND table_schema != 'information_schema'
		  AND LOWER(table_name) LIKE '%' || $1 || '%'
		ORDER BY table_schema, table_name
		LIMIT 20
	`, query)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, viewName string
		if err := rows.Scan(&schemaName, &viewName); err != nil {
			continue
		}
		results = append(results, SearchResult{
			Type:        "pgview",
			Name:        viewName,
			ServerName:  serviceName,
			ServiceName: serviceName,
			SchemaName:  schemaName,
			Tags:        []string{"PostgreSQL", "View", schemaName},
			Icon:        "\uf06e", // eye icon
			Description: schemaName + "." + viewName,
		})
	}

	return results
}

// searchPGColumns searches for PostgreSQL columns matching the query
func (s *Server) searchPGColumns(db *sql.DB, serviceName, query string) []SearchResult {
	var results []SearchResult

	rows, err := db.Query(`
		SELECT table_schema, table_name, column_name, data_type
		FROM information_schema.columns
		WHERE table_schema NOT LIKE 'pg_%'
		  AND table_schema != 'information_schema'
		  AND LOWER(column_name) LIKE '%' || $1 || '%'
		ORDER BY table_schema, table_name, ordinal_position
		LIMIT 30
	`, query)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, tableName, columnName, dataType string
		if err := rows.Scan(&schemaName, &tableName, &columnName, &dataType); err != nil {
			continue
		}
		results = append(results, SearchResult{
			Type:        "pgcolumn",
			Name:        columnName,
			ServerName:  serviceName,
			ServiceName: serviceName,
			SchemaName:  schemaName,
			TableName:   tableName,
			DataType:    dataType,
			Tags:        []string{"PostgreSQL", "Column", dataType},
			Icon:        "\uf0db", // columns icon
			Description: schemaName + "." + tableName + "." + columnName,
		})
	}

	return results
}

// searchPGFunctions searches for PostgreSQL functions matching the query
func (s *Server) searchPGFunctions(db *sql.DB, serviceName, query string) []SearchResult {
	var results []SearchResult

	rows, err := db.Query(`
		SELECT n.nspname AS schema_name,
		       p.proname AS function_name,
		       pg_catalog.pg_get_function_result(p.oid) AS return_type
		FROM pg_catalog.pg_proc p
		LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'
		  AND LOWER(p.proname) LIKE '%' || $1 || '%'
		ORDER BY n.nspname, p.proname
		LIMIT 20
	`, query)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, funcName, returnType string
		if err := rows.Scan(&schemaName, &funcName, &returnType); err != nil {
			continue
		}
		results = append(results, SearchResult{
			Type:        "pgfunction",
			Name:        funcName,
			ServerName:  serviceName,
			ServiceName: serviceName,
			SchemaName:  schemaName,
			DataType:    returnType,
			Tags:        []string{"PostgreSQL", "Function"},
			Icon:        "\uf121", // code icon
			Description: schemaName + "." + funcName + "()",
		})
	}

	return results
}

// searchS3Entities searches S3 connections and buckets
func (s *Server) searchS3Entities(query string) []SearchResult {
	var results []SearchResult

	for _, conn := range s.config.S3Connections {
		// Search connection name
		if matchesQuery(conn.Name, query) {
			results = append(results, SearchResult{
				Type:           "s3connection",
				Name:           conn.Name,
				ServerName:     conn.Name,
				S3ConnectionID: conn.ID,
				Tags:           []string{"S3", "Connection"},
				Icon:           "\uf0c2", // cloud icon
				Description:    conn.Endpoint,
			})
		}

		// Try to list buckets and search them
		client := s.getS3Client(conn.ID)
		if client == nil {
			continue
		}

		buckets, err := client.ListBuckets(context.Background())
		if err != nil {
			continue
		}

		for _, bucket := range buckets {
			if matchesQuery(bucket.Name, query) {
				results = append(results, SearchResult{
					Type:           "s3bucket",
					Name:           bucket.Name,
					ServerName:     conn.Name,
					S3ConnectionID: conn.ID,
					S3Bucket:       bucket.Name,
					Tags:           []string{"S3", "Bucket"},
					Icon:           "\uf0e8", // bucket/sitemap icon
					Description:    conn.Name + "/" + bucket.Name,
				})
			}
		}
	}

	return results
}

// searchQGISProjects searches QGIS projects
func (s *Server) searchQGISProjects(query string) []SearchResult {
	var results []SearchResult

	for _, project := range s.config.QGISProjects {
		// Search project name and title
		if matchesQuery(project.Name, query) || (project.Title != "" && matchesQuery(project.Title, query)) {
			results = append(results, SearchResult{
				Type:            "qgisproject",
				Name:            project.Name,
				ServerName:      "QGIS Projects",
				QGISProjectID:   project.ID,
				QGISProjectPath: project.Path,
				Tags:            []string{"QGIS", "Project"},
				Icon:            "\uf279", // map icon
				Description:     project.Path,
			})
		}
	}

	return results
}

// searchGeoNodeEntities searches GeoNode connections and resources
func (s *Server) searchGeoNodeEntities(query string) []SearchResult {
	var results []SearchResult

	for _, conn := range s.config.GeoNodeConnections {
		// Search connection name
		if matchesQuery(conn.Name, query) {
			results = append(results, SearchResult{
				Type:                "geonodeconnection",
				Name:                conn.Name,
				ServerName:          conn.Name,
				GeoNodeConnectionID: conn.ID,
				GeoNodeURL:          conn.URL,
				Tags:                []string{"GeoNode", "Connection"},
				Icon:                "\uf0ac", // globe icon
				Description:         conn.URL,
			})
		}

		// Get GeoNode client and search resources
		client := s.getGeoNodeClient(conn.ID)
		if client == nil {
			continue
		}

		// Search datasets - fetch all and filter locally
		datasetsResp, err := client.GetDatasets(1, 100)
		if err == nil {
			for _, dataset := range datasetsResp.Datasets {
				// Local filtering by query
				if matchesQuery(dataset.Title, query) || matchesQuery(dataset.Alternate, query) {
					results = append(results, SearchResult{
						Type:                "geonodedataset",
						Name:                dataset.Title,
						ServerName:          conn.Name,
						GeoNodeConnectionID: conn.ID,
						GeoNodeResourcePK:   dataset.PK.Int(),
						GeoNodeAlternate:    dataset.Alternate,
						GeoNodeURL:          conn.URL,
						Tags:                []string{"GeoNode", "Dataset"},
						Icon:                "\uf5fd", // layers icon
						Description:         dataset.Alternate,
					})
				}
			}
		}

		// Search maps - fetch all and filter locally
		mapsResp, err := client.GetMaps(1, 50)
		if err == nil {
			for _, m := range mapsResp.Maps {
				// Local filtering by query
				if matchesQuery(m.Title, query) {
					results = append(results, SearchResult{
						Type:                "geonodemap",
						Name:                m.Title,
						ServerName:          conn.Name,
						GeoNodeConnectionID: conn.ID,
						GeoNodeResourcePK:   m.PK.Int(),
						GeoNodeURL:          conn.URL,
						Tags:                []string{"GeoNode", "Map"},
						Icon:                "\uf279", // map icon
					})
				}
			}
		}

		// Search documents - fetch all and filter locally
		docsResp, err := client.GetDocuments(1, 50)
		if err == nil {
			for _, doc := range docsResp.Documents {
				// Local filtering by query
				if matchesQuery(doc.Title, query) {
					results = append(results, SearchResult{
						Type:                "geonodedocument",
						Name:                doc.Title,
						ServerName:          conn.Name,
						GeoNodeConnectionID: conn.ID,
						GeoNodeResourcePK:   doc.PK.Int(),
						GeoNodeURL:          conn.URL,
						Tags:                []string{"GeoNode", "Document"},
						Icon:                "\uf15b", // file icon
					})
				}
			}
		}
	}

	return results
}
