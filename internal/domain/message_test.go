package domain

import (
	"insider-message-system/pkg/constants/enums/messagestatus"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	to := "+905551111111"
	content := "Test message"

	message := NewMessage(to, content)

	assert.NotNil(t, message)
	assert.NotEqual(t, "", message.ID.String())
	assert.Equal(t, to, message.To)
	assert.Equal(t, content, message.Content)
	assert.Equal(t, messagestatus.Pending, message.Status)
	assert.WithinDuration(t, time.Now(), message.CreatedAt, time.Second)
	assert.Nil(t, message.SentAt)
	assert.Nil(t, message.MessageID)
	assert.Nil(t, message.FailureReason)
}

func TestMessage_MarkAsSent(t *testing.T) {
	message := NewMessage("+905551111111", "Test message")
	messageID := "test-message-id"

	beforeMark := time.Now()
	message.MarkAsSent(messageID)
	afterMark := time.Now()

	assert.Equal(t, messagestatus.Sent, message.Status)
	assert.NotNil(t, message.SentAt)
	assert.WithinDuration(t, beforeMark, *message.SentAt, afterMark.Sub(beforeMark))
	assert.NotNil(t, message.MessageID)
	assert.Equal(t, messageID, *message.MessageID)
	assert.Nil(t, message.FailureReason)
}

func TestMessage_MarkAsFailed(t *testing.T) {
	message := NewMessage("+905551111111", "Test message")
	reason := "Webhook timeout"

	message.MarkAsFailed(reason)

	assert.Equal(t, messagestatus.Failed, message.Status)
	assert.Nil(t, message.SentAt)
	assert.Nil(t, message.MessageID)
	assert.NotNil(t, message.FailureReason)
	assert.Equal(t, reason, *message.FailureReason)
}

func TestMessage_IsValidForSending(t *testing.T) {
	tests := []struct {
		name     string
		message  *Message
		expected bool
	}{
		{
			name:     "valid pending message",
			message:  NewMessage("+905551111111", "Valid message"),
			expected: true,
		},
		{
			name: "already sent message",
			message: &Message{
				Status:  messagestatus.Sent,
				Content: "Valid message",
			},
			expected: false,
		},
		{
			name: "failed message",
			message: &Message{
				Status:  messagestatus.Failed,
				Content: "Valid message",
			},
			expected: false,
		},
		{
			name: "message too long",
			message: &Message{
				Status:  messagestatus.Pending,
				Content: "This is a very long message that exceeds the 160 character limit. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			},
			expected: false,
		},
		{
			name: "message at character limit",
			message: &Message{
				Status:  messagestatus.Pending,
				Content: "This message is exactly 160 characters long to test the boundary condition. This should be valid for sending since it meets the character limit perfectly.",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.IsValidForSending()
			assert.Equal(t, tt.expected, result)
		})
	}
}
