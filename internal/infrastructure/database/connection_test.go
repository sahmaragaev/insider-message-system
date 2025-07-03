package database

import (
	"errors"
	"insider-message-system/pkg/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, 5, cfg.MaxAttempts)
}

func Test_retryWithBackoff_Success(t *testing.T) {
	calls := 0
	err := retryWithBackoff(func() error {
		calls++
		if calls < 3 {
			return errors.New("fail")
		}
		return nil
	}, DefaultRetryConfig())
	assert.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func Test_retryWithBackoff_Failure(t *testing.T) {
	calls := 0
	err := retryWithBackoff(func() error {
		calls++
		return errors.New("fail")
	}, RetryConfig{MaxAttempts: 2, InitialDelay: 1 * time.Millisecond, MaxDelay: 1 * time.Millisecond, BackoffMultiplier: 1})
	assert.Error(t, err)
	assert.Equal(t, 2, calls)
}

func configToRetryConfig(attempts int) RetryConfig {
	return RetryConfig{MaxAttempts: attempts, InitialDelay: 1 * time.Millisecond, MaxDelay: 1 * time.Millisecond, BackoffMultiplier: 1}
}

func TestNewConnection_SuccessAndClose(t *testing.T) {
	cfg := config.DatabaseConfig{DSN: ":memory:", MaxOpenConnections: 1, MaxIdleConnections: 1, ConnMaxLifetime: time.Second}
	conn, err := NewConnectionWithDialector(cfg, sqlite.Open(":memory:"))
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.NoError(t, conn.Close())
}

func TestNewConnection_Failure(t *testing.T) {
	cfg := config.DatabaseConfig{DSN: "file:nonexistent?mode=invalid"}
	conn, err := NewConnection(cfg)
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestDB_Migrate_And_HealthCheck(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	conn := &DB{DB: db}
	assert.NoError(t, conn.Migrate())
	assert.NoError(t, conn.HealthCheck())
}
