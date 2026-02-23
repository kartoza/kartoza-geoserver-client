package qfieldcloud

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kartoza/kartoza-cloudbench/internal/config"
)

// newTestClient returns a Client wired to the provided test server URL.
func newTestClient(serverURL string) *Client {
	conn := &config.QFieldCloudConnection{
		URL:   serverURL,
		Token: "test-token",
	}
	return NewClient(conn)
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/auth/login/" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(LoginResponse{Token: "mytoken123"})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	token, err := c.Login("user", "pass")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if token != "mytoken123" {
		t.Errorf("expected token 'mytoken123', got '%s'", token)
	}
}

func TestGetCurrentUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/user/" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(User{Username: "alice", Email: "alice@example.com"})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	user, err := c.GetCurrentUser()
	if err != nil {
		t.Fatalf("GetCurrentUser failed: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("expected username 'alice', got '%s'", user.Username)
	}
}

func TestTestConnection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(User{Username: "alice"})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if err := c.TestConnection(); err != nil {
		t.Errorf("TestConnection should succeed: %v", err)
	}
}

func TestTestConnectionFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if err := c.TestConnection(); err == nil {
		t.Error("TestConnection should fail for a 401 response")
	}
}

// ─── Projects ─────────────────────────────────────────────────────────────────

func TestListProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		projects := []Project{
			{ID: "proj-1", Name: "alpha", Owner: "alice"},
			{ID: "proj-2", Name: "beta", Owner: "bob"},
		}
		json.NewEncoder(w).Encode(projects)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	projects, err := c.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Name != "alpha" {
		t.Errorf("expected first project to be 'alpha', got '%s'", projects[0].Name)
	}
}

func TestGetProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/projects/") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(Project{ID: "proj-1", Name: "alpha", Owner: "alice"})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	project, err := c.GetProject("proj-1")
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if project.ID != "proj-1" {
		t.Errorf("expected project ID 'proj-1', got '%s'", project.ID)
	}
}

func TestCreateProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req ProjectCreate
		json.NewDecoder(r.Body).Decode(&req)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Project{ID: "new-proj", Name: req.Name, Owner: "alice"})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	project, err := c.CreateProject(ProjectCreate{Name: "new-project"})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if project.Name != "new-project" {
		t.Errorf("expected name 'new-project', got '%s'", project.Name)
	}
}

func TestDeleteProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if err := c.DeleteProject("proj-1"); err != nil {
		t.Errorf("DeleteProject failed: %v", err)
	}
}

// ─── Files ────────────────────────────────────────────────────────────────────

func TestListFiles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files := []ProjectFile{
			{Name: "project.qgs", Size: 1024},
			{Name: "data/layer.gpkg", Size: 4096},
		}
		json.NewEncoder(w).Encode(files)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	files, err := c.ListFiles("proj-1")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestDeleteFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if err := c.DeleteFile("proj-1", "project.qgs"); err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}
}

// ─── Jobs ─────────────────────────────────────────────────────────────────────

func TestListJobs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs := []Job{
			{ID: "job-1", Type: "package", Status: "finished"},
		}
		json.NewEncoder(w).Encode(jobs)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	jobs, err := c.ListJobs("proj-1")
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Status != "finished" {
		t.Errorf("expected status 'finished', got '%s'", jobs[0].Status)
	}
}

func TestCreateJob(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Job{ID: "job-new", Type: "package", Status: "queued"})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	job, err := c.CreateJob(JobCreate{ProjectID: "proj-1", Type: "package"})
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}
	if job.Status != "queued" {
		t.Errorf("expected status 'queued', got '%s'", job.Status)
	}
}

// ─── Collaborators ────────────────────────────────────────────────────────────

func TestListCollaborators(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collabs := []Collaborator{
			{Username: "bob", Role: "editor"},
		}
		json.NewEncoder(w).Encode(collabs)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	collabs, err := c.ListCollaborators("proj-1")
	if err != nil {
		t.Fatalf("ListCollaborators failed: %v", err)
	}
	if len(collabs) != 1 || collabs[0].Username != "bob" {
		t.Errorf("unexpected collaborators: %+v", collabs)
	}
}

func TestRemoveCollaborator(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if err := c.RemoveCollaborator("proj-1", "bob"); err != nil {
		t.Errorf("RemoveCollaborator failed: %v", err)
	}
}

// ─── Deltas ───────────────────────────────────────────────────────────────────

func TestListDeltas(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deltas := []Delta{
			{ID: "delta-1", Status: "applied"},
			{ID: "delta-2", Status: "pending"},
		}
		json.NewEncoder(w).Encode(deltas)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	deltas, err := c.ListDeltas("proj-1")
	if err != nil {
		t.Fatalf("ListDeltas failed: %v", err)
	}
	if len(deltas) != 2 {
		t.Errorf("expected 2 deltas, got %d", len(deltas))
	}
}

// ─── Client defaults ──────────────────────────────────────────────────────────

func TestNewClientDefaultURL(t *testing.T) {
	conn := &config.QFieldCloudConnection{
		Token: "tok",
	}
	c := NewClient(conn)
	if c.baseURL != defaultBaseURL {
		t.Errorf("expected default base URL '%s', got '%s'", defaultBaseURL, c.baseURL)
	}
}

func TestNewClientCustomURL(t *testing.T) {
	conn := &config.QFieldCloudConnection{
		URL:   "https://custom.qfield.io/",
		Token: "tok",
	}
	c := NewClient(conn)
	// Trailing slash should be stripped
	if c.baseURL != "https://custom.qfield.io" {
		t.Errorf("expected URL without trailing slash, got '%s'", c.baseURL)
	}
}
