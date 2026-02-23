package webserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/qfieldcloud"
)

// handleQFieldCloudConnections handles QFieldCloud connection CRUD.
// GET  /api/qfieldcloud/connections        - list connections
// POST /api/qfieldcloud/connections        - create connection
func (s *Server) handleQFieldCloudConnections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listQFieldCloudConnections(w, r)
	case http.MethodPost:
		s.createQFieldCloudConnection(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleQFieldCloudConnectionByID handles single-connection operations and sub-resources.
// GET    /api/qfieldcloud/connections/{id}           - get connection
// PUT    /api/qfieldcloud/connections/{id}           - update connection
// DELETE /api/qfieldcloud/connections/{id}           - delete connection
// POST   /api/qfieldcloud/connections/{id}/test      - test connection
// GET    /api/qfieldcloud/connections/{id}/projects  - list projects
// POST   /api/qfieldcloud/connections/{id}/projects  - create project
// GET    /api/qfieldcloud/connections/{id}/projects/{projectId}              - get project
// PATCH  /api/qfieldcloud/connections/{id}/projects/{projectId}              - update project
// DELETE /api/qfieldcloud/connections/{id}/projects/{projectId}              - delete project
// GET    /api/qfieldcloud/connections/{id}/projects/{projectId}/files        - list files
// POST   /api/qfieldcloud/connections/{id}/projects/{projectId}/files        - upload file
// DELETE /api/qfieldcloud/connections/{id}/projects/{projectId}/files/{fn}   - delete file
// GET    /api/qfieldcloud/connections/{id}/projects/{projectId}/download/{fn} - download file
// GET    /api/qfieldcloud/connections/{id}/projects/{projectId}/jobs         - list jobs
// POST   /api/qfieldcloud/connections/{id}/projects/{projectId}/jobs         - create job
// GET    /api/qfieldcloud/connections/{id}/projects/{projectId}/collaborators - list collaborators
// POST   /api/qfieldcloud/connections/{id}/projects/{projectId}/collaborators - add collaborator
// PATCH  /api/qfieldcloud/connections/{id}/projects/{projectId}/collaborators/{username} - update
// DELETE /api/qfieldcloud/connections/{id}/projects/{projectId}/collaborators/{username} - remove
// GET    /api/qfieldcloud/connections/{id}/projects/{projectId}/deltas       - list deltas
func (s *Server) handleQFieldCloudConnectionByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/qfieldcloud/connections/")
	parts := strings.SplitN(path, "/", -1)
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Connection ID required", http.StatusBadRequest)
		return
	}
	connID := parts[0]

	// /api/qfieldcloud/connections/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.getQFieldCloudConnection(w, r, connID)
		case http.MethodPut:
			s.updateQFieldCloudConnection(w, r, connID)
		case http.MethodDelete:
			s.deleteQFieldCloudConnection(w, r, connID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/qfieldcloud/connections/{id}/{sub}
	switch parts[1] {
	case "test":
		s.testQFieldCloudConnection(w, r, connID)
	case "projects":
		s.handleQFieldCloudProjects(w, r, connID, parts[2:])
	default:
		http.Error(w, "Unknown sub-resource", http.StatusNotFound)
	}
}

// handleQFieldCloudProjects routes project-level requests.
func (s *Server) handleQFieldCloudProjects(w http.ResponseWriter, r *http.Request, connID string, rest []string) {
	if len(rest) == 0 {
		// /connections/{id}/projects
		switch r.Method {
		case http.MethodGet:
			s.listQFieldCloudProjects(w, r, connID)
		case http.MethodPost:
			s.createQFieldCloudProject(w, r, connID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	projectID := rest[0]

	if len(rest) == 1 {
		// /connections/{id}/projects/{projectId}
		switch r.Method {
		case http.MethodGet:
			s.getQFieldCloudProject(w, r, connID, projectID)
		case http.MethodPatch:
			s.updateQFieldCloudProject(w, r, connID, projectID)
		case http.MethodDelete:
			s.deleteQFieldCloudProject(w, r, connID, projectID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /connections/{id}/projects/{projectId}/{resource}[/...]
	resource := rest[1]
	switch resource {
	case "files":
		s.handleQFieldCloudFiles(w, r, connID, projectID, rest[2:])
	case "download":
		s.handleQFieldCloudDownload(w, r, connID, projectID, rest[2:])
	case "jobs":
		s.handleQFieldCloudJobs(w, r, connID, projectID, rest[2:])
	case "collaborators":
		s.handleQFieldCloudCollaborators(w, r, connID, projectID, rest[2:])
	case "deltas":
		s.handleQFieldCloudDeltas(w, r, connID, projectID)
	default:
		http.Error(w, "Unknown project sub-resource", http.StatusNotFound)
	}
}

// ============================================================================
// Connection CRUD helpers
// ============================================================================

func (s *Server) listQFieldCloudConnections(w http.ResponseWriter, r *http.Request) {
	conns := s.config.QFieldCloudConnections
	if conns == nil {
		conns = []config.QFieldCloudConnection{}
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
	s.jsonResponse(w, safe)
}

func (s *Server) createQFieldCloudConnection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Token    string `json:"token,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	// Default URL to the public QFieldCloud instance
	if input.URL == "" {
		input.URL = "https://app.qfield.cloud"
	}

	conn := config.QFieldCloudConnection{
		ID:       uuid.New().String(),
		Name:     input.Name,
		URL:      strings.TrimSuffix(input.URL, "/"),
		Username: input.Username,
		Password: input.Password,
		Token:    input.Token,
		IsActive: true,
	}

	client := qfieldcloud.NewClient(&conn)
	if err := client.TestConnection(); err != nil {
		http.Error(w, "Connection test failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.config.AddQFieldCloudConnection(conn)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.qfieldcloudClientsMu.Lock()
	s.qfieldcloudClients[conn.ID] = client
	s.qfieldcloudClientsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       conn.ID,
		"name":     conn.Name,
		"url":      conn.URL,
		"username": conn.Username,
	})
}

func (s *Server) getQFieldCloudConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetQFieldCloudConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	s.jsonResponse(w, map[string]interface{}{
		"id":        conn.ID,
		"name":      conn.Name,
		"url":       conn.URL,
		"username":  conn.Username,
		"has_token": conn.Token != "",
		"is_active": conn.IsActive,
	})
}

func (s *Server) updateQFieldCloudConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetQFieldCloudConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	var input struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Token    string `json:"token,omitempty"`
		IsActive *bool  `json:"is_active,omitempty"`
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
	if input.IsActive != nil {
		conn.IsActive = *input.IsActive
	}

	s.config.UpdateQFieldCloudConnection(*conn)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.qfieldcloudClientsMu.Lock()
	s.qfieldcloudClients[conn.ID] = qfieldcloud.NewClient(conn)
	s.qfieldcloudClientsMu.Unlock()

	s.jsonResponse(w, map[string]interface{}{
		"id":       conn.ID,
		"name":     conn.Name,
		"url":      conn.URL,
		"username": conn.Username,
	})
}

func (s *Server) deleteQFieldCloudConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetQFieldCloudConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	s.config.RemoveQFieldCloudConnection(connID)
	if err := s.config.Save(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.qfieldcloudClientsMu.Lock()
	delete(s.qfieldcloudClients, connID)
	s.qfieldcloudClientsMu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) testQFieldCloudConnection(w http.ResponseWriter, r *http.Request, connID string) {
	conn := s.config.GetQFieldCloudConnection(connID)
	if conn == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	client := qfieldcloud.NewClient(conn)
	if err := client.TestConnection(); err != nil {
		s.jsonResponse(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	s.jsonResponse(w, map[string]interface{}{"success": true})
}

// handleQFieldCloudTestConnectionDirect tests a connection without saving.
func (s *Server) handleQFieldCloudTestConnectionDirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		URL      string `json:"url"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Token    string `json:"token,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.URL == "" {
		input.URL = "https://app.qfield.cloud"
	}
	conn := &config.QFieldCloudConnection{
		URL:      strings.TrimSuffix(input.URL, "/"),
		Username: input.Username,
		Password: input.Password,
		Token:    input.Token,
	}
	client := qfieldcloud.NewClient(conn)
	if err := client.TestConnection(); err != nil {
		s.jsonResponse(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	s.jsonResponse(w, map[string]interface{}{"success": true})
}

// getQFieldCloudClient returns (or lazily creates) the QFieldCloud client for a connection.
func (s *Server) getQFieldCloudClient(connID string) *qfieldcloud.Client {
	s.qfieldcloudClientsMu.RLock()
	client, ok := s.qfieldcloudClients[connID]
	s.qfieldcloudClientsMu.RUnlock()
	if ok {
		return client
	}
	conn := s.config.GetQFieldCloudConnection(connID)
	if conn == nil {
		return nil
	}
	client = qfieldcloud.NewClient(conn)
	s.qfieldcloudClientsMu.Lock()
	s.qfieldcloudClients[connID] = client
	s.qfieldcloudClientsMu.Unlock()
	return client
}

// ============================================================================
// Projects
// ============================================================================

func (s *Server) listQFieldCloudProjects(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	projects, err := client.ListProjects()
	if err != nil {
		s.jsonError(w, "Failed to list projects: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, projects)
}

func (s *Server) createQFieldCloudProject(w http.ResponseWriter, r *http.Request, connID string) {
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	var req qfieldcloud.ProjectCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}
	project, err := client.CreateProject(req)
	if err != nil {
		s.jsonError(w, "Failed to create project: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (s *Server) getQFieldCloudProject(w http.ResponseWriter, r *http.Request, connID, projectID string) {
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	project, err := client.GetProject(projectID)
	if err != nil {
		s.jsonError(w, "Failed to get project: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, project)
}

func (s *Server) updateQFieldCloudProject(w http.ResponseWriter, r *http.Request, connID, projectID string) {
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	var req qfieldcloud.ProjectUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	project, err := client.UpdateProject(projectID, req)
	if err != nil {
		s.jsonError(w, "Failed to update project: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, project)
}

func (s *Server) deleteQFieldCloudProject(w http.ResponseWriter, r *http.Request, connID, projectID string) {
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	if err := client.DeleteProject(projectID); err != nil {
		s.jsonError(w, "Failed to delete project: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Files
// ============================================================================

func (s *Server) handleQFieldCloudFiles(w http.ResponseWriter, r *http.Request, connID, projectID string, rest []string) {
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	if len(rest) == 0 {
		switch r.Method {
		case http.MethodGet:
			files, err := client.ListFiles(projectID)
			if err != nil {
				s.jsonError(w, "Failed to list files: "+err.Error(), http.StatusInternalServerError)
				return
			}
			s.jsonResponse(w, files)
		case http.MethodPost:
			// Upload file via multipart form
			if err := r.ParseMultipartForm(64 << 20); err != nil {
				http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
				return
			}
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "No file provided", http.StatusBadRequest)
				return
			}
			defer file.Close()

			remoteName := r.FormValue("filename")
			if remoteName == "" {
				remoteName = header.Filename
			}
			if err := client.UploadFile(projectID, remoteName, file, header.Size); err != nil {
				s.jsonError(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "filename": remoteName})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /files/{filename} - DELETE only
	filename := strings.Join(rest, "/")
	if r.Method == http.MethodDelete {
		if err := client.DeleteFile(projectID, filename); err != nil {
			s.jsonError(w, "Failed to delete file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleQFieldCloudDownload serves a file download.
// GET /connections/{id}/projects/{projectId}/download/{filename...}
func (s *Server) handleQFieldCloudDownload(w http.ResponseWriter, r *http.Request, connID, projectID string, rest []string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if len(rest) == 0 {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	filename := strings.Join(rest, "/")
	data, err := client.DownloadFile(projectID, filename)
	if err != nil {
		s.jsonError(w, "Download failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

// ============================================================================
// Jobs
// ============================================================================

func (s *Server) handleQFieldCloudJobs(w http.ResponseWriter, r *http.Request, connID, projectID string, rest []string) {
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	if len(rest) == 0 {
		switch r.Method {
		case http.MethodGet:
			jobs, err := client.ListJobs(projectID)
			if err != nil {
				s.jsonError(w, "Failed to list jobs: "+err.Error(), http.StatusInternalServerError)
				return
			}
			s.jsonResponse(w, jobs)
		case http.MethodPost:
			var req qfieldcloud.JobCreate
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			req.ProjectID = projectID
			job, err := client.CreateJob(req)
			if err != nil {
				s.jsonError(w, "Failed to create job: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(job)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	jobID := rest[0]
	if r.Method == http.MethodGet {
		job, err := client.GetJob(jobID)
		if err != nil {
			s.jsonError(w, "Failed to get job: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, job)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// ============================================================================
// Collaborators
// ============================================================================

func (s *Server) handleQFieldCloudCollaborators(w http.ResponseWriter, r *http.Request, connID, projectID string, rest []string) {
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	if len(rest) == 0 {
		switch r.Method {
		case http.MethodGet:
			collabs, err := client.ListCollaborators(projectID)
			if err != nil {
				s.jsonError(w, "Failed to list collaborators: "+err.Error(), http.StatusInternalServerError)
				return
			}
			s.jsonResponse(w, collabs)
		case http.MethodPost:
			var req qfieldcloud.CollaboratorCreate
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			collab, err := client.AddCollaborator(projectID, req)
			if err != nil {
				s.jsonError(w, "Failed to add collaborator: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(collab)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	username := rest[0]
	switch r.Method {
	case http.MethodPatch:
		var req qfieldcloud.CollaboratorUpdate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		collab, err := client.UpdateCollaborator(projectID, username, req)
		if err != nil {
			s.jsonError(w, "Failed to update collaborator: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, collab)
	case http.MethodDelete:
		if err := client.RemoveCollaborator(projectID, username); err != nil {
			s.jsonError(w, "Failed to remove collaborator: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ============================================================================
// Deltas
// ============================================================================

func (s *Server) handleQFieldCloudDeltas(w http.ResponseWriter, r *http.Request, connID, projectID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	client := s.getQFieldCloudClient(connID)
	if client == nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	deltas, err := client.ListDeltas(projectID)
	if err != nil {
		s.jsonError(w, "Failed to list deltas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, deltas)
}
