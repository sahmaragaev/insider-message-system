package apidocs

import (
	"time"

	"insider-message-system/internal/common/pagination"

	"github.com/google/uuid"
)

// Swagger response types for documentation purposes only
// These types mirror the actual response structure but with specific typing for better API docs

type MessageCreatedResponse struct {
	Status bool        `json:"status" example:"true"`
	Msg    string      `json:"msg" example:"Message created successfully"`
	Data   MessageData `json:"data"`
}

type MessageData struct {
	ID            uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	To            string     `json:"to" example:"+905551111111"`
	Content       string     `json:"content" example:"Insider - Project"`
	Status        string     `json:"status" example:"pending"`
	CreatedAt     time.Time  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	SentAt        *time.Time `json:"sent_at,omitempty" example:"2024-01-15T10:35:00Z"`
	MessageID     *string    `json:"message_id,omitempty" example:"msg_123456"`
	FailureReason *string    `json:"failure_reason,omitempty"`
}

type MessagesListResponse struct {
	Status bool             `json:"status" example:"true"`
	Msg    string           `json:"msg" example:"Request processed successfully"`
	Data   MessagesListData `json:"data"`
}

type MessagesListData struct {
	Data       []MessageData             `json:"data"`
	Pagination pagination.PaginationInfo `json:"pagination"`
}

type SchedulerResponse struct {
	Status bool          `json:"status" example:"true"`
	Msg    string        `json:"msg" example:"Scheduler started successfully"`
	Data   SchedulerData `json:"data"`
}

type SchedulerData struct {
	Status  string `json:"status" example:"running"`
	Message string `json:"message" example:"Scheduler started successfully"`
}

type ErrorResponse struct {
	Status bool      `json:"status" example:"false"`
	Msg    string    `json:"msg" example:"Validation failed"`
	Data   ErrorData `json:"data,omitempty"`
}

type ErrorData struct {
	Code    string `json:"code,omitempty" example:"VALIDATION_ERROR"`
	Details string `json:"details,omitempty" example:"Invalid request parameters"`
}
