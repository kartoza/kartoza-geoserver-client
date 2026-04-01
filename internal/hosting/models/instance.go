package models

import (
	"encoding/json"
	"time"
)

// InstanceStatus represents the status of a deployed instance.
type InstanceStatus string

const (
	InstanceStatusPending     InstanceStatus = "pending"
	InstanceStatusDeploying   InstanceStatus = "deploying"
	InstanceStatusStartingUp  InstanceStatus = "starting_up"
	InstanceStatusOnline      InstanceStatus = "online"
	InstanceStatusOffline     InstanceStatus = "offline"
	InstanceStatusError       InstanceStatus = "error"
	InstanceStatusMaintenance InstanceStatus = "maintenance"
	InstanceStatusDeleting    InstanceStatus = "deleting"
	InstanceStatusDeleted     InstanceStatus = "deleted"
)

// HealthStatus represents the health state of an instance.
type HealthStatus string

const (
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// Instance represents a deployed service instance.
type Instance struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	SalesOrderID    string         `json:"sales_order_id"`
	ProductID       string         `json:"product_id"`
	PackageID       string         `json:"package_id"`
	ClusterID       string         `json:"cluster_id"`
	Name            string         `json:"name"`
	DisplayName     string         `json:"display_name,omitempty"`
	Status          InstanceStatus `json:"status"`
	URL             string         `json:"url,omitempty"`
	InternalURL     string         `json:"internal_url,omitempty"`
	HealthEndpoint  string         `json:"health_endpoint,omitempty"`
	VaultPath       string         `json:"vault_path,omitempty"`
	AdminUsername   string         `json:"admin_username,omitempty"`
	HealthStatus    HealthStatus   `json:"health_status"`
	HealthMessage   string         `json:"health_message,omitempty"`
	LastHealthCheck *time.Time     `json:"last_health_check,omitempty"`
	CPUUsage        *float64       `json:"cpu_usage,omitempty"`
	MemoryUsage     *float64       `json:"memory_usage,omitempty"`
	StorageUsage    *float64       `json:"storage_usage,omitempty"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	DeletedAt       *time.Time     `json:"deleted_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Related data (populated by queries)
	User         *User         `json:"user,omitempty"`
	Product      *Product      `json:"product,omitempty"`
	Package      *Package      `json:"package,omitempty"`
	Cluster      *Cluster      `json:"cluster,omitempty"`
	Subscription *Subscription `json:"subscription,omitempty"`
}

// IsOperational returns true if the instance is running and accessible.
func (i *Instance) IsOperational() bool {
	return i.Status == InstanceStatusOnline || i.Status == InstanceStatusStartingUp
}

// CanBeDeleted returns true if the instance can be deleted.
func (i *Instance) CanBeDeleted() bool {
	return i.Status != InstanceStatusDeleting && i.Status != InstanceStatusDeleted
}

// CanBeRestarted returns true if the instance can be restarted.
func (i *Instance) CanBeRestarted() bool {
	return i.Status == InstanceStatusOnline || i.Status == InstanceStatusOffline || i.Status == InstanceStatusError
}

// StatusBadgeColor returns a CSS color class for the status.
func (i *Instance) StatusBadgeColor() string {
	switch i.Status {
	case InstanceStatusOnline:
		return "green"
	case InstanceStatusStartingUp, InstanceStatusDeploying:
		return "yellow"
	case InstanceStatusOffline, InstanceStatusMaintenance:
		return "gray"
	case InstanceStatusError:
		return "red"
	case InstanceStatusDeleting, InstanceStatusDeleted:
		return "gray"
	default:
		return "gray"
	}
}

// InstanceCredentials represents the access credentials for an instance.
type InstanceCredentials struct {
	URL           string            `json:"url"`
	AdminUsername string            `json:"admin_username"`
	AdminPassword string            `json:"admin_password"`
	DatabaseHost  string            `json:"database_host,omitempty"`
	DatabasePort  int               `json:"database_port,omitempty"`
	DatabaseName  string            `json:"database_name,omitempty"`
	DatabaseUser  string            `json:"database_user,omitempty"`
	DatabasePass  string            `json:"database_pass,omitempty"`
	Extra         map[string]string `json:"extra,omitempty"`
}

// ActivityType represents the type of deployment activity.
type ActivityType string

const (
	ActivityTypeCreate    ActivityType = "create"
	ActivityTypeDelete    ActivityType = "delete"
	ActivityTypeRestart   ActivityType = "restart"
	ActivityTypeUpgrade   ActivityType = "upgrade"
	ActivityTypeDowngrade ActivityType = "downgrade"
	ActivityTypeBackup    ActivityType = "backup"
	ActivityTypeRestore   ActivityType = "restore"
	ActivityTypeScale     ActivityType = "scale"
)

// ActivityStatus represents the status of a deployment activity.
type ActivityStatus string

const (
	ActivityStatusPending   ActivityStatus = "pending"
	ActivityStatusRunning   ActivityStatus = "running"
	ActivityStatusSuccess   ActivityStatus = "success"
	ActivityStatusError     ActivityStatus = "error"
	ActivityStatusCancelled ActivityStatus = "cancelled"
)

// Activity represents a deployment or management activity on an instance.
type Activity struct {
	ID                 string          `json:"id"`
	InstanceID         string          `json:"instance_id"`
	UserID             string          `json:"user_id,omitempty"`
	ActivityType       ActivityType    `json:"activity_type"`
	Status             ActivityStatus  `json:"status"`
	JenkinsBuildNumber *int            `json:"jenkins_build_number,omitempty"`
	JenkinsBuildURL    string          `json:"jenkins_build_url,omitempty"`
	ArgoCDAppName      string          `json:"argocd_app_name,omitempty"`
	ErrorMessage       string          `json:"error_message,omitempty"`
	Metadata           json.RawMessage `json:"metadata,omitempty"`
	StartedAt          *time.Time      `json:"started_at,omitempty"`
	CompletedAt        *time.Time      `json:"completed_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`

	// Related data
	Instance *Instance `json:"instance,omitempty"`
	User     *User     `json:"user,omitempty"`
}

// Duration returns the duration of the activity if completed.
func (a *Activity) Duration() time.Duration {
	if a.StartedAt == nil {
		return 0
	}
	if a.CompletedAt == nil {
		return time.Since(*a.StartedAt)
	}
	return a.CompletedAt.Sub(*a.StartedAt)
}

// CreateInstanceRequest represents a request to create an instance (internal).
type CreateInstanceRequest struct {
	UserID       string `json:"user_id"`
	SalesOrderID string `json:"sales_order_id"`
	ProductID    string `json:"product_id"`
	PackageID    string `json:"package_id"`
	ClusterID    string `json:"cluster_id"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name,omitempty"`
}

// InstanceSummary provides a summary view of an instance.
type InstanceSummary struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	DisplayName  string         `json:"display_name"`
	ProductName  string         `json:"product_name"`
	PackageName  string         `json:"package_name"`
	ClusterName  string         `json:"cluster_name"`
	Status       InstanceStatus `json:"status"`
	HealthStatus HealthStatus   `json:"health_status"`
	URL          string         `json:"url"`
	CreatedAt    time.Time      `json:"created_at"`
}
