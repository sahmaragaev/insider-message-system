package errors

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewError(code, message string, status int) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

func NewErrorWithDetails(code, message, details string, status int) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
		Status:  status,
	}
}

var (
	ErrInternalServer = NewError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrBadRequest     = NewError("BAD_REQUEST", "Bad request", http.StatusBadRequest)
	ErrNotFound       = NewError("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrUnauthorized   = NewError("UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
	ErrValidation     = NewError("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest)

	ErrMessageNotFound         = NewError("MESSAGE_NOT_FOUND", "Message not found", http.StatusNotFound)
	ErrMessageSendFailed       = NewError("MESSAGE_SEND_FAILED", "Failed to send message", http.StatusInternalServerError)
	ErrSchedulerNotRunning     = NewError("SCHEDULER_NOT_RUNNING", "Scheduler is not running", http.StatusBadRequest)
	ErrSchedulerAlreadyRunning = NewError("SCHEDULER_ALREADY_RUNNING", "Scheduler is already running", http.StatusBadRequest)
	ErrInvalidMessageContent   = NewError("INVALID_MESSAGE_CONTENT", "Message content exceeds character limit", http.StatusBadRequest)
	ErrDatabaseConnection      = NewError("DATABASE_CONNECTION_ERROR", "Database connection failed", http.StatusInternalServerError)
	ErrCacheConnection         = NewError("CACHE_CONNECTION_ERROR", "Cache connection failed", http.StatusInternalServerError)
)

func WrapError(err error, code, message string, status int) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: err.Error(),
		Status:  status,
	}
}
