// Package handlers provides HTTP handlers for the hosting platform API.
package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/auth"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/email"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService  *auth.Service
	middleware   *auth.Middleware
	emailService *email.Service
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		middleware:  auth.NewMiddleware(authService),
	}
}

// NewAuthHandlerWithEmail creates a new auth handler with email support.
func NewAuthHandlerWithEmail(authService *auth.Service, emailService *email.Service) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		middleware:   auth.NewMiddleware(authService),
		emailService: emailService,
	}
}

// Middleware returns the auth middleware.
func (h *AuthHandler) Middleware() *auth.Middleware {
	return h.middleware
}

// HandleRegister handles POST /api/v1/auth/register
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailExists):
			jsonError(w, "Email already registered", http.StatusConflict)
		case errors.Is(err, auth.ErrInvalidEmail):
			jsonError(w, "Invalid email address", http.StatusBadRequest)
		default:
			// Check if it's a validation error
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// Send welcome email asynchronously
	if h.emailService != nil {
		go func() {
			if err := h.emailService.SendWelcome(r.Context(), req.Email, req.FirstName); err != nil {
				log.Printf("Failed to send welcome email to %s: %v", req.Email, err)
			}
		}()
	}

	jsonResponse(w, resp, http.StatusCreated)
}

// HandleLogin handles POST /api/v1/auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		case errors.Is(err, auth.ErrUserDisabled):
			jsonError(w, "Account is disabled", http.StatusForbidden)
		default:
			jsonError(w, "Login failed", http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, resp, http.StatusOK)
}

// HandleRefresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		jsonError(w, "Refresh token required", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrExpiredToken):
			jsonError(w, "Refresh token has expired", http.StatusUnauthorized)
		case errors.Is(err, auth.ErrInvalidToken):
			jsonError(w, "Invalid refresh token", http.StatusUnauthorized)
		case errors.Is(err, auth.ErrUserDisabled):
			jsonError(w, "Account is disabled", http.StatusForbidden)
		default:
			jsonError(w, "Token refresh failed", http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, resp, http.StatusOK)
}

// HandleProfile handles GET/PUT /api/v1/auth/profile
func (h *AuthHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getProfile(w, r, userID)
	case http.MethodPut:
		h.updateProfile(w, r, userID)
	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AuthHandler) getProfile(w http.ResponseWriter, r *http.Request, userID string) {
	profile, err := h.authService.GetUserProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			jsonError(w, "User not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, profile, http.StatusOK)
}

func (h *AuthHandler) updateProfile(w http.ResponseWriter, r *http.Request, userID string) {
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authService.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			jsonError(w, "User not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, user, http.StatusOK)
}

// HandleChangePassword handles POST /api/v1/auth/change-password
func (h *AuthHandler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authService.ChangePassword(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			jsonError(w, "Current password is incorrect", http.StatusBadRequest)
		case errors.Is(err, auth.ErrUserNotFound):
			jsonError(w, "User not found", http.StatusNotFound)
		default:
			// Password validation error
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	jsonResponse(w, map[string]string{"message": "Password changed successfully"}, http.StatusOK)
}

// HandleRequestPasswordReset handles POST /api/v1/auth/reset-password
func (h *AuthHandler) HandleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		jsonError(w, "Email required", http.StatusBadRequest)
		return
	}

	// Request password reset token
	resetInfo, _ := h.authService.RequestPasswordReset(r.Context(), req.Email)

	// Send password reset email asynchronously if token was generated
	if resetInfo != nil && h.emailService != nil {
		go func() {
			if err := h.emailService.SendPasswordReset(r.Context(), resetInfo.Email, resetInfo.FirstName, resetInfo.Token); err != nil {
				log.Printf("Failed to send password reset email to %s: %v", resetInfo.Email, err)
			}
		}()
	}

	// Always return success to not reveal if email exists
	jsonResponse(w, map[string]string{
		"message": "If the email exists, a password reset link will be sent",
	}, http.StatusOK)
}

// HandleConfirmPasswordReset handles POST /api/v1/auth/reset-confirm
func (h *AuthHandler) HandleConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		jsonError(w, "Token and new password required", http.StatusBadRequest)
		return
	}

	err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			jsonError(w, "Invalid or expired reset token", http.StatusBadRequest)
		case errors.Is(err, auth.ErrTokenExpired):
			jsonError(w, "Reset token has expired", http.StatusBadRequest)
		default:
			// Password validation error
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	jsonResponse(w, map[string]string{"message": "Password reset successfully"}, http.StatusOK)
}

// HandleLogout handles POST /api/v1/auth/logout
// Note: With JWT, logout is typically handled client-side by removing the token.
// This endpoint can be used to invalidate refresh tokens if stored server-side.
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// With stateless JWT, we just acknowledge the logout
	// The client should discard the tokens
	jsonResponse(w, map[string]string{"message": "Logged out successfully"}, http.StatusOK)
}

// Helper functions for JSON responses

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
