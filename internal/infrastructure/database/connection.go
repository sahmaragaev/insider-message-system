package database

import (
	"fmt"
	"insider-message-system/internal/domain"
	"insider-message-system/pkg/config"
	"insider-message-system/pkg/constants/enums/messagestatus"
	"insider-message-system/pkg/logger"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// DB wraps the GORM database connection for the application.
type DB struct {
	*gorm.DB
}

// RetryConfig configures retry logic for database operations.
type RetryConfig struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       5,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

func retryWithBackoff(operation func() error, config RetryConfig) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if err := operation(); err == nil {
			if attempt > 1 {
				logger.Info(fmt.Sprintf("Operation succeeded after %d attempts", attempt))
			}
			return nil
		} else {
			lastErr = err
			logger.Warn(fmt.Sprintf("Attempt %d failed: %v", attempt, err))
		}

		if attempt == config.MaxAttempts {
			break
		}

		logger.Info(fmt.Sprintf("Retrying in %v...", delay))
		time.Sleep(delay)

		delay = time.Duration(float64(delay) * config.BackoffMultiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// Add a NewConnectionWithDialector for testability
func NewConnectionWithDialector(cfg config.DatabaseConfig, dialector gorm.Dialector) (*DB, error) {
	retryConfig := DefaultRetryConfig()

	var db *gorm.DB
	var err error

	err = retryWithBackoff(func() error {
		gormConfig := &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Silent),
		}

		db, err = gorm.Open(dialector, gormConfig)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}

		sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}

		return nil
	}, retryConfig)

	if err != nil {
		return nil, err
	}

	logger.Info("Database connection established successfully")

	return &DB{
		DB: db,
	}, nil
}

// Update NewConnection to call NewConnectionWithDialector with Postgres by default
func NewConnection(cfg config.DatabaseConfig) (*DB, error) {
	return NewConnectionWithDialector(cfg, postgres.Open(cfg.DSN))
}

// Migrate runs database migrations and creates necessary enums and indexes.
func (db *DB) Migrate() error {
	retryConfig := DefaultRetryConfig()

	return retryWithBackoff(func() error {
		err := db.AutoMigrate(&domain.Message{})
		if err != nil {
			return fmt.Errorf("failed to run auto migration: %w", err)
		}

		// Only run enum/index creation for Postgres
		dialect := strings.ToLower(db.Dialector.Name())
		if dialect == "postgres" {
			if err := db.createEnumIfNotExists(); err != nil {
				return fmt.Errorf("failed to create enum type: %w", err)
			}
			if err := db.createIndexes(); err != nil {
				return fmt.Errorf("failed to create indexes: %w", err)
			}
		}

		return nil
	}, retryConfig)
}

func (db *DB) createEnumIfNotExists() error {
	return db.Exec(fmt.Sprintf(`
		DO $$ 
		BEGIN
			CREATE TYPE message_status AS ENUM ('%s', '%s', '%s');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;
	`, messagestatus.Pending, messagestatus.Sent, messagestatus.Failed)).Error
}

func (db *DB) createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status)",
		"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages(sent_at)",
		"CREATE INDEX IF NOT EXISTS idx_messages_status_created ON messages(status, created_at) WHERE status = '" + messagestatus.Pending.String() + "'",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return err
		}
	}

	return nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	logger.Info("Closing database connection")
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// HealthCheck pings the database to verify connectivity.
func (db *DB) HealthCheck() error {
	retryConfig := RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
	}

	return retryWithBackoff(func() error {
		sqlDB, err := db.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	}, retryConfig)
}
