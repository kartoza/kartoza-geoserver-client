package models

import (
	"time"
)

// DiscountType represents the type of discount.
type DiscountType string

const (
	DiscountTypePercent DiscountType = "percent"
	DiscountTypeFixed   DiscountType = "fixed"
)

// DurationType represents how long the discount applies.
type DurationType string

const (
	DurationTypeOnce      DurationType = "once"
	DurationTypeRepeating DurationType = "repeating"
	DurationTypeForever   DurationType = "forever"
)

// CouponGroup represents a group of coupons with shared discount rules.
type CouponGroup struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	Description       string       `json:"description,omitempty"`
	DiscountType      DiscountType `json:"discount_type"`
	DiscountPercent   *float64     `json:"discount_percent,omitempty"`
	DiscountAmount    *float64     `json:"discount_amount,omitempty"`
	Currency          string       `json:"currency"`
	DurationType      DurationType `json:"duration_type"`
	DurationMonths    *int         `json:"duration_months,omitempty"`
	AppliesToProducts []string     `json:"applies_to_products,omitempty"`
	MinOrderAmount    *float64     `json:"min_order_amount,omitempty"`
	ValidFrom         *time.Time   `json:"valid_from,omitempty"`
	ValidUntil        *time.Time   `json:"valid_until,omitempty"`
	MaxUses           *int         `json:"max_uses,omitempty"`
	MaxUsesPerUser    int          `json:"max_uses_per_user"`
	CurrentUses       int          `json:"current_uses"`
	IsActive          bool         `json:"is_active"`
	StripeCouponID    string       `json:"stripe_coupon_id,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
}

// IsValid checks if the coupon group is currently valid.
func (g *CouponGroup) IsValid() bool {
	if !g.IsActive {
		return false
	}
	now := time.Now()
	if g.ValidFrom != nil && now.Before(*g.ValidFrom) {
		return false
	}
	if g.ValidUntil != nil && now.After(*g.ValidUntil) {
		return false
	}
	if g.MaxUses != nil && g.CurrentUses >= *g.MaxUses {
		return false
	}
	return true
}

// CalculateDiscount calculates the discount for a given amount.
func (g *CouponGroup) CalculateDiscount(amount float64) float64 {
	switch g.DiscountType {
	case DiscountTypePercent:
		if g.DiscountPercent != nil {
			return amount * (*g.DiscountPercent / 100)
		}
	case DiscountTypeFixed:
		if g.DiscountAmount != nil {
			if *g.DiscountAmount > amount {
				return amount
			}
			return *g.DiscountAmount
		}
	}
	return 0
}

// Coupon represents an individual coupon code.
type Coupon struct {
	ID             string     `json:"id"`
	GroupID        string     `json:"group_id"`
	Code           string     `json:"code"`
	UsedByUserID   *string    `json:"used_by_user_id,omitempty"`
	UsedForOrderID *string    `json:"used_for_order_id,omitempty"`
	UsedAt         *time.Time `json:"used_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`

	// Related data
	Group *CouponGroup `json:"group,omitempty"`
}

// IsUsed returns true if the coupon has been used.
func (c *Coupon) IsUsed() bool {
	return c.UsedAt != nil
}

// IsAvailable returns true if the coupon can be used.
func (c *Coupon) IsAvailable() bool {
	return c.IsActive && !c.IsUsed()
}

// ValidateCouponRequest represents a request to validate a coupon.
type ValidateCouponRequest struct {
	Code      string `json:"code"`
	ProductID string `json:"product_id,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
}

// ValidateCouponResponse represents the result of coupon validation.
type ValidateCouponResponse struct {
	Valid           bool     `json:"valid"`
	Message         string   `json:"message,omitempty"`
	DiscountType    string   `json:"discount_type,omitempty"`
	DiscountPercent *float64 `json:"discount_percent,omitempty"`
	DiscountAmount  *float64 `json:"discount_amount,omitempty"`
	CouponID        string   `json:"coupon_id,omitempty"`
}

// CreateCouponGroupRequest represents a request to create a coupon group.
type CreateCouponGroupRequest struct {
	Name              string       `json:"name"`
	Description       string       `json:"description,omitempty"`
	DiscountType      DiscountType `json:"discount_type"`
	DiscountPercent   *float64     `json:"discount_percent,omitempty"`
	DiscountAmount    *float64     `json:"discount_amount,omitempty"`
	Currency          string       `json:"currency"`
	DurationType      DurationType `json:"duration_type"`
	DurationMonths    *int         `json:"duration_months,omitempty"`
	AppliesToProducts []string     `json:"applies_to_products,omitempty"`
	MinOrderAmount    *float64     `json:"min_order_amount,omitempty"`
	ValidFrom         *time.Time   `json:"valid_from,omitempty"`
	ValidUntil        *time.Time   `json:"valid_until,omitempty"`
	MaxUses           *int         `json:"max_uses,omitempty"`
	MaxUsesPerUser    int          `json:"max_uses_per_user"`
	CodeCount         int          `json:"code_count"` // Number of codes to generate
	CodePrefix        string       `json:"code_prefix,omitempty"`
}
