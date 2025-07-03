package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.NotNil(t, cfg)
}

func TestNewClient(t *testing.T) {
	cfg := DefaultConfig()
	client := NewClient(cfg)
	assert.NotNil(t, client)
}

func TestNewClientWithAuth(t *testing.T) {
	cfg := DefaultConfig()
	client := NewClientWithAuth("key", cfg)
	assert.NotNil(t, client)
}

func TestNewClientWithCustomAuth(t *testing.T) {
	cfg := DefaultConfig()
	client := NewClientWithCustomAuth("header", "value", cfg)
	assert.NotNil(t, client)
}
