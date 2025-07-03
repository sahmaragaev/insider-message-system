package services

import (
	"context"
	"insider-message-system/pkg/logger"

	"go.uber.org/zap"
)

type MessageProcessor interface {
	ProcessMessages(ctx context.Context) error
}

type messageProcessor struct {
	messageService Message
	batchSize      int
}

func NewMessageProcessor(messageService Message, batchSize int) MessageProcessor {
	return &messageProcessor{
		messageService: messageService,
		batchSize:      batchSize,
	}
}

func (p *messageProcessor) ProcessMessages(ctx context.Context) error {
	logger.Debug("Processing messages", zap.Int("batch_size", p.batchSize))

	return p.messageService.SendPendingMessages(ctx, p.batchSize)
}
