package preview

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

// LayerInfo contains information about a layer to preview
type LayerInfo struct {
	Name        string `json:"name"`
	Workspace   string `json:"workspace"`
	StoreName   string `json:"store_name"`
	StoreType   string `json:"store_type"` // "datastore" or "coveragestore"
	GeoServerURL string `json:"geoserver_url"`
	Type        string `json:"type"` // "vector" or "raster"
}

// Server provides an embedded HTTP server for layer previews
type Server struct {
	server    *http.Server
	listener  net.Listener
	layer     *LayerInfo
	mu        sync.RWMutex
	running   bool
	port      int
}

// NewServer creates a new preview server
func NewServer() *Server {
	return &Server{}
}

// Start starts the preview server and returns the URL
func (s *Server) Start(layer *LayerInfo) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If already running, update layer and return existing URL
	if s.running {
		s.layer = layer
		return fmt.Sprintf("http://localhost:%d", s.port), nil
	}

	s.layer = layer

	// Find an available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("failed to find available port: %w", err)
	}
	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	// Create file server for static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		listener.Close()
		return "", fmt.Errorf("failed to access static files: %w", err)
	}

	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// API endpoint to get layer info
	mux.HandleFunc("/api/layer", s.handleLayerInfo)

	s.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			// Log error but don't crash
			fmt.Printf("Preview server error: %v\n", err)
		}
	}()

	s.running = true
	url := fmt.Sprintf("http://localhost:%d", s.port)
	return url, nil
}

// Stop stops the preview server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.running = false
	return nil
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetPort returns the port the server is running on
func (s *Server) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

// handleLayerInfo returns the current layer info as JSON
func (s *Server) handleLayerInfo(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	layer := s.layer
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if layer == nil {
		http.Error(w, `{"error": "no layer configured"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(layer)
}

// OpenBrowser opens the default browser with the given URL
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
