// Package models provides data models for the hosting platform.
package models

import (
	"time"
)

// User represents a registered user in the hosting platform.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	FirstName    string    `json:"first_name,omitempty"`
	LastName     string    `json:"last_name,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	IsActive     bool      `json:"is_active"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	if u.FirstName == "" {
		return u.LastName
	}
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

// UserToken represents an authentication or reset token.
type UserToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"` // "auth", "refresh", "reset"
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the token has expired.
func (t *UserToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// UserBillingInfo represents a user's billing information.
type UserBillingInfo struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	CompanyName        string    `json:"company_name,omitempty"`
	TaxNumber          string    `json:"tax_number,omitempty"`
	AddressLine1       string    `json:"address_line1,omitempty"`
	AddressLine2       string    `json:"address_line2,omitempty"`
	City               string    `json:"city,omitempty"`
	State              string    `json:"state,omitempty"`
	Country            string    `json:"country,omitempty"`
	PostalCode         string    `json:"postal_code,omitempty"`
	Phone              string    `json:"phone,omitempty"`
	StripeCustomerID   string    `json:"stripe_customer_id,omitempty"`
	PaystackCustomerID string    `json:"paystack_customer_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CreateUserRequest represents a request to create a new user.
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// UpdateUserRequest represents a request to update a user.
type UpdateUserRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// UpdatePasswordRequest represents a request to change password.
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// UserProfile represents a user's public profile with additional information.
type UserProfile struct {
	User
	BillingInfo    *UserBillingInfo `json:"billing_info,omitempty"`
	InstanceCount  int              `json:"instance_count"`
	ActiveInstance int              `json:"active_instances"`
}
