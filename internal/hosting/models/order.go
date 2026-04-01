package models

import (
	"time"
)

// BillingCycle represents the billing frequency.
type BillingCycle string

const (
	BillingMonthly BillingCycle = "monthly"
	BillingYearly  BillingCycle = "yearly"
)

// OrderStatus represents the status of a sales order.
type OrderStatus string

const (
	OrderStatusPending        OrderStatus = "pending"
	OrderStatusPaymentPending OrderStatus = "payment_pending"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusDeploying      OrderStatus = "deploying"
	OrderStatusDeployed       OrderStatus = "deployed"
	OrderStatusCancelled      OrderStatus = "cancelled"
	OrderStatusRefunded       OrderStatus = "refunded"
	OrderStatusFailed         OrderStatus = "failed"
)

// SalesOrder represents a customer order for a hosted service.
type SalesOrder struct {
	ID                    string       `json:"id"`
	UserID                string       `json:"user_id"`
	PackageID             string       `json:"package_id"`
	ClusterID             string       `json:"cluster_id"`
	AppName               string       `json:"app_name"`
	BillingCycle          BillingCycle `json:"billing_cycle"`
	Status                OrderStatus  `json:"status"`
	SubtotalAmount        float64      `json:"subtotal_amount"`
	DiscountAmount        float64      `json:"discount_amount"`
	TaxAmount             float64      `json:"tax_amount"`
	TotalAmount           float64      `json:"total_amount"`
	Currency              string       `json:"currency"`
	CouponID              *string      `json:"coupon_id,omitempty"`
	PaymentMethod         string       `json:"payment_method,omitempty"`
	PaymentID             string       `json:"payment_id,omitempty"`
	StripeSessionID       string       `json:"stripe_session_id,omitempty"`
	StripePaymentIntentID string       `json:"stripe_payment_intent_id,omitempty"`
	PaystackReference     string       `json:"paystack_reference,omitempty"`
	Notes                 string       `json:"notes,omitempty"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`

	// Related data (populated by queries)
	User     *User     `json:"user,omitempty"`
	Package  *Package  `json:"package,omitempty"`
	Cluster  *Cluster  `json:"cluster,omitempty"`
	Coupon   *Coupon   `json:"coupon,omitempty"`
	Instance *Instance `json:"instance,omitempty"`
}

// IsPending returns true if the order is awaiting payment.
func (o *SalesOrder) IsPending() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusPaymentPending
}

// CanBeCancelled returns true if the order can be cancelled.
func (o *SalesOrder) CanBeCancelled() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusPaymentPending
}

// CreateOrderRequest represents a request to create a new order.
type CreateOrderRequest struct {
	PackageID    string       `json:"package_id"`
	ClusterID    string       `json:"cluster_id"`
	AppName      string       `json:"app_name"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	CouponCode   string       `json:"coupon_code,omitempty"`
}

// CheckoutRequest represents a request to initiate checkout.
type CheckoutRequest struct {
	PaymentMethod string `json:"payment_method"` // "stripe" or "paystack"
	SuccessURL    string `json:"success_url"`
	CancelURL     string `json:"cancel_url"`
}

// CheckoutResponse represents the response from initiating checkout.
type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id,omitempty"`
	Reference   string `json:"reference,omitempty"`
}

// OrderSummary provides a summary view of an order.
type OrderSummary struct {
	ID           string       `json:"id"`
	ProductName  string       `json:"product_name"`
	PackageName  string       `json:"package_name"`
	AppName      string       `json:"app_name"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	TotalAmount  float64      `json:"total_amount"`
	Currency     string       `json:"currency"`
	Status       OrderStatus  `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
}
