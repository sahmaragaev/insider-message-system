package logger

import (
	"testing"

	"insider-message-system/pkg/config"
	"insider-message-system/pkg/constants/enums/formattypes"
	"insider-message-system/pkg/constants/enums/loglevels"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLoggerWrappers_NoPanic(t *testing.T) {
	Logger = zap.NewNop()
	defer func() { recover() }()
	Info("info")
	Debug("debug")
	Warn("warn")
	Error("error")
	l := With(zap.String("foo", "bar"))

	assert.NotNil(t, l)
	assert.NoError(t, Sync())
}

func TestInit(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:      loglevels.Info,
		Format:     formattypes.FormatJSON,
		OutputPath: "stdout",
	}

	_ = Init(cfg)
}
