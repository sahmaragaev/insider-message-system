package usecases

import (
	"context"
	"insider-message-system/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMessageService struct {
	mock.Mock
}

func (m *mockMessageService) GetSentMessages(ctx context.Context, page, pageSize int) ([]*domain.Message, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*domain.Message), args.Get(1).(int64), args.Error(2)
}

func (m *mockMessageService) CreateMessage(ctx context.Context, req domain.MessageRequest) (*domain.Message, error) {
	return nil, nil
}

func (m *mockMessageService) SendPendingMessages(ctx context.Context, limit int) error {
	return nil
}

func TestGetMessagesUseCase_Execute(t *testing.T) {
	mockSvc := new(mockMessageService)
	uc := NewGetMessagesUseCase(mockSvc)

	messages := []*domain.Message{
		{ID: uuid.New(), Content: "Hello"},
		{ID: uuid.New(), Content: "World"},
	}
	total := int64(2)
	mockSvc.On("GetSentMessages", mock.Anything, 1, 10).Return(messages, total, nil)

	request := GetMessagesRequest{Page: 1, Limit: 10}
	resp, err := uc.Execute(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, messages, resp.Data)
	assert.Equal(t, 1, resp.Pagination.Page)
	assert.Equal(t, 10, resp.Pagination.PageSize)
	assert.Equal(t, int64(2), resp.Pagination.Total)
	assert.Equal(t, 1, resp.Pagination.TotalPages)

	mockSvc.AssertExpectations(t)
}
