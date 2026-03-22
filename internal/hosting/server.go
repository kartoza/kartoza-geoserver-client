// Package hosting provides the web server for the hosting platform.
package hosting

import (
	"log"
	"net/http"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/admin"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/auth"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/db"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/deploy"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/email"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/handlers"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/health"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/payment"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/repository"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/vault"
)

// Config holds configuration for the hosting server.
type Config struct {
	DatabaseURL string
	JWTSecret   string

	// Stripe configuration
	StripeSecretKey     string
	StripeWebhookSecret string
	StripePublicKey     string

	// Paystack configuration
	PaystackSecretKey string
	PaystackPublicKey string

	// Vault configuration
	VaultAddr  string
	VaultToken string

	// Jenkins configuration
	JenkinsURL   string
	JenkinsUser  string
	JenkinsToken string

	// Email configuration
	EmailProvider string // sendgrid, resend, smtp, console
	EmailAPIKey   string
	EmailFrom     string
	EmailFromName string
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string

	// Base URL for links in emails
	BaseURL string

	// Base domain for instance URLs
	BaseDomain string
}

// Server represents the hosting platform server.
type Server struct {
	db             *db.DB
	authService    *auth.Service
	authMiddleware *auth.Middleware
	paymentService *payment.Service
	deployService  *deploy.Service
	emailService   *email.Service
	vaultClient    *vault.Client
	healthChecker  *health.Checker

	// Repositories
	userRepo     *repository.UserRepository
	productRepo  *repository.ProductRepository
	orderRepo    *repository.OrderRepository
	instanceRepo *repository.InstanceRepository

	// Handlers
	authHandler     *handlers.AuthHandler
	productHandler  *handlers.ProductHandler
	orderHandler    *handlers.OrderHandler
	instanceHandler *handlers.InstanceHandler
	adminHandler    *admin.Handler
	adminService    *admin.Service
}

// NewServer creates a new hosting server.
func NewServer(cfg Config) (*Server, error) {
	// Connect to database
	database, err := db.NewFromDSN(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Create repositories
	userRepo := repository.NewUserRepository(database)
	productRepo := repository.NewProductRepository(database)
	orderRepo := repository.NewOrderRepository(database)
	instanceRepo := repository.NewInstanceRepository(database)

	// Create JWT config
	jwtConfig := auth.JWTConfig{
		SecretKey:     cfg.JWTSecret,
		Issuer:        "kartoza-cloudbench",
		AccessExpiry:  24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	if cfg.JWTSecret == "" {
		jwtConfig = auth.DefaultJWTConfig()
	}

	// Create services
	authService := auth.NewService(userRepo, jwtConfig)
	authMiddleware := auth.NewMiddleware(authService)

	// Create email service
	emailConfig := email.Config{
		Provider:       cfg.EmailProvider,
		FromEmail:      cfg.EmailFrom,
		FromName:       cfg.EmailFromName,
		BaseURL:        cfg.BaseURL,
		SendGridAPIKey: cfg.EmailAPIKey,
		ResendAPIKey:   cfg.EmailAPIKey,
		SMTPHost:       cfg.SMTPHost,
		SMTPPort:       cfg.SMTPPort,
		SMTPUsername:   cfg.SMTPUsername,
		SMTPPassword:   cfg.SMTPPassword,
	}
	if emailConfig.Provider == "" {
		emailConfig.Provider = "console"
	}
	emailService, err := email.NewService(emailConfig)
	if err != nil {
		log.Printf("Warning: Failed to create email service: %v", err)
	}

	// Create payment service
	paymentConfig := payment.Config{
		StripeSecretKey:      cfg.StripeSecretKey,
		StripeWebhookSecret:  cfg.StripeWebhookSecret,
		StripePublishableKey: cfg.StripePublicKey,
		PaystackSecretKey:    cfg.PaystackSecretKey,
		PaystackPublicKey:    cfg.PaystackPublicKey,
		DefaultCurrency:      "USD",
	}
	paymentService := payment.NewService(paymentConfig)

	// Create vault client
	var vaultClient *vault.Client
	if cfg.VaultAddr != "" && cfg.VaultToken != "" {
		vaultClient = vault.NewClient(vault.Config{
			Address: cfg.VaultAddr,
			Token:   cfg.VaultToken,
		})
	}

	// Create deploy service
	deployConfig := deploy.ServiceConfig{
		GeoServerJobName: "deploy-geoserver",
		GeoNodeJobName:   "deploy-geonode",
		PostGISJobName:   "deploy-postgis",
		QueueTimeout:     5 * time.Minute,
		BuildTimeout:     30 * time.Minute,
		BaseDomain:       cfg.BaseDomain,
	}
	deployService := deploy.NewService(
		nil, // Jenkins client (optional, configured separately)
		vaultClient,
		instanceRepo,
		orderRepo,
		productRepo,
		deployConfig,
	)
	if emailService != nil {
		deployService.SetEmailService(emailService)
		deployService.SetUserRepository(userRepo)
	}

	// Create health checker
	healthConfig := health.Config{
		Interval: 5 * time.Minute,
		Timeout:  10 * time.Second,
	}
	healthChecker := health.NewChecker(instanceRepo, healthConfig)

	// Create admin service
	adminService := admin.NewService(userRepo, instanceRepo, orderRepo, productRepo)

	// Create handlers
	var authHandler *handlers.AuthHandler
	if emailService != nil {
		authHandler = handlers.NewAuthHandlerWithEmail(authService, emailService)
	} else {
		authHandler = handlers.NewAuthHandler(authService)
	}
	productHandler := handlers.NewProductHandler(productRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, productRepo, userRepo, paymentService)
	if emailService != nil {
		orderHandler.SetEmailService(emailService)
	}
	instanceHandler := handlers.NewInstanceHandler(instanceRepo, productRepo, deployService, healthChecker, vaultClient)
	adminHandler := admin.NewHandler(adminService)

	return &Server{
		db:              database,
		authService:     authService,
		authMiddleware:  authMiddleware,
		paymentService:  paymentService,
		deployService:   deployService,
		emailService:    emailService,
		vaultClient:     vaultClient,
		healthChecker:   healthChecker,
		userRepo:        userRepo,
		productRepo:     productRepo,
		orderRepo:       orderRepo,
		instanceRepo:    instanceRepo,
		authHandler:     authHandler,
		productHandler:  productHandler,
		orderHandler:    orderHandler,
		instanceHandler: instanceHandler,
		adminHandler:    adminHandler,
		adminService:    adminService,
	}, nil
}

// SetupRoutes registers all hosting API routes on the given mux.
func (s *Server) SetupRoutes(mux *http.ServeMux) {
	// Auth routes (public)
	mux.HandleFunc("/api/v1/auth/register", s.authHandler.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", s.authHandler.HandleLogin)
	mux.HandleFunc("/api/v1/auth/logout", s.authMiddleware.RequireAuthFunc(s.authHandler.HandleLogout))
	mux.HandleFunc("/api/v1/auth/refresh", s.authHandler.HandleRefresh)
	mux.HandleFunc("/api/v1/auth/profile", s.authMiddleware.RequireAuthFunc(s.authHandler.HandleProfile))
	mux.HandleFunc("/api/v1/auth/reset-password", s.authHandler.HandleRequestPasswordReset)
	mux.HandleFunc("/api/v1/auth/reset-confirm", s.authHandler.HandleConfirmPasswordReset)

	// Product routes (public)
	mux.HandleFunc("/api/v1/products", s.productHandler.HandleProducts)
	mux.HandleFunc("/api/v1/products/", s.productHandler.HandleProductBySlug)
	mux.HandleFunc("/api/v1/packages/", s.productHandler.HandlePackageByID)
	mux.HandleFunc("/api/v1/clusters", s.productHandler.HandleClusters)

	// Order routes (protected)
	mux.HandleFunc("/api/v1/orders", s.authMiddleware.RequireAuthFunc(s.handleOrders))
	mux.HandleFunc("/api/v1/orders/", s.authMiddleware.RequireAuthFunc(s.handleOrderByID))

	// Instance routes (protected)
	mux.HandleFunc("/api/v1/instances", s.authMiddleware.RequireAuthFunc(s.instanceHandler.HandleListInstances))
	mux.HandleFunc("/api/v1/instances/", s.authMiddleware.RequireAuthFunc(s.handleInstanceByID))

	// Payment webhooks (public but verified by signature)
	mux.HandleFunc("/api/v1/webhooks/stripe", s.orderHandler.HandleStripeWebhook)
	mux.HandleFunc("/api/v1/webhooks/paystack", s.orderHandler.HandlePaystackWebhook)

	// Admin routes (protected + admin only)
	mux.HandleFunc("/api/v1/admin/dashboard", s.authMiddleware.RequireAdminFunc(s.adminHandler.HandleDashboardStats))
	mux.HandleFunc("/api/v1/admin/users", s.authMiddleware.RequireAdminFunc(s.adminHandler.HandleListUsers))
	mux.HandleFunc("/api/v1/admin/users/", s.authMiddleware.RequireAdminFunc(s.handleAdminUser))
	mux.HandleFunc("/api/v1/admin/instances", s.authMiddleware.RequireAdminFunc(s.adminHandler.HandleListInstances))
	mux.HandleFunc("/api/v1/admin/orders", s.authMiddleware.RequireAdminFunc(s.adminHandler.HandleListOrders))
	mux.HandleFunc("/api/v1/admin/analytics/revenue", s.authMiddleware.RequireAdminFunc(s.adminHandler.HandleRevenueChart))
	mux.HandleFunc("/api/v1/admin/health", s.authMiddleware.RequireAdminFunc(s.adminHandler.HandleSystemHealth))
}

// handleAdminUser routes between GET and PUT for individual users.
func (s *Server) handleAdminUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.adminHandler.HandleGetUser(w, r)
	case http.MethodPut:
		s.adminHandler.HandleUpdateUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleOrders routes between GET and POST for orders.
func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.orderHandler.HandleListOrders(w, r)
	case http.MethodPost:
		s.orderHandler.HandleCreateOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleOrderByID routes operations on individual orders.
func (s *Server) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.orderHandler.HandleGetOrder(w, r)
	case http.MethodDelete:
		s.orderHandler.HandleCancelOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleInstanceByID routes operations on individual instances.
func (s *Server) handleInstanceByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.instanceHandler.HandleGetInstance(w, r)
	case http.MethodDelete:
		s.instanceHandler.HandleDeleteInstance(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Close closes the server and its resources.
func (s *Server) Close() error {
	if s.healthChecker != nil {
		s.healthChecker.Stop()
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
