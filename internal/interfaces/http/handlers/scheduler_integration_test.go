//go:build integration
// +build integration

package handlers_test

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"insider-message-system/internal/application/services"
	"insider-message-system/internal/application/usecases"
	"insider-message-system/internal/domain"
	"insider-message-system/internal/infrastructure/database"
	apphttp "insider-message-system/internal/interfaces/http"
	"insider-message-system/internal/interfaces/http/handlers"
	"insider-message-system/pkg/config"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockProcessor struct{}

func (m *mockProcessor) ProcessMessages(ctx context.Context) error {
	return nil
}

func setupSchedulerTestServer(t *testing.T) *httptest.Server {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	dbWrapper := &database.DB{DB: db}
	dbWrapper.AutoMigrate(&domain.Message{})

	schedulerService := services.NewScheduler(config.SchedulerConfig{
		Interval: 100 * time.Millisecond,
	})
	schedulerService.SetMessageProcessor(&mockProcessor{})
	controlSchedulerUC := usecases.NewControlSchedulerUseCase(schedulerService)
	schedulerHandler := handlers.NewSchedulerHandler(controlSchedulerUC)

	routeConfig := apphttp.RouteConfig{
		SchedulerHandler: schedulerHandler,
	}
	_, router := apphttp.NewServer(config.ServerConfig{}, routeConfig)
	return httptest.NewServer(router)
}

func TestIntegration_SchedulerStatus(t *testing.T) {
	ts := setupSchedulerTestServer(t)
	defer ts.Close()

	resp, err := stdhttp.Get(ts.URL + "/api/v1/scheduler/status")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)

	var response struct {
		Status bool   `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.True(t, response.Status)
	require.Contains(t, []string{"running", "stopped"}, response.Data.Status)
}

func TestIntegration_SchedulerStartStop(t *testing.T) {
	ts := setupSchedulerTestServer(t)
	defer ts.Close()

	// Start scheduler
	resp, err := stdhttp.Post(ts.URL+"/api/v1/scheduler/start", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)

	// Stop scheduler
	resp2, err := stdhttp.Post(ts.URL+"/api/v1/scheduler/stop", "application/json", nil)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, stdhttp.StatusOK, resp2.StatusCode)
}

func TestIntegration_Scheduler_DoubleStartStop(t *testing.T) {
	ts := setupSchedulerTestServer(t)
	defer ts.Close()

	// Start scheduler once
	resp, err := stdhttp.Post(ts.URL+"/api/v1/scheduler/start", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)

	// Start scheduler again (should fail or return conflict)
	resp2, err := stdhttp.Post(ts.URL+"/api/v1/scheduler/start", "application/json", nil)
	require.NoError(t, err)
	defer resp2.Body.Close()
	// Acceptable: 409 Conflict, 400 Bad Request, or 500 Internal Server Error
	require.Contains(t, []int{stdhttp.StatusConflict, stdhttp.StatusBadRequest, stdhttp.StatusInternalServerError}, resp2.StatusCode)

	// Stop scheduler once
	resp3, err := stdhttp.Post(ts.URL+"/api/v1/scheduler/stop", "application/json", nil)
	require.NoError(t, err)
	defer resp3.Body.Close()
	require.Equal(t, stdhttp.StatusOK, resp3.StatusCode)

	// Stop scheduler again (should fail or return conflict)
	resp4, err := stdhttp.Post(ts.URL+"/api/v1/scheduler/stop", "application/json", nil)
	require.NoError(t, err)
	defer resp4.Body.Close()
	require.Contains(t, []int{stdhttp.StatusConflict, stdhttp.StatusBadRequest, stdhttp.StatusInternalServerError}, resp4.StatusCode)
}
