package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"insider-message-system/internal/domain"
	"insider-message-system/pkg/config"
	"insider-message-system/pkg/errors"
	"insider-message-system/pkg/logger"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CacheService defines the interface for caching message data in Redis.
type CacheService interface {
	SetMessageCache(ctx context.Context, messageID uuid.UUID, entry domain.CacheEntry) error
	GetMessageCache(ctx context.Context, messageID uuid.UUID) (*domain.CacheEntry, error)
	DeleteMessageCache(ctx context.Context, messageID uuid.UUID) error
}

type redisClient interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type cacheService struct {
	client      redisClient
	closeClient *redis.Client
}

// NewCacheService creates a new Redis cache service with the given configuration.
func NewCacheService(cfg config.RedisConfig) (CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established successfully")

	return &cacheService{client: client, closeClient: client}, nil
}

func (c *cacheService) SetMessageCache(ctx context.Context, messageID uuid.UUID, entry domain.CacheEntry) error {
	key := fmt.Sprintf("message:%s", messageID.String())

	data, err := json.Marshal(entry)
	if err != nil {
		logger.Error("Failed to marshal cache entry", zap.Error(err), zap.String("message_id", messageID.String()))
		return errors.WrapError(err, "CACHE_ERROR", "Failed to marshal cache entry", 500)
	}

	err = c.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		logger.Error("Failed to set cache entry", zap.Error(err), zap.String("message_id", messageID.String()))
		return errors.WrapError(err, "CACHE_ERROR", "Failed to set cache entry", 500)
	}

	logger.Debug("Cache entry set successfully", zap.String("message_id", messageID.String()))
	return nil
}

func (c *cacheService) GetMessageCache(ctx context.Context, messageID uuid.UUID) (*domain.CacheEntry, error) {
	key := fmt.Sprintf("message:%s", messageID.String())

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		logger.Error("Failed to get cache entry", zap.Error(err), zap.String("message_id", messageID.String()))
		return nil, errors.WrapError(err, "CACHE_ERROR", "Failed to get cache entry", 500)
	}

	var entry domain.CacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		logger.Error("Failed to unmarshal cache entry", zap.Error(err), zap.String("message_id", messageID.String()))
		return nil, errors.WrapError(err, "CACHE_ERROR", "Failed to unmarshal cache entry", 500)
	}

	return &entry, nil
}

func (c *cacheService) DeleteMessageCache(ctx context.Context, messageID uuid.UUID) error {
	key := fmt.Sprintf("message:%s", messageID.String())

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		logger.Error("Failed to delete cache entry", zap.Error(err), zap.String("message_id", messageID.String()))
		return errors.WrapError(err, "CACHE_ERROR", "Failed to delete cache entry", 500)
	}

	logger.Debug("Cache entry deleted successfully", zap.String("message_id", messageID.String()))
	return nil
}

func (c *cacheService) Close() error {
	logger.Info("Closing Redis connection")
	if c.closeClient != nil {
		return c.closeClient.Close()
	}
	return nil
}

func (c *cacheService) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if c.closeClient != nil {
		return c.closeClient.Ping(ctx).Err()
	}
	return nil
}
