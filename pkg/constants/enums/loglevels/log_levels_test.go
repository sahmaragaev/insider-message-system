package loglevels

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLogLevel_String(t *testing.T) {
	assert.Equal(t, "info", Info.String())
	assert.Equal(t, "debug", Debug.String())
}

func TestLogLevel_IsValid(t *testing.T) {
	assert.True(t, Info.IsValid())
	assert.True(t, Debug.IsValid())
	assert.False(t, LogLevel("invalid").IsValid())
}

func TestLogLevel_ToZapLevel(t *testing.T) {
	assert.Equal(t, zapcore.InfoLevel, Info.ToZapLevel())
	assert.Equal(t, zapcore.DebugLevel, Debug.ToZapLevel())
}

func TestFromString(t *testing.T) {
	assert.Equal(t, Info, FromString("info"))
	assert.Equal(t, Debug, FromString("debug"))
	assert.Equal(t, Info, FromString("unknown")) // Default case
}
