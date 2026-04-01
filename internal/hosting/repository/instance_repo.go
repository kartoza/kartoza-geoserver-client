package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/db"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
)

// InstanceRepository provides data access for instances and activities.
type InstanceRepository struct {
	db *db.DB
}

// AdminInstanceListOptions specifies options for listing instances in admin.
type AdminInstanceListOptions struct {
	Page     int
	PageSize int
	Search   string
	Status   string
	UserID   string
	Product  string
}

// AdminInstanceListItem represents an instance in admin listings.
type AdminInstanceListItem struct {
	ID              string
	Name            string
	UserID          string
	UserEmail       string
	ProductName     string
	PackageName     string
	Status          models.InstanceStatus
	HealthStatus    models.HealthStatus
	URL             string
	CreatedAt       time.Time
	LastHealthCheck *time.Time
}

// InstanceStats represents aggregate instance statistics.
type InstanceStats struct {
	TotalInstances  int
	OnlineInstances int
	ByStatus        map[string]int
}

// NewInstanceRepository creates a new instance repository.
func NewInstanceRepository(database *db.DB) *InstanceRepository {
	return &InstanceRepository{db: database}
}

// CreateInstance creates a new instance.
func (r *InstanceRepository) CreateInstance(ctx context.Context, inst *models.Instance) error {
	query := `
		INSERT INTO instances (
			user_id, sales_order_id, product_id, package_id, cluster_id,
			name, display_name, status, url, internal_url, vault_path, admin_username
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, health_status, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		inst.UserID, inst.SalesOrderID, inst.ProductID, inst.PackageID, inst.ClusterID,
		inst.Name, nullString(inst.DisplayName), inst.Status,
		nullString(inst.URL), nullString(inst.InternalURL),
		nullString(inst.VaultPath), nullString(inst.AdminUsername),
	).Scan(&inst.ID, &inst.HealthStatus, &inst.CreatedAt, &inst.UpdatedAt)
}

// GetInstanceByID retrieves an instance by ID.
func (r *InstanceRepository) GetInstanceByID(ctx context.Context, id string) (*models.Instance, error) {
	query := `
		SELECT id, user_id, sales_order_id, product_id, package_id, cluster_id,
		       name, display_name, status, url, internal_url, vault_path, admin_username,
		       health_status, health_message, last_health_check,
		       cpu_usage, memory_usage, storage_usage, expires_at, deleted_at,
		       created_at, updated_at
		FROM instances WHERE id = $1
	`
	inst := &models.Instance{}
	var displayName, url, internalURL, vaultPath, adminUsername, healthMessage sql.NullString
	var lastHealthCheck, expiresAt, deletedAt sql.NullTime
	var cpuUsage, memUsage, storageUsage sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inst.ID, &inst.UserID, &inst.SalesOrderID, &inst.ProductID, &inst.PackageID,
		&inst.ClusterID, &inst.Name, &displayName, &inst.Status,
		&url, &internalURL, &vaultPath, &adminUsername,
		&inst.HealthStatus, &healthMessage, &lastHealthCheck,
		&cpuUsage, &memUsage, &storageUsage, &expiresAt, &deletedAt,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	inst.DisplayName = displayName.String
	inst.URL = url.String
	inst.InternalURL = internalURL.String
	inst.VaultPath = vaultPath.String
	inst.AdminUsername = adminUsername.String
	inst.HealthMessage = healthMessage.String
	if lastHealthCheck.Valid {
		inst.LastHealthCheck = &lastHealthCheck.Time
	}
	if cpuUsage.Valid {
		inst.CPUUsage = &cpuUsage.Float64
	}
	if memUsage.Valid {
		inst.MemoryUsage = &memUsage.Float64
	}
	if storageUsage.Valid {
		inst.StorageUsage = &storageUsage.Float64
	}
	if expiresAt.Valid {
		inst.ExpiresAt = &expiresAt.Time
	}
	if deletedAt.Valid {
		inst.DeletedAt = &deletedAt.Time
	}

	return inst, nil
}

// GetInstanceByName retrieves an instance by cluster and name.
func (r *InstanceRepository) GetInstanceByName(ctx context.Context, clusterID, name string) (*models.Instance, error) {
	query := `
		SELECT id, user_id, sales_order_id, product_id, package_id, cluster_id,
		       name, display_name, status, url, internal_url, vault_path, admin_username,
		       health_status, health_message, last_health_check,
		       cpu_usage, memory_usage, storage_usage, expires_at, deleted_at,
		       created_at, updated_at
		FROM instances WHERE cluster_id = $1 AND name = $2
	`
	inst := &models.Instance{}
	var displayName, url, internalURL, vaultPath, adminUsername, healthMessage sql.NullString
	var lastHealthCheck, expiresAt, deletedAt sql.NullTime
	var cpuUsage, memUsage, storageUsage sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, clusterID, name).Scan(
		&inst.ID, &inst.UserID, &inst.SalesOrderID, &inst.ProductID, &inst.PackageID,
		&inst.ClusterID, &inst.Name, &displayName, &inst.Status,
		&url, &internalURL, &vaultPath, &adminUsername,
		&inst.HealthStatus, &healthMessage, &lastHealthCheck,
		&cpuUsage, &memUsage, &storageUsage, &expiresAt, &deletedAt,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	inst.DisplayName = displayName.String
	inst.URL = url.String
	inst.InternalURL = internalURL.String
	inst.VaultPath = vaultPath.String
	inst.AdminUsername = adminUsername.String
	inst.HealthMessage = healthMessage.String
	if lastHealthCheck.Valid {
		inst.LastHealthCheck = &lastHealthCheck.Time
	}
	if cpuUsage.Valid {
		inst.CPUUsage = &cpuUsage.Float64
	}
	if memUsage.Valid {
		inst.MemoryUsage = &memUsage.Float64
	}
	if storageUsage.Valid {
		inst.StorageUsage = &storageUsage.Float64
	}
	if expiresAt.Valid {
		inst.ExpiresAt = &expiresAt.Time
	}
	if deletedAt.Valid {
		inst.DeletedAt = &deletedAt.Time
	}

	return inst, nil
}

// UpdateInstanceStatus updates the status of an instance.
func (r *InstanceRepository) UpdateInstanceStatus(ctx context.Context, id string, status models.InstanceStatus) error {
	query := `UPDATE instances SET status = $2 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateInstanceHealth updates the health information for an instance.
func (r *InstanceRepository) UpdateInstanceHealth(ctx context.Context, id string, healthStatus models.HealthStatus, message string) error {
	query := `
		UPDATE instances SET
			health_status = $2,
			health_message = $3,
			last_health_check = $4
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, healthStatus, nullString(message), time.Now())
	return err
}

// UpdateInstanceUsage updates resource usage for an instance.
func (r *InstanceRepository) UpdateInstanceUsage(ctx context.Context, id string, cpu, memory, storage *float64) error {
	query := `
		UPDATE instances SET
			cpu_usage = COALESCE($2, cpu_usage),
			memory_usage = COALESCE($3, memory_usage),
			storage_usage = COALESCE($4, storage_usage)
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, cpu, memory, storage)
	return err
}

// UpdateInstanceURL updates the URL for an instance.
func (r *InstanceRepository) UpdateInstanceURL(ctx context.Context, id, url, internalURL string) error {
	query := `
		UPDATE instances SET
			url = COALESCE($2, url),
			internal_url = COALESCE($3, internal_url)
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, nullString(url), nullString(internalURL))
	return err
}

// UpdateInstanceVault updates the vault path and admin username for an instance.
func (r *InstanceRepository) UpdateInstanceVault(ctx context.Context, id, vaultPath, adminUsername string) error {
	query := `
		UPDATE instances SET
			vault_path = COALESCE($2, vault_path),
			admin_username = COALESCE($3, admin_username)
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, nullString(vaultPath), nullString(adminUsername))
	return err
}

// SoftDeleteInstance marks an instance as deleted.
func (r *InstanceRepository) SoftDeleteInstance(ctx context.Context, id string) error {
	query := `
		UPDATE instances SET
			status = 'deleted',
			deleted_at = $2
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

// ListUserInstances retrieves instances for a user.
func (r *InstanceRepository) ListUserInstances(ctx context.Context, userID string, includeDeleted bool) ([]*models.Instance, error) {
	query := `
		SELECT id, user_id, sales_order_id, product_id, package_id, cluster_id,
		       name, display_name, status, url, internal_url, vault_path, admin_username,
		       health_status, health_message, last_health_check,
		       cpu_usage, memory_usage, storage_usage, expires_at, deleted_at,
		       created_at, updated_at
		FROM instances WHERE user_id = $1
	`
	if !includeDeleted {
		query += ` AND status != 'deleted'`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInstances(rows)
}

// ListInstancesByCluster retrieves instances for a cluster.
func (r *InstanceRepository) ListInstancesByCluster(ctx context.Context, clusterID string) ([]*models.Instance, error) {
	query := `
		SELECT id, user_id, sales_order_id, product_id, package_id, cluster_id,
		       name, display_name, status, url, internal_url, vault_path, admin_username,
		       health_status, health_message, last_health_check,
		       cpu_usage, memory_usage, storage_usage, expires_at, deleted_at,
		       created_at, updated_at
		FROM instances WHERE cluster_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInstances(rows)
}

// ListInstancesNeedingHealthCheck retrieves instances that need a health check.
func (r *InstanceRepository) ListInstancesNeedingHealthCheck(ctx context.Context, olderThan time.Duration) ([]*models.Instance, error) {
	cutoff := time.Now().Add(-olderThan)
	query := `
		SELECT id, user_id, sales_order_id, product_id, package_id, cluster_id,
		       name, display_name, status, url, internal_url, vault_path, admin_username,
		       health_status, health_message, last_health_check,
		       cpu_usage, memory_usage, storage_usage, expires_at, deleted_at,
		       created_at, updated_at
		FROM instances
		WHERE status IN ('online', 'offline', 'error')
		  AND (last_health_check IS NULL OR last_health_check < $1)
		ORDER BY last_health_check NULLS FIRST
	`
	rows, err := r.db.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInstances(rows)
}

func (r *InstanceRepository) scanInstances(rows *sql.Rows) ([]*models.Instance, error) {
	var instances []*models.Instance
	for rows.Next() {
		inst := &models.Instance{}
		var displayName, url, internalURL, vaultPath, adminUsername, healthMessage sql.NullString
		var lastHealthCheck, expiresAt, deletedAt sql.NullTime
		var cpuUsage, memUsage, storageUsage sql.NullFloat64

		if err := rows.Scan(
			&inst.ID, &inst.UserID, &inst.SalesOrderID, &inst.ProductID, &inst.PackageID,
			&inst.ClusterID, &inst.Name, &displayName, &inst.Status,
			&url, &internalURL, &vaultPath, &adminUsername,
			&inst.HealthStatus, &healthMessage, &lastHealthCheck,
			&cpuUsage, &memUsage, &storageUsage, &expiresAt, &deletedAt,
			&inst.CreatedAt, &inst.UpdatedAt,
		); err != nil {
			return nil, err
		}

		inst.DisplayName = displayName.String
		inst.URL = url.String
		inst.InternalURL = internalURL.String
		inst.VaultPath = vaultPath.String
		inst.AdminUsername = adminUsername.String
		inst.HealthMessage = healthMessage.String
		if lastHealthCheck.Valid {
			inst.LastHealthCheck = &lastHealthCheck.Time
		}
		if cpuUsage.Valid {
			inst.CPUUsage = &cpuUsage.Float64
		}
		if memUsage.Valid {
			inst.MemoryUsage = &memUsage.Float64
		}
		if storageUsage.Valid {
			inst.StorageUsage = &storageUsage.Float64
		}
		if expiresAt.Valid {
			inst.ExpiresAt = &expiresAt.Time
		}
		if deletedAt.Valid {
			inst.DeletedAt = &deletedAt.Time
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// CreateActivity creates a new deployment activity.
func (r *InstanceRepository) CreateActivity(ctx context.Context, activity *models.Activity) error {
	query := `
		INSERT INTO activities (instance_id, user_id, activity_type, status, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		activity.InstanceID, nullString(activity.UserID),
		activity.ActivityType, activity.Status, activity.Metadata,
	).Scan(&activity.ID, &activity.CreatedAt)
}

// GetActivityByID retrieves an activity by ID.
func (r *InstanceRepository) GetActivityByID(ctx context.Context, id string) (*models.Activity, error) {
	query := `
		SELECT id, instance_id, user_id, activity_type, status,
		       jenkins_build_number, jenkins_build_url, argocd_app_name,
		       error_message, metadata, started_at, completed_at, created_at
		FROM activities WHERE id = $1
	`
	activity := &models.Activity{}
	var userID, jenkinsBuildURL, argocdAppName, errorMessage sql.NullString
	var jenkinsBuildNum sql.NullInt32
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&activity.ID, &activity.InstanceID, &userID, &activity.ActivityType, &activity.Status,
		&jenkinsBuildNum, &jenkinsBuildURL, &argocdAppName,
		&errorMessage, &activity.Metadata, &startedAt, &completedAt, &activity.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	activity.UserID = userID.String
	if jenkinsBuildNum.Valid {
		v := int(jenkinsBuildNum.Int32)
		activity.JenkinsBuildNumber = &v
	}
	activity.JenkinsBuildURL = jenkinsBuildURL.String
	activity.ArgoCDAppName = argocdAppName.String
	activity.ErrorMessage = errorMessage.String
	if startedAt.Valid {
		activity.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		activity.CompletedAt = &completedAt.Time
	}

	return activity, nil
}

// UpdateActivityStatus updates the status of an activity.
func (r *InstanceRepository) UpdateActivityStatus(ctx context.Context, id string, status models.ActivityStatus, errorMessage string) error {
	var query string
	var args []interface{}

	if status == models.ActivityStatusRunning {
		query = `
			UPDATE activities SET
				status = $2,
				started_at = COALESCE(started_at, $3)
			WHERE id = $1
		`
		args = []interface{}{id, status, time.Now()}
	} else if status == models.ActivityStatusSuccess || status == models.ActivityStatusError || status == models.ActivityStatusCancelled {
		query = `
			UPDATE activities SET
				status = $2,
				error_message = $3,
				completed_at = $4
			WHERE id = $1
		`
		args = []interface{}{id, status, nullString(errorMessage), time.Now()}
	} else {
		query = `UPDATE activities SET status = $2 WHERE id = $1`
		args = []interface{}{id, status}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// UpdateActivityJenkins updates Jenkins information for an activity.
func (r *InstanceRepository) UpdateActivityJenkins(ctx context.Context, id string, buildNumber int, buildURL string) error {
	query := `
		UPDATE activities SET
			jenkins_build_number = $2,
			jenkins_build_url = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, buildNumber, nullString(buildURL))
	return err
}

// ListInstanceActivities retrieves activities for an instance.
func (r *InstanceRepository) ListInstanceActivities(ctx context.Context, instanceID string, limit int) ([]*models.Activity, error) {
	query := `
		SELECT id, instance_id, user_id, activity_type, status,
		       jenkins_build_number, jenkins_build_url, argocd_app_name,
		       error_message, metadata, started_at, completed_at, created_at
		FROM activities WHERE instance_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, instanceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		var userID, jenkinsBuildURL, argocdAppName, errorMessage sql.NullString
		var jenkinsBuildNum sql.NullInt32
		var startedAt, completedAt sql.NullTime

		if err := rows.Scan(
			&activity.ID, &activity.InstanceID, &userID, &activity.ActivityType, &activity.Status,
			&jenkinsBuildNum, &jenkinsBuildURL, &argocdAppName,
			&errorMessage, &activity.Metadata, &startedAt, &completedAt, &activity.CreatedAt,
		); err != nil {
			return nil, err
		}

		activity.UserID = userID.String
		if jenkinsBuildNum.Valid {
			v := int(jenkinsBuildNum.Int32)
			activity.JenkinsBuildNumber = &v
		}
		activity.JenkinsBuildURL = jenkinsBuildURL.String
		activity.ArgoCDAppName = argocdAppName.String
		activity.ErrorMessage = errorMessage.String
		if startedAt.Valid {
			activity.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			activity.CompletedAt = &completedAt.Time
		}
		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

// GetPendingActivities retrieves activities that are pending or running.
func (r *InstanceRepository) GetPendingActivities(ctx context.Context) ([]*models.Activity, error) {
	query := `
		SELECT id, instance_id, user_id, activity_type, status,
		       jenkins_build_number, jenkins_build_url, argocd_app_name,
		       error_message, metadata, started_at, completed_at, created_at
		FROM activities WHERE status IN ('pending', 'running')
		ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		var userID, jenkinsBuildURL, argocdAppName, errorMessage sql.NullString
		var jenkinsBuildNum sql.NullInt32
		var startedAt, completedAt sql.NullTime

		if err := rows.Scan(
			&activity.ID, &activity.InstanceID, &userID, &activity.ActivityType, &activity.Status,
			&jenkinsBuildNum, &jenkinsBuildURL, &argocdAppName,
			&errorMessage, &activity.Metadata, &startedAt, &completedAt, &activity.CreatedAt,
		); err != nil {
			return nil, err
		}

		activity.UserID = userID.String
		if jenkinsBuildNum.Valid {
			v := int(jenkinsBuildNum.Int32)
			activity.JenkinsBuildNumber = &v
		}
		activity.JenkinsBuildURL = jenkinsBuildURL.String
		activity.ArgoCDAppName = argocdAppName.String
		activity.ErrorMessage = errorMessage.String
		if startedAt.Valid {
			activity.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			activity.CompletedAt = &completedAt.Time
		}
		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

// GetInstanceSummaries retrieves instance summaries with product/package names.
func (r *InstanceRepository) GetInstanceSummaries(ctx context.Context, userID string) ([]*models.InstanceSummary, error) {
	query := `
		SELECT i.id, i.name, COALESCE(i.display_name, i.name) as display_name,
		       p.name as product_name, pk.name as package_name, c.name as cluster_name,
		       i.status, i.health_status, i.url, i.created_at
		FROM instances i
		JOIN products p ON i.product_id = p.id
		JOIN packages pk ON i.package_id = pk.id
		JOIN clusters c ON i.cluster_id = c.id
		WHERE i.user_id = $1 AND i.status != 'deleted'
		ORDER BY i.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.InstanceSummary
	for rows.Next() {
		s := &models.InstanceSummary{}
		var url sql.NullString
		if err := rows.Scan(
			&s.ID, &s.Name, &s.DisplayName, &s.ProductName, &s.PackageName,
			&s.ClusterName, &s.Status, &s.HealthStatus, &url, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		s.URL = url.String
		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
}

// InstanceNameExists checks if an instance name is already used in a cluster.
func (r *InstanceRepository) InstanceNameExists(ctx context.Context, clusterID, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM instances WHERE cluster_id = $1 AND name = $2 AND status != 'deleted')`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, clusterID, name).Scan(&exists)
	return exists, err
}

// Create is an alias for CreateInstance.
func (r *InstanceRepository) Create(ctx context.Context, inst *models.Instance) error {
	return r.CreateInstance(ctx, inst)
}

// GetByID is an alias for GetInstanceByID.
func (r *InstanceRepository) GetByID(ctx context.Context, id string) (*models.Instance, error) {
	return r.GetInstanceByID(ctx, id)
}

// UpdateStatus is an alias for UpdateInstanceStatus.
func (r *InstanceRepository) UpdateStatus(ctx context.Context, id string, status models.InstanceStatus) error {
	return r.UpdateInstanceStatus(ctx, id, status)
}

// UpdateHealthStatus updates the health status for an instance.
func (r *InstanceRepository) UpdateHealthStatus(ctx context.Context, id string, healthStatus string) error {
	query := `
		UPDATE instances SET
			health_status = $2,
			last_health_check = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, healthStatus, time.Now())
	return err
}

// UpdateActivityBuildURL updates the Jenkins build URL for an activity.
func (r *InstanceRepository) UpdateActivityBuildURL(ctx context.Context, id string, buildURL string) error {
	query := `UPDATE activities SET jenkins_build_url = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, buildURL)
	return err
}

// GetActiveInstances retrieves all instances that should be health checked.
func (r *InstanceRepository) GetActiveInstances(ctx context.Context) ([]*models.Instance, error) {
	query := `
		SELECT id, user_id, sales_order_id, product_id, package_id, cluster_id,
		       name, display_name, status, url, internal_url, vault_path, admin_username,
		       health_status, health_message, last_health_check,
		       cpu_usage, memory_usage, storage_usage, expires_at, deleted_at,
		       created_at, updated_at
		FROM instances
		WHERE status IN ('online', 'starting_up', 'offline', 'error')
		ORDER BY last_health_check NULLS FIRST
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInstances(rows)
}

// GetInstanceStats returns aggregate instance statistics.
func (r *InstanceRepository) GetInstanceStats(ctx context.Context) (*InstanceStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'online') as online,
			status,
			COUNT(*) as status_count
		FROM instances
		WHERE status != 'deleted'
		GROUP BY GROUPING SETS ((), (status))
	`

	stats := &InstanceStats{
		ByStatus: make(map[string]int),
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var total, online int
		var status sql.NullString
		var statusCount int

		if err := rows.Scan(&total, &online, &status, &statusCount); err != nil {
			return nil, err
		}

		if status.Valid {
			stats.ByStatus[status.String] = statusCount
		} else {
			stats.TotalInstances = total
			stats.OnlineInstances = online
		}
	}

	return stats, rows.Err()
}

// ListInstancesAdmin retrieves instances for admin listing with filtering and pagination.
func (r *InstanceRepository) ListInstancesAdmin(ctx context.Context, opts AdminInstanceListOptions) ([]AdminInstanceListItem, int, error) {
	// Build WHERE clause
	conditions := []string{"i.status != 'deleted'"}
	args := []interface{}{}
	argNum := 1

	if opts.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(i.name ILIKE $%d OR i.display_name ILIKE $%d OR u.email ILIKE $%d)",
			argNum, argNum, argNum,
		))
		args = append(args, "%"+opts.Search+"%")
		argNum++
	}

	if opts.Status != "" {
		conditions = append(conditions, fmt.Sprintf("i.status = $%d", argNum))
		args = append(args, opts.Status)
		argNum++
	}

	if opts.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("i.user_id = $%d", argNum))
		args = append(args, opts.UserID)
		argNum++
	}

	if opts.Product != "" {
		conditions = append(conditions, fmt.Sprintf("p.slug = $%d", argNum))
		args = append(args, opts.Product)
		argNum++
	}

	whereClause := "WHERE " + conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM instances i
		JOIN users u ON i.user_id = u.id
		JOIN products p ON i.product_id = p.id
		%s
	`, whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (opts.Page - 1) * opts.PageSize

	// Main query
	query := fmt.Sprintf(`
		SELECT i.id, i.name, i.user_id, u.email as user_email,
		       p.name as product_name, pk.name as package_name,
		       i.status, i.health_status, COALESCE(i.url, '') as url,
		       i.created_at, i.last_health_check
		FROM instances i
		JOIN users u ON i.user_id = u.id
		JOIN products p ON i.product_id = p.id
		JOIN packages pk ON i.package_id = pk.id
		%s
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, opts.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var instances []AdminInstanceListItem
	for rows.Next() {
		var inst AdminInstanceListItem
		var lastHealthCheck sql.NullTime
		if err := rows.Scan(
			&inst.ID, &inst.Name, &inst.UserID, &inst.UserEmail,
			&inst.ProductName, &inst.PackageName, &inst.Status, &inst.HealthStatus,
			&inst.URL, &inst.CreatedAt, &lastHealthCheck,
		); err != nil {
			return nil, 0, err
		}
		if lastHealthCheck.Valid {
			inst.LastHealthCheck = &lastHealthCheck.Time
		}
		instances = append(instances, inst)
	}

	return instances, total, rows.Err()
}
