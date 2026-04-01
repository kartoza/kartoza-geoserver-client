// Package repository provides data access layer for the hosting platform.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/db"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
)

// Common errors
var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
)

// UserRepository provides data access for users.
type UserRepository struct {
	db *db.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(database *db.DB) *UserRepository {
	return &UserRepository{db: database}
}

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, avatar_url, is_active, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.AvatarURL, user.IsActive, user.IsAdmin,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url,
		       is_active, is_admin, created_at, updated_at
		FROM users WHERE id = $1
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.AvatarURL, &user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url,
		       is_active, is_admin, created_at, updated_at
		FROM users WHERE email = $1
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.AvatarURL, &user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates a user.
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			email = $2, first_name = $3, last_name = $4, avatar_url = $5,
			is_active = $6, is_admin = $7
		WHERE id = $1
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Email, user.FirstName, user.LastName,
		user.AvatarURL, user.IsActive, user.IsAdmin,
	).Scan(&user.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}

// UpdatePassword updates a user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, passwordHash)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete soft-deletes a user by setting is_active to false.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET is_active = false WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves users with pagination.
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE is_active = true`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get users
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url,
		       is_active, is_admin, created_at, updated_at
		FROM users WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
			&user.AvatarURL, &user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	return users, total, rows.Err()
}

// EmailExists checks if an email is already registered.
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

// CreateToken creates a user token.
func (r *UserRepository) CreateToken(ctx context.Context, token *models.UserToken) error {
	query := `
		INSERT INTO user_tokens (user_id, token, token_type, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		token.UserID, token.Token, token.TokenType, token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

// GetToken retrieves a token by its value.
func (r *UserRepository) GetToken(ctx context.Context, token string) (*models.UserToken, error) {
	query := `
		SELECT id, user_id, token, token_type, expires_at, created_at
		FROM user_tokens WHERE token = $1
	`
	t := &models.UserToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&t.ID, &t.UserID, &t.Token, &t.TokenType, &t.ExpiresAt, &t.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

// DeleteToken deletes a token.
func (r *UserRepository) DeleteToken(ctx context.Context, token string) error {
	query := `DELETE FROM user_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

// DeleteExpiredTokens removes all expired tokens.
func (r *UserRepository) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	query := `DELETE FROM user_tokens WHERE expires_at < $1`
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteUserTokens deletes all tokens for a user.
func (r *UserRepository) DeleteUserTokens(ctx context.Context, userID string) error {
	query := `DELETE FROM user_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// GetBillingInfo retrieves billing info for a user.
func (r *UserRepository) GetBillingInfo(ctx context.Context, userID string) (*models.UserBillingInfo, error) {
	query := `
		SELECT id, user_id, company_name, tax_number, address_line1, address_line2,
		       city, state, country, postal_code, phone, stripe_customer_id,
		       paystack_customer_id, created_at, updated_at
		FROM user_billing_info WHERE user_id = $1
	`
	info := &models.UserBillingInfo{}
	var companyName, taxNumber, addr1, addr2, city, state, country, postal, phone, stripeID, paystackID sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&info.ID, &info.UserID, &companyName, &taxNumber, &addr1, &addr2,
		&city, &state, &country, &postal, &phone, &stripeID, &paystackID,
		&info.CreatedAt, &info.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	info.CompanyName = companyName.String
	info.TaxNumber = taxNumber.String
	info.AddressLine1 = addr1.String
	info.AddressLine2 = addr2.String
	info.City = city.String
	info.State = state.String
	info.Country = country.String
	info.PostalCode = postal.String
	info.Phone = phone.String
	info.StripeCustomerID = stripeID.String
	info.PaystackCustomerID = paystackID.String
	return info, nil
}

// UpsertBillingInfo creates or updates billing info for a user.
func (r *UserRepository) UpsertBillingInfo(ctx context.Context, info *models.UserBillingInfo) error {
	query := `
		INSERT INTO user_billing_info (
			user_id, company_name, tax_number, address_line1, address_line2,
			city, state, country, postal_code, phone, stripe_customer_id, paystack_customer_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (user_id) DO UPDATE SET
			company_name = EXCLUDED.company_name,
			tax_number = EXCLUDED.tax_number,
			address_line1 = EXCLUDED.address_line1,
			address_line2 = EXCLUDED.address_line2,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			country = EXCLUDED.country,
			postal_code = EXCLUDED.postal_code,
			phone = EXCLUDED.phone,
			stripe_customer_id = COALESCE(EXCLUDED.stripe_customer_id, user_billing_info.stripe_customer_id),
			paystack_customer_id = COALESCE(EXCLUDED.paystack_customer_id, user_billing_info.paystack_customer_id)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		info.UserID, nullString(info.CompanyName), nullString(info.TaxNumber),
		nullString(info.AddressLine1), nullString(info.AddressLine2),
		nullString(info.City), nullString(info.State), nullString(info.Country),
		nullString(info.PostalCode), nullString(info.Phone),
		nullString(info.StripeCustomerID), nullString(info.PaystackCustomerID),
	).Scan(&info.ID, &info.CreatedAt, &info.UpdatedAt)
}

// Helper to convert empty strings to NULL
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// AdminUserListOptions specifies options for listing users in admin.
type AdminUserListOptions struct {
	Page     int
	PageSize int
	Search   string
	SortBy   string
	SortDir  string
	IsActive *bool
	IsAdmin  *bool
}

// AdminUserListItem represents a user in admin listings.
type AdminUserListItem struct {
	ID            string
	Email         string
	FirstName     string
	LastName      string
	IsActive      bool
	IsAdmin       bool
	InstanceCount int
	TotalSpent    float64
	CreatedAt     time.Time
	LastLoginAt   *time.Time
}

// UserStats represents aggregate user statistics.
type UserStats struct {
	TotalUsers  int
	ActiveUsers int
}

// Ping checks database connectivity.
func (r *UserRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// GetUserStats returns aggregate user statistics.
func (r *UserRepository) GetUserStats(ctx context.Context) (*UserStats, error) {
	query := `
		SELECT
			COUNT(*) as total_users,
			COUNT(*) FILTER (WHERE is_active = true) as active_users
		FROM users
	`
	stats := &UserStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(&stats.TotalUsers, &stats.ActiveUsers)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// ListUsersAdmin retrieves users for admin listing with filtering and pagination.
func (r *UserRepository) ListUsersAdmin(ctx context.Context, opts AdminUserListOptions) ([]AdminUserListItem, int, error) {
	// Build WHERE clause
	conditions := []string{"1=1"}
	args := []interface{}{}
	argNum := 1

	if opts.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)",
			argNum, argNum, argNum,
		))
		args = append(args, "%"+opts.Search+"%")
		argNum++
	}

	if opts.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argNum))
		args = append(args, *opts.IsActive)
		argNum++
	}

	if opts.IsAdmin != nil {
		conditions = append(conditions, fmt.Sprintf("is_admin = $%d", argNum))
		args = append(args, *opts.IsAdmin)
		argNum++
	}

	whereClause := "WHERE " + conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "created_at"
	if opts.SortBy != "" {
		// Whitelist allowed sort columns
		switch opts.SortBy {
		case "email", "first_name", "last_name", "created_at", "is_active", "is_admin":
			orderBy = opts.SortBy
		}
	}
	orderDir := "DESC"
	if opts.SortDir == "asc" {
		orderDir = "ASC"
	}

	// Calculate offset
	offset := (opts.Page - 1) * opts.PageSize

	// Main query with subqueries for instance count and total spent
	query := fmt.Sprintf(`
		SELECT u.id, u.email, u.first_name, u.last_name, u.is_active, u.is_admin,
			   u.created_at,
			   COALESCE((SELECT COUNT(*) FROM instances WHERE user_id = u.id AND status != 'deleted'), 0) as instance_count,
			   COALESCE((SELECT SUM(total_amount) FROM sales_orders WHERE user_id = u.id AND status = 'paid'), 0) as total_spent,
			   (SELECT MAX(created_at) FROM user_tokens WHERE user_id = u.id) as last_login_at
		FROM users u
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, orderDir, argNum, argNum+1)

	args = append(args, opts.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []AdminUserListItem
	for rows.Next() {
		var u AdminUserListItem
		var lastLoginAt sql.NullTime
		if err := rows.Scan(
			&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.IsActive, &u.IsAdmin,
			&u.CreatedAt, &u.InstanceCount, &u.TotalSpent, &lastLoginAt,
		); err != nil {
			return nil, 0, err
		}
		if lastLoginAt.Valid {
			u.LastLoginAt = &lastLoginAt.Time
		}
		users = append(users, u)
	}

	return users, total, rows.Err()
}

// GetUserProfile retrieves a user with billing info and instance counts.
func (r *UserRepository) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile := &models.UserProfile{User: *user}

	// Get billing info (may not exist)
	billing, err := r.GetBillingInfo(ctx, userID)
	if err == nil {
		profile.BillingInfo = billing
	}

	// Get instance counts
	countQuery := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'online') as active
		FROM instances WHERE user_id = $1 AND status != 'deleted'
	`
	err = r.db.QueryRowContext(ctx, countQuery, userID).Scan(
		&profile.InstanceCount, &profile.ActiveInstance,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get instance counts: %w", err)
	}

	return profile, nil
}
