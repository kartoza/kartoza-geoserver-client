package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ResendProvider implements the Provider interface using Resend.
type ResendProvider struct {
	apiKey     string
	httpClient *http.Client
}

// NewResendProvider creates a new Resend email provider.
func NewResendProvider(apiKey string) *ResendProvider {
	return &ResendProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the provider name.
func (p *ResendProvider) Name() string {
	return "resend"
}

// resendRequest represents the Resend API request structure.
type resendRequest struct {
	From        string              `json:"from"`
	To          []string            `json:"to"`
	Subject     string              `json:"subject"`
	HTML        string              `json:"html,omitempty"`
	Text        string              `json:"text,omitempty"`
	ReplyTo     string              `json:"reply_to,omitempty"`
	Attachments []resendAttachment  `json:"attachments,omitempty"`
}

type resendAttachment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// Send sends an email using Resend.
func (p *ResendProvider) Send(ctx context.Context, msg *Message) error {
	if p.apiKey == "" {
		return fmt.Errorf("Resend API key not configured")
	}

	// Build from address
	from := msg.From
	if msg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", msg.FromName, msg.From)
	}

	// Build request
	req := resendRequest{
		From:    from,
		To:      msg.To,
		Subject: msg.Subject,
		HTML:    msg.HTMLBody,
		Text:    msg.TextBody,
		ReplyTo: msg.ReplyTo,
	}

	// Add attachments
	for _, att := range msg.Attachments {
		req.Attachments = append(req.Attachments, resendAttachment{
			Filename: att.Filename,
			Content:  string(att.Content),
		})
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal Resend request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create Resend request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send Resend request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
