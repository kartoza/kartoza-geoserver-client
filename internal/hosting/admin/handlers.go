package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/auth"
)

// Handler handles admin HTTP requests.
type Handler struct {
	service *Service
}

// NewHandler creates a new admin handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// AdminOnly is middleware that requires admin access.
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAdmin(r.Context()) {
			jsonError(w, "Admin access required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// HandleDashboardStats handles GET /api/v1/admin/dashboard
func (h *Handler) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		jsonError(w, "Failed to get dashboard stats", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, stats, http.StatusOK)
}

// HandleListUsers handles GET /api/v1/admin/users
func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	opts := UserListOptions{
		Page:     getIntParam(r, "page", 1),
		PageSize: getIntParam(r, "page_size", 20),
		Search:   r.URL.Query().Get("search"),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
	}

	if activeStr := r.URL.Query().Get("is_active"); activeStr != "" {
		active := activeStr == "true"
		opts.IsActive = &active
	}
	if adminStr := r.URL.Query().Get("is_admin"); adminStr != "" {
		admin := adminStr == "true"
		opts.IsAdmin = &admin
	}

	result, err := h.service.ListUsers(r.Context(), opts)
	if err != nil {
		jsonError(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, result, http.StatusOK)
}

// HandleGetUser handles GET /api/v1/admin/users/{id}
func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := extractPathParam(r.URL.Path, "/api/v1/admin/users/")
	if userID == "" {
		jsonError(w, "User ID required", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserDetails(r.Context(), userID)
	if err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, user, http.StatusOK)
}

// HandleUpdateUser handles PUT /api/v1/admin/users/{id}
func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := extractPathParam(r.URL.Path, "/api/v1/admin/users/")
	if userID == "" {
		jsonError(w, "User ID required", http.StatusBadRequest)
		return
	}

	var req UpdateUserAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.UpdateUser(r.Context(), userID, req)
	if err != nil {
		jsonError(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// Log admin action
	adminID := auth.GetUserID(r.Context())
	h.service.LogAction(r.Context(), adminID, "update_user", "user", userID, "", r.RemoteAddr)

	jsonResponse(w, user, http.StatusOK)
}

// HandleListInstances handles GET /api/v1/admin/instances
func (h *Handler) HandleListInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	opts := InstanceListOptions{
		Page:     getIntParam(r, "page", 1),
		PageSize: getIntParam(r, "page_size", 20),
		Search:   r.URL.Query().Get("search"),
		Status:   r.URL.Query().Get("status"),
		UserID:   r.URL.Query().Get("user_id"),
		Product:  r.URL.Query().Get("product"),
	}

	result, err := h.service.ListInstances(r.Context(), opts)
	if err != nil {
		jsonError(w, "Failed to list instances", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, result, http.StatusOK)
}

// HandleListOrders handles GET /api/v1/admin/orders
func (h *Handler) HandleListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	opts := OrderListOptions{
		Page:     getIntParam(r, "page", 1),
		PageSize: getIntParam(r, "page_size", 20),
		Status:   r.URL.Query().Get("status"),
		UserID:   r.URL.Query().Get("user_id"),
	}

	result, err := h.service.ListOrders(r.Context(), opts)
	if err != nil {
		jsonError(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, result, http.StatusOK)
}

// HandleRevenueChart handles GET /api/v1/admin/analytics/revenue
func (h *Handler) HandleRevenueChart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "day"
	}

	data, err := h.service.GetRevenueChart(r.Context(), period, groupBy)
	if err != nil {
		jsonError(w, "Failed to get revenue data", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"data":     data,
		"period":   period,
		"group_by": groupBy,
	}, http.StatusOK)
}

// HandleSystemHealth handles GET /api/v1/admin/health
func (h *Handler) HandleSystemHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health, err := h.service.GetSystemHealth(r.Context())
	if err != nil {
		jsonError(w, "Failed to get system health", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, health, http.StatusOK)
}

// Helper functions

func jsonResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func getIntParam(r *http.Request, name string, defaultVal int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func extractPathParam(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}
	remaining := path[len(prefix):]
	// Remove trailing slashes or additional path segments
	for i, c := range remaining {
		if c == '/' {
			return remaining[:i]
		}
	}
	return remaining
}
