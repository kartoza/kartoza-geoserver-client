package email

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// ConsoleProvider implements the Provider interface by logging to console.
// This is useful for development and testing.
type ConsoleProvider struct{}

// NewConsoleProvider creates a new console email provider.
func NewConsoleProvider() *ConsoleProvider {
	return &ConsoleProvider{}
}

// Name returns the provider name.
func (p *ConsoleProvider) Name() string {
	return "console"
}

// Send logs the email to console instead of sending it.
func (p *ConsoleProvider) Send(ctx context.Context, msg *Message) error {
	log.Printf(`
================================================================================
EMAIL (Console Provider - Not Actually Sent)
================================================================================
From:    %s <%s>
To:      %s
Subject: %s
Reply-To: %s
--------------------------------------------------------------------------------
%s
================================================================================
`,
		msg.FromName,
		msg.From,
		strings.Join(msg.To, ", "),
		msg.Subject,
		msg.ReplyTo,
		truncateBody(msg.HTMLBody, 500),
	)

	return nil
}

// truncateBody truncates the body for console output.
func truncateBody(body string, maxLen int) string {
	// Strip HTML tags for cleaner console output
	stripped := stripHTMLTags(body)

	if len(stripped) <= maxLen {
		return stripped
	}
	return stripped[:maxLen] + "...\n[truncated]"
}

// stripHTMLTags removes HTML tags from a string for cleaner console output.
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false

	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}

	// Clean up multiple newlines and whitespace
	lines := strings.Split(result.String(), "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, "\n")
}

// MockProvider implements the Provider interface for testing.
type MockProvider struct {
	SentMessages []*Message
	ShouldFail   bool
	FailError    error
}

// NewMockProvider creates a new mock email provider for testing.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		SentMessages: make([]*Message, 0),
	}
}

// Name returns the provider name.
func (p *MockProvider) Name() string {
	return "mock"
}

// Send records the message or returns an error if ShouldFail is true.
func (p *MockProvider) Send(ctx context.Context, msg *Message) error {
	if p.ShouldFail {
		if p.FailError != nil {
			return p.FailError
		}
		return fmt.Errorf("mock send failure")
	}

	p.SentMessages = append(p.SentMessages, msg)
	return nil
}

// Reset clears all sent messages.
func (p *MockProvider) Reset() {
	p.SentMessages = make([]*Message, 0)
	p.ShouldFail = false
	p.FailError = nil
}

// LastMessage returns the last sent message, or nil if none.
func (p *MockProvider) LastMessage() *Message {
	if len(p.SentMessages) == 0 {
		return nil
	}
	return p.SentMessages[len(p.SentMessages)-1]
}

// MessageCount returns the number of sent messages.
func (p *MockProvider) MessageCount() int {
	return len(p.SentMessages)
}
