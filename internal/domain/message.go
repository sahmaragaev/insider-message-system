package domain

import (
	"insider-message-system/pkg/constants/enums/messagestatus"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Message represents a message entity in the system.
type Message struct {
	ID            uuid.UUID                   `json:"id" db:"id"`
	To            string                      `json:"to" db:"to"`
	Content       string                      `json:"content" db:"content"`
	Status        messagestatus.MessageStatus `json:"status" db:"status"`
	CreatedAt     time.Time                   `json:"created_at" db:"created_at"`
	SentAt        *time.Time                  `json:"sent_at,omitempty" db:"sent_at"`
	MessageID     *string                     `json:"message_id,omitempty" db:"message_id"`
	FailureReason *string                     `json:"failure_reason,omitempty" db:"failure_reason"`
}

// MessageRequest contains the data required to create a new message.
type MessageRequest struct {
	To      string `json:"to" binding:"required" example:"+905551111111"`
	Content string `json:"content" binding:"required,max=160" example:"Insider - Project"`
}

// MessageResponse represents the response returned after creating a message.
type MessageResponse struct {
	Message   string `json:"message" example:"Accepted"`
	MessageID string `json:"messageId" example:"67f2f8a8-ea58-4ed0-a6f9-ff217df4d849"`
}

// WebhookRequest represents the payload sent to a webhook.
type WebhookRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

// CacheEntry represents a cached message entry with its sent time.
type CacheEntry struct {
	MessageID string    `json:"message_id"`
	SentAt    time.Time `json:"sent_at"`
}

// NewMessage creates a new Message instance with the given recipient and content.
func NewMessage(to, content string) *Message {
	return &Message{
		ID:        uuid.New(),
		To:        to,
		Content:   content,
		Status:    messagestatus.Pending,
		CreatedAt: time.Now(),
	}
}

// MarkAsSent marks the message as sent and sets the sent time and message ID.
func (m *Message) MarkAsSent(messageID string) {
	now := time.Now()
	m.Status = messagestatus.Sent
	m.SentAt = &now
	m.MessageID = &messageID
}

// MarkAsFailed marks the message as failed and sets the failure reason.
func (m *Message) MarkAsFailed(reason string) {
	m.Status = messagestatus.Failed
	m.FailureReason = &reason
}

// IsValidForSending checks if the message is valid for sending (pending status and content length).
func (m *Message) IsValidForSending() bool {
	return m.Status == messagestatus.Pending && len(m.Content) <= 160
}

// TableName returns the database table name for the Message model.
func (Message) TableName() string {
	return "messages"
}

// BeforeCreate is a GORM hook that sets the ID before creating a new message record.
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
