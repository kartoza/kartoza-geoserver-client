// Package email provides email notification services for the hosting platform.
package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

// Provider represents an email service provider.
type Provider interface {
	// Send sends an email.
	Send(ctx context.Context, msg *Message) error
	// Name returns the provider name.
	Name() string
}

// Message represents an email message.
type Message struct {
	To          []string
	From        string
	FromName    string
	Subject     string
	HTMLBody    string
	TextBody    string
	ReplyTo     string
	Attachments []Attachment
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// Service handles email notifications.
type Service struct {
	provider    Provider
	fromEmail   string
	fromName    string
	baseURL     string
	templates   *template.Template
}

// Config holds email service configuration.
type Config struct {
	Provider    string // "sendgrid", "resend", "smtp"
	FromEmail   string
	FromName    string
	BaseURL     string // Base URL for links in emails

	// SendGrid settings
	SendGridAPIKey string

	// Resend settings
	ResendAPIKey string

	// SMTP settings (fallback)
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

// NewService creates a new email service.
func NewService(cfg Config) (*Service, error) {
	var provider Provider
	var err error

	switch cfg.Provider {
	case "sendgrid":
		provider = NewSendGridProvider(cfg.SendGridAPIKey)
	case "resend":
		provider = NewResendProvider(cfg.ResendAPIKey)
	case "smtp":
		provider, err = NewSMTPProvider(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to create SMTP provider: %w", err)
		}
	case "console":
		provider = NewConsoleProvider()
	default:
		// Default to console provider for development
		provider = NewConsoleProvider()
	}

	// Load templates
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	return &Service{
		provider:  provider,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
		baseURL:   cfg.BaseURL,
		templates: tmpl,
	}, nil
}

// renderTemplate renders an email template with the given data.
func (s *Service) renderTemplate(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}
	return buf.String(), nil
}

// WelcomeData holds data for the welcome email template.
type WelcomeData struct {
	FirstName   string
	Email       string
	LoginURL    string
	SupportURL  string
	CurrentYear int
}

// SendWelcome sends a welcome email to a new user.
func (s *Service) SendWelcome(ctx context.Context, email, firstName string) error {
	data := WelcomeData{
		FirstName:   firstName,
		Email:       email,
		LoginURL:    s.baseURL + "/login",
		SupportURL:  s.baseURL + "/support",
		CurrentYear: time.Now().Year(),
	}

	html, err := s.renderTemplate("welcome.html", data)
	if err != nil {
		return err
	}

	msg := &Message{
		To:       []string{email},
		From:     s.fromEmail,
		FromName: s.fromName,
		Subject:  "Welcome to Kartoza Geospatial Hosting!",
		HTMLBody: html,
	}

	return s.provider.Send(ctx, msg)
}

// InstanceReadyData holds data for the instance ready email template.
type InstanceReadyData struct {
	FirstName      string
	InstanceName   string
	ProductName    string
	InstanceURL    string
	Username       string
	Password       string
	DashboardURL   string
	DocumentionURL string
	SupportURL     string
	CurrentYear    int
}

// SendInstanceReady sends an email when an instance is ready.
func (s *Service) SendInstanceReady(ctx context.Context, email, firstName, instanceName, productName, instanceURL, username, password string) error {
	data := InstanceReadyData{
		FirstName:      firstName,
		InstanceName:   instanceName,
		ProductName:    productName,
		InstanceURL:    instanceURL,
		Username:       username,
		Password:       password,
		DashboardURL:   s.baseURL + "/dashboard",
		DocumentionURL: s.baseURL + "/docs",
		SupportURL:     s.baseURL + "/support",
		CurrentYear:    time.Now().Year(),
	}

	html, err := s.renderTemplate("instance_ready.html", data)
	if err != nil {
		return err
	}

	msg := &Message{
		To:       []string{email},
		From:     s.fromEmail,
		FromName: s.fromName,
		Subject:  fmt.Sprintf("Your %s Instance is Ready!", productName),
		HTMLBody: html,
	}

	return s.provider.Send(ctx, msg)
}

// PaymentReminderData holds data for the payment reminder email template.
type PaymentReminderData struct {
	FirstName       string
	InstanceName    string
	ProductName     string
	ExpirationDate  string
	DaysRemaining   int
	RenewalURL      string
	BillingURL      string
	SupportURL      string
	CurrentYear     int
}

// SendPaymentReminder sends a payment reminder email.
func (s *Service) SendPaymentReminder(ctx context.Context, email, firstName, instanceName, productName string, expirationDate time.Time, daysRemaining int) error {
	data := PaymentReminderData{
		FirstName:      firstName,
		InstanceName:   instanceName,
		ProductName:    productName,
		ExpirationDate: expirationDate.Format("January 2, 2006"),
		DaysRemaining:  daysRemaining,
		RenewalURL:     s.baseURL + "/billing/renew",
		BillingURL:     s.baseURL + "/billing",
		SupportURL:     s.baseURL + "/support",
		CurrentYear:    time.Now().Year(),
	}

	html, err := s.renderTemplate("payment_reminder.html", data)
	if err != nil {
		return err
	}

	var subject string
	if daysRemaining <= 1 {
		subject = fmt.Sprintf("URGENT: Your %s subscription expires tomorrow!", productName)
	} else if daysRemaining <= 7 {
		subject = fmt.Sprintf("Reminder: Your %s subscription expires in %d days", productName, daysRemaining)
	} else {
		subject = fmt.Sprintf("Your %s subscription renews soon", productName)
	}

	msg := &Message{
		To:       []string{email},
		From:     s.fromEmail,
		FromName: s.fromName,
		Subject:  subject,
		HTMLBody: html,
	}

	return s.provider.Send(ctx, msg)
}

// PasswordResetData holds data for the password reset email template.
type PasswordResetData struct {
	FirstName    string
	ResetURL     string
	ExpiresIn    string
	SupportURL   string
	CurrentYear  int
}

// SendPasswordReset sends a password reset email.
func (s *Service) SendPasswordReset(ctx context.Context, email, firstName, resetToken string) error {
	data := PasswordResetData{
		FirstName:   firstName,
		ResetURL:    fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, resetToken),
		ExpiresIn:   "1 hour",
		SupportURL:  s.baseURL + "/support",
		CurrentYear: time.Now().Year(),
	}

	html, err := s.renderTemplate("password_reset.html", data)
	if err != nil {
		return err
	}

	msg := &Message{
		To:       []string{email},
		From:     s.fromEmail,
		FromName: s.fromName,
		Subject:  "Reset Your Password",
		HTMLBody: html,
	}

	return s.provider.Send(ctx, msg)
}

// OrderConfirmationData holds data for the order confirmation email template.
type OrderConfirmationData struct {
	FirstName      string
	OrderID        string
	ProductName    string
	PackageName    string
	BillingCycle   string
	TotalAmount    string
	InstanceName   string
	DashboardURL   string
	SupportURL     string
	CurrentYear    int
}

// SendOrderConfirmation sends an order confirmation email.
func (s *Service) SendOrderConfirmation(ctx context.Context, email, firstName, orderID, productName, packageName, billingCycle, totalAmount, instanceName string) error {
	data := OrderConfirmationData{
		FirstName:    firstName,
		OrderID:      orderID,
		ProductName:  productName,
		PackageName:  packageName,
		BillingCycle: billingCycle,
		TotalAmount:  totalAmount,
		InstanceName: instanceName,
		DashboardURL: s.baseURL + "/dashboard",
		SupportURL:   s.baseURL + "/support",
		CurrentYear:  time.Now().Year(),
	}

	html, err := s.renderTemplate("order_confirmation.html", data)
	if err != nil {
		return err
	}

	msg := &Message{
		To:       []string{email},
		From:     s.fromEmail,
		FromName: s.fromName,
		Subject:  fmt.Sprintf("Order Confirmation - %s %s", productName, packageName),
		HTMLBody: html,
	}

	return s.provider.Send(ctx, msg)
}
