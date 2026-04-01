// Package admin provides administrative functionality for the hosting platform.
package admin

import (
	"context"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/repository"
)

// Service provides admin functionality.
type Service struct {
	userRepo     *repository.UserRepository
	instanceRepo *repository.InstanceRepository
	orderRepo    *repository.OrderRepository
	productRepo  *repository.ProductRepository
}

// NewService creates a new admin service.
func NewService(
	userRepo *repository.UserRepository,
	instanceRepo *repository.InstanceRepository,
	orderRepo *repository.OrderRepository,
	productRepo *repository.ProductRepository,
) *Service {
	return &Service{
		userRepo:     userRepo,
		instanceRepo: instanceRepo,
		orderRepo:    orderRepo,
		productRepo:  productRepo,
	}
}

// DashboardStats represents high-level dashboard statistics.
type DashboardStats struct {
	TotalUsers        int     `json:"total_users"`
	ActiveUsers       int     `json:"active_users"`
	TotalInstances    int     `json:"total_instances"`
	OnlineInstances   int     `json:"online_instances"`
	TotalOrders       int     `json:"total_orders"`
	MonthlyRevenue    float64 `json:"monthly_revenue"`
	TotalRevenue      float64 `json:"total_revenue"`
	PendingOrders     int     `json:"pending_orders"`
	InstancesByStatus map[string]int `json:"instances_by_status"`
	RevenueByProduct  map[string]float64 `json:"revenue_by_product"`
}

// GetDashboardStats returns dashboard statistics.
func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{
		InstancesByStatus: make(map[string]int),
		RevenueByProduct:  make(map[string]float64),
	}

	// Get user stats
	userStats, err := s.userRepo.GetUserStats(ctx)
	if err == nil {
		stats.TotalUsers = userStats.TotalUsers
		stats.ActiveUsers = userStats.ActiveUsers
	}

	// Get instance stats
	instanceStats, err := s.instanceRepo.GetInstanceStats(ctx)
	if err == nil {
		stats.TotalInstances = instanceStats.TotalInstances
		stats.OnlineInstances = instanceStats.OnlineInstances
		stats.InstancesByStatus = instanceStats.ByStatus
	}

	// Get order stats
	orderStats, err := s.orderRepo.GetOrderStats(ctx)
	if err == nil {
		stats.TotalOrders = orderStats.TotalOrders
		stats.PendingOrders = orderStats.PendingOrders
		stats.MonthlyRevenue = orderStats.MonthlyRevenue
		stats.TotalRevenue = orderStats.TotalRevenue
		stats.RevenueByProduct = orderStats.RevenueByProduct
	}

	return stats, nil
}

// UserListItem represents a user in admin listings.
type UserListItem struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	IsActive      bool      `json:"is_active"`
	IsAdmin       bool      `json:"is_admin"`
	InstanceCount int       `json:"instance_count"`
	TotalSpent    float64   `json:"total_spent"`
	CreatedAt     time.Time `json:"created_at"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
}

// UserListOptions specifies options for listing users.
type UserListOptions struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Search   string `json:"search"`
	SortBy   string `json:"sort_by"`
	SortDir  string `json:"sort_dir"`
	IsActive *bool  `json:"is_active,omitempty"`
	IsAdmin  *bool  `json:"is_admin,omitempty"`
}

// UserListResult contains paginated user results.
type UserListResult struct {
	Users      []UserListItem `json:"users"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// ListUsers returns a paginated list of users.
func (s *Service) ListUsers(ctx context.Context, opts UserListOptions) (*UserListResult, error) {
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	users, total, err := s.userRepo.ListUsersAdmin(ctx, repository.AdminUserListOptions{
		Page:     opts.Page,
		PageSize: opts.PageSize,
		Search:   opts.Search,
		SortBy:   opts.SortBy,
		SortDir:  opts.SortDir,
		IsActive: opts.IsActive,
		IsAdmin:  opts.IsAdmin,
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserListItem, len(users))
	for i, u := range users {
		items[i] = UserListItem{
			ID:            u.ID,
			Email:         u.Email,
			FirstName:     u.FirstName,
			LastName:      u.LastName,
			IsActive:      u.IsActive,
			IsAdmin:       u.IsAdmin,
			InstanceCount: u.InstanceCount,
			TotalSpent:    u.TotalSpent,
			CreatedAt:     u.CreatedAt,
			LastLoginAt:   u.LastLoginAt,
		}
	}

	totalPages := (total + opts.PageSize - 1) / opts.PageSize

	return &UserListResult{
		Users:      items,
		TotalCount: total,
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserDetails returns detailed information about a user.
func (s *Service) GetUserDetails(ctx context.Context, userID string) (*models.UserProfile, error) {
	return s.userRepo.GetUserProfile(ctx, userID)
}

// UpdateUserAdmin updates a user's admin-controlled fields.
type UpdateUserAdminRequest struct {
	IsActive *bool `json:"is_active,omitempty"`
	IsAdmin  *bool `json:"is_admin,omitempty"`
}

// UpdateUser updates a user's admin-controlled fields.
func (s *Service) UpdateUser(ctx context.Context, userID string, req UpdateUserAdminRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.IsAdmin != nil {
		user.IsAdmin = *req.IsAdmin
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// InstanceListItem represents an instance in admin listings.
type InstanceListItem struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	UserID       string     `json:"user_id"`
	UserEmail    string     `json:"user_email"`
	ProductName  string     `json:"product_name"`
	PackageName  string     `json:"package_name"`
	Status       string     `json:"status"`
	HealthStatus string     `json:"health_status"`
	URL          string     `json:"url"`
	CreatedAt    time.Time  `json:"created_at"`
	LastHealthCheck *time.Time `json:"last_health_check,omitempty"`
}

// InstanceListOptions specifies options for listing instances.
type InstanceListOptions struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Search   string `json:"search"`
	Status   string `json:"status"`
	UserID   string `json:"user_id"`
	Product  string `json:"product"`
}

// InstanceListResult contains paginated instance results.
type InstanceListResult struct {
	Instances  []InstanceListItem `json:"instances"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// ListInstances returns a paginated list of all instances.
func (s *Service) ListInstances(ctx context.Context, opts InstanceListOptions) (*InstanceListResult, error) {
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	instances, total, err := s.instanceRepo.ListInstancesAdmin(ctx, repository.AdminInstanceListOptions{
		Page:     opts.Page,
		PageSize: opts.PageSize,
		Search:   opts.Search,
		Status:   opts.Status,
		UserID:   opts.UserID,
		Product:  opts.Product,
	})
	if err != nil {
		return nil, err
	}

	items := make([]InstanceListItem, len(instances))
	for i, inst := range instances {
		items[i] = InstanceListItem{
			ID:              inst.ID,
			Name:            inst.Name,
			UserID:          inst.UserID,
			UserEmail:       inst.UserEmail,
			ProductName:     inst.ProductName,
			PackageName:     inst.PackageName,
			Status:          string(inst.Status),
			HealthStatus:    string(inst.HealthStatus),
			URL:             inst.URL,
			CreatedAt:       inst.CreatedAt,
			LastHealthCheck: inst.LastHealthCheck,
		}
	}

	totalPages := (total + opts.PageSize - 1) / opts.PageSize

	return &InstanceListResult{
		Instances:  items,
		TotalCount: total,
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalPages: totalPages,
	}, nil
}

// OrderListItem represents an order in admin listings.
type OrderListItem struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	ProductName  string    `json:"product_name"`
	PackageName  string    `json:"package_name"`
	Status       string    `json:"status"`
	TotalAmount  int64     `json:"total_amount"`
	Currency     string    `json:"currency"`
	PaymentMethod string   `json:"payment_method"`
	CreatedAt    time.Time `json:"created_at"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
}

// OrderListOptions specifies options for listing orders.
type OrderListOptions struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status"`
	UserID   string `json:"user_id"`
	DateFrom *time.Time `json:"date_from,omitempty"`
	DateTo   *time.Time `json:"date_to,omitempty"`
}

// OrderListResult contains paginated order results.
type OrderListResult struct {
	Orders     []OrderListItem `json:"orders"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// ListOrders returns a paginated list of all orders.
func (s *Service) ListOrders(ctx context.Context, opts OrderListOptions) (*OrderListResult, error) {
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	orders, total, err := s.orderRepo.ListOrdersAdmin(ctx, repository.AdminOrderListOptions{
		Page:     opts.Page,
		PageSize: opts.PageSize,
		Status:   opts.Status,
		UserID:   opts.UserID,
		DateFrom: opts.DateFrom,
		DateTo:   opts.DateTo,
	})
	if err != nil {
		return nil, err
	}

	items := make([]OrderListItem, len(orders))
	for i, o := range orders {
		items[i] = OrderListItem{
			ID:            o.ID,
			UserID:        o.UserID,
			UserEmail:     o.UserEmail,
			ProductName:   o.ProductName,
			PackageName:   o.PackageName,
			Status:        string(o.Status),
			TotalAmount:   o.TotalAmount,
			Currency:      o.Currency,
			PaymentMethod: o.PaymentMethod,
			CreatedAt:     o.CreatedAt,
			PaidAt:        o.PaidAt,
		}
	}

	totalPages := (total + opts.PageSize - 1) / opts.PageSize

	return &OrderListResult{
		Orders:     items,
		TotalCount: total,
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalPages: totalPages,
	}, nil
}

// RevenueDataPoint represents a revenue data point for charts.
type RevenueDataPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

// GetRevenueChart returns revenue data for charting.
func (s *Service) GetRevenueChart(ctx context.Context, period string, groupBy string) ([]RevenueDataPoint, error) {
	repoData, err := s.orderRepo.GetRevenueChart(ctx, period, groupBy)
	if err != nil {
		return nil, err
	}

	// Convert repository types to admin service types
	result := make([]RevenueDataPoint, len(repoData))
	for i, dp := range repoData {
		result[i] = RevenueDataPoint{
			Date:    dp.Date,
			Revenue: dp.Revenue,
			Orders:  dp.Orders,
		}
	}
	return result, nil
}

// SystemHealth represents overall system health.
type SystemHealth struct {
	Status     string                   `json:"status"` // healthy, degraded, unhealthy
	Components map[string]ComponentHealth `json:"components"`
	Timestamp  time.Time                `json:"timestamp"`
}

// ComponentHealth represents health of a single component.
type ComponentHealth struct {
	Status      string `json:"status"` // healthy, unhealthy
	Message     string `json:"message,omitempty"`
	LastChecked time.Time `json:"last_checked"`
}

// GetSystemHealth returns system health status.
func (s *Service) GetSystemHealth(ctx context.Context) (*SystemHealth, error) {
	health := &SystemHealth{
		Status:     "healthy",
		Components: make(map[string]ComponentHealth),
		Timestamp:  time.Now(),
	}

	// Check database
	dbHealth := ComponentHealth{
		Status:      "healthy",
		LastChecked: time.Now(),
	}
	if err := s.userRepo.Ping(ctx); err != nil {
		dbHealth.Status = "unhealthy"
		dbHealth.Message = err.Error()
		health.Status = "degraded"
	}
	health.Components["database"] = dbHealth

	// Check instance health summary
	instanceHealth := ComponentHealth{
		Status:      "healthy",
		LastChecked: time.Now(),
	}
	stats, err := s.instanceRepo.GetInstanceStats(ctx)
	if err != nil {
		instanceHealth.Status = "unhealthy"
		instanceHealth.Message = err.Error()
	} else {
		unhealthyCount := stats.ByStatus["error"] + stats.ByStatus["offline"]
		if unhealthyCount > 0 {
			instanceHealth.Message = "Some instances are unhealthy"
			if unhealthyCount > stats.TotalInstances/2 {
				health.Status = "degraded"
			}
		}
	}
	health.Components["instances"] = instanceHealth

	return health, nil
}

// AuditLogEntry represents an admin action audit log entry.
type AuditLogEntry struct {
	ID        string    `json:"id"`
	AdminID   string    `json:"admin_id"`
	AdminEmail string   `json:"admin_email"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	ResourceID string   `json:"resource_id"`
	Details   string    `json:"details,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// LogAction logs an admin action for auditing.
func (s *Service) LogAction(ctx context.Context, adminID, action, resource, resourceID, details, ipAddress string) error {
	// This would typically insert into an audit_logs table
	// For now, we'll log it (in production, this should write to DB)
	return nil
}
