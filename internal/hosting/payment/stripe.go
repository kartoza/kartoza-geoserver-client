package payment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

// StripeProvider implements the Provider interface for Stripe.
type StripeProvider struct {
	secretKey     string
	webhookSecret string
}

// NewStripeProvider creates a new Stripe provider.
func NewStripeProvider(secretKey, webhookSecret string) *StripeProvider {
	stripe.Key = secretKey
	return &StripeProvider{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
	}
}

// Name returns the provider name.
func (p *StripeProvider) Name() string {
	return "stripe"
}

// IsConfigured returns true if the provider is properly configured.
func (p *StripeProvider) IsConfigured() bool {
	return p.secretKey != ""
}

// CreateCheckoutSession creates a new Stripe checkout session.
func (p *StripeProvider) CreateCheckoutSession(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error) {
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(int64(req.Quantity)),
			},
		},
	}

	// Set mode
	switch req.Mode {
	case "subscription":
		params.Mode = stripe.String(string(stripe.CheckoutSessionModeSubscription))
		if req.TrialDays > 0 {
			params.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
				TrialPeriodDays: stripe.Int64(int64(req.TrialDays)),
			}
		}
	default:
		params.Mode = stripe.String(string(stripe.CheckoutSessionModePayment))
	}

	// Set customer
	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	} else if req.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(req.CustomerEmail)
	}

	// Set metadata
	if len(req.Metadata) > 0 {
		params.Metadata = req.Metadata
	}

	// Apply coupon
	if req.CouponID != "" {
		params.Discounts = []*stripe.CheckoutSessionDiscountParams{
			{
				Coupon: stripe.String(req.CouponID),
			},
		}
	}

	s, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &CheckoutResponse{
		SessionID:   s.ID,
		CheckoutURL: s.URL,
	}, nil
}

// VerifyWebhook verifies a Stripe webhook signature.
func (p *StripeProvider) VerifyWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, p.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidWebhook, err)
	}

	we := &WebhookEvent{
		Type: string(event.Type),
	}

	// Parse event data based on type
	switch event.Type {
	case "checkout.session.completed":
		var cs stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &cs); err == nil {
			we.SessionID = cs.ID
			if cs.Customer != nil {
				we.CustomerID = cs.Customer.ID
			}
			we.Status = string(cs.Status)
			we.Metadata = cs.Metadata
			if cs.PaymentIntent != nil {
				we.PaymentIntentID = cs.PaymentIntent.ID
			}
			if cs.Subscription != nil {
				we.SubscriptionID = cs.Subscription.ID
			}
			if cs.AmountTotal != 0 {
				we.Amount = cs.AmountTotal
				we.Currency = string(cs.Currency)
			}
		}

	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err == nil {
			we.SubscriptionID = sub.ID
			if sub.Customer != nil {
				we.CustomerID = sub.Customer.ID
			}
			we.Status = string(sub.Status)
			we.Metadata = sub.Metadata
		}

	case "invoice.paid", "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err == nil {
			if inv.Customer != nil {
				we.CustomerID = inv.Customer.ID
			}
			we.Amount = inv.AmountPaid
			we.Currency = string(inv.Currency)
			// Get subscription from line items
			if inv.Lines != nil && len(inv.Lines.Data) > 0 {
				lineItem := inv.Lines.Data[0]
				if lineItem.Subscription != nil {
					we.SubscriptionID = lineItem.Subscription.ID
				}
			}
		}
	}

	return we, nil
}

// GetPaymentStatus gets the status of a Stripe checkout session.
func (p *StripeProvider) GetPaymentStatus(ctx context.Context, sessionID string) (*PaymentStatus, error) {
	s, err := session.Get(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	status := &PaymentStatus{
		Status:   string(s.Status),
		Currency: string(s.Currency),
		Paid:     s.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid,
	}

	if s.PaymentIntent != nil {
		status.PaymentIntentID = s.PaymentIntent.ID
	}
	if s.AmountTotal != 0 {
		status.Amount = s.AmountTotal
	}

	return status, nil
}

// CreateSubscription creates a new Stripe subscription.
func (p *StripeProvider) CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(req.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(req.PriceID),
			},
		},
	}

	if req.TrialDays > 0 {
		params.TrialPeriodDays = stripe.Int64(int64(req.TrialDays))
	}

	// Apply coupon via discount
	if req.CouponID != "" {
		params.AddExpand("latest_invoice")
	}

	if len(req.Metadata) > 0 {
		params.Metadata = req.Metadata
	}

	if req.PaymentBehavior != "" {
		params.PaymentBehavior = stripe.String(req.PaymentBehavior)
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	resp := &SubscriptionResponse{
		SubscriptionID: sub.ID,
		Status:         string(sub.Status),
	}

	// Get period from subscription items (v82 API moved these fields to items)
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		resp.CurrentPeriodStart = item.CurrentPeriodStart
		resp.CurrentPeriodEnd = item.CurrentPeriodEnd
	}

	if sub.TrialEnd != 0 {
		resp.TrialEnd = &sub.TrialEnd
	}

	return resp, nil
}

// CancelSubscription cancels a Stripe subscription.
func (p *StripeProvider) CancelSubscription(ctx context.Context, subscriptionID string, immediate bool) error {
	if immediate {
		_, err := subscription.Cancel(subscriptionID, nil)
		return err
	}

	_, err := subscription.Update(subscriptionID, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	})
	return err
}

// CreateCustomer creates a Stripe customer.
func (p *StripeProvider) CreateCustomer(ctx context.Context, email, name string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}
	if name != "" {
		params.Name = stripe.String(name)
	}

	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	return c.ID, nil
}
