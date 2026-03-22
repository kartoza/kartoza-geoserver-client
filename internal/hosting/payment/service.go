// Package payment provides payment processing functionality.
package payment

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrPaymentFailed     = errors.New("payment failed")
	ErrInvalidWebhook    = errors.New("invalid webhook signature")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidProvider   = errors.New("invalid payment provider")
	ErrProviderNotSetup  = errors.New("payment provider not configured")
	ErrSubscriptionError = errors.New("subscription error")
)

// Provider represents a payment provider.
type Provider interface {
	// Name returns the provider name
	Name() string

	// CreateCheckoutSession creates a new checkout session
	CreateCheckoutSession(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)

	// VerifyWebhook verifies a webhook signature and returns the event
	VerifyWebhook(payload []byte, signature string) (*WebhookEvent, error)

	// GetPaymentStatus gets the status of a payment
	GetPaymentStatus(ctx context.Context, sessionID string) (*PaymentStatus, error)

	// CreateSubscription creates a new subscription
	CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error)

	// CancelSubscription cancels a subscription
	CancelSubscription(ctx context.Context, subscriptionID string, immediate bool) error

	// CreateCustomer creates a customer in the payment provider
	CreateCustomer(ctx context.Context, email, name string) (string, error)

	// IsConfigured returns true if the provider is properly configured
	IsConfigured() bool
}

// CheckoutRequest represents a checkout session request.
type CheckoutRequest struct {
	OrderID      string
	CustomerID   string
	CustomerEmail string
	PriceID      string
	Quantity     int
	Mode         string // "payment" or "subscription"
	SuccessURL   string
	CancelURL    string
	Metadata     map[string]string
	TrialDays    int
	CouponID     string
}

// CheckoutResponse represents a checkout session response.
type CheckoutResponse struct {
	SessionID   string
	CheckoutURL string
	Reference   string // For Paystack
}

// WebhookEvent represents a webhook event from a payment provider.
type WebhookEvent struct {
	Type            string
	SessionID       string
	PaymentIntentID string
	SubscriptionID  string
	CustomerID      string
	Status          string
	Metadata        map[string]string
	Amount          int64
	Currency        string
}

// PaymentStatus represents the status of a payment.
type PaymentStatus struct {
	Status          string
	PaymentIntentID string
	Amount          int64
	Currency        string
	Paid            bool
}

// SubscriptionRequest represents a subscription creation request.
type SubscriptionRequest struct {
	CustomerID     string
	PriceID        string
	TrialDays      int
	CouponID       string
	Metadata       map[string]string
	PaymentBehavior string // "default_incomplete" or "allow_incomplete"
}

// SubscriptionResponse represents a subscription response.
type SubscriptionResponse struct {
	SubscriptionID     string
	Status             string
	CurrentPeriodStart int64
	CurrentPeriodEnd   int64
	TrialEnd           *int64
}

// Config holds payment configuration.
type Config struct {
	StripeSecretKey       string
	StripeWebhookSecret   string
	StripePublishableKey  string
	PaystackSecretKey     string
	PaystackPublicKey     string
	PaystackWebhookSecret string
	DefaultCurrency       string
}

// Service manages payment processing.
type Service struct {
	stripe   Provider
	paystack Provider
	config   Config
}

// NewService creates a new payment service.
func NewService(config Config) *Service {
	s := &Service{
		config: config,
	}

	// Initialize Stripe if configured
	if config.StripeSecretKey != "" {
		s.stripe = NewStripeProvider(config.StripeSecretKey, config.StripeWebhookSecret)
	}

	// Initialize Paystack if configured
	if config.PaystackSecretKey != "" {
		s.paystack = NewPaystackProvider(config.PaystackSecretKey, config.PaystackWebhookSecret)
	}

	return s
}

// GetProvider returns the specified payment provider.
func (s *Service) GetProvider(name string) (Provider, error) {
	switch name {
	case "stripe":
		if s.stripe == nil || !s.stripe.IsConfigured() {
			return nil, ErrProviderNotSetup
		}
		return s.stripe, nil
	case "paystack":
		if s.paystack == nil || !s.paystack.IsConfigured() {
			return nil, ErrProviderNotSetup
		}
		return s.paystack, nil
	default:
		return nil, ErrInvalidProvider
	}
}

// AvailableProviders returns a list of configured payment providers.
func (s *Service) AvailableProviders() []string {
	providers := []string{}
	if s.stripe != nil && s.stripe.IsConfigured() {
		providers = append(providers, "stripe")
	}
	if s.paystack != nil && s.paystack.IsConfigured() {
		providers = append(providers, "paystack")
	}
	return providers
}

// StripePublishableKey returns the Stripe publishable key for frontend use.
func (s *Service) StripePublishableKey() string {
	return s.config.StripePublishableKey
}

// PaystackPublicKey returns the Paystack public key for frontend use.
func (s *Service) PaystackPublicKey() string {
	return s.config.PaystackPublicKey
}
