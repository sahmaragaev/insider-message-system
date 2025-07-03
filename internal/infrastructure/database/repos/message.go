package repos

import (
	"context"
	"insider-message-system/internal/domain"
	"insider-message-system/internal/infrastructure/database"
	"insider-message-system/pkg/constants/enums/messagestatus"
	"insider-message-system/pkg/errors"
	"insider-message-system/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Message defines the interface for message repository operations.
type Message interface {
	Create(ctx context.Context, message *domain.Message) error
	GetPendingMessages(ctx context.Context, limit int) ([]*domain.Message, error)
	GetSentMessages(ctx context.Context, offset, limit int) ([]*domain.Message, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status messagestatus.MessageStatus, messageID *string, failureReason *string) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Message, error)
	GetTotalSentCount(ctx context.Context) (int64, error)
}

type message struct {
	db *database.DB
}

// NewMessage creates a new message repository with the given database connection.
func NewMessage(db *database.DB) Message {
	return &message{db: db}
}

func (r *message) Create(ctx context.Context, message *domain.Message) error {
	result := r.db.WithContext(ctx).Create(message)
	if result.Error != nil {
		logger.Error("Failed to create message", zap.Error(result.Error), zap.String("message_id", message.ID.String()))
		return errors.WrapError(result.Error, "DATABASE_ERROR", "Failed to create message", 500)
	}

	logger.Info("Message created successfully", zap.String("message_id", message.ID.String()))
	return nil
}

func (r *message) GetPendingMessages(ctx context.Context, limit int) ([]*domain.Message, error) {
	var messages []*domain.Message

	result := r.db.WithContext(ctx).
		Where("status = ?", messagestatus.Pending).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages)

	if result.Error != nil {
		logger.Error("Failed to get pending messages", zap.Error(result.Error))
		return nil, errors.WrapError(result.Error, "DATABASE_ERROR", "Failed to get pending messages", 500)
	}

	return messages, nil
}

func (r *message) GetSentMessages(ctx context.Context, offset, limit int) ([]*domain.Message, error) {
	var messages []*domain.Message

	result := r.db.WithContext(ctx).
		Where("status = ?", messagestatus.Sent).
		Order("sent_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages)

	if result.Error != nil {
		logger.Error("Failed to get sent messages", zap.Error(result.Error))
		return nil, errors.WrapError(result.Error, "DATABASE_ERROR", "Failed to get sent messages", 500)
	}

	return messages, nil
}

func (r *message) UpdateStatus(ctx context.Context, id uuid.UUID, status messagestatus.MessageStatus, messageID *string, failureReason *string) error {
	updates := map[string]any{
		"status":         status,
		"message_id":     messageID,
		"failure_reason": failureReason,
	}

	if status == messagestatus.Sent {
		updates["sent_at"] = gorm.Expr("NOW()")
	}

	result := r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		logger.Error("Failed to update message status", zap.Error(result.Error), zap.String("message_id", id.String()))
		return errors.WrapError(result.Error, "DATABASE_ERROR", "Failed to update message status", 500)
	}

	if result.RowsAffected == 0 {
		return errors.ErrMessageNotFound
	}

	logger.Info("Message status updated successfully",
		zap.String("message_id", id.String()),
		zap.String("status", string(status)))
	return nil
}

func (r *message) GetByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	var message domain.Message

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&message)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.ErrMessageNotFound
		}
		logger.Error("Failed to get message by ID", zap.Error(result.Error), zap.String("message_id", id.String()))
		return nil, errors.WrapError(result.Error, "DATABASE_ERROR", "Failed to get message", 500)
	}

	return &message, nil
}

func (r *message) GetTotalSentCount(ctx context.Context) (int64, error) {
	var count int64

	result := r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("status = ?", messagestatus.Sent).
		Count(&count)

	if result.Error != nil {
		logger.Error("Failed to get total sent count", zap.Error(result.Error))
		return 0, errors.WrapError(result.Error, "DATABASE_ERROR", "Failed to get total sent count", 500)
	}

	return count, nil
}
