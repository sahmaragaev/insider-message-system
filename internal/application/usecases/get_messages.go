package usecases

import (
	"context"
	"insider-message-system/internal/application/services"
	"insider-message-system/internal/common/pagination"
	"insider-message-system/internal/domain"
	"insider-message-system/pkg/logger"

	"go.uber.org/zap"
)

// GetMessagesRequest represents the query parameters for retrieving sent messages.
type GetMessagesRequest struct {
	Page  int `form:"page" binding:"omitempty,min=1" example:"1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
}

// GetMessagesResponse is the response structure for a paginated list of sent messages.
type GetMessagesResponse struct {
	Data       []*domain.Message         `json:"data"`
	Pagination pagination.PaginationInfo `json:"pagination"`
}

// GetMessagesUseCase handles the business logic for retrieving sent messages.
type GetMessagesUseCase struct {
	messageService services.Message
}

// NewGetMessagesUseCase creates a new GetMessagesUseCase with the given message service.
func NewGetMessagesUseCase(messageService services.Message) *GetMessagesUseCase {
	return &GetMessagesUseCase{
		messageService: messageService,
	}
}

// Execute retrieves sent messages based on the request parameters.
func (uc *GetMessagesUseCase) Execute(ctx context.Context, request GetMessagesRequest) (*GetMessagesResponse, error) {
	logger.Info("Getting sent messages",
		zap.Int("page", request.Page),
		zap.Int("limit", request.Limit))

	messages, total, err := uc.messageService.GetSentMessages(ctx, request.Page, request.Limit)
	if err != nil {
		logger.Error("Failed to get sent messages in use case", zap.Error(err))
		return nil, err
	}

	totalPages := int((total + int64(request.Limit) - 1) / int64(request.Limit))

	response := &GetMessagesResponse{
		Data: messages,
		Pagination: pagination.PaginationInfo{
			Page:       request.Page,
			PageSize:   request.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	logger.Info("Retrieved sent messages successfully",
		zap.Int("count", len(messages)),
		zap.Int64("total", total))

	return response, nil
}
