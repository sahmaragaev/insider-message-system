package services

import (
	"context"
	"insider-message-system/internal/domain"
	"insider-message-system/internal/infrastructure/database/repos"
	"insider-message-system/internal/infrastructure/redis"
	"insider-message-system/internal/infrastructure/webhook"
	"insider-message-system/pkg/constants/enums/messagestatus"
	"insider-message-system/pkg/errors"
	"insider-message-system/pkg/logger"

	"go.uber.org/zap"
)

// Message defines the interface for message-related business logic.
type Message interface {
	CreateMessage(ctx context.Context, request domain.MessageRequest) (*domain.Message, error)
	GetSentMessages(ctx context.Context, page, pageSize int) ([]*domain.Message, int64, error)
	SendPendingMessages(ctx context.Context, limit int) error
}

type message struct {
	messageRepo    repos.Message
	webhookService webhook.Client
	cacheService   redis.CacheService
}

// NewMessage creates a new Message service with the given dependencies.
func NewMessage(
	messageRepo repos.Message,
	webhookService webhook.Client,
	cacheService redis.CacheService,
) Message {
	return &message{
		messageRepo:    messageRepo,
		webhookService: webhookService,
		cacheService:   cacheService,
	}
}

func (s *message) CreateMessage(ctx context.Context, request domain.MessageRequest) (*domain.Message, error) {
	if len(request.Content) > 160 {
		return nil, errors.ErrInvalidMessageContent
	}

	message := domain.NewMessage(request.To, request.Content)

	if err := s.messageRepo.Create(ctx, message); err != nil {
		logger.Error("Failed to create message", zap.Error(err))
		return nil, err
	}

	logger.Info("Message created successfully", zap.String("message_id", message.ID.String()))
	return message, nil
}

func (s *message) GetSentMessages(ctx context.Context, page, pageSize int) ([]*domain.Message, int64, error) {
	offset := (page - 1) * pageSize

	messages, err := s.messageRepo.GetSentMessages(ctx, offset, pageSize)
	if err != nil {
		logger.Error("Failed to get sent messages", zap.Error(err))
		return nil, 0, err
	}

	total, err := s.messageRepo.GetTotalSentCount(ctx)
	if err != nil {
		logger.Error("Failed to get total sent count", zap.Error(err))
		return nil, 0, err
	}

	return messages, total, nil
}

func (s *message) SendPendingMessages(ctx context.Context, limit int) error {
	messages, err := s.messageRepo.GetPendingMessages(ctx, limit)
	if err != nil {
		logger.Error("Failed to get pending messages", zap.Error(err))
		return err
	}

	if len(messages) == 0 {
		logger.Debug("No pending messages to send")
		return nil
	}

	logger.Info("Processing pending messages", zap.Int("count", len(messages)))

	for _, message := range messages {
		if err := s.sendSingleMessage(ctx, message); err != nil {
			logger.Error("Failed to send message",
				zap.Error(err),
				zap.String("message_id", message.ID.String()))
			continue
		}
	}

	return nil
}

func (s *message) sendSingleMessage(ctx context.Context, message *domain.Message) error {
	if !message.IsValidForSending() {
		logger.Warn("Message is not valid for sending",
			zap.String("message_id", message.ID.String()),
			zap.String("status", string(message.Status)))
		return nil
	}

	webhookRequest := domain.WebhookRequest{
		To:      message.To,
		Content: message.Content,
	}

	response, err := s.webhookService.SendMessage(ctx, webhookRequest)
	if err != nil {
		failureReason := err.Error()
		if err := s.messageRepo.UpdateStatus(ctx, message.ID, messagestatus.Failed, nil, &failureReason); err != nil {
			logger.Error("Failed to update message status to failed",
				zap.Error(err),
				zap.String("message_id", message.ID.String()))
		}
		return err
	}

	if err := s.messageRepo.UpdateStatus(ctx, message.ID, messagestatus.Sent, &response.MessageID, nil); err != nil {
		logger.Error("Failed to update message status to sent",
			zap.Error(err),
			zap.String("message_id", message.ID.String()))
		return err
	}

	if s.cacheService != nil {
		cacheEntry := domain.CacheEntry{
			MessageID: response.MessageID,
			SentAt:    message.CreatedAt,
		}

		if err := s.cacheService.SetMessageCache(ctx, message.ID, cacheEntry); err != nil {
			logger.Warn("Failed to cache message",
				zap.Error(err),
				zap.String("message_id", message.ID.String()))
		}
	}

	logger.Info("Message sent successfully",
		zap.String("message_id", message.ID.String()),
		zap.String("webhook_message_id", response.MessageID))

	return nil
}
