package webserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/mergin"
)

// handleMerginMapsConnections handles Mergin Maps connection CRUD operations.
// GET  /api/mergin/connections  – list all connections
// POST /api/mergin/connections  – create a new connection
func (s *Server) handleMerginMapsConnections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listMerginMapsConnections(w, r)
	case http.MethodPost:
		s.createMerginMapsConnection(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMerginMapsConnectionByID handles single Mergin Maps connection operations.
// GET    /api/mergin/connections/{id}         – get connection details
// PUT    /api/mergin/connections/{id}         – update connection
// DELETE /api/mergin/connections/{id}         – delete connection
// GET    /api/mergin/connections/{id}/test    – test connection
// GET    /api/mergin/connections/{id}/projects – list projects
func (s *Server) handleMerginMapsConnectionByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/mergin/connections/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Connection ID required", http.StatusBadRequest)
		return
	}

	connID := parts[0]

	if len(parts) > 1 {
		switch parts[1] {
		case "test":
			s.testMerginMapsConnection(w, r, connID)
		case "projects":
			s.getMerginMapsProjects(w, r, connID)
		default:
			http.Error(w, "Unknown sub-resource", http.StatusNotFound)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getMerginMapsConnection(w, r, connID)
	case http.MethodPut:
		s.updateMerginMapsConnection(w, r, connID)
	case http.MethodDelete:
		s.deleteMerginMapsConnection(w, r, connID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMerginMapsTestConnection tests a Mergin Maps connection without saving it.
// POST /api/mergin/connections/test
func (s *Server) handleMerginMapsTestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn := &config.MerginMapsConnection{
		URL:      input.URL,
		Username: input.Username,
		Password: input.Password,
		Token:    input.Token,
	}
	client := mergin.NewClient(conn)

	if err := client.TestConnection(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// listMerginMapsConnections returns all Mergin Maps connections (passwords masked).
func (s *Server) listMerginMapsConnections(w http.ResponseWriter, r *http.Request) {
	conns := s.config.MerginMapsConnections
	if conns == nil {
		conns = []config.MerginMapsConnection{}
	}

	safe := make([]map[string]interface{}, len(conns))
	for i, c := range conns {
		safe[i] = map[string]interface{}{
			"id":        c.ID,
			"name":      c.Name,
			"url":       c.URL,
			"username":  c.Username,
			"has_token": c.Token != "",
			"is_active": c.IsActive,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safe)
}

// createMerginMapsConnection creates a new Mergin Maps connection.
func (s *Server) createMerginMapsConnection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	serverURL := input.URL
	if serverURL == "" {
		serverURL = mergin.DefaultURL
	}

	conn := config.MerginMapsConnection{
		ID:       uuid.New().String(),
		Name:     input.Name,
		URL:      strings.TrimSuffix(serverURL, "/"),
		Username: input.Username,
		Password: input.Password,
		Token:    input.Token,
		IsActive: true,
	}

	client := mergin.NewClient(&conn)
	if err := client.TestConnection(); err != nil {
		http.Error(w, "Connection test failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.config.AddMerginMapsConnection(conn)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       conn.ID,
		"name":     conn.Name,
		"url":      conn.URL,
		"username": conn.Username,
	})
}

// getMerginMapsConnection returns a single Mergin Maps connection.
func (s *Server) getMerginMapsConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetMerginMapsConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        conn.ID,
		"name":      conn.Name,
		"url":       conn.URL,
		"username":  conn.Username,
		"has_token": conn.Token != "",
		"is_active": conn.IsActive,
	})
}

// updateMerginMapsConnection updates an existing Mergin Maps connection.
func (s *Server) updateMerginMapsConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetMerginMapsConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	var input struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name != "" {
		conn.Name = input.Name
	}
	if input.URL != "" {
		conn.URL = strings.TrimSuffix(input.URL, "/")
	}
	if input.Username != "" {
		conn.Username = input.Username
	}
	if input.Password != "" {
		conn.Password = input.Password
	}
	if input.Token != "" {
		conn.Token = input.Token
	}

	s.config.UpdateMerginMapsConnection(*conn)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        conn.ID,
		"name":      conn.Name,
		"url":       conn.URL,
		"username":  conn.Username,
		"has_token": conn.Token != "",
		"is_active": conn.IsActive,
	})
}

// deleteMerginMapsConnection deletes a Mergin Maps connection.
func (s *Server) deleteMerginMapsConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetMerginMapsConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	s.config.RemoveMerginMapsConnection(connID)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// testMerginMapsConnection tests an existing saved Mergin Maps connection.
func (s *Server) testMerginMapsConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetMerginMapsConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	client := mergin.NewClient(conn)
	if err := client.TestConnection(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// getMerginMapsProjects lists projects for a Mergin Maps connection.
// Optional query param: namespace (filter to one workspace/namespace)
func (s *Server) getMerginMapsProjects(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetMerginMapsConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	namespace := r.URL.Query().Get("namespace")

	client := mergin.NewClient(conn)
	result, err := client.GetProjects(1, 100, namespace)
	if err != nil {
		http.Error(w, "Failed to fetch projects: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
