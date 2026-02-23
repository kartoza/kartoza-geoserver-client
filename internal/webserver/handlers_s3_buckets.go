package webserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/s3client"
)

// handleS3Buckets handles bucket operations
func (s *Server) handleS3Buckets(w http.ResponseWriter, r *http.Request, connID string, pathParts []string) {
	client := s.getS3Client(connID)
	if client == nil {
		s.jsonError(w, "S3 connection not found", http.StatusNotFound)
		return
	}

	// /api/s3/connections/{connId}/buckets
	if len(pathParts) == 0 || pathParts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			s.listS3Buckets(w, r, client)
		case http.MethodPost:
			s.createS3Bucket(w, r, client)
		case http.MethodOptions:
			s.handleCORS(w)
		default:
			s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	bucketName := pathParts[0]

	// /api/s3/connections/{connId}/buckets/{bucket}/objects
	if len(pathParts) >= 2 && pathParts[1] == "objects" {
		s.handleS3Objects(w, r, client, bucketName)
		return
	}

	// /api/s3/connections/{connId}/buckets/{bucket}/presign
	if len(pathParts) >= 2 && pathParts[1] == "presign" {
		s.handleS3Presign(w, r, client, bucketName)
		return
	}

	// /api/s3/connections/{connId}/buckets/{bucket}
	switch r.Method {
	case http.MethodDelete:
		s.deleteS3Bucket(w, r, client, bucketName)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listS3Buckets lists all buckets
func (s *Server) listS3Buckets(w http.ResponseWriter, r *http.Request, client *s3client.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]S3BucketResponse, len(buckets))
	for i, b := range buckets {
		response[i] = S3BucketResponse{
			Name:         b.Name,
			CreationDate: b.CreationDate.Format(time.RFC3339),
		}
	}

	s.jsonResponse(w, response)
}

// createS3Bucket creates a new bucket
func (s *Server) createS3Bucket(w http.ResponseWriter, r *http.Request, client *s3client.Client) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.jsonError(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.CreateBucket(ctx, req.Name); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	s.jsonResponse(w, S3BucketResponse{
		Name:         req.Name,
		CreationDate: time.Now().Format(time.RFC3339),
	})
}

// deleteS3Bucket deletes a bucket
func (s *Server) deleteS3Bucket(w http.ResponseWriter, r *http.Request, client *s3client.Client, bucketName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.DeleteBucket(ctx, bucketName); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleS3Presign generates a presigned URL
func (s *Server) handleS3Presign(w http.ResponseWriter, r *http.Request, client *s3client.Client, bucketName string) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		s.jsonError(w, "Object key is required", http.StatusBadRequest)
		return
	}

	// Default 1 hour expiry
	expires := 1 * time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url, err := client.GetPresignedURL(ctx, bucketName, key, expires)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]string{
		"url":     url,
		"expires": time.Now().Add(expires).Format(time.RFC3339),
	})
}
