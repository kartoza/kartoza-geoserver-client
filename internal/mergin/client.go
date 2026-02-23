// Package mergin provides a client for the Mergin Maps REST API.
// The implementation is modelled after the official Mergin Maps Python API client
// (https://github.com/MerginMaps/python-api-client).
package mergin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/config"
)

const (
	// DefaultURL is the URL of the Mergin Maps public SaaS instance.
	DefaultURL = "https://app.merginmaps.com"
)

// Client is a Mergin Maps API client.
type Client struct {
	baseURL    string
	username   string
	password   string
	token      string // Bearer token (without "Bearer " prefix)
	httpClient *http.Client
}

// NewClient creates a new Mergin Maps API client from a connection config.
func NewClient(conn *config.MerginMapsConnection) *Client {
	baseURL := conn.URL
	if baseURL == "" {
		baseURL = DefaultURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		baseURL:  baseURL,
		username: conn.Username,
		password: conn.Password,
		token:    conn.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an authenticated HTTP request.
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.baseURL + path

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// Login authenticates with the Mergin Maps server and returns a session token.
// This mirrors the MerginClient.login() method from the Python client.
func (c *Client) Login(login, password string) (string, error) {
	params := map[string]string{"login": login, "password": password}
	body, _ := json.Marshal(params)

	req, err := http.NewRequest("POST", c.baseURL+"/v1/auth/login", strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Session struct {
			Token  string `json:"token"`
			Expire string `json:"expire"`
		} `json:"session"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	return result.Session.Token, nil
}

// TestConnection tests that the Mergin Maps server is reachable and credentials are valid.
func (c *Client) TestConnection() error {
	// If we have a token, verify it against /v1/user/profile
	// If we have username/password, try to authenticate first
	if c.token != "" {
		resp, err := c.doRequest("GET", "/v1/user/profile", nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("authentication failed (status %d): %s", resp.StatusCode, string(body))
		}
		return nil
	}

	if c.username != "" && c.password != "" {
		_, err := c.Login(c.username, c.password)
		return err
	}

	// No credentials – just check the server is alive
	resp, err := c.httpClient.Get(c.baseURL + "/v1/server/config")
	if err != nil {
		return fmt.Errorf("server not reachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error (status %d)", resp.StatusCode)
	}
	return nil
}

// UserInfo holds basic information about the authenticated user.
type UserInfo struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
}

// GetUserInfo returns profile information for the authenticated user.
// Mirrors MerginClient.user_info() from the Python client.
func (c *Client) GetUserInfo() (*UserInfo, error) {
	resp, err := c.doRequest("GET", "/v1/user/profile", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}
	return &info, nil
}

// Workspace holds information about a Mergin Maps workspace (namespace / organisation).
type Workspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// WorkspacesListResponse is the raw JSON response from /v1/workspaces.
type WorkspacesListResponse []Workspace

// GetWorkspaces lists all workspaces available to the authenticated user.
// Mirrors MerginClient.workspaces_list() from the Python client.
func (c *Client) GetWorkspaces() ([]Workspace, error) {
	resp, err := c.doRequest("GET", "/v1/workspaces", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result []Workspace
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode workspaces: %w", err)
	}
	return result, nil
}

// ProjectAccess holds the access control lists for a project.
type ProjectAccess struct {
	OwnersNames  []string `json:"ownersnames"`
	WritersNames []string `json:"writersnames"`
	ReadersNames []string `json:"readersnames"`
}

// Project holds metadata for a single Mergin Maps project.
// Mirrors the project dict returned by the Python client.
type Project struct {
	ID          string        `json:"id"`
	Namespace   string        `json:"namespace"`
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated,omitempty"`
	DiskUsage   int64         `json:"disk_usage"`
	Tags        []string      `json:"tags,omitempty"`
	Public      bool          `json:"public"`
	Access      ProjectAccess `json:"access,omitempty"`
}

// ProjectsResponse is the paginated response from /v1/project/paginated.
type ProjectsResponse struct {
	Projects []Project `json:"projects"`
	Count    int       `json:"count"`
}

// GetProjects lists projects accessible to the authenticated user.
// Mirrors MerginClient.paginated_projects_list() from the Python client.
func (c *Client) GetProjects(page, perPage int, namespace string) (*ProjectsResponse, error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("per_page", fmt.Sprintf("%d", perPage))
	if namespace != "" {
		params.Set("only_namespace", namespace)
	}

	path := "/v1/project/paginated?" + params.Encode()
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ProjectsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode projects: %w", err)
	}
	return &result, nil
}

// GetProject returns detailed information about a single project.
// Mirrors MerginClient.project_info() from the Python client.
func (c *Client) GetProject(namespace, name string) (*Project, error) {
	path := fmt.Sprintf("/v1/project/%s/%s", url.PathEscape(namespace), url.PathEscape(name))
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode project: %w", err)
	}
	return &project, nil
}

// ProjectFile represents a file tracked inside a Mergin Maps project.
type ProjectFile struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	MD5Sum  string `json:"md5sum,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// GetProjectFiles returns the list of files in the latest version of a project.
func (c *Client) GetProjectFiles(namespace, name string) ([]ProjectFile, error) {
	// The files are embedded in the project detail response under a "files" key.
	path := fmt.Sprintf("/v1/project/%s/%s", url.PathEscape(namespace), url.PathEscape(name))
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode project response: %w", err)
	}

	filesRaw, ok := raw["files"]
	if !ok {
		// The server did not return a files field – return an empty list rather than failing.
		return []ProjectFile{}, nil
	}

	var files []ProjectFile
	if err := json.Unmarshal(filesRaw, &files); err != nil {
		return nil, fmt.Errorf("failed to decode project files: %w", err)
	}
	return files, nil
}
