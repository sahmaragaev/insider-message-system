package http

import (
	"context"
	"testing"
	"time"

	"insider-message-system/internal/interfaces/http/handlers"
	"insider-message-system/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.ServerConfig{
		Host:         "localhost",
		Port:         "8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: &handlers.SchedulerHandler{},
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	server, router := NewServer(cfg, routeConfig)

	assert.NotNil(t, server)
	assert.NotNil(t, router)
	assert.Equal(t, "localhost:8080", server.Addr())
	assert.Equal(t, cfg, server.config)
}

func TestNewServer_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.ServerConfig{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: &handlers.SchedulerHandler{},
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	server, _ := NewServer(cfg, routeConfig)

	assert.Equal(t, "0.0.0.0:8080", server.Addr())
}

func TestServer_Addr(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.ServerConfig{
		Host:         "127.0.0.1",
		Port:         "9090",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: &handlers.SchedulerHandler{},
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	server, _ := NewServer(cfg, routeConfig)

	assert.Equal(t, "127.0.0.1:9090", server.Addr())
}

func TestServer_Stop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.ServerConfig{
		Host:         "localhost",
		Port:         "0", // random port
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: &handlers.SchedulerHandler{},
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	server, _ := NewServer(cfg, routeConfig)

	// Start server in background
	go func() {
		_ = server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	ctx := context.Background()
	err := server.Stop(ctx)

	assert.NoError(t, err)
}

func TestServer_Stop_WithTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.ServerConfig{
		Host:         "localhost",
		Port:         "0", // random port
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: &handlers.SchedulerHandler{},
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	server, _ := NewServer(cfg, routeConfig)

	// Start server in background
	go func() {
		_ = server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Stop(ctx)

	assert.NoError(t, err)
}
