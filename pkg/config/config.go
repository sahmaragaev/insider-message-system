package config

import (
	"insider-message-system/pkg/constants/enums/formattypes"
	"insider-message-system/pkg/constants/enums/loglevels"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Redis          RedisConfig          `mapstructure:"redis"`
	Webhook        WebhookConfig        `mapstructure:"webhook"`
	Scheduler      SchedulerConfig      `mapstructure:"scheduler"`
	Logger         LoggerConfig         `mapstructure:"logger"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Driver             string        `mapstructure:"driver"`
	DSN                string        `mapstructure:"dsn"`
	MaxOpenConnections int           `mapstructure:"max_open_connections"`
	MaxIdleConnections int           `mapstructure:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type WebhookConfig struct {
	URL     string        `mapstructure:"url"`
	AuthKey string        `mapstructure:"auth_key"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type SchedulerConfig struct {
	Interval   time.Duration `mapstructure:"interval"`
	BatchSize  int           `mapstructure:"batch_size"`
	MaxRetries int           `mapstructure:"max_retries"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
	AutoStart  bool          `mapstructure:"auto_start"`
}

type LoggerConfig struct {
	Level      loglevels.LogLevel     `mapstructure:"level"`
	Format     formattypes.FormatType `mapstructure:"format"`
	OutputPath string                 `mapstructure:"output_path"`
}

type CircuitBreakerConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	FailureRate   float64       `mapstructure:"failure_rate"`
	MinRequests   int           `mapstructure:"min_requests"`
	HalfOpenAfter time.Duration `mapstructure:"half_open_after"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	setupDefaults()
	setupEnvironmentVariables()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if isContainerEnvironment() {
				setupContainerDefaults()
			}
		} else {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setupDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.max_open_connections", 25)
	viper.SetDefault("database.max_idle_connections", 5)
	viper.SetDefault("database.conn_max_lifetime", "15m")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("webhook.timeout", "30s")

	viper.SetDefault("scheduler.interval", "2m")
	viper.SetDefault("scheduler.batch_size", 2)
	viper.SetDefault("scheduler.max_retries", 3)
	viper.SetDefault("scheduler.retry_delay", "5s")
	viper.SetDefault("scheduler.auto_start", false)

	viper.SetDefault("logger.level", loglevels.Info)
	viper.SetDefault("logger.format", formattypes.FormatJSON)
	viper.SetDefault("logger.output_path", "stdout")

	viper.SetDefault("circuit_breaker.enabled", true)
	viper.SetDefault("circuit_breaker.failure_rate", 0.5)
	viper.SetDefault("circuit_breaker.min_requests", 10)
	viper.SetDefault("circuit_breaker.half_open_after", "10s")
}

func setupEnvironmentVariables() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("")

	viper.BindEnv("database.dsn", "DATABASE_DSN")
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("webhook.url", "WEBHOOK_URL")
	viper.BindEnv("webhook.auth_key", "WEBHOOK_AUTH_KEY")
	viper.BindEnv("scheduler.auto_start", "SCHEDULER_AUTO_START")
	viper.BindEnv("scheduler.interval", "SCHEDULER_INTERVAL")
	viper.BindEnv("scheduler.batch_size", "SCHEDULER_BATCH_SIZE")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("logger.level", "LOG_LEVEL")
	viper.BindEnv("logger.format", "LOG_FORMAT")
	viper.BindEnv("logger.output_path", "LOG_OUTPUT_PATH")
	viper.BindEnv("circuit_breaker.enabled", "CIRCUIT_BREAKER_ENABLED")
	viper.BindEnv("circuit_breaker.failure_rate", "CIRCUIT_BREAKER_FAILURE_RATE")
	viper.BindEnv("circuit_breaker.min_requests", "CIRCUIT_BREAKER_MIN_REQUESTS")
	viper.BindEnv("circuit_breaker.half_open_after", "CIRCUIT_BREAKER_HALF_OPEN_AFTER")
}

func setupContainerDefaults() {
	if viper.GetString("database.dsn") == "" {
		viper.SetDefault("database.dsn", "postgres://postgres:postgres@postgres:5432/insider_messages?sslmode=disable")
	}

	if viper.GetString("redis.host") == "" {
		viper.SetDefault("redis.host", "redis")
	}
}

func isContainerEnvironment() bool {
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return true
	}

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	if os.Getenv("DATABASE_DSN") != "" || os.Getenv("REDIS_HOST") != "" {
		return true
	}

	return false
}
