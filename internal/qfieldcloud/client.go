// Package qfieldcloud provides a client for the QFieldCloud REST API.
// API reference: https://docs.qfield.org/reference/qfieldcloud/api/
package qfieldcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/config"
)

const defaultBaseURL = "https://app.qfield.cloud"

// Client is a QFieldCloud API client.
type Client struct {
	baseURL    string
	token      string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new QFieldCloud API client from a connection config.
func NewClient(conn *config.QFieldCloudConnection) *Client {
	baseURL := strings.TrimSuffix(conn.URL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		baseURL:  baseURL,
		token:    conn.Token,
		username: conn.Username,
		password: conn.Password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication.
func (c *Client) doRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	reqURL := c.baseURL + path

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Token "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// doJSON performs a JSON request and decodes the response body into dest.
func (c *Client) doJSON(method, path string, reqBody, dest interface{}) error {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	resp, err := c.doRequest(method, path, body, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if dest != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dest); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// ============================================================================
// Authentication
// ============================================================================

// LoginResponse is returned by the login endpoint.
type LoginResponse struct {
	Token string `json:"token"`
}

// Login authenticates with username/password and returns a token.
func (c *Client) Login(username, password string) (string, error) {
	payload := map[string]string{
		"username": username,
		"password": password,
	}
	var result LoginResponse
	if err := c.doJSON("POST", "/api/v1/auth/login/", payload, &result); err != nil {
		return "", err
	}
	return result.Token, nil
}

// Logout invalidates the current token.
func (c *Client) Logout() error {
	return c.doJSON("POST", "/api/v1/auth/logout/", nil, nil)
}

// User represents a QFieldCloud user account.
type User struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// GetCurrentUser returns details about the authenticated user.
func (c *Client) GetCurrentUser() (*User, error) {
	var user User
	if err := c.doJSON("GET", "/api/v1/auth/user/", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// TestConnection verifies that the connection is working.
func (c *Client) TestConnection() error {
	_, err := c.GetCurrentUser()
	return err
}

// ============================================================================
// Projects
// ============================================================================

// Project represents a QFieldCloud project.
type Project struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Owner                 string `json:"owner"`
	Description           string `json:"description"`
	IsPublic              bool   `json:"is_public"`
	CanRepackage          bool   `json:"can_repackage"`
	NeedsRepackaging      bool   `json:"needs_repackaging"`
	StorageSize           int64  `json:"file_storage_bytes"`
	Status                string `json:"status"`
	LastPackagedAt        string `json:"last_packaged_at"`
	DataLastPackagedAt    string `json:"data_last_packaged_at"`
	OwnerUsername         string `json:"owner_username,omitempty"`
	UserRole              string `json:"user_role,omitempty"`
}

// ProjectCreate holds the required fields for creating a project.
type ProjectCreate struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public,omitempty"`
}

// ProjectUpdate holds the fields for updating a project.
type ProjectUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IsPublic    *bool  `json:"is_public,omitempty"`
}

// ListProjects returns all projects accessible to the authenticated user.
func (c *Client) ListProjects() ([]Project, error) {
	var projects []Project
	if err := c.doJSON("GET", "/api/v1/projects/", nil, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject returns a single project by ID.
func (c *Client) GetProject(projectID string) (*Project, error) {
	var project Project
	if err := c.doJSON("GET", "/api/v1/projects/"+projectID+"/", nil, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProject creates a new project.
func (c *Client) CreateProject(req ProjectCreate) (*Project, error) {
	var project Project
	if err := c.doJSON("POST", "/api/v1/projects/", req, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProject updates an existing project.
func (c *Client) UpdateProject(projectID string, req ProjectUpdate) (*Project, error) {
	var project Project
	if err := c.doJSON("PATCH", "/api/v1/projects/"+projectID+"/", req, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject deletes a project.
func (c *Client) DeleteProject(projectID string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/projects/"+projectID+"/", nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ============================================================================
// Files
// ============================================================================

// ProjectFile represents a file within a QFieldCloud project.
type ProjectFile struct {
	Name               string    `json:"name"`
	Size               int64     `json:"size"`
	SHA256             string    `json:"sha256"`
	LastModified       string    `json:"last_modified"`
	IsPackagingFile    bool      `json:"is_packaging_file"`
	VersionsCount      int       `json:"versions_count"`
}

// ListFiles returns all files for a project.
func (c *Client) ListFiles(projectID string) ([]ProjectFile, error) {
	var files []ProjectFile
	if err := c.doJSON("GET", "/api/v1/files/"+projectID+"/", nil, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// UploadFile uploads a file to a project.
// filename is the remote path within the project (e.g. "mymap.qgs" or "data/layer.gpkg").
func (c *Client) UploadFile(projectID, filename string, content io.Reader, contentSize int64) error {
	// Build multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, content); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Use a longer timeout for file uploads
	uploadClient := &http.Client{Timeout: 10 * time.Minute}
	reqURL := c.baseURL + "/api/v1/files/" + projectID + "/" + url.PathEscape(filename) + "/"
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Token "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := uploadClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// DownloadFile downloads a file from a project and returns its content.
func (c *Client) DownloadFile(projectID, filename string) ([]byte, error) {
	path := "/api/v1/files/" + projectID + "/" + url.PathEscape(filename) + "/"

	// Use a longer timeout for file downloads
	downloadClient := &http.Client{Timeout: 10 * time.Minute}
	reqURL := c.baseURL + path
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("Accept", "*/*")
	if c.token != "" {
		req.Header.Set("Authorization", "Token "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return data, nil
}

// DeleteFile deletes a file from a project.
func (c *Client) DeleteFile(projectID, filename string) error {
	path := "/api/v1/files/" + projectID + "/" + url.PathEscape(filename) + "/"
	resp, err := c.doRequest("DELETE", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ============================================================================
// Jobs
// ============================================================================

// Job represents a QFieldCloud processing job.
type Job struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Output      string `json:"output,omitempty"`
	Feedback    string `json:"feedback,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	FinishedAt  string `json:"finished_at,omitempty"`
}

// JobCreate holds the fields for creating/triggering a job.
type JobCreate struct {
	ProjectID string `json:"project_id"`
	Type      string `json:"type"` // "process_projectfile", "apply_deltas", "package", etc.
}

// ListJobs returns all jobs for a project.
func (c *Client) ListJobs(projectID string) ([]Job, error) {
	path := "/api/v1/jobs/"
	if projectID != "" {
		path += "?project_id=" + url.QueryEscape(projectID)
	}
	var jobs []Job
	if err := c.doJSON("GET", path, nil, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// GetJob returns a single job by ID.
func (c *Client) GetJob(jobID string) (*Job, error) {
	var job Job
	if err := c.doJSON("GET", "/api/v1/jobs/"+jobID+"/", nil, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// CreateJob triggers a new job (e.g. repackage a project).
func (c *Client) CreateJob(req JobCreate) (*Job, error) {
	var job Job
	if err := c.doJSON("POST", "/api/v1/jobs/", req, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// ============================================================================
// Collaborators
// ============================================================================

// Collaborator represents a project collaborator.
type Collaborator struct {
	Username  string `json:"collaborator"`
	Role      string `json:"role"` // "admin", "manager", "editor", "reporter", "viewer"
	ProjectID string `json:"project_id,omitempty"`
}

// CollaboratorCreate holds the fields for adding a collaborator.
type CollaboratorCreate struct {
	Collaborator string `json:"collaborator"`
	Role         string `json:"role"`
}

// CollaboratorUpdate holds the fields for updating a collaborator's role.
type CollaboratorUpdate struct {
	Role string `json:"role"`
}

// ListCollaborators returns all collaborators for a project.
func (c *Client) ListCollaborators(projectID string) ([]Collaborator, error) {
	var collaborators []Collaborator
	if err := c.doJSON("GET", "/api/v1/collaborators/"+projectID+"/", nil, &collaborators); err != nil {
		return nil, err
	}
	return collaborators, nil
}

// AddCollaborator adds a collaborator to a project.
func (c *Client) AddCollaborator(projectID string, req CollaboratorCreate) (*Collaborator, error) {
	var collaborator Collaborator
	if err := c.doJSON("POST", "/api/v1/collaborators/"+projectID+"/", req, &collaborator); err != nil {
		return nil, err
	}
	return &collaborator, nil
}

// UpdateCollaborator updates a collaborator's role.
func (c *Client) UpdateCollaborator(projectID, username string, req CollaboratorUpdate) (*Collaborator, error) {
	var collaborator Collaborator
	if err := c.doJSON("PATCH", "/api/v1/collaborators/"+projectID+"/"+username+"/", req, &collaborator); err != nil {
		return nil, err
	}
	return &collaborator, nil
}

// RemoveCollaborator removes a collaborator from a project.
func (c *Client) RemoveCollaborator(projectID, username string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/collaborators/"+projectID+"/"+username+"/", nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remove collaborator failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ============================================================================
// Deltas (offline data synchronisation)
// ============================================================================

// Delta represents a single change set pushed from a mobile device.
type Delta struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	ClientID    string `json:"client_id"`
	Status      string `json:"status"` // "pending", "applied", "conflict", "not_applied", "error"
	Output      string `json:"output,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListDeltas returns all deltas for a project.
func (c *Client) ListDeltas(projectID string) ([]Delta, error) {
	var deltas []Delta
	if err := c.doJSON("GET", "/api/v1/deltas/"+projectID+"/", nil, &deltas); err != nil {
		return nil, err
	}
	return deltas, nil
}

// GetDelta returns a single delta by ID.
func (c *Client) GetDelta(projectID, deltaID string) (*Delta, error) {
	var delta Delta
	if err := c.doJSON("GET", "/api/v1/deltas/"+projectID+"/"+deltaID+"/", nil, &delta); err != nil {
		return nil, err
	}
	return &delta, nil
}

// ============================================================================
// Packages
// ============================================================================

// Package represents a project data package (downloadable snapshot for mobile).
type Package struct {
	ProjectID    string `json:"project_id"`
	Status       string `json:"status"`
	PackagedAt   string `json:"packaged_at"`
	DataSize     int64  `json:"data_size"`
	PackageSize  int64  `json:"package_size"`
}

// GetLatestPackage returns the latest package for a project.
func (c *Client) GetLatestPackage(projectID string) (*Package, error) {
	var pkg Package
	if err := c.doJSON("GET", "/api/v1/packages/"+projectID+"/latest/", nil, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// ============================================================================
// Members (organisation members)
// ============================================================================

// Member represents an organisation member.
type Member struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// ListMembers returns all members of the organisation.
func (c *Client) ListMembers() ([]Member, error) {
	var members []Member
	if err := c.doJSON("GET", "/api/v1/members/", nil, &members); err != nil {
		return nil, err
	}
	return members, nil
}
