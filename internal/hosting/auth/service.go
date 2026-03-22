package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/repository"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrEmailExists        = errors.New("email already registered")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrTokenExpired       = errors.New("token has expired")
)

// Service provides authentication functionality.
type Service struct {
	userRepo   *repository.UserRepository
	jwtManager *JWTManager
}

// NewService creates a new auth service.
func NewService(userRepo *repository.UserRepository, jwtConfig JWTConfig) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtManager: NewJWTManager(jwtConfig),
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response.
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Validate email
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Check if email exists
	exists, err := s.userRepo.EmailExists(ctx, strings.ToLower(req.Email))
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Validate password
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	// Hash password
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        strings.ToLower(req.Email),
		PasswordHash: hash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
		IsAdmin:      false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// Login authenticates a user and returns tokens.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserDisabled
	}

	// Verify password
	if !VerifyPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// RefreshTokens generates new tokens using a refresh token.
func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserDisabled
	}

	// Generate new tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// ValidateToken validates an access token and returns the claims.
func (s *Service) ValidateToken(token string) (*Claims, error) {
	return s.jwtManager.ValidateAccessToken(token)
}

// GetUserByID retrieves a user by ID.
func (s *Service) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateProfile updates a user's profile.
func (s *Service) UpdateProfile(ctx context.Context, userID string, req models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// ChangePassword changes a user's password.
func (s *Service) ChangePassword(ctx context.Context, userID string, req models.UpdatePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if !VerifyPassword(req.CurrentPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	// Hash new password
	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// PasswordResetInfo contains information about a password reset request.
type PasswordResetInfo struct {
	Token     string
	FirstName string
	Email     string
}

// RequestPasswordReset creates a password reset token.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) (*PasswordResetInfo, error) {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Don't reveal if email exists
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate reset token
	token, err := GenerateRandomToken(32)
	if err != nil {
		return nil, err
	}

	// Store token
	resetToken := &models.UserToken{
		UserID:    user.ID,
		Token:     token,
		TokenType: "reset",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.userRepo.CreateToken(ctx, resetToken); err != nil {
		return nil, fmt.Errorf("failed to create reset token: %w", err)
	}

	return &PasswordResetInfo{
		Token:     token,
		FirstName: user.FirstName,
		Email:     user.Email,
	}, nil
}

// ResetPassword resets a user's password using a reset token.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Get token
	resetToken, err := s.userRepo.GetToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidToken
		}
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Check if expired
	if resetToken.IsExpired() {
		// Clean up expired token
		_ = s.userRepo.DeleteToken(ctx, token)
		return ErrTokenExpired
	}

	// Check token type
	if resetToken.TokenType != "reset" {
		return ErrInvalidToken
	}

	// Validate new password
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete used token
	_ = s.userRepo.DeleteToken(ctx, token)

	// Delete all user tokens (invalidate existing sessions)
	_ = s.userRepo.DeleteUserTokens(ctx, resetToken.UserID)

	return nil
}

// GetUserProfile retrieves a user's profile with billing info and instance counts.
func (s *Service) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	return s.userRepo.GetUserProfile(ctx, userID)
}

// validateEmail validates an email address format.
func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	// Simple email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}
