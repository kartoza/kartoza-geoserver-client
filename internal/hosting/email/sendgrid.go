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

// SendGridProvider implements the Provider interface using SendGrid.
type SendGridProvider struct {
	apiKey     string
	httpClient *http.Client
}

// NewSendGridProvider creates a new SendGrid email provider.
func NewSendGridProvider(apiKey string) *SendGridProvider {
	return &SendGridProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the provider name.
func (p *SendGridProvider) Name() string {
	return "sendgrid"
}

// sendGridRequest represents the SendGrid API request structure.
type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	ReplyTo          *sendGridEmail            `json:"reply_to,omitempty"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
	Attachments      []sendGridAttachment      `json:"attachments,omitempty"`
}

type sendGridPersonalization struct {
	To []sendGridEmail `json:"to"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sendGridAttachment struct {
	Content     string `json:"content"`
	Type        string `json:"type"`
	Filename    string `json:"filename"`
	Disposition string `json:"disposition"`
}

// Send sends an email using SendGrid.
func (p *SendGridProvider) Send(ctx context.Context, msg *Message) error {
	if p.apiKey == "" {
		return fmt.Errorf("SendGrid API key not configured")
	}

	// Build recipients
	recipients := make([]sendGridEmail, len(msg.To))
	for i, to := range msg.To {
		recipients[i] = sendGridEmail{Email: to}
	}

	// Build content
	var content []sendGridContent
	if msg.TextBody != "" {
		content = append(content, sendGridContent{
			Type:  "text/plain",
			Value: msg.TextBody,
		})
	}
	if msg.HTMLBody != "" {
		content = append(content, sendGridContent{
			Type:  "text/html",
			Value: msg.HTMLBody,
		})
	}

	// Build request
	req := sendGridRequest{
		Personalizations: []sendGridPersonalization{
			{To: recipients},
		},
		From: sendGridEmail{
			Email: msg.From,
			Name:  msg.FromName,
		},
		Subject: msg.Subject,
		Content: content,
	}

	if msg.ReplyTo != "" {
		req.ReplyTo = &sendGridEmail{Email: msg.ReplyTo}
	}

	// Add attachments
	for _, att := range msg.Attachments {
		req.Attachments = append(req.Attachments, sendGridAttachment{
			Content:     string(att.Content),
			Type:        att.ContentType,
			Filename:    att.Filename,
			Disposition: "attachment",
		})
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal SendGrid request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create SendGrid request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send SendGrid request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SendGrid API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
