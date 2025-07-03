package messagestatus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageStatus_String(t *testing.T) {
	assert.Equal(t, "pending", Pending.String())
	assert.Equal(t, "sent", Sent.String())
}

func TestMessageStatus_IsValid(t *testing.T) {
	assert.True(t, Pending.IsValid())
	assert.True(t, Sent.IsValid())
	assert.False(t, MessageStatus("invalid").IsValid())
}

func TestFromString(t *testing.T) {
	assert.Equal(t, Pending, FromString("pending"))
	assert.Equal(t, Sent, FromString("sent"))
	assert.Equal(t, Pending, FromString("unknown")) // Default case
}
