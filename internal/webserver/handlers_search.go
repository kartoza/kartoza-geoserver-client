package webserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

// SearchResult represents a single search result
type SearchResult struct {
	Type         string   `json:"type"`         // workspace, datastore, coveragestore, layer, style, layergroup
	Name         string   `json:"name"`
	Workspace    string   `json:"workspace,omitempty"`
	StoreName    string   `json:"storeName,omitempty"`
	StoreType    string   `json:"storeType,omitempty"`
	ConnectionID string   `json:"connectionId"`
	ServerName   string   `json:"serverName"`
	Tags         []string `json:"tags"`         // Additional tags/badges
	Description  string   `json:"description,omitempty"`
	Icon         string   `json:"icon"`         // Nerd font icon codepoint
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
