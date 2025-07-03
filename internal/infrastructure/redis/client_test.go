package redis

import (
	"context"
	"errors"
	"insider-message-system/internal/domain"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRedisClient struct {
	mock.Mock
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

type mockCloseClient struct {
	mock.Mock
}

func (m *mockCloseClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockPingClient struct {
	mock.Mock
}

func (m *mockPingClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func TestCacheService_SetGetDelete(t *testing.T) {
	m := new(mockRedisClient)
	cs := &cacheService{client: m}
	ctx := context.Background()
	id := uuid.New()
	entry := domain.CacheEntry{MessageID: "mid", SentAt: time.Now()}
	key := "message:" + id.String()

	setCmd := redis.NewStatusCmd(ctx)
	m.On("Set", ctx, key, mock.Anything, mock.Anything).Return(setCmd).Once()
	setCmd.SetVal("OK")
	setCmd.SetErr(nil)
	err := cs.SetMessageCache(ctx, id, entry)
	assert.NoError(t, err)

	jsonData := `{"message_id":"mid","sent_at":"2020-01-01T00:00:00Z"}`
	getCmd := redis.NewStringCmd(ctx)
	getCmd.SetVal(jsonData)
	getCmd.SetErr(nil)
	m.On("Get", ctx, key).Return(getCmd).Once()
	got, err := cs.GetMessageCache(ctx, id)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "mid", got.MessageID)

	delCmd := redis.NewIntCmd(ctx)
	delCmd.SetVal(1)
	delCmd.SetErr(nil)
	m.On("Del", ctx, []string{key}).Return(delCmd).Once()
	err = cs.DeleteMessageCache(ctx, id)
	assert.NoError(t, err)
}

func TestCacheService_Close(t *testing.T) {
	cs := &cacheService{closeClient: nil}
	assert.NoError(t, cs.Close())
}

func TestCacheService_HealthCheck(t *testing.T) {
	cs := &cacheService{closeClient: nil}
	assert.NoError(t, cs.HealthCheck())
}

func TestCacheService_SetMessageCache_RedisError(t *testing.T) {
	m := new(mockRedisClient)
	cs := &cacheService{client: m}
	ctx := context.Background()
	id := uuid.New()
	entry := domain.CacheEntry{MessageID: "mid"}
	key := "message:" + id.String()

	setCmd := redis.NewStatusCmd(ctx)
	setCmd.SetErr(errors.New("redis fail"))
	m.On("Set", ctx, key, mock.Anything, mock.Anything).Return(setCmd).Once()
	err := cs.SetMessageCache(ctx, id, entry)
	assert.Error(t, err)
}

func TestCacheService_GetMessageCache_RedisNil(t *testing.T) {
	m := new(mockRedisClient)
	cs := &cacheService{client: m}
	ctx := context.Background()
	id := uuid.New()
	key := "message:" + id.String()

	getCmd := redis.NewStringCmd(ctx)
	getCmd.SetErr(redis.Nil)
	m.On("Get", ctx, key).Return(getCmd).Once()
	got, err := cs.GetMessageCache(ctx, id)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestCacheService_GetMessageCache_RedisError(t *testing.T) {
	m := new(mockRedisClient)
	cs := &cacheService{client: m}
	ctx := context.Background()
	id := uuid.New()
	key := "message:" + id.String()

	getCmd := redis.NewStringCmd(ctx)
	getCmd.SetErr(errors.New("redis fail"))
	m.On("Get", ctx, key).Return(getCmd).Once()
	got, err := cs.GetMessageCache(ctx, id)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestCacheService_GetMessageCache_UnmarshalError(t *testing.T) {
	m := new(mockRedisClient)
	cs := &cacheService{client: m}
	ctx := context.Background()
	id := uuid.New()
	key := "message:" + id.String()

	getCmd := redis.NewStringCmd(ctx)
	getCmd.SetVal("not-json")
	getCmd.SetErr(nil)
	m.On("Get", ctx, key).Return(getCmd).Once()
	got, err := cs.GetMessageCache(ctx, id)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestCacheService_DeleteMessageCache_RedisError(t *testing.T) {
	m := new(mockRedisClient)
	cs := &cacheService{client: m}
	ctx := context.Background()
	id := uuid.New()
	key := "message:" + id.String()

	delCmd := redis.NewIntCmd(ctx)
	delCmd.SetErr(errors.New("redis fail"))
	m.On("Del", ctx, []string{key}).Return(delCmd).Once()
	err := cs.DeleteMessageCache(ctx, id)
	assert.Error(t, err)
}
