package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

type uploadSession struct {
	SessionID   string
	ConnID      string
	Workspace   string
	Filename    string
	TotalSize   int64
	ChunkSize   int64
	TotalChunks int
	Received    map[int]bool
	TempDir     string
	CreatedAt   time.Time

	// GeoServer upload progress
	gsTotal atomic.Int64
	gsSent  atomic.Int64
	gsDone  atomic.Bool
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]*uploadSession
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		sessions: make(map[string]*uploadSession),
	}
}

func (ss *sessionStore) get(id string) *uploadSession {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.sessions[id]
}

func (ss *sessionStore) set(sess *uploadSession) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[sess.SessionID] = sess
}

func (ss *sessionStore) delete(id string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, id)
}

const defaultChunkSize int64 = 5 * 1024 * 1024

type progressReader struct {
	r    io.Reader
	sess *uploadSession
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.r.Read(p)
	if n > 0 {
		pr.sess.gsSent.Add(int64(n))
	}
	return
}

type chunkUploadInitRequest struct {
	ConnID    string `json:"connId"`
	Workspace string `json:"workspace"`
	Filename  string `json:"filename"`
	TotalSize int64  `json:"totalSize"`
	ChunkSize int64  `json:"chunkSize,omitempty"`
}

type chunkUploadInitResponse struct {
	SessionID   string `json:"sessionId"`
	ChunkSize   int64  `json:"chunkSize"`
	TotalChunks int    `json:"totalChunks"`
}

type chunkUploadChunkResponse struct {
	Received    int `json:"received"`
	TotalChunks int `json:"totalChunks"`
}

type chunkUploadCompleteRequest struct {
	SessionID string `json:"sessionId"`
}

func (s *Server) handleChunkUploadInit(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}
	if r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chunkUploadInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ConnID == "" || req.Workspace == "" || req.Filename == "" || req.TotalSize <= 0 {
		s.jsonError(w, "connId, workspace, filename, and totalSize are required", http.StatusBadRequest)
		return
	}

	if s.getClient(req.ConnID) == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}

	totalChunks := int((req.TotalSize + chunkSize - 1) / chunkSize)

	tempDir, err := os.MkdirTemp("", "chunked-upload-*")
	if err != nil {
		s.jsonError(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}

	sessionID := fmt.Sprintf("sess_%d", time.Now().UnixNano())
	sess := &uploadSession{
		SessionID:   sessionID,
		ConnID:      req.ConnID,
		Workspace:   req.Workspace,
		Filename:    req.Filename,
		TotalSize:   req.TotalSize,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		Received:    make(map[int]bool),
		TempDir:     tempDir,
		CreatedAt:   time.Now(),
	}
	s.sessions.set(sess)

	s.jsonResponse(w, chunkUploadInitResponse{
		SessionID:   sessionID,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
	})
}

// POST /api/upload/chunk
func (s *Server) handleChunkUploadChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}
	if r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		s.jsonError(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	sessionID := r.FormValue("sessionId")
	if sessionID == "" {
		s.jsonError(w, "sessionId is required", http.StatusBadRequest)
		return
	}

	chunkIndexStr := r.FormValue("chunkIndex")
	if chunkIndexStr == "" {
		s.jsonError(w, "chunkIndex is required", http.StatusBadRequest)
		return
	}
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		s.jsonError(w, "Invalid chunkIndex", http.StatusBadRequest)
		return
	}

	sess := s.sessions.get(sessionID)
	if sess == nil {
		s.jsonError(w, "Session not found", http.StatusNotFound)
		return
	}

	if chunkIndex >= sess.TotalChunks {
		s.jsonError(w, "chunkIndex out of range", http.StatusBadRequest)
		return
	}

	chunkFile, _, err := r.FormFile("chunk")
	if err != nil {
		s.jsonError(w, "No chunk data provided", http.StatusBadRequest)
		return
	}
	defer chunkFile.Close()

	chunkPath := filepath.Join(sess.TempDir, fmt.Sprintf("chunk-%05d", chunkIndex))
	out, err := os.Create(chunkPath)
	if err != nil {
		s.jsonError(w, "Failed to save chunk", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, chunkFile); err != nil {
		s.jsonError(w, "Failed to write chunk", http.StatusInternalServerError)
		return
	}

	s.sessions.mu.Lock()
	sess.Received[chunkIndex] = true
	received := len(sess.Received)
	s.sessions.mu.Unlock()

	s.jsonResponse(w, chunkUploadChunkResponse{
		Received:    received,
		TotalChunks: sess.TotalChunks,
	})
}

// POST /api/upload/complete
func (s *Server) handleChunkUploadComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}
	if r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chunkUploadCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sess := s.sessions.get(req.SessionID)
	if sess == nil {
		s.jsonError(w, "Session not found", http.StatusNotFound)
		return
	}

	client := s.getClient(sess.ConnID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	// Collect and sort chunk files.
	entries, err := os.ReadDir(sess.TempDir)
	if err != nil {
		s.jsonError(w, "Failed to read temp directory", http.StatusInternalServerError)
		return
	}

	var chunkFiles []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "chunk-") {
			chunkFiles = append(chunkFiles, filepath.Join(sess.TempDir, entry.Name()))
		}
	}
	sort.Strings(chunkFiles)

	if len(chunkFiles) != sess.TotalChunks {
		s.jsonError(w, fmt.Sprintf("Expected %d chunks, got %d", sess.TotalChunks, len(chunkFiles)), http.StatusBadRequest)
		return
	}

	// Assemble chunks into the final file.
	assembledPath := filepath.Join(sess.TempDir, sess.Filename)
	assembled, err := os.Create(assembledPath)
	if err != nil {
		s.jsonError(w, "Failed to create assembled file", http.StatusInternalServerError)
		return
	}

	for _, chunkPath := range chunkFiles {
		cf, err := os.Open(chunkPath)
		if err != nil {
			assembled.Close()
			s.jsonError(w, "Failed to read chunk file", http.StatusInternalServerError)
			return
		}
		_, copyErr := io.Copy(assembled, cf)
		cf.Close()
		if copyErr != nil {
			assembled.Close()
			s.jsonError(w, "Failed to assemble chunks", http.StatusInternalServerError)
			return
		}
	}
	assembled.Close()

	defer func() {
		os.RemoveAll(sess.TempDir)
		s.sessions.delete(sess.SessionID)
	}()

	fileType := detectFileType(sess.Filename)
	if !fileType.CanUpload() {
		s.jsonError(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	storeName := sanitizeStoreName(strings.TrimSuffix(sess.Filename, filepath.Ext(sess.Filename)))

	// Open assembled file and prepare progress-tracked reader for GeoServer upload.
	assembledFile, err := os.Open(assembledPath)
	if err != nil {
		s.jsonError(w, "Failed to open assembled file", http.StatusInternalServerError)
		return
	}
	defer assembledFile.Close()

	sess.gsTotal.Store(sess.TotalSize)
	sess.gsSent.Store(0)
	pr := &progressReader{r: assembledFile, sess: sess}

	var uploadErr error
	var storeType string

	switch fileType {
	case models.FileTypeShapefile:
		uploadErr = client.UploadShapefileFrom(sess.Workspace, storeName, pr, sess.TotalSize)
		storeType = "datastore"
	case models.FileTypeGeoTIFF:
		uploadErr = client.UploadGeoTIFFFrom(sess.Workspace, storeName, pr, sess.TotalSize)
		storeType = "coveragestore"
	case models.FileTypeGeoPackage:
		uploadErr = client.UploadGeoPackageFrom(sess.Workspace, storeName, pr, sess.TotalSize)
		storeType = "datastore"
	case models.FileTypeSLD, models.FileTypeCSS:
		format := "sld"
		if fileType == models.FileTypeCSS {
			format = "css"
		}
		// Styles are small; use file path variant (no progress tracking needed).
		assembledFile.Close()
		uploadErr = client.UploadStyle(sess.Workspace, storeName, assembledPath, format)
		storeType = "style"
	default:
		s.jsonError(w, "Unsupported file type for upload", http.StatusBadRequest)
		return
	}
	sess.gsDone.Store(true)

	if uploadErr != nil {
		s.jsonResponse(w, UploadResponse{
			Success: false,
			Message: uploadErr.Error(),
		})
		return
	}

	s.jsonResponse(w, UploadResponse{
		Success:   true,
		Message:   fmt.Sprintf("Successfully uploaded %s", sess.Filename),
		StoreName: storeName,
		StoreType: storeType,
	})
}

// POST /api/upload/session/{id} and /api/upload/session/{id}/progress
func (s *Server) handleUploadSession(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/progress") {
		s.handleChunkUploadProgress(w, r)
		return
	}
	s.handleChunkUploadCancel(w, r)
}

// DELETE /api/upload/session/{id}
func (s *Server) handleChunkUploadCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}
	if r.Method != http.MethodDelete {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := strings.TrimPrefix(r.URL.Path, "/api/upload/session/")
	sessionID = strings.TrimSuffix(sessionID, "/")

	if sessionID == "" {
		s.jsonError(w, "Session ID required", http.StatusBadRequest)
		return
	}

	sess := s.sessions.get(sessionID)
	if sess != nil {
		os.RemoveAll(sess.TempDir)
		s.sessions.delete(sessionID)
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/upload/session/{id}/progress
func (s *Server) handleChunkUploadProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.handleCORS(w)
		return
	}
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := strings.TrimPrefix(r.URL.Path, "/api/upload/session/")
	sessionID = strings.TrimSuffix(sessionID, "/progress")

	if sessionID == "" {
		s.jsonError(w, "Session ID required", http.StatusBadRequest)
		return
	}

	sess := s.sessions.get(sessionID)
	if sess == nil {
		s.jsonResponse(w, map[string]interface{}{
			"sent":  int64(0),
			"total": int64(0),
			"done":  true,
		})
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"sent":  sess.gsSent.Load(),
		"total": sess.gsTotal.Load(),
		"done":  sess.gsDone.Load(),
	})
}
