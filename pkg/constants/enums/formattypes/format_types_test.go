package formattypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatType_String(t *testing.T) {
	assert.Equal(t, "json", FormatJSON.String())
	assert.Equal(t, "console", FormatConsole.String())
}

func TestFormatType_IsValid(t *testing.T) {
	assert.True(t, FormatJSON.IsValid())
	assert.True(t, FormatConsole.IsValid())
	assert.False(t, FormatType("invalid").IsValid())
}

func TestFromString(t *testing.T) {
	assert.Equal(t, FormatJSON, FromString("json"))
	assert.Equal(t, FormatConsole, FromString("console"))
	assert.Equal(t, FormatJSON, FromString("unknown")) // Default case
}
