package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/db"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
)

// OrderRepository provides data access for orders and subscriptions.
type OrderRepository struct {
	db *db.DB
}

// AdminOrderListOptions specifies options for listing orders in admin.
type AdminOrderListOptions struct {
	Page     int
	PageSize int
	Status   string
	UserID   string
	DateFrom *time.Time
	DateTo   *time.Time
}

// AdminOrderListItem represents an order in admin listings.
type AdminOrderListItem struct {
	ID            string
	UserID        string
	UserEmail     string
	ProductName   string
	PackageName   string
	Status        models.OrderStatus
	TotalAmount   int64
	Currency      string
	PaymentMethod string
	CreatedAt     time.Time
	PaidAt        *time.Time
}

// OrderStats represents aggregate order statistics.
type OrderStats struct {
	TotalOrders      int
	PendingOrders    int
	MonthlyRevenue   float64
	TotalRevenue     float64
	RevenueByProduct map[string]float64
}

// RevenueDataPoint represents a revenue data point for charts.
type RevenueDataPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

// NewOrderRepository creates a new order repository.
func NewOrderRepository(database *db.DB) *OrderRepository {
	return &OrderRepository{db: database}
}

// CreateOrder creates a new sales order.
func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.SalesOrder) error {
	query := `
		INSERT INTO sales_orders (
			user_id, package_id, cluster_id, app_name, billing_cycle, status,
			subtotal_amount, discount_amount, tax_amount, total_amount, currency,
			coupon_id, payment_method, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		order.UserID, order.PackageID, order.ClusterID, order.AppName,
		order.BillingCycle, order.Status, order.SubtotalAmount, order.DiscountAmount,
		order.TaxAmount, order.TotalAmount, order.Currency, order.CouponID,
		nullString(order.PaymentMethod), nullString(order.Notes),
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
}

// GetOrderByID retrieves an order by ID.
func (r *OrderRepository) GetOrderByID(ctx context.Context, id string) (*models.SalesOrder, error) {
	query := `
		SELECT id, user_id, package_id, cluster_id, app_name, billing_cycle, status,
		       subtotal_amount, discount_amount, tax_amount, total_amount, currency,
		       coupon_id, payment_method, payment_id, stripe_session_id,
		       stripe_payment_intent_id, paystack_reference, notes, created_at, updated_at
		FROM sales_orders WHERE id = $1
	`
	order := &models.SalesOrder{}
	var couponID, paymentMethod, paymentID, stripeSession, stripeIntent, paystackRef, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.UserID, &order.PackageID, &order.ClusterID, &order.AppName,
		&order.BillingCycle, &order.Status, &order.SubtotalAmount, &order.DiscountAmount,
		&order.TaxAmount, &order.TotalAmount, &order.Currency, &couponID,
		&paymentMethod, &paymentID, &stripeSession, &stripeIntent, &paystackRef,
		&notes, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if couponID.Valid {
		order.CouponID = &couponID.String
	}
	order.PaymentMethod = paymentMethod.String
	order.PaymentID = paymentID.String
	order.StripeSessionID = stripeSession.String
	order.StripePaymentIntentID = stripeIntent.String
	order.PaystackReference = paystackRef.String
	order.Notes = notes.String

	return order, nil
}

// GetOrderByStripeSession retrieves an order by Stripe session ID.
func (r *OrderRepository) GetOrderByStripeSession(ctx context.Context, sessionID string) (*models.SalesOrder, error) {
	query := `
		SELECT id, user_id, package_id, cluster_id, app_name, billing_cycle, status,
		       subtotal_amount, discount_amount, tax_amount, total_amount, currency,
		       coupon_id, payment_method, payment_id, stripe_session_id,
		       stripe_payment_intent_id, paystack_reference, notes, created_at, updated_at
		FROM sales_orders WHERE stripe_session_id = $1
	`
	order := &models.SalesOrder{}
	var couponID, paymentMethod, paymentID, stripeSession, stripeIntent, paystackRef, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&order.ID, &order.UserID, &order.PackageID, &order.ClusterID, &order.AppName,
		&order.BillingCycle, &order.Status, &order.SubtotalAmount, &order.DiscountAmount,
		&order.TaxAmount, &order.TotalAmount, &order.Currency, &couponID,
		&paymentMethod, &paymentID, &stripeSession, &stripeIntent, &paystackRef,
		&notes, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if couponID.Valid {
		order.CouponID = &couponID.String
	}
	order.PaymentMethod = paymentMethod.String
	order.PaymentID = paymentID.String
	order.StripeSessionID = stripeSession.String
	order.StripePaymentIntentID = stripeIntent.String
	order.PaystackReference = paystackRef.String
	order.Notes = notes.String

	return order, nil
}

// UpdateOrderStatus updates the status of an order.
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id string, status models.OrderStatus) error {
	query := `UPDATE sales_orders SET status = $2 WHERE id = $1`
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

// UpdateOrderPayment updates payment information for an order.
func (r *OrderRepository) UpdateOrderPayment(ctx context.Context, id string, paymentMethod, paymentID, stripeSession, stripeIntent, paystackRef string) error {
	query := `
		UPDATE sales_orders SET
			payment_method = COALESCE($2, payment_method),
			payment_id = COALESCE($3, payment_id),
			stripe_session_id = COALESCE($4, stripe_session_id),
			stripe_payment_intent_id = COALESCE($5, stripe_payment_intent_id),
			paystack_reference = COALESCE($6, paystack_reference)
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id,
		nullString(paymentMethod), nullString(paymentID),
		nullString(stripeSession), nullString(stripeIntent), nullString(paystackRef),
	)
	return err
}

// ListUserOrders retrieves orders for a user.
func (r *OrderRepository) ListUserOrders(ctx context.Context, userID string, limit, offset int) ([]*models.SalesOrder, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM sales_orders WHERE user_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, package_id, cluster_id, app_name, billing_cycle, status,
		       subtotal_amount, discount_amount, tax_amount, total_amount, currency,
		       coupon_id, payment_method, payment_id, stripe_session_id,
		       stripe_payment_intent_id, paystack_reference, notes, created_at, updated_at
		FROM sales_orders WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*models.SalesOrder
	for rows.Next() {
		order := &models.SalesOrder{}
		var couponID, paymentMethod, paymentID, stripeSession, stripeIntent, paystackRef, notes sql.NullString

		if err := rows.Scan(
			&order.ID, &order.UserID, &order.PackageID, &order.ClusterID, &order.AppName,
			&order.BillingCycle, &order.Status, &order.SubtotalAmount, &order.DiscountAmount,
			&order.TaxAmount, &order.TotalAmount, &order.Currency, &couponID,
			&paymentMethod, &paymentID, &stripeSession, &stripeIntent, &paystackRef,
			&notes, &order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if couponID.Valid {
			order.CouponID = &couponID.String
		}
		order.PaymentMethod = paymentMethod.String
		order.PaymentID = paymentID.String
		order.StripeSessionID = stripeSession.String
		order.StripePaymentIntentID = stripeIntent.String
		order.PaystackReference = paystackRef.String
		order.Notes = notes.String
		orders = append(orders, order)
	}

	return orders, total, rows.Err()
}

// CreateSubscription creates a new subscription.
func (r *OrderRepository) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			user_id, instance_id, package_id, stripe_subscription_id, stripe_customer_id,
			paystack_subscription_id, paystack_customer_id, billing_cycle, status,
			current_period_start, current_period_end, trial_start, trial_end
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		sub.UserID, sub.InstanceID, sub.PackageID,
		nullString(sub.StripeSubscriptionID), nullString(sub.StripeCustomerID),
		nullString(sub.PaystackSubscriptionID), nullString(sub.PaystackCustomerID),
		sub.BillingCycle, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.TrialStart, sub.TrialEnd,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
}

// GetSubscriptionByID retrieves a subscription by ID.
func (r *OrderRepository) GetSubscriptionByID(ctx context.Context, id string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, instance_id, package_id, stripe_subscription_id, stripe_customer_id,
		       paystack_subscription_id, paystack_customer_id, billing_cycle, status,
		       current_period_start, current_period_end, cancel_at_period_end, cancelled_at,
		       trial_start, trial_end, created_at, updated_at
		FROM subscriptions WHERE id = $1
	`
	sub := &models.Subscription{}
	var stripeSub, stripeCust, paystackSub, paystackCust sql.NullString
	var cancelledAt, trialStart, trialEnd sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID, &sub.UserID, &sub.InstanceID, &sub.PackageID,
		&stripeSub, &stripeCust, &paystackSub, &paystackCust,
		&sub.BillingCycle, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &cancelledAt, &trialStart, &trialEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	sub.StripeSubscriptionID = stripeSub.String
	sub.StripeCustomerID = stripeCust.String
	sub.PaystackSubscriptionID = paystackSub.String
	sub.PaystackCustomerID = paystackCust.String
	if cancelledAt.Valid {
		sub.CancelledAt = &cancelledAt.Time
	}
	if trialStart.Valid {
		sub.TrialStart = &trialStart.Time
	}
	if trialEnd.Valid {
		sub.TrialEnd = &trialEnd.Time
	}

	return sub, nil
}

// GetSubscriptionByStripeID retrieves a subscription by Stripe subscription ID.
func (r *OrderRepository) GetSubscriptionByStripeID(ctx context.Context, stripeID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, instance_id, package_id, stripe_subscription_id, stripe_customer_id,
		       paystack_subscription_id, paystack_customer_id, billing_cycle, status,
		       current_period_start, current_period_end, cancel_at_period_end, cancelled_at,
		       trial_start, trial_end, created_at, updated_at
		FROM subscriptions WHERE stripe_subscription_id = $1
	`
	sub := &models.Subscription{}
	var stripeSub, stripeCust, paystackSub, paystackCust sql.NullString
	var cancelledAt, trialStart, trialEnd sql.NullTime

	err := r.db.QueryRowContext(ctx, query, stripeID).Scan(
		&sub.ID, &sub.UserID, &sub.InstanceID, &sub.PackageID,
		&stripeSub, &stripeCust, &paystackSub, &paystackCust,
		&sub.BillingCycle, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &cancelledAt, &trialStart, &trialEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	sub.StripeSubscriptionID = stripeSub.String
	sub.StripeCustomerID = stripeCust.String
	sub.PaystackSubscriptionID = paystackSub.String
	sub.PaystackCustomerID = paystackCust.String
	if cancelledAt.Valid {
		sub.CancelledAt = &cancelledAt.Time
	}
	if trialStart.Valid {
		sub.TrialStart = &trialStart.Time
	}
	if trialEnd.Valid {
		sub.TrialEnd = &trialEnd.Time
	}

	return sub, nil
}

// UpdateSubscriptionStatus updates the status of a subscription.
func (r *OrderRepository) UpdateSubscriptionStatus(ctx context.Context, id string, status models.SubscriptionStatus) error {
	query := `UPDATE subscriptions SET status = $2 WHERE id = $1`
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

// UpdateSubscriptionPeriod updates the billing period of a subscription.
func (r *OrderRepository) UpdateSubscriptionPeriod(ctx context.Context, id string, start, end time.Time) error {
	query := `
		UPDATE subscriptions SET
			current_period_start = $2,
			current_period_end = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, start, end)
	return err
}

// CancelSubscription marks a subscription for cancellation.
func (r *OrderRepository) CancelSubscription(ctx context.Context, id string, immediate bool) error {
	now := time.Now()
	if immediate {
		query := `
			UPDATE subscriptions SET
				status = 'cancelled',
				cancelled_at = $2,
				cancel_at_period_end = false
			WHERE id = $1
		`
		_, err := r.db.ExecContext(ctx, query, id, now)
		return err
	}

	query := `
		UPDATE subscriptions SET
			cancel_at_period_end = true,
			cancelled_at = $2
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, now)
	return err
}

// ListUserSubscriptions retrieves subscriptions for a user.
func (r *OrderRepository) ListUserSubscriptions(ctx context.Context, userID string) ([]*models.Subscription, error) {
	query := `
		SELECT id, user_id, instance_id, package_id, stripe_subscription_id, stripe_customer_id,
		       paystack_subscription_id, paystack_customer_id, billing_cycle, status,
		       current_period_start, current_period_end, cancel_at_period_end, cancelled_at,
		       trial_start, trial_end, created_at, updated_at
		FROM subscriptions WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.Subscription
	for rows.Next() {
		sub := &models.Subscription{}
		var stripeSub, stripeCust, paystackSub, paystackCust sql.NullString
		var cancelledAt, trialStart, trialEnd sql.NullTime

		if err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.InstanceID, &sub.PackageID,
			&stripeSub, &stripeCust, &paystackSub, &paystackCust,
			&sub.BillingCycle, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
			&sub.CancelAtPeriodEnd, &cancelledAt, &trialStart, &trialEnd,
			&sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, err
		}

		sub.StripeSubscriptionID = stripeSub.String
		sub.StripeCustomerID = stripeCust.String
		sub.PaystackSubscriptionID = paystackSub.String
		sub.PaystackCustomerID = paystackCust.String
		if cancelledAt.Valid {
			sub.CancelledAt = &cancelledAt.Time
		}
		if trialStart.Valid {
			sub.TrialStart = &trialStart.Time
		}
		if trialEnd.Valid {
			sub.TrialEnd = &trialEnd.Time
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

// GetSubscriptionByInstance retrieves a subscription for an instance.
func (r *OrderRepository) GetSubscriptionByInstance(ctx context.Context, instanceID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, instance_id, package_id, stripe_subscription_id, stripe_customer_id,
		       paystack_subscription_id, paystack_customer_id, billing_cycle, status,
		       current_period_start, current_period_end, cancel_at_period_end, cancelled_at,
		       trial_start, trial_end, created_at, updated_at
		FROM subscriptions WHERE instance_id = $1
	`
	sub := &models.Subscription{}
	var stripeSub, stripeCust, paystackSub, paystackCust sql.NullString
	var cancelledAt, trialStart, trialEnd sql.NullTime

	err := r.db.QueryRowContext(ctx, query, instanceID).Scan(
		&sub.ID, &sub.UserID, &sub.InstanceID, &sub.PackageID,
		&stripeSub, &stripeCust, &paystackSub, &paystackCust,
		&sub.BillingCycle, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &cancelledAt, &trialStart, &trialEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	sub.StripeSubscriptionID = stripeSub.String
	sub.StripeCustomerID = stripeCust.String
	sub.PaystackSubscriptionID = paystackSub.String
	sub.PaystackCustomerID = paystackCust.String
	if cancelledAt.Valid {
		sub.CancelledAt = &cancelledAt.Time
	}
	if trialStart.Valid {
		sub.TrialStart = &trialStart.Time
	}
	if trialEnd.Valid {
		sub.TrialEnd = &trialEnd.Time
	}

	return sub, nil
}

// GetExpiringSubscriptions retrieves subscriptions expiring within the given days.
func (r *OrderRepository) GetExpiringSubscriptions(ctx context.Context, withinDays int) ([]*models.Subscription, error) {
	query := `
		SELECT id, user_id, instance_id, package_id, stripe_subscription_id, stripe_customer_id,
		       paystack_subscription_id, paystack_customer_id, billing_cycle, status,
		       current_period_start, current_period_end, cancel_at_period_end, cancelled_at,
		       trial_start, trial_end, created_at, updated_at
		FROM subscriptions
		WHERE status = 'active'
		  AND current_period_end <= $1
		  AND current_period_end > NOW()
		ORDER BY current_period_end
	`
	expiryDate := time.Now().AddDate(0, 0, withinDays)
	rows, err := r.db.QueryContext(ctx, query, expiryDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.Subscription
	for rows.Next() {
		sub := &models.Subscription{}
		var stripeSub, stripeCust, paystackSub, paystackCust sql.NullString
		var cancelledAt, trialStart, trialEnd sql.NullTime

		if err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.InstanceID, &sub.PackageID,
			&stripeSub, &stripeCust, &paystackSub, &paystackCust,
			&sub.BillingCycle, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
			&sub.CancelAtPeriodEnd, &cancelledAt, &trialStart, &trialEnd,
			&sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, err
		}

		sub.StripeSubscriptionID = stripeSub.String
		sub.StripeCustomerID = stripeCust.String
		sub.PaystackSubscriptionID = paystackSub.String
		sub.PaystackCustomerID = paystackCust.String
		if cancelledAt.Valid {
			sub.CancelledAt = &cancelledAt.Time
		}
		if trialStart.Valid {
			sub.TrialStart = &trialStart.Time
		}
		if trialEnd.Valid {
			sub.TrialEnd = &trialEnd.Time
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

// GetOrderSummaries retrieves order summaries with product/package names.
func (r *OrderRepository) GetOrderSummaries(ctx context.Context, userID string, limit, offset int) ([]*models.OrderSummary, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM sales_orders WHERE user_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT so.id, p.name as product_name, pk.name as package_name,
		       so.app_name, so.billing_cycle, so.total_amount, so.currency,
		       so.status, so.created_at
		FROM sales_orders so
		JOIN packages pk ON so.package_id = pk.id
		JOIN products p ON pk.product_id = p.id
		WHERE so.user_id = $1
		ORDER BY so.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var summaries []*models.OrderSummary
	for rows.Next() {
		s := &models.OrderSummary{}
		if err := rows.Scan(
			&s.ID, &s.ProductName, &s.PackageName, &s.AppName,
			&s.BillingCycle, &s.TotalAmount, &s.Currency, &s.Status, &s.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		summaries = append(summaries, s)
	}

	return summaries, total, rows.Err()
}

// GetOrderStats returns aggregate order statistics.
func (r *OrderRepository) GetOrderStats(ctx context.Context) (*OrderStats, error) {
	stats := &OrderStats{
		RevenueByProduct: make(map[string]float64),
	}

	// Basic stats
	basicQuery := `
		SELECT
			COUNT(*) as total_orders,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_orders,
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'paid'), 0) as total_revenue,
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'paid' AND created_at >= date_trunc('month', NOW())), 0) as monthly_revenue
		FROM sales_orders
	`
	err := r.db.QueryRowContext(ctx, basicQuery).Scan(
		&stats.TotalOrders, &stats.PendingOrders, &stats.TotalRevenue, &stats.MonthlyRevenue,
	)
	if err != nil {
		return nil, err
	}

	// Revenue by product
	productQuery := `
		SELECT p.name, COALESCE(SUM(so.total_amount), 0) as revenue
		FROM sales_orders so
		JOIN packages pk ON so.package_id = pk.id
		JOIN products p ON pk.product_id = p.id
		WHERE so.status = 'paid'
		GROUP BY p.name
	`
	rows, err := r.db.QueryContext(ctx, productQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var revenue float64
		if err := rows.Scan(&name, &revenue); err != nil {
			return nil, err
		}
		stats.RevenueByProduct[name] = revenue
	}

	return stats, rows.Err()
}

// ListOrdersAdmin retrieves orders for admin listing with filtering and pagination.
func (r *OrderRepository) ListOrdersAdmin(ctx context.Context, opts AdminOrderListOptions) ([]AdminOrderListItem, int, error) {
	// Build WHERE clause
	conditions := []string{"1=1"}
	args := []interface{}{}
	argNum := 1

	if opts.Status != "" {
		conditions = append(conditions, fmt.Sprintf("so.status = $%d", argNum))
		args = append(args, opts.Status)
		argNum++
	}

	if opts.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("so.user_id = $%d", argNum))
		args = append(args, opts.UserID)
		argNum++
	}

	if opts.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("so.created_at >= $%d", argNum))
		args = append(args, *opts.DateFrom)
		argNum++
	}

	if opts.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("so.created_at <= $%d", argNum))
		args = append(args, *opts.DateTo)
		argNum++
	}

	whereClause := "WHERE " + conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM sales_orders so
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
		SELECT so.id, so.user_id, u.email as user_email,
		       p.name as product_name, pk.name as package_name,
		       so.status, so.total_amount, so.currency,
		       COALESCE(so.payment_method, '') as payment_method,
		       so.created_at
		FROM sales_orders so
		JOIN users u ON so.user_id = u.id
		JOIN packages pk ON so.package_id = pk.id
		JOIN products p ON pk.product_id = p.id
		%s
		ORDER BY so.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, opts.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []AdminOrderListItem
	for rows.Next() {
		var o AdminOrderListItem
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.UserEmail, &o.ProductName, &o.PackageName,
			&o.Status, &o.TotalAmount, &o.Currency, &o.PaymentMethod, &o.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, rows.Err()
}

// GetRevenueChart returns revenue data for charting.
func (r *OrderRepository) GetRevenueChart(ctx context.Context, period string, groupBy string) ([]RevenueDataPoint, error) {
	// Determine date range based on period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "7d":
		startDate = now.AddDate(0, 0, -7)
	case "30d":
		startDate = now.AddDate(0, 0, -30)
	case "90d":
		startDate = now.AddDate(0, 0, -90)
	case "1y":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, 0, -30) // Default to 30 days
	}

	// Determine grouping
	var dateFormat string
	switch groupBy {
	case "day":
		dateFormat = "YYYY-MM-DD"
	case "week":
		dateFormat = "IYYY-IW"
	case "month":
		dateFormat = "YYYY-MM"
	default:
		dateFormat = "YYYY-MM-DD"
	}

	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(created_at, '%s') as date,
			COALESCE(SUM(total_amount), 0) as revenue,
			COUNT(*) as orders
		FROM sales_orders
		WHERE status = 'paid' AND created_at >= $1
		GROUP BY TO_CHAR(created_at, '%s')
		ORDER BY date
	`, dateFormat, dateFormat)

	rows, err := r.db.QueryContext(ctx, query, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataPoints []RevenueDataPoint
	for rows.Next() {
		var dp RevenueDataPoint
		if err := rows.Scan(&dp.Date, &dp.Revenue, &dp.Orders); err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, dp)
	}

	return dataPoints, rows.Err()
}
