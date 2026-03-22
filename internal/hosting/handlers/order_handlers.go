package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/auth"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/email"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/payment"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/repository"
)

// OrderHandler handles order-related HTTP requests.
type OrderHandler struct {
	orderRepo    *repository.OrderRepository
	productRepo  *repository.ProductRepository
	userRepo     *repository.UserRepository
	paymentSvc   *payment.Service
	emailService *email.Service
}

// NewOrderHandler creates a new order handler.
func NewOrderHandler(
	orderRepo *repository.OrderRepository,
	productRepo *repository.ProductRepository,
	userRepo *repository.UserRepository,
	paymentSvc *payment.Service,
) *OrderHandler {
	return &OrderHandler{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
		paymentSvc:  paymentSvc,
	}
}

// SetEmailService sets the email service for notifications.
func (h *OrderHandler) SetEmailService(emailService *email.Service) {
	h.emailService = emailService
}

// HandleCreateOrder handles POST /api/v1/orders
func (h *OrderHandler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate package exists
	pkg, err := h.productRepo.GetPackageByID(r.Context(), req.PackageID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Package not found", http.StatusBadRequest)
			return
		}
		jsonError(w, "Failed to get package", http.StatusInternalServerError)
		return
	}

	if !pkg.IsAvailable {
		jsonError(w, "Package is not available", http.StatusBadRequest)
		return
	}

	// Validate cluster exists
	cluster, err := h.productRepo.GetClusterByID(r.Context(), req.ClusterID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Cluster not found", http.StatusBadRequest)
			return
		}
		jsonError(w, "Failed to get cluster", http.StatusInternalServerError)
		return
	}

	if !cluster.IsActive || !cluster.HasCapacity() {
		jsonError(w, "Cluster is not available", http.StatusBadRequest)
		return
	}

	// Validate app name
	if req.AppName == "" {
		jsonError(w, "App name is required", http.StatusBadRequest)
		return
	}

	// Check if app name is unique in cluster
	exists, err := h.checkAppNameExists(r.Context(), req.ClusterID, req.AppName)
	if err != nil {
		jsonError(w, "Failed to validate app name", http.StatusInternalServerError)
		return
	}
	if exists {
		jsonError(w, "App name already exists in this cluster", http.StatusConflict)
		return
	}

	// Calculate pricing
	var subtotal float64
	if req.BillingCycle == models.BillingYearly {
		subtotal = pkg.PriceYearly
	} else {
		subtotal = pkg.PriceMonthly
	}

	// TODO: Handle coupon validation and discount calculation

	order := &models.SalesOrder{
		UserID:         userID,
		PackageID:      req.PackageID,
		ClusterID:      req.ClusterID,
		AppName:        req.AppName,
		BillingCycle:   req.BillingCycle,
		Status:         models.OrderStatusPending,
		SubtotalAmount: subtotal,
		DiscountAmount: 0,
		TaxAmount:      0,
		TotalAmount:    subtotal,
		Currency:       "USD",
	}

	if err := h.orderRepo.CreateOrder(r.Context(), order); err != nil {
		jsonError(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Include related data in response
	order.Package = pkg
	order.Cluster = cluster

	jsonResponse(w, order, http.StatusCreated)
}

// HandleGetOrder handles GET /api/v1/orders/{id}
func (h *OrderHandler) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract order ID from path
	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	orderID = strings.Split(orderID, "/")[0]
	if orderID == "" {
		jsonError(w, "Order ID required", http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.GetOrderByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Order not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	// Verify ownership (unless admin)
	if order.UserID != userID && !auth.IsAdmin(r.Context()) {
		jsonError(w, "Order not found", http.StatusNotFound)
		return
	}

	// Populate related data
	if pkg, err := h.productRepo.GetPackageByID(r.Context(), order.PackageID); err == nil {
		order.Package = pkg
		if product, err := h.productRepo.GetProductByID(r.Context(), pkg.ProductID); err == nil {
			order.Package.Product = product
		}
	}
	if cluster, err := h.productRepo.GetClusterByID(r.Context(), order.ClusterID); err == nil {
		order.Cluster = cluster
	}

	jsonResponse(w, order, http.StatusOK)
}

// HandleListOrders handles GET /api/v1/orders
func (h *OrderHandler) HandleListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse pagination
	limit := 20
	offset := 0
	// TODO: Parse from query params

	summaries, total, err := h.orderRepo.GetOrderSummaries(r.Context(), userID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"orders": summaries,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}, http.StatusOK)
}

// HandleCheckout handles POST /api/v1/orders/{id}/checkout/{provider}
func (h *OrderHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract order ID and provider from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "checkout" {
		jsonError(w, "Invalid path", http.StatusBadRequest)
		return
	}
	orderID := parts[0]
	providerName := parts[2]

	// Get order
	order, err := h.orderRepo.GetOrderByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Order not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	// Verify ownership
	if order.UserID != userID {
		jsonError(w, "Order not found", http.StatusNotFound)
		return
	}

	// Check order status
	if !order.IsPending() {
		jsonError(w, "Order is not pending", http.StatusBadRequest)
		return
	}

	// Get user for email
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		jsonError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Get package for price ID
	pkg, err := h.productRepo.GetPackageByID(r.Context(), order.PackageID)
	if err != nil {
		jsonError(w, "Failed to get package", http.StatusInternalServerError)
		return
	}

	// Parse request
	var req models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get payment provider
	provider, err := h.paymentSvc.GetProvider(providerName)
	if err != nil {
		jsonError(w, "Payment provider not available", http.StatusBadRequest)
		return
	}

	// Get price ID based on billing cycle
	var priceID string
	if order.BillingCycle == models.BillingYearly {
		if providerName == "stripe" {
			priceID = pkg.StripePriceYearlyID
		} else {
			priceID = pkg.PaystackPlanYearlyID
		}
	} else {
		if providerName == "stripe" {
			priceID = pkg.StripePriceMonthlyID
		} else {
			priceID = pkg.PaystackPlanMonthlyID
		}
	}

	if priceID == "" {
		jsonError(w, "Price not configured for this package", http.StatusBadRequest)
		return
	}

	// Create checkout session
	checkoutReq := payment.CheckoutRequest{
		OrderID:       order.ID,
		CustomerEmail: user.Email,
		PriceID:       priceID,
		Quantity:      1,
		Mode:          "subscription",
		SuccessURL:    req.SuccessURL + "?session_id={CHECKOUT_SESSION_ID}",
		CancelURL:     req.CancelURL,
		Metadata: map[string]string{
			"order_id": order.ID,
			"user_id":  userID,
		},
	}

	checkoutResp, err := provider.CreateCheckoutSession(r.Context(), checkoutReq)
	if err != nil {
		jsonError(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	// Update order with payment info
	_ = h.orderRepo.UpdateOrderPayment(r.Context(), order.ID,
		providerName, "", checkoutResp.SessionID, "", checkoutResp.Reference)
	_ = h.orderRepo.UpdateOrderStatus(r.Context(), order.ID, models.OrderStatusPaymentPending)

	jsonResponse(w, models.CheckoutResponse{
		CheckoutURL: checkoutResp.CheckoutURL,
		SessionID:   checkoutResp.SessionID,
		Reference:   checkoutResp.Reference,
	}, http.StatusOK)
}

// HandleCancelOrder handles POST /api/v1/orders/{id}/cancel
func (h *OrderHandler) HandleCancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract order ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	parts := strings.Split(path, "/")
	orderID := parts[0]

	order, err := h.orderRepo.GetOrderByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Order not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	// Verify ownership
	if order.UserID != userID && !auth.IsAdmin(r.Context()) {
		jsonError(w, "Order not found", http.StatusNotFound)
		return
	}

	if !order.CanBeCancelled() {
		jsonError(w, "Order cannot be cancelled", http.StatusBadRequest)
		return
	}

	if err := h.orderRepo.UpdateOrderStatus(r.Context(), order.ID, models.OrderStatusCancelled); err != nil {
		jsonError(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "Order cancelled"}, http.StatusOK)
}

// HandleStripeWebhook handles POST /api/v1/webhooks/stripe
func (h *OrderHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("Stripe-Signature")

	provider, err := h.paymentSvc.GetProvider("stripe")
	if err != nil {
		jsonError(w, "Stripe not configured", http.StatusInternalServerError)
		return
	}

	event, err := provider.VerifyWebhook(payload, signature)
	if err != nil {
		jsonError(w, "Invalid webhook signature", http.StatusBadRequest)
		return
	}

	// Process the event
	if err := h.processPaymentEvent(r.Context(), event); err != nil {
		// Log the error but return success to Stripe
		// to prevent retries for application-level errors
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandlePaystackWebhook handles POST /api/v1/webhooks/paystack
func (h *OrderHandler) HandlePaystackWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("X-Paystack-Signature")

	provider, err := h.paymentSvc.GetProvider("paystack")
	if err != nil {
		jsonError(w, "Paystack not configured", http.StatusInternalServerError)
		return
	}

	event, err := provider.VerifyWebhook(payload, signature)
	if err != nil {
		jsonError(w, "Invalid webhook signature", http.StatusBadRequest)
		return
	}

	// Process the event
	if err := h.processPaymentEvent(r.Context(), event); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandlePaymentConfig returns payment provider configuration for frontend
func (h *OrderHandler) HandlePaymentConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"providers":              h.paymentSvc.AvailableProviders(),
		"stripe_publishable_key": h.paymentSvc.StripePublishableKey(),
		"paystack_public_key":    h.paymentSvc.PaystackPublicKey(),
	}, http.StatusOK)
}

// processPaymentEvent processes a payment webhook event.
func (h *OrderHandler) processPaymentEvent(ctx context.Context, event *payment.WebhookEvent) error {
	switch event.Type {
	case "checkout.session.completed", "charge.success":
		// Get order from metadata or session
		orderID := event.Metadata["order_id"]
		if orderID == "" && event.SessionID != "" {
			order, err := h.orderRepo.GetOrderByStripeSession(ctx, event.SessionID)
			if err == nil {
				orderID = order.ID
			}
		}

		if orderID != "" {
			// Update order status
			if err := h.orderRepo.UpdateOrderStatus(ctx, orderID, models.OrderStatusPaid); err != nil {
				return err
			}

			// Update payment info
			_ = h.orderRepo.UpdateOrderPayment(ctx, orderID,
				"", event.PaymentIntentID, "", event.PaymentIntentID, "")

			// Send order confirmation email
			h.sendOrderConfirmationEmail(ctx, orderID)

			// TODO: Trigger deployment
		}

	case "customer.subscription.deleted":
		// Handle subscription cancellation
		// TODO: Update subscription status, potentially mark instance for deletion

	case "invoice.payment_failed":
		// Handle failed payment
		// TODO: Update subscription status, send notification
	}

	return nil
}

// sendOrderConfirmationEmail sends an order confirmation email.
func (h *OrderHandler) sendOrderConfirmationEmail(ctx context.Context, orderID string) {
	if h.emailService == nil {
		return
	}

	// Get order details
	order, err := h.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		log.Printf("Failed to get order for email: %v", err)
		return
	}

	// Get user
	user, err := h.userRepo.GetByID(ctx, order.UserID)
	if err != nil {
		log.Printf("Failed to get user for email: %v", err)
		return
	}

	// Get product and package
	pkg, err := h.productRepo.GetPackageByID(ctx, order.PackageID)
	if err != nil {
		log.Printf("Failed to get package for email: %v", err)
		return
	}

	product, err := h.productRepo.GetProductByID(ctx, pkg.ProductID)
	if err != nil {
		log.Printf("Failed to get product for email: %v", err)
		return
	}

	// Format total amount
	totalAmount := fmt.Sprintf("$%.2f", float64(order.TotalAmount)/100)

	// Determine billing cycle
	billingCycle := "Monthly"
	if order.BillingCycle == "yearly" {
		billingCycle = "Yearly"
	}

	// Send email
	err = h.emailService.SendOrderConfirmation(
		ctx,
		user.Email,
		user.FirstName,
		order.ID,
		product.Name,
		pkg.Name,
		billingCycle,
		totalAmount,
		order.AppName,
	)
	if err != nil {
		log.Printf("Failed to send order confirmation email: %v", err)
	}
}

// checkAppNameExists checks if an app name already exists in a cluster.
func (h *OrderHandler) checkAppNameExists(ctx context.Context, clusterID, appName string) (bool, error) {
	// This would typically check the instances table
	// For now, return false
	return false, nil
}

// Context type for avoiding import cycles
type contextKey string
