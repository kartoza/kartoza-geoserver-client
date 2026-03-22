package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPProvider implements the Provider interface using SMTP.
type SMTPProvider struct {
	host     string
	port     int
	username string
	password string
	auth     smtp.Auth
}

// NewSMTPProvider creates a new SMTP email provider.
func NewSMTPProvider(host string, port int, username, password string) (*SMTPProvider, error) {
	if host == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}

	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	return &SMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		auth:     auth,
	}, nil
}

// Name returns the provider name.
func (p *SMTPProvider) Name() string {
	return "smtp"
}

// Send sends an email using SMTP.
func (p *SMTPProvider) Send(ctx context.Context, msg *Message) error {
	// Build email headers and body
	var builder strings.Builder

	// From header
	if msg.FromName != "" {
		builder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", msg.FromName, msg.From))
	} else {
		builder.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
	}

	// To header
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	// Reply-To header
	if msg.ReplyTo != "" {
		builder.WriteString(fmt.Sprintf("Reply-To: %s\r\n", msg.ReplyTo))
	}

	// Subject header
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))

	// MIME headers
	builder.WriteString("MIME-Version: 1.0\r\n")

	if msg.HTMLBody != "" && msg.TextBody != "" {
		// Multipart message
		boundary := "----=_NextPart_001_0001_01D12345.6789ABCD"
		builder.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		builder.WriteString("\r\n")

		// Plain text part
		builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(msg.TextBody)
		builder.WriteString("\r\n")

		// HTML part
		builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(msg.HTMLBody)
		builder.WriteString("\r\n")

		// End boundary
		builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if msg.HTMLBody != "" {
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(msg.HTMLBody)
	} else {
		builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(msg.TextBody)
	}

	emailBody := builder.String()
	addr := fmt.Sprintf("%s:%d", p.host, p.port)

	// Use TLS for port 465, STARTTLS for other ports
	if p.port == 465 {
		return p.sendWithTLS(addr, msg.From, msg.To, []byte(emailBody))
	}

	return smtp.SendMail(addr, p.auth, msg.From, msg.To, []byte(emailBody))
}

// sendWithTLS sends email using implicit TLS (port 465).
func (p *SMTPProvider) sendWithTLS(addr, from string, to []string, body []byte) error {
	tlsConfig := &tls.Config{
		ServerName: p.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, p.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate if credentials provided
	if p.auth != nil {
		if err := client.Auth(p.auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to start data: %w", err)
	}

	_, err = w.Write(body)
	if err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data: %w", err)
	}

	return client.Quit()
}
