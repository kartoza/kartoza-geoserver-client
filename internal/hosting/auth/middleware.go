package auth

import (
	"context"
	"net/http"
	"strings"
)

// ContextKey is the type for context keys.
type ContextKey string

const (
	// ContextUserIDKey is the context key for user ID.
	ContextUserIDKey ContextKey = "user_id"
	// ContextUserEmailKey is the context key for user email.
	ContextUserEmailKey ContextKey = "user_email"
	// ContextIsAdminKey is the context key for admin status.
	ContextIsAdminKey ContextKey = "is_admin"
	// ContextClaimsKey is the context key for full claims.
	ContextClaimsKey ContextKey = "claims"
)

// Middleware provides HTTP middleware for authentication.
type Middleware struct {
	service *Service
}

// NewMiddleware creates a new auth middleware.
func NewMiddleware(service *Service) *Middleware {
	return &Middleware{service: service}
}

// RequireAuth is middleware that requires a valid JWT token.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		claims, err := m.service.ValidateToken(token)
		if err != nil {
			if err == ErrExpiredToken {
				http.Error(w, `{"error": "Token has expired"}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextUserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ContextUserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, ContextIsAdminKey, claims.IsAdmin)
		ctx = context.WithValue(ctx, ContextClaimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin is middleware that requires an admin user.
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value(ContextIsAdminKey).(bool)
		if !ok || !isAdmin {
			http.Error(w, `{"error": "Admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

// OptionalAuth is middleware that validates a token if present but doesn't require it.
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.service.ValidateToken(token)
		if err != nil {
			// Token is invalid but not required, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Add claims to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextUserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ContextUserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, ContextIsAdminKey, claims.IsAdmin)
		ctx = context.WithValue(ctx, ContextClaimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuthFunc wraps a handler function with authentication requirement.
func (m *Middleware) RequireAuthFunc(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireAuth(next).ServeHTTP
}

// RequireAdminFunc wraps a handler function with admin requirement.
func (m *Middleware) RequireAdminFunc(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireAdmin(next).ServeHTTP
}

// extractToken extracts the JWT token from the Authorization header.
func extractToken(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check query parameter as fallback (for WebSocket connections, etc.)
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	return ""
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextUserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserEmail extracts the user email from the request context.
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(ContextUserEmailKey).(string); ok {
		return email
	}
	return ""
}

// IsAdmin checks if the current user is an admin.
func IsAdmin(ctx context.Context) bool {
	if isAdmin, ok := ctx.Value(ContextIsAdminKey).(bool); ok {
		return isAdmin
	}
	return false
}

// GetClaims extracts the full claims from the request context.
func GetClaims(ctx context.Context) *Claims {
	if claims, ok := ctx.Value(ContextClaimsKey).(*Claims); ok {
		return claims
	}
	return nil
}

// IsAuthenticated checks if the request context has valid authentication.
func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != ""
}
