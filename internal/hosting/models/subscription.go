package models

import (
	"time"
)

// SubscriptionStatus represents the status of a subscription.
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusUnpaid   SubscriptionStatus = "unpaid"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
)

// Subscription represents a recurring payment subscription.
type Subscription struct {
	ID                     string             `json:"id"`
	UserID                 string             `json:"user_id"`
	InstanceID             string             `json:"instance_id"`
	PackageID              string             `json:"package_id"`
	StripeSubscriptionID   string             `json:"stripe_subscription_id,omitempty"`
	StripeCustomerID       string             `json:"stripe_customer_id,omitempty"`
	PaystackSubscriptionID string             `json:"paystack_subscription_id,omitempty"`
	PaystackCustomerID     string             `json:"paystack_customer_id,omitempty"`
	BillingCycle           BillingCycle       `json:"billing_cycle"`
	Status                 SubscriptionStatus `json:"status"`
	CurrentPeriodStart     time.Time          `json:"current_period_start"`
	CurrentPeriodEnd       time.Time          `json:"current_period_end"`
	CancelAtPeriodEnd      bool               `json:"cancel_at_period_end"`
	CancelledAt            *time.Time         `json:"cancelled_at,omitempty"`
	TrialStart             *time.Time         `json:"trial_start,omitempty"`
	TrialEnd               *time.Time         `json:"trial_end,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at"`

	// Related data (populated by queries)
	User     *User     `json:"user,omitempty"`
	Instance *Instance `json:"instance,omitempty"`
	Package  *Package  `json:"package,omitempty"`
}

// IsActive returns true if the subscription is currently active.
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrialing
}

// DaysRemaining returns the number of days until the current period ends.
func (s *Subscription) DaysRemaining() int {
	if s.CurrentPeriodEnd.IsZero() {
		return 0
	}
	remaining := time.Until(s.CurrentPeriodEnd)
	if remaining < 0 {
		return 0
	}
	return int(remaining.Hours() / 24)
}

// IsExpiringSoon returns true if subscription expires within the given days.
func (s *Subscription) IsExpiringSoon(days int) bool {
	return s.DaysRemaining() <= days && s.DaysRemaining() > 0
}

// IsTrialing returns true if the subscription is in trial period.
func (s *Subscription) IsTrialing() bool {
	if s.TrialEnd == nil {
		return false
	}
	return time.Now().Before(*s.TrialEnd)
}

// TrialDaysRemaining returns the number of days remaining in the trial.
func (s *Subscription) TrialDaysRemaining() int {
	if s.TrialEnd == nil {
		return 0
	}
	remaining := time.Until(*s.TrialEnd)
	if remaining < 0 {
		return 0
	}
	return int(remaining.Hours() / 24)
}

// CancelSubscriptionRequest represents a request to cancel a subscription.
type CancelSubscriptionRequest struct {
	CancelImmediately bool   `json:"cancel_immediately"`
	Reason            string `json:"reason,omitempty"`
}

// SubscriptionSummary provides a summary view of a subscription.
type SubscriptionSummary struct {
	ID                string             `json:"id"`
	InstanceName      string             `json:"instance_name"`
	ProductName       string             `json:"product_name"`
	PackageName       string             `json:"package_name"`
	Status            SubscriptionStatus `json:"status"`
	BillingCycle      BillingCycle       `json:"billing_cycle"`
	CurrentPeriodEnd  time.Time          `json:"current_period_end"`
	CancelAtPeriodEnd bool               `json:"cancel_at_period_end"`
}
