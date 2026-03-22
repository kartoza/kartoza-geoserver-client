package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/auth"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/deploy"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/health"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/repository"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/vault"
)

// InstanceHandler handles instance-related HTTP requests.
type InstanceHandler struct {
	instanceRepo *repository.InstanceRepository
	productRepo  *repository.ProductRepository
	deployService *deploy.Service
	healthChecker *health.Checker
	vaultClient   *vault.Client
}

// NewInstanceHandler creates a new instance handler.
func NewInstanceHandler(
	instanceRepo *repository.InstanceRepository,
	productRepo *repository.ProductRepository,
	deployService *deploy.Service,
	healthChecker *health.Checker,
	vaultClient *vault.Client,
) *InstanceHandler {
	return &InstanceHandler{
		instanceRepo:  instanceRepo,
		productRepo:   productRepo,
		deployService: deployService,
		healthChecker: healthChecker,
		vaultClient:   vaultClient,
	}
}

// HandleListInstances handles GET /api/v1/instances
func (h *InstanceHandler) HandleListInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get instance summaries for the user
	summaries, err := h.instanceRepo.GetInstanceSummaries(r.Context(), userID)
	if err != nil {
		jsonError(w, "Failed to list instances", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"instances": summaries,
		"total":     len(summaries),
	}, http.StatusOK)
}

// HandleGetInstance handles GET /api/v1/instances/{id}
func (h *InstanceHandler) HandleGetInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract instance ID from path
	instanceID := strings.TrimPrefix(r.URL.Path, "/api/v1/instances/")
	instanceID = strings.Split(instanceID, "/")[0]
	if instanceID == "" {
		jsonError(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	instance, err := h.instanceRepo.GetInstanceByID(r.Context(), instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Instance not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get instance", http.StatusInternalServerError)
		return
	}

	// Verify ownership (unless admin)
	if instance.UserID != userID && !auth.IsAdmin(r.Context()) {
		jsonError(w, "Instance not found", http.StatusNotFound)
		return
	}

	// Populate related data
	if product, err := h.productRepo.GetProductByID(r.Context(), instance.ProductID); err == nil {
		instance.Product = product
	}
	if pkg, err := h.productRepo.GetPackageByID(r.Context(), instance.PackageID); err == nil {
		instance.Package = pkg
	}
	if cluster, err := h.productRepo.GetClusterByID(r.Context(), instance.ClusterID); err == nil {
		instance.Cluster = cluster
	}

	jsonResponse(w, instance, http.StatusOK)
}

// HandleGetInstanceCredentials handles GET /api/v1/instances/{id}/credentials
func (h *InstanceHandler) HandleGetInstanceCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract instance ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/instances/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "credentials" {
		jsonError(w, "Invalid path", http.StatusBadRequest)
		return
	}
	instanceID := parts[0]

	instance, err := h.instanceRepo.GetInstanceByID(r.Context(), instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Instance not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get instance", http.StatusInternalServerError)
		return
	}

	// Verify ownership (unless admin)
	if instance.UserID != userID && !auth.IsAdmin(r.Context()) {
		jsonError(w, "Instance not found", http.StatusNotFound)
		return
	}

	// Instance must be operational
	if !instance.IsOperational() {
		jsonError(w, "Instance is not operational", http.StatusBadRequest)
		return
	}

	// Get credentials from Vault
	if instance.VaultPath == "" {
		jsonError(w, "Credentials not available", http.StatusNotFound)
		return
	}

	creds, err := h.vaultClient.ReadCredentials(r.Context(), instance.VaultPath)
	if err != nil {
		jsonError(w, "Failed to get credentials", http.StatusInternalServerError)
		return
	}

	// Build response
	resp := &models.InstanceCredentials{
		URL:           instance.URL,
		AdminUsername: creds.AdminUsername,
		AdminPassword: creds.AdminPassword,
		DatabaseHost:  creds.DatabaseHost,
		DatabasePort:  creds.DatabasePort,
		DatabaseName:  creds.DatabaseName,
		DatabaseUser:  creds.DatabaseUser,
		DatabasePass:  creds.DatabasePass,
		Extra:         creds.Extra,
	}

	jsonResponse(w, resp, http.StatusOK)
}

// HandleGetInstanceHealth handles GET /api/v1/instances/{id}/health
func (h *InstanceHandler) HandleGetInstanceHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract instance ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/instances/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "health" {
		jsonError(w, "Invalid path", http.StatusBadRequest)
		return
	}
	instanceID := parts[0]

	instance, err := h.instanceRepo.GetInstanceByID(r.Context(), instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Instance not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get instance", http.StatusInternalServerError)
		return
	}

	// Verify ownership (unless admin)
	if instance.UserID != userID && !auth.IsAdmin(r.Context()) {
		jsonError(w, "Instance not found", http.StatusNotFound)
		return
	}

	// Get health check result
	result := h.healthChecker.GetResult(instanceID)
	if result == nil {
		// Perform an immediate check
		result = h.healthChecker.CheckInstance(r.Context(), instance)
	}

	jsonResponse(w, map[string]interface{}{
		"instance_id":   result.InstanceID,
		"status":        result.Status,
		"response_time": result.ResponseTime.Milliseconds(),
		"status_code":   result.StatusCode,
		"error":         result.Error,
		"checked_at":    result.CheckedAt,
	}, http.StatusOK)
}

// HandleRestartInstance handles POST /api/v1/instances/{id}/restart
func (h *InstanceHandler) HandleRestartInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract instance ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/instances/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "restart" {
		jsonError(w, "Invalid path", http.StatusBadRequest)
		return
	}
	instanceID := parts[0]

	instance, err := h.instanceRepo.GetInstanceByID(r.Context(), instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Instance not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get instance", http.StatusInternalServerError)
		return
	}

	// Verify ownership (unless admin)
	if instance.UserID != userID && !auth.IsAdmin(r.Context()) {
		jsonError(w, "Instance not found", http.StatusNotFound)
		return
	}

	// Check if instance can be restarted
	if !instance.CanBeRestarted() {
		jsonError(w, "Instance cannot be restarted in its current state", http.StatusBadRequest)
		return
	}

	// Trigger restart
	if err := h.deployService.Restart(r.Context(), instanceID); err != nil {
		jsonError(w, "Failed to restart instance", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "Restart initiated"}, http.StatusAccepted)
}

// HandleDeleteInstance handles DELETE /api/v1/instances/{id}
func (h *InstanceHandler) HandleDeleteInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract instance ID from path
	instanceID := strings.TrimPrefix(r.URL.Path, "/api/v1/instances/")
	instanceID = strings.Split(instanceID, "/")[0]
	if instanceID == "" {
		jsonError(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	instance, err := h.instanceRepo.GetInstanceByID(r.Context(), instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Instance not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get instance", http.StatusInternalServerError)
		return
	}

	// Verify ownership (unless admin)
	if instance.UserID != userID && !auth.IsAdmin(r.Context()) {
		jsonError(w, "Instance not found", http.StatusNotFound)
		return
	}

	// Check if instance can be deleted
	if !instance.CanBeDeleted() {
		jsonError(w, "Instance cannot be deleted in its current state", http.StatusBadRequest)
		return
	}

	// Trigger deletion
	if err := h.deployService.Delete(r.Context(), instanceID); err != nil {
		jsonError(w, "Failed to delete instance", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "Deletion initiated"}, http.StatusAccepted)
}

// HandleListActivities handles GET /api/v1/instances/{id}/activities
func (h *InstanceHandler) HandleListActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract instance ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/instances/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "activities" {
		jsonError(w, "Invalid path", http.StatusBadRequest)
		return
	}
	instanceID := parts[0]

	instance, err := h.instanceRepo.GetInstanceByID(r.Context(), instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Instance not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get instance", http.StatusInternalServerError)
		return
	}

	// Verify ownership (unless admin)
	if instance.UserID != userID && !auth.IsAdmin(r.Context()) {
		jsonError(w, "Instance not found", http.StatusNotFound)
		return
	}

	// Get activities
	activities, err := h.instanceRepo.ListInstanceActivities(r.Context(), instanceID, 50)
	if err != nil {
		jsonError(w, "Failed to list activities", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"activities": activities,
		"total":      len(activities),
	}, http.StatusOK)
}

// HandleArgoCDWebhook handles POST /api/v1/webhooks/argocd
func (h *InstanceHandler) HandleArgoCDWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload deploy.ArgoCDWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.deployService.HandleArgoCDWebhook(r.Context(), payload); err != nil {
		// Log but don't fail - ArgoCD will retry
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleHealthSummary handles GET /api/v1/admin/instances/health (admin only)
func (h *InstanceHandler) HandleHealthSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !auth.IsAdmin(r.Context()) {
		jsonError(w, "Admin access required", http.StatusForbidden)
		return
	}

	summary := h.healthChecker.GetSummary()
	jsonResponse(w, summary, http.StatusOK)
}
