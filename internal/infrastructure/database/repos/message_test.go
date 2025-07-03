package repos

import (
	"context"
	"insider-message-system/internal/domain"
	"insider-message-system/pkg/constants/enums/messagestatus"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"insider-message-system/internal/infrastructure/database"
)

type mockMessageRepo struct{}

func (m *mockMessageRepo) Create(ctx context.Context, msg *domain.Message) error { return nil }
func (m *mockMessageRepo) GetSentMessages(ctx context.Context, offset, limit int) ([]*domain.Message, error) {
	return nil, nil
}
func (m *mockMessageRepo) GetPendingMessages(ctx context.Context, limit int) ([]*domain.Message, error) {
	return nil, nil
}
func (m *mockMessageRepo) GetTotalSentCount(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockMessageRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status messagestatus.MessageStatus, messageID, failureReason *string) error {
	return nil
}
func (m *mockMessageRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	return nil, nil
}

func TestMessageRepoInterface(t *testing.T) {
	var repo Message = &mockMessageRepo{}
	assert.NotNil(t, repo)
}

func setupTestDB(t *testing.T) *database.DB {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = gdb.AutoMigrate(&domain.Message{})
	require.NoError(t, err)
	return &database.DB{DB: gdb}
}

func TestMessageRepo_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMessage(db)
	msg := &domain.Message{
		ID:        uuid.New(),
		To:        "+1234567890",
		Content:   "hello",
		Status:    messagestatus.Pending,
		CreatedAt: time.Now(),
	}
	err := repo.Create(context.Background(), msg)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), msg.ID)
	require.NoError(t, err)
	require.Equal(t, msg.To, got.To)
	require.Equal(t, msg.Content, got.Content)
}

func TestMessageRepo_GetPendingMessages(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMessage(db)
	msg := &domain.Message{
		ID:        uuid.New(),
		To:        "+1234567890",
		Content:   "pending",
		Status:    messagestatus.Pending,
		CreatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), msg))

	pending, err := repo.GetPendingMessages(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, "pending", pending[0].Content)
}

func TestMessageRepo_GetSentMessages(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMessage(db)
	msg := &domain.Message{
		ID:        uuid.New(),
		To:        "+1234567890",
		Content:   "sent",
		Status:    messagestatus.Sent,
		CreatedAt: time.Now(),
		SentAt:    ptrTime(time.Now()),
	}
	require.NoError(t, repo.Create(context.Background(), msg))

	sent, err := repo.GetSentMessages(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Len(t, sent, 1)
	require.Equal(t, "sent", sent[0].Content)
}

func TestMessageRepo_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMessage(db)
	msg := &domain.Message{
		ID:        uuid.New(),
		To:        "+1234567890",
		Content:   "to update",
		Status:    messagestatus.Pending,
		CreatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), msg))

	messageID := "external-id"
	failureReason := "failed to send"
	err := repo.UpdateStatus(context.Background(), msg.ID, messagestatus.Failed, &messageID, &failureReason)
	require.NoError(t, err)

	updated, err := repo.GetByID(context.Background(), msg.ID)
	require.NoError(t, err)
	require.Equal(t, messagestatus.Failed, updated.Status)
	require.Equal(t, &messageID, updated.MessageID)
	require.Equal(t, &failureReason, updated.FailureReason)
}

func TestMessageRepo_GetTotalSentCount(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMessage(db)
	msg := &domain.Message{
		ID:        uuid.New(),
		To:        "+1234567890",
		Content:   "sent count",
		Status:    messagestatus.Sent,
		CreatedAt: time.Now(),
		SentAt:    ptrTime(time.Now()),
	}
	require.NoError(t, repo.Create(context.Background(), msg))

	count, err := repo.GetTotalSentCount(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
