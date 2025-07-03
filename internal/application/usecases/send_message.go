package usecases

import (
	"context"
	"insider-message-system/internal/application/services"
	"insider-message-system/internal/domain"
	"insider-message-system/pkg/logger"

	"errors"

	"go.uber.org/zap"
)

// SendMessageUseCase handles the business logic for sending a new message.
type SendMessageUseCase struct {
	messageService services.Message
}

// NewSendMessageUseCase creates a new SendMessageUseCase with the given message service.
func NewSendMessageUseCase(messageService services.Message) *SendMessageUseCase {
	return &SendMessageUseCase{
		messageService: messageService,
	}
}

// Execute sends a new message based on the provided request.
func (uc *SendMessageUseCase) Execute(ctx context.Context, request domain.MessageRequest) (*domain.Message, error) {
	logger.Info("Creating new message",
		zap.String("to", request.To),
		zap.Int("content_length", len(request.Content)))

	message, err := uc.messageService.CreateMessage(ctx, request)
	if err != nil {
		logger.Error("Failed to create message in use case", zap.Error(err))
		return nil, err
	}

	if message == nil {
		logger.Warn("CreateMessage returned nil message with no error")
		return nil, errors.New("internal: message is nil")
	}

	logger.Info("Message created successfully via use case", zap.String("message_id", message.ID.String()))
	return message, nil
}
