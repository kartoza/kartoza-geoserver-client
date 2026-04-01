package models

import (
	"encoding/json"
	"time"
)

// TicketStatus represents the status of a support ticket.
type TicketStatus string

const (
	TicketStatusOpen            TicketStatus = "open"
	TicketStatusInProgress      TicketStatus = "in_progress"
	TicketStatusWaitingResponse TicketStatus = "waiting_response"
	TicketStatusResolved        TicketStatus = "resolved"
	TicketStatusClosed          TicketStatus = "closed"
)

// TicketPriority represents the priority of a support ticket.
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityNormal TicketPriority = "normal"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

// Ticket represents a support ticket.
type Ticket struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	InstanceID  *string        `json:"instance_id,omitempty"`
	Subject     string         `json:"subject"`
	Description string         `json:"description,omitempty"`
	Status      TicketStatus   `json:"status"`
	Priority    TicketPriority `json:"priority"`
	Category    string         `json:"category,omitempty"`
	AssignedTo  *string        `json:"assigned_to,omitempty"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	ClosedAt    *time.Time     `json:"closed_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// Related data
	User       *User            `json:"user,omitempty"`
	Instance   *Instance        `json:"instance,omitempty"`
	AssignedUser *User          `json:"assigned_user,omitempty"`
	Messages   []TicketMessage  `json:"messages,omitempty"`
}

// IsOpen returns true if the ticket is still open.
func (t *Ticket) IsOpen() bool {
	return t.Status != TicketStatusResolved && t.Status != TicketStatusClosed
}

// TicketMessage represents a message in a support ticket.
type TicketMessage struct {
	ID          string          `json:"id"`
	TicketID    string          `json:"ticket_id"`
	UserID      string          `json:"user_id"`
	IsStaff     bool            `json:"is_staff"`
	IsInternal  bool            `json:"is_internal"`
	Message     string          `json:"message"`
	Attachments json.RawMessage `json:"attachments,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`

	// Related data
	User *User `json:"user,omitempty"`
}

// CreateTicketRequest represents a request to create a ticket.
type CreateTicketRequest struct {
	InstanceID  *string        `json:"instance_id,omitempty"`
	Subject     string         `json:"subject"`
	Description string         `json:"description"`
	Priority    TicketPriority `json:"priority"`
	Category    string         `json:"category,omitempty"`
}

// AddTicketMessageRequest represents a request to add a message to a ticket.
type AddTicketMessageRequest struct {
	Message    string   `json:"message"`
	IsInternal bool     `json:"is_internal,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

// UpdateTicketRequest represents a request to update a ticket.
type UpdateTicketRequest struct {
	Status     *TicketStatus   `json:"status,omitempty"`
	Priority   *TicketPriority `json:"priority,omitempty"`
	AssignedTo *string         `json:"assigned_to,omitempty"`
}

// Notification represents a user notification.
type Notification struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Title            string    `json:"title"`
	Message          string    `json:"message"`
	NotificationType string    `json:"notification_type"`
	LinkURL          string    `json:"link_url,omitempty"`
	IsRead           bool      `json:"is_read"`
	ReadAt           *time.Time `json:"read_at,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// NotificationType constants
const (
	NotificationTypeInfo     = "info"
	NotificationTypeSuccess  = "success"
	NotificationTypeWarning  = "warning"
	NotificationTypeError    = "error"
	NotificationTypeInstance = "instance"
	NotificationTypeBilling  = "billing"
	NotificationTypeSystem   = "system"
)

// AuditLogEntry represents an entry in the audit log.
type AuditLogEntry struct {
	ID         string          `json:"id"`
	UserID     *string         `json:"user_id,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *string         `json:"entity_id,omitempty"`
	OldValues  json.RawMessage `json:"old_values,omitempty"`
	NewValues  json.RawMessage `json:"new_values,omitempty"`
	IPAddress  string          `json:"ip_address,omitempty"`
	UserAgent  string          `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`

	// Related data
	User *User `json:"user,omitempty"`
}
