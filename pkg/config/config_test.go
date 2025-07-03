package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSetupDefaults(t *testing.T) {
	viper.Reset()
	setupDefaults()
	assert.Equal(t, "8080", viper.GetString("server.port"))
	assert.Equal(t, "postgres", viper.GetString("database.driver"))
	assert.Equal(t, "localhost", viper.GetString("redis.host"))
	assert.Equal(t, "json", viper.GetString("logger.format"))
}

func TestIsContainerEnvironment(t *testing.T) {
	os.Setenv("DOCKER_CONTAINER", "true")
	assert.True(t, isContainerEnvironment())
	os.Unsetenv("DOCKER_CONTAINER")
}

func TestLoggerConfigStruct(t *testing.T) {
	cfg := LoggerConfig{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	}
	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
	assert.Equal(t, "stdout", cfg.OutputPath)
}
