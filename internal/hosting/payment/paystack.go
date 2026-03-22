package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const paystackBaseURL = "https://api.paystack.co"

// PaystackProvider implements the Provider interface for Paystack.
type PaystackProvider struct {
	secretKey     string
	webhookSecret string
	httpClient    *http.Client
}

// NewPaystackProvider creates a new Paystack provider.
func NewPaystackProvider(secretKey, webhookSecret string) *PaystackProvider {
	return &PaystackProvider{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the provider name.
func (p *PaystackProvider) Name() string {
	return "paystack"
}

// IsConfigured returns true if the provider is properly configured.
func (p *PaystackProvider) IsConfigured() bool {
	return p.secretKey != ""
}

// CreateCheckoutSession creates a new Paystack transaction initialization.
func (p *PaystackProvider) CreateCheckoutSession(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error) {
	// Paystack uses kobo (1/100 of Naira) or cents for USD
	payload := map[string]interface{}{
		"email":        req.CustomerEmail,
		"callback_url": req.SuccessURL,
		"metadata":     req.Metadata,
	}

	// Set amount if it's a one-time payment
	if req.Mode != "subscription" {
		// Note: amount would need to be passed differently for Paystack
		// This is a simplified implementation
		payload["callback_url"] = req.SuccessURL
	} else {
		// For subscriptions, use a plan
		if req.PriceID != "" {
			payload["plan"] = req.PriceID
		}
	}

	resp, err := p.doRequest(ctx, "POST", "/transaction/initialize", payload)
	if err != nil {
		return nil, err
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response from Paystack")
	}

	return &CheckoutResponse{
		Reference:   data["reference"].(string),
		CheckoutURL: data["authorization_url"].(string),
	}, nil
}

// VerifyWebhook verifies a Paystack webhook signature.
func (p *PaystackProvider) VerifyWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	// Compute HMAC
	mac := hmac.New(sha512.New, []byte(p.secretKey))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedSig), []byte(signature)) {
		return nil, ErrInvalidWebhook
	}

	// Parse the event
	var event struct {
		Event string                 `json:"event"`
		Data  map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	we := &WebhookEvent{
		Type: event.Event,
	}

	// Parse based on event type
	switch event.Event {
	case "charge.success":
		if ref, ok := event.Data["reference"].(string); ok {
			we.SessionID = ref
		}
		if customer, ok := event.Data["customer"].(map[string]interface{}); ok {
			if id, ok := customer["customer_code"].(string); ok {
				we.CustomerID = id
			}
		}
		if amount, ok := event.Data["amount"].(float64); ok {
			we.Amount = int64(amount)
		}
		if currency, ok := event.Data["currency"].(string); ok {
			we.Currency = currency
		}
		we.Status = "success"

	case "subscription.create", "subscription.disable":
		if code, ok := event.Data["subscription_code"].(string); ok {
			we.SubscriptionID = code
		}
		if customer, ok := event.Data["customer"].(map[string]interface{}); ok {
			if id, ok := customer["customer_code"].(string); ok {
				we.CustomerID = id
			}
		}
		if status, ok := event.Data["status"].(string); ok {
			we.Status = status
		}
	}

	if meta, ok := event.Data["metadata"].(map[string]interface{}); ok {
		we.Metadata = make(map[string]string)
		for k, v := range meta {
			if s, ok := v.(string); ok {
				we.Metadata[k] = s
			}
		}
	}

	return we, nil
}

// GetPaymentStatus gets the status of a Paystack transaction.
func (p *PaystackProvider) GetPaymentStatus(ctx context.Context, reference string) (*PaymentStatus, error) {
	resp, err := p.doRequest(ctx, "GET", "/transaction/verify/"+reference, nil)
	if err != nil {
		return nil, err
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response from Paystack")
	}

	status := &PaymentStatus{
		Status: data["status"].(string),
		Paid:   data["status"].(string) == "success",
	}

	if amount, ok := data["amount"].(float64); ok {
		status.Amount = int64(amount)
	}
	if currency, ok := data["currency"].(string); ok {
		status.Currency = currency
	}

	return status, nil
}

// CreateSubscription creates a new Paystack subscription.
func (p *PaystackProvider) CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error) {
	payload := map[string]interface{}{
		"customer": req.CustomerID,
		"plan":     req.PriceID,
	}

	resp, err := p.doRequest(ctx, "POST", "/subscription", payload)
	if err != nil {
		return nil, err
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response from Paystack")
	}

	return &SubscriptionResponse{
		SubscriptionID: data["subscription_code"].(string),
		Status:         data["status"].(string),
	}, nil
}

// CancelSubscription cancels a Paystack subscription.
func (p *PaystackProvider) CancelSubscription(ctx context.Context, subscriptionID string, immediate bool) error {
	payload := map[string]interface{}{
		"code":  subscriptionID,
		"token": subscriptionID, // Paystack uses the email token
	}

	_, err := p.doRequest(ctx, "POST", "/subscription/disable", payload)
	return err
}

// CreateCustomer creates a Paystack customer.
func (p *PaystackProvider) CreateCustomer(ctx context.Context, email, name string) (string, error) {
	payload := map[string]interface{}{
		"email": email,
	}
	if name != "" {
		// Split name into first and last
		payload["first_name"] = name
	}

	resp, err := p.doRequest(ctx, "POST", "/customer", payload)
	if err != nil {
		return "", err
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response from Paystack")
	}

	return data["customer_code"].(string), nil
}

// doRequest performs an HTTP request to the Paystack API.
func (p *PaystackProvider) doRequest(ctx context.Context, method, path string, payload interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, paystackBaseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		message := "request failed"
		if msg, ok := result["message"].(string); ok {
			message = msg
		}
		return nil, fmt.Errorf("Paystack error: %s", message)
	}

	return result, nil
}
