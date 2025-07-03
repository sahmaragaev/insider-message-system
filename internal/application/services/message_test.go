package services

import (
	"context"
	"errors"
	"insider-message-system/internal/domain"
	"insider-message-system/pkg/circuitbreaker"
	"insider-message-system/pkg/constants/enums/messagestatus"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMessageRepository struct {
	mock.Mock
}

func (m *mockMessageRepository) Create(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *mockMessageRepository) GetPendingMessages(ctx context.Context, limit int) ([]*domain.Message, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*domain.Message), args.Error(1)
}

func (m *mockMessageRepository) GetSentMessages(ctx context.Context, offset, limit int) ([]*domain.Message, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*domain.Message), args.Error(1)
}

func (m *mockMessageRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status messagestatus.MessageStatus,
	messageID *string,
	failureReason *string,
) error {
	args := m.Called(ctx, id, status, messageID, failureReason)
	return args.Error(0)
}

func (m *mockMessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *mockMessageRepository) GetTotalSentCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

type mockWebhookService struct {
	mock.Mock
}

func (m *mockWebhookService) SendMessage(ctx context.Context, request domain.WebhookRequest) (*domain.MessageResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*domain.MessageResponse), args.Error(1)
}

func (m *mockWebhookService) GetCircuitBreakerMetrics() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *mockWebhookService) GetCircuitBreakerState() circuitbreaker.State {
	args := m.Called()
	return args.Get(0).(circuitbreaker.State)
}

type mockCacheService struct {
	mock.Mock
}

func (m *mockCacheService) SetMessageCache(ctx context.Context, messageID uuid.UUID, entry domain.CacheEntry) error {
	args := m.Called(ctx, messageID, entry)
	return args.Error(0)
}

func (m *mockCacheService) GetMessageCache(ctx context.Context, messageID uuid.UUID) (*domain.CacheEntry, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).(*domain.CacheEntry), args.Error(1)
}

func (m *mockCacheService) DeleteMessageCache(ctx context.Context, messageID uuid.UUID) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func TestMessageService_CreateMessage(t *testing.T) {
	tests := []struct {
		name        string
		request     domain.MessageRequest
		setupMocks  func(*mockMessageRepository)
		expectError bool
	}{
		{
			name: "successful message creation",
			request: domain.MessageRequest{
				To:      "+905551111111",
				Content: "Test message",
			},
			setupMocks: func(repo *mockMessageRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "content too long",
			request: domain.MessageRequest{
				To:      "+905551111111",
				Content: "This is a very long message that exceeds the 160 character limit. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			},
			setupMocks:  func(repo *mockMessageRepository) {},
			expectError: true,
		},
		{
			name: "repository error",
			request: domain.MessageRequest{
				To:      "+905551111111",
				Content: "Test message",
			},
			setupMocks: func(repo *mockMessageRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMessageRepository{}
			mockWebhook := &mockWebhookService{}
			mockCache := &mockCacheService{}

			tt.setupMocks(mockRepo)

			service := NewMessage(mockRepo, mockWebhook, mockCache)
			ctx := context.Background()

			message, err := service.CreateMessage(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, message)
				assert.Equal(t, tt.request.To, message.To)
				assert.Equal(t, tt.request.Content, message.Content)
				assert.Equal(t, messagestatus.Pending, message.Status)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageService_SendPendingMessages(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		setupMocks  func(*mockMessageRepository, *mockWebhookService, *mockCacheService)
		expectError bool
	}{
		{
			name:  "no pending messages",
			limit: 2,
			setupMocks: func(repo *mockMessageRepository, webhook *mockWebhookService, cache *mockCacheService) {
				repo.On("GetPendingMessages", mock.Anything, 2).Return([]*domain.Message{}, nil)
			},
			expectError: false,
		},
		{
			name:  "successful message sending",
			limit: 1,
			setupMocks: func(repo *mockMessageRepository, webhook *mockWebhookService, cache *mockCacheService) {
				message := &domain.Message{
					ID:        uuid.New(),
					To:        "+905551111111",
					Content:   "Test message",
					Status:    messagestatus.Pending,
					CreatedAt: time.Now(),
				}

				repo.On("GetPendingMessages", mock.Anything, 1).Return([]*domain.Message{message}, nil)

				webhookResponse := &domain.MessageResponse{
					Message:   "Accepted",
					MessageID: "test-message-id",
				}
				webhook.On("SendMessage", mock.Anything, mock.AnythingOfType("domain.WebhookRequest")).Return(webhookResponse, nil)

				repo.On("UpdateStatus", mock.Anything, message.ID, messagestatus.Sent, &webhookResponse.MessageID, (*string)(nil)).Return(nil)

				cache.On("SetMessageCache", mock.Anything, message.ID, mock.AnythingOfType("domain.CacheEntry")).Return(nil)
			},
			expectError: false,
		},
		{
			name:  "webhook error",
			limit: 1,
			setupMocks: func(repo *mockMessageRepository, webhook *mockWebhookService, cache *mockCacheService) {
				message := &domain.Message{
					ID:        uuid.New(),
					To:        "+905551111111",
					Content:   "Test message",
					Status:    messagestatus.Pending,
					CreatedAt: time.Now(),
				}

				repo.On("GetPendingMessages", mock.Anything, 1).Return([]*domain.Message{message}, nil)

				webhook.On("SendMessage", mock.Anything, mock.AnythingOfType("domain.WebhookRequest")).Return((*domain.MessageResponse)(nil), errors.New("webhook error"))

				failureReason := "webhook error"
				repo.On("UpdateStatus", mock.Anything, message.ID, messagestatus.Failed, (*string)(nil), &failureReason).Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMessageRepository{}
			mockWebhook := &mockWebhookService{}
			mockCache := &mockCacheService{}

			tt.setupMocks(mockRepo, mockWebhook, mockCache)

			service := NewMessage(mockRepo, mockWebhook, mockCache)
			ctx := context.Background()

			err := service.SendPendingMessages(ctx, tt.limit)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockWebhook.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestMessageService_GetSentMessages(t *testing.T) {
	type args struct {
		page     int
		pageSize int
	}

	tests := []struct {
		name        string
		args        args
		setupMocks  func(*mockMessageRepository)
		expectError bool
	}{
		{
			name: "success",
			args: args{page: 2, pageSize: 5},
			setupMocks: func(repo *mockMessageRepository) {
				repo.On("GetSentMessages", mock.Anything, 5, 5).Return([]*domain.Message{{ID: uuid.New()}}, nil)
				repo.On("GetTotalSentCount", mock.Anything).Return(int64(10), nil)
			},
			expectError: false,
		},
		{
			name: "repo error on GetSentMessages",
			args: args{page: 1, pageSize: 10},
			setupMocks: func(repo *mockMessageRepository) {
				repo.On("GetSentMessages", mock.Anything, 0, 10).Return([]*domain.Message(nil), errors.New("db error"))
			},
			expectError: true,
		},
		{
			name: "repo error on GetTotalSentCount",
			args: args{page: 1, pageSize: 10},
			setupMocks: func(repo *mockMessageRepository) {
				repo.On("GetSentMessages", mock.Anything, 0, 10).Return([]*domain.Message{}, nil)
				repo.On("GetTotalSentCount", mock.Anything).Return(int64(0), errors.New("count error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMessageRepository{}
			mockWebhook := &mockWebhookService{}
			mockCache := &mockCacheService{}

			tt.setupMocks(mockRepo)

			service := NewMessage(mockRepo, mockWebhook, mockCache)
			ctx := context.Background()

			msgs, total, err := service.GetSentMessages(ctx, tt.args.page, tt.args.pageSize)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, msgs)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msgs)
				assert.Equal(t, int64(10), total)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageService_SendPendingMessages_UpdateStatusFails(t *testing.T) {
	mockRepo := &mockMessageRepository{}
	mockWebhook := &mockWebhookService{}
	mockCache := &mockCacheService{}

	message := &domain.Message{
		ID:        uuid.New(),
		To:        "+905551111111",
		Content:   "Test message",
		Status:    messagestatus.Pending,
		CreatedAt: time.Now(),
	}
	mockRepo.On("GetPendingMessages", mock.Anything, 1).Return([]*domain.Message{message}, nil)
	webhookResponse := &domain.MessageResponse{Message: "Accepted", MessageID: "test-message-id"}
	mockWebhook.On("SendMessage", mock.Anything, mock.AnythingOfType("domain.WebhookRequest")).Return(webhookResponse, nil)
	mockRepo.On("UpdateStatus", mock.Anything, message.ID, messagestatus.Sent, &webhookResponse.MessageID, (*string)(nil)).Return(errors.New("update error"))
	mockCache.On("SetMessageCache", mock.Anything, message.ID, mock.AnythingOfType("domain.CacheEntry")).Return(nil).Maybe()

	service := NewMessage(mockRepo, mockWebhook, mockCache)
	err := service.SendPendingMessages(context.Background(), 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockWebhook.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestMessageService_SendPendingMessages_CacheServiceNil(t *testing.T) {
	mockRepo := &mockMessageRepository{}
	mockWebhook := &mockWebhookService{}

	message := &domain.Message{
		ID:        uuid.New(),
		To:        "+905551111111",
		Content:   "Test message",
		Status:    messagestatus.Pending,
		CreatedAt: time.Now(),
	}
	mockRepo.On("GetPendingMessages", mock.Anything, 1).Return([]*domain.Message{message}, nil)
	webhookResponse := &domain.MessageResponse{Message: "Accepted", MessageID: "test-message-id"}
	mockWebhook.On("SendMessage", mock.Anything, mock.AnythingOfType("domain.WebhookRequest")).Return(webhookResponse, nil)
	mockRepo.On("UpdateStatus", mock.Anything, message.ID, messagestatus.Sent, &webhookResponse.MessageID, (*string)(nil)).Return(nil)

	service := NewMessage(mockRepo, mockWebhook, nil)
	err := service.SendPendingMessages(context.Background(), 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockWebhook.AssertExpectations(t)
}

func TestMessageService_sendSingleMessage_InvalidMessage(t *testing.T) {
	mockRepo := &mockMessageRepository{}
	mockWebhook := &mockWebhookService{}
	mockCache := &mockCacheService{}

	invalidMsg := &domain.Message{
		ID:        uuid.New(),
		To:        "+905551111111",
		Content:   "This message is too long to be valid for sending because it exceeds the 160 character limit. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		Status:    messagestatus.Pending,
		CreatedAt: time.Now(),
	}

	service := NewMessage(mockRepo, mockWebhook, mockCache)
	// Should not panic or call webhook/repo
	err := service.(*message).sendSingleMessage(context.Background(), invalidMsg)
	assert.NoError(t, err)
}
