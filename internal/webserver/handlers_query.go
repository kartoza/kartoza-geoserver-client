package webserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/llm"
	"github.com/kartoza/kartoza-cloudbench/internal/query"
)

// handleQueryBuilder handles /api/query/build - build SQL from visual definition
func (s *Server) handleQueryBuilder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var def query.QueryDefinition
	if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	qb := query.FromDefinition(def)
	sql, args, err := qb.Build()
	if err != nil {
		s.jsonError(w, "Failed to build query: "+err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"sql":  sql,
		"args": args,
	})
}

// handleQueryExecute handles /api/query/execute - execute a visual query
func (s *Server) handleQueryExecute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Definition  query.QueryDefinition `json:"definition"`
		ServiceName string                `json:"service_name"`
		MaxRows     int                   `json:"max_rows"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ServiceName == "" {
		s.jsonError(w, "service_name is required", http.StatusBadRequest)
		return
	}

	// Build the SQL
	qb := query.FromDefinition(req.Definition)
	if req.MaxRows > 0 {
		qb.Limit(req.MaxRows)
	}
	sql, _, err := qb.Build()
	if err != nil {
		s.jsonError(w, "Failed to build query: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Execute using the LLM executor (safe, read-only)
	executor := llm.NewQueryExecutor(
		llm.WithMaxRows(req.MaxRows),
		llm.WithTimeout(30*time.Second),
		llm.WithReadOnly(true),
	)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, req.ServiceName, sql)
	if err != nil {
		s.jsonError(w, "Query execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"sql":     sql,
		"result":  result,
	})
}

// handleQuerySave handles /api/query/save - save a query definition
func (s *Server) handleQuerySave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name        string                `json:"name"`
		ServiceName string                `json:"service_name"`
		Definition  query.QueryDefinition `json:"definition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	// Store the query definition in config
	req.Definition.Name = req.Name
	s.config.SaveQuery(req.ServiceName, req.Name, req.Definition)
	if err := s.config.Save(); err != nil {
		s.jsonError(w, "Failed to save query", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Query saved",
		"name":    req.Name,
	})
}

// handleQueryList handles /api/query/list - list saved queries
func (s *Server) handleQueryList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("service")
	queries := s.config.GetQueries(serviceName)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"queries": queries,
	})
}

// handleQueryDelete handles /api/query/delete - delete a saved query
func (s *Server) handleQueryDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("service")
	queryName := r.URL.Query().Get("name")

	if queryName == "" {
		s.jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	s.config.DeleteQuery(serviceName, queryName)
	if err := s.config.Save(); err != nil {
		s.jsonError(w, "Failed to delete query", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleQuery handles /api/query/* routes
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/query")
	path = strings.Trim(path, "/")

	switch path {
	case "build":
		s.handleQueryBuilder(w, r)
	case "execute":
		s.handleQueryExecute(w, r)
	case "save":
		s.handleQuerySave(w, r)
	case "list":
		s.handleQueryList(w, r)
	case "delete":
		s.handleQueryDelete(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
