package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"insider-message-system/internal/application/services"
	"insider-message-system/internal/application/usecases"
	"insider-message-system/internal/domain"
	"insider-message-system/internal/interfaces/http/handlers"
	"insider-message-system/pkg/circuitbreaker"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock SchedulerService for safe ControlSchedulerUseCase
type mockSchedulerService struct{}

func (m *mockSchedulerService) Start(ctx context.Context) error                         { return nil }
func (m *mockSchedulerService) Stop() error                                             { return nil }
func (m *mockSchedulerService) IsRunning() bool                                         { return true }
func (m *mockSchedulerService) SetMessageProcessor(processor services.MessageProcessor) {}

// Mock webhook client for testing
type mockWebhookClient struct{}

func (m *mockWebhookClient) SendMessage(ctx context.Context, request domain.WebhookRequest) (*domain.MessageResponse, error) {
	return nil, nil
}

func (m *mockWebhookClient) GetCircuitBreakerMetrics() map[string]any {
	return map[string]any{
		"state":   "OPEN",
		"enabled": true,
	}
}

func (m *mockWebhookClient) GetCircuitBreakerState() circuitbreaker.State {
	return circuitbreaker.StateOpen
}

func newMockSchedulerHandler() *handlers.SchedulerHandler {
	mockService := &mockSchedulerService{}
	uc := usecases.NewControlSchedulerUseCase(mockService)
	return handlers.NewSchedulerHandler(uc)
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: newMockSchedulerHandler(),
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	SetupRoutes(router, routeConfig)

	// Test that routes are registered
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "insider-message-system")
}

func TestSetupRoutes_WithWebhookClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock webhook client
	mockWebhookClient := &mockWebhookClient{}

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: newMockSchedulerHandler(),
		WebhookClient:    mockWebhookClient,
		AuthKey:          "test-key",
	}

	SetupRoutes(router, routeConfig)

	// Test circuit breaker status endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/circuit-breaker/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "OPEN")
}

func TestSetupRoutes_WithoutWebhookClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: newMockSchedulerHandler(),
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	SetupRoutes(router, routeConfig)

	// Test circuit breaker status endpoint without webhook client
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/circuit-breaker/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "not configured")
}

func TestSetupRoutes_MessageRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: newMockSchedulerHandler(),
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	SetupRoutes(router, routeConfig)

	// Test message routes exist (they will return 404 since handlers are not properly mocked)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/messages", nil)
	router.ServeHTTP(w, req)

	// Should not be 404 (route exists) but might be 400 due to missing body
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestSetupRoutes_SchedulerRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	routeConfig := RouteConfig{
		MessageHandler:   &handlers.MessageHandler{},
		SchedulerHandler: newMockSchedulerHandler(),
		WebhookClient:    nil,
		AuthKey:          "test-key",
	}

	SetupRoutes(router, routeConfig)

	// Test scheduler routes exist
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scheduler/status", nil)
	router.ServeHTTP(w, req)

	// Should not be 404 (route exists)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
