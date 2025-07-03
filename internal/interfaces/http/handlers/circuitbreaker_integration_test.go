//go:build integration
// +build integration

package handlers_test

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	apphttp "insider-message-system/internal/interfaces/http"
	"insider-message-system/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestIntegration_CircuitBreakerStatus(t *testing.T) {
	_, router := apphttp.NewServer(config.ServerConfig{}, apphttp.RouteConfig{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := stdhttp.Get(ts.URL + "/api/v1/circuit-breaker/status")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)

	var response struct {
		Status bool `json:"status"`
		Data   struct {
			Enabled bool   `json:"enabled"`
			Message string `json:"message"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.True(t, response.Status)
	require.False(t, response.Data.Enabled)
	require.Contains(t, response.Data.Message, "not configured")
}
