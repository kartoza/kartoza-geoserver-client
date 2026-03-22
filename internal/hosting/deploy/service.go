// Package deploy provides deployment functionality for hosted instances.
package deploy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/email"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/repository"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/vault"
)

// Service orchestrates instance deployment.
type Service struct {
	jenkins      *JenkinsClient
	vault        *vault.Client
	instanceRepo *repository.InstanceRepository
	orderRepo    *repository.OrderRepository
	productRepo  *repository.ProductRepository
	userRepo     *repository.UserRepository
	emailService *email.Service
	config       ServiceConfig
}

// ServiceConfig holds deployment service configuration.
type ServiceConfig struct {
	// Jenkins job names per product
	GeoServerJobName string
	GeoNodeJobName   string
	PostGISJobName   string

	// Timeouts
	QueueTimeout time.Duration
	BuildTimeout time.Duration

	// Base domain for instance URLs
	BaseDomain string
}

// NewService creates a new deployment service.
func NewService(
	jenkins *JenkinsClient,
	vaultClient *vault.Client,
	instanceRepo *repository.InstanceRepository,
	orderRepo *repository.OrderRepository,
	productRepo *repository.ProductRepository,
	config ServiceConfig,
) *Service {
	if config.QueueTimeout == 0 {
		config.QueueTimeout = 5 * time.Minute
	}
	if config.BuildTimeout == 0 {
		config.BuildTimeout = 30 * time.Minute
	}

	return &Service{
		jenkins:      jenkins,
		vault:        vaultClient,
		instanceRepo: instanceRepo,
		orderRepo:    orderRepo,
		productRepo:  productRepo,
		config:       config,
	}
}

// SetEmailService sets the email service for notifications.
func (s *Service) SetEmailService(emailService *email.Service) {
	s.emailService = emailService
}

// SetUserRepository sets the user repository for looking up user info.
func (s *Service) SetUserRepository(userRepo *repository.UserRepository) {
	s.userRepo = userRepo
}

// DeployRequest represents a request to deploy a new instance.
type DeployRequest struct {
	OrderID   string
	UserID    string
	ProductID string
	PackageID string
	ClusterID string
	AppName   string
}

// DeployResult represents the result of a deployment.
type DeployResult struct {
	InstanceID string
	ActivityID string
	URL        string
	Status     string
	Error      error
}

// Deploy initiates deployment of a new instance.
func (s *Service) Deploy(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	result := &DeployResult{}

	// Get product details
	product, err := s.productRepo.GetProductByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Get package details
	pkg, err := s.productRepo.GetPackageByID(ctx, req.PackageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	// Get cluster details
	cluster, err := s.productRepo.GetClusterByID(ctx, req.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	// Generate instance URL
	instanceURL := fmt.Sprintf("https://%s.%s", req.AppName, cluster.Domain)

	// Create instance record
	instance := &models.Instance{
		UserID:       req.UserID,
		SalesOrderID: req.OrderID,
		ProductID:    req.ProductID,
		PackageID:    req.PackageID,
		ClusterID:    req.ClusterID,
		Name:         req.AppName,
		Status:       models.InstanceStatusDeploying,
		URL:          instanceURL,
		VaultPath:    fmt.Sprintf("cloudbench/instances/%s/%s", req.UserID, req.AppName),
	}

	if err := s.instanceRepo.Create(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}
	result.InstanceID = instance.ID
	result.URL = instanceURL

	// Create activity record
	activity := &models.Activity{
		InstanceID:   instance.ID,
		ActivityType: models.ActivityTypeCreate,
		Status:       models.ActivityStatusPending,
	}
	if err := s.instanceRepo.CreateActivity(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}
	result.ActivityID = activity.ID

	// Update order status
	if err := s.orderRepo.UpdateOrderStatus(ctx, req.OrderID, models.OrderStatusDeploying); err != nil {
		// Log but don't fail
		fmt.Printf("failed to update order status: %v\n", err)
	}

	// Determine Jenkins job name
	jobName := s.getJobName(product.Slug)
	if jobName == "" {
		return nil, fmt.Errorf("no Jenkins job configured for product: %s", product.Slug)
	}

	// Build Jenkins parameters
	jenkinsParams := map[string]string{
		"APP_NAME":      req.AppName,
		"CLUSTER":       cluster.Code,
		"PRODUCT":       product.Slug,
		"PACKAGE":       pkg.Slug,
		"CPU_LIMIT":     pkg.CPULimit,
		"MEMORY_LIMIT":  pkg.MemoryLimit,
		"STORAGE_LIMIT": pkg.StorageLimit,
		"USER_ID":       req.UserID,
		"ORDER_ID":      req.OrderID,
		"INSTANCE_ID":   instance.ID,
		"VAULT_PATH":    instance.VaultPath,
	}

	// Trigger Jenkins build
	queueID, err := s.jenkins.TriggerBuild(ctx, TriggerBuildParams{
		JobName:    jobName,
		Parameters: jenkinsParams,
	})
	if err != nil {
		// Update activity with error
		s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusError, err.Error())
		s.instanceRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusError)
		result.Error = fmt.Errorf("failed to trigger build: %w", err)
		return result, result.Error
	}

	// Update activity with build info
	s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusRunning, "")

	// Start async monitoring of the build
	go s.monitorBuild(context.Background(), instance.ID, activity.ID, jobName, queueID)

	result.Status = string(models.InstanceStatusDeploying)
	return result, nil
}

// monitorBuild monitors a Jenkins build and updates status.
func (s *Service) monitorBuild(ctx context.Context, instanceID, activityID, jobName string, queueID int) {
	// Wait for build to start
	buildNumber, err := s.jenkins.WaitForBuild(ctx, queueID, s.config.QueueTimeout)
	if err != nil {
		s.instanceRepo.UpdateActivityStatus(ctx, activityID, models.ActivityStatusError, err.Error())
		s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatusError)
		return
	}

	// Update activity with build URL
	buildURL := fmt.Sprintf("%s/job/%s/%d", s.jenkins.baseURL, jobName, buildNumber)
	s.instanceRepo.UpdateActivityBuildURL(ctx, activityID, buildURL)

	// Wait for build to complete
	buildInfo, err := s.jenkins.WaitForBuildCompletion(ctx, jobName, buildNumber, s.config.BuildTimeout)
	if err != nil {
		s.instanceRepo.UpdateActivityStatus(ctx, activityID, models.ActivityStatusError, err.Error())
		s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatusError)
		return
	}

	// Check build result
	if buildInfo.Result == "SUCCESS" {
		s.instanceRepo.UpdateActivityStatus(ctx, activityID, models.ActivityStatusSuccess, "")
		s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatusStartingUp)

		// Update order to deployed
		instance, _ := s.instanceRepo.GetByID(ctx, instanceID)
		if instance != nil && instance.SalesOrderID != "" {
			s.orderRepo.UpdateOrderStatus(ctx, instance.SalesOrderID, models.OrderStatusDeployed)
		}

		// Send instance ready email
		s.sendInstanceReadyEmail(ctx, instanceID)
	} else {
		errMsg := fmt.Sprintf("build failed with result: %s", buildInfo.Result)
		s.instanceRepo.UpdateActivityStatus(ctx, activityID, models.ActivityStatusError, errMsg)
		s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatusError)
	}
}

// sendInstanceReadyEmail sends an email notification when an instance is ready.
func (s *Service) sendInstanceReadyEmail(ctx context.Context, instanceID string) {
	if s.emailService == nil || s.userRepo == nil {
		return
	}

	// Get instance
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		log.Printf("Failed to get instance for email: %v", err)
		return
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, instance.UserID)
	if err != nil {
		log.Printf("Failed to get user for email: %v", err)
		return
	}

	// Get product
	product, err := s.productRepo.GetProductByID(ctx, instance.ProductID)
	if err != nil {
		log.Printf("Failed to get product for email: %v", err)
		return
	}

	// Get credentials from Vault
	username := "admin"
	password := ""
	if s.vault != nil && instance.VaultPath != "" {
		creds, err := s.vault.ReadCredentials(ctx, instance.VaultPath)
		if err == nil && creds != nil {
			username = creds.AdminUsername
			password = creds.AdminPassword
		}
	}

	// Send email
	err = s.emailService.SendInstanceReady(
		ctx,
		user.Email,
		user.FirstName,
		instance.Name,
		product.Name,
		instance.URL,
		username,
		password,
	)
	if err != nil {
		log.Printf("Failed to send instance ready email: %v", err)
	}
}

// Delete initiates deletion of an instance.
func (s *Service) Delete(ctx context.Context, instanceID string) error {
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// Update status to deleting
	if err := s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatusDeleting); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Create activity record
	activity := &models.Activity{
		InstanceID:   instanceID,
		ActivityType: models.ActivityTypeDelete,
		Status:       models.ActivityStatusPending,
	}
	if err := s.instanceRepo.CreateActivity(ctx, activity); err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}

	// Get product for job name
	product, err := s.productRepo.GetProductByID(ctx, instance.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// Get cluster
	cluster, err := s.productRepo.GetClusterByID(ctx, instance.ClusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	// Trigger delete job
	jobName := s.getDeleteJobName(product.Slug)
	if jobName == "" {
		// Use generic delete job
		jobName = "cloudbench-delete-instance"
	}

	jenkinsParams := map[string]string{
		"APP_NAME":    instance.Name,
		"CLUSTER":     cluster.Code,
		"PRODUCT":     product.Slug,
		"INSTANCE_ID": instanceID,
		"VAULT_PATH":  instance.VaultPath,
	}

	queueID, err := s.jenkins.TriggerBuild(ctx, TriggerBuildParams{
		JobName:    jobName,
		Parameters: jenkinsParams,
	})
	if err != nil {
		s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusError, err.Error())
		return fmt.Errorf("failed to trigger delete job: %w", err)
	}

	// Update activity
	s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusRunning, "")

	// Monitor deletion
	go s.monitorDeletion(context.Background(), instanceID, activity.ID, jobName, queueID)

	return nil
}

// monitorDeletion monitors a deletion job.
func (s *Service) monitorDeletion(ctx context.Context, instanceID, activityID, jobName string, queueID int) {
	// Wait for build to start
	buildNumber, err := s.jenkins.WaitForBuild(ctx, queueID, s.config.QueueTimeout)
	if err != nil {
		s.instanceRepo.UpdateActivityStatus(ctx, activityID, models.ActivityStatusError, err.Error())
		return
	}

	// Wait for build to complete
	buildInfo, err := s.jenkins.WaitForBuildCompletion(ctx, jobName, buildNumber, s.config.BuildTimeout)
	if err != nil {
		s.instanceRepo.UpdateActivityStatus(ctx, activityID, models.ActivityStatusError, err.Error())
		return
	}

	if buildInfo.Result == "SUCCESS" {
		s.instanceRepo.UpdateActivityStatus(ctx, activityID, models.ActivityStatusSuccess, "")
		s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatusDeleted)

		// Delete credentials from Vault
		instance, _ := s.instanceRepo.GetByID(ctx, instanceID)
		if instance != nil && instance.VaultPath != "" {
			s.vault.DeleteCredentials(ctx, instance.VaultPath)
		}
	} else {
		errMsg := fmt.Sprintf("delete job failed with result: %s", buildInfo.Result)
		s.instanceRepo.UpdateActivityStatus(ctx, activityID, models.ActivityStatusError, errMsg)
	}
}

// Restart initiates a restart of an instance.
func (s *Service) Restart(ctx context.Context, instanceID string) error {
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// Create activity record
	activity := &models.Activity{
		InstanceID:   instanceID,
		ActivityType: models.ActivityTypeRestart,
		Status:       models.ActivityStatusPending,
	}
	if err := s.instanceRepo.CreateActivity(ctx, activity); err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}

	// Get cluster
	cluster, err := s.productRepo.GetClusterByID(ctx, instance.ClusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	// Trigger restart job
	jenkinsParams := map[string]string{
		"APP_NAME":    instance.Name,
		"CLUSTER":     cluster.Code,
		"INSTANCE_ID": instanceID,
	}

	queueID, err := s.jenkins.TriggerBuild(ctx, TriggerBuildParams{
		JobName:    "cloudbench-restart-instance",
		Parameters: jenkinsParams,
	})
	if err != nil {
		s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusError, err.Error())
		return fmt.Errorf("failed to trigger restart job: %w", err)
	}

	s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusRunning, "")

	// Monitor restart (simpler than deploy)
	go func() {
		ctx := context.Background()
		buildNumber, err := s.jenkins.WaitForBuild(ctx, queueID, s.config.QueueTimeout)
		if err != nil {
			s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusError, err.Error())
			return
		}

		buildInfo, err := s.jenkins.WaitForBuildCompletion(ctx, "cloudbench-restart-instance", buildNumber, 10*time.Minute)
		if err != nil {
			s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusError, err.Error())
			return
		}

		if buildInfo.Result == "SUCCESS" {
			s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusSuccess, "")
		} else {
			s.instanceRepo.UpdateActivityStatus(ctx, activity.ID, models.ActivityStatusError, buildInfo.Result)
		}
	}()

	return nil
}

// getJobName returns the Jenkins job name for a product.
func (s *Service) getJobName(productSlug string) string {
	switch productSlug {
	case "geoserver":
		if s.config.GeoServerJobName != "" {
			return s.config.GeoServerJobName
		}
		return "cloudbench-deploy-geoserver"
	case "geonode":
		if s.config.GeoNodeJobName != "" {
			return s.config.GeoNodeJobName
		}
		return "cloudbench-deploy-geonode"
	case "postgis":
		if s.config.PostGISJobName != "" {
			return s.config.PostGISJobName
		}
		return "cloudbench-deploy-postgis"
	default:
		return ""
	}
}

// getDeleteJobName returns the Jenkins delete job name for a product.
func (s *Service) getDeleteJobName(productSlug string) string {
	switch productSlug {
	case "geoserver":
		return "cloudbench-delete-geoserver"
	case "geonode":
		return "cloudbench-delete-geonode"
	case "postgis":
		return "cloudbench-delete-postgis"
	default:
		return "cloudbench-delete-instance"
	}
}

// HandleArgoCDWebhook processes ArgoCD webhook notifications.
func (s *Service) HandleArgoCDWebhook(ctx context.Context, payload ArgoCDWebhook) error {
	// Extract instance ID from annotations or labels
	instanceID := payload.Application.Metadata.Labels["instance-id"]
	if instanceID == "" {
		return fmt.Errorf("no instance-id in webhook payload")
	}

	// Get the instance
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	// Update status based on ArgoCD sync status
	switch payload.Application.Status.Health.Status {
	case "Healthy":
		if instance.Status == models.InstanceStatusStartingUp {
			s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatusOnline)
		}
	case "Degraded":
		s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatusError)
	case "Progressing":
		// Still deploying, no action needed
	}

	return nil
}

// ArgoCDWebhook represents an ArgoCD webhook payload.
type ArgoCDWebhook struct {
	Application struct {
		Metadata struct {
			Name   string            `json:"name"`
			Labels map[string]string `json:"labels"`
		} `json:"metadata"`
		Status struct {
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
			Sync struct {
				Status string `json:"status"`
			} `json:"sync"`
		} `json:"status"`
	} `json:"application"`
}
