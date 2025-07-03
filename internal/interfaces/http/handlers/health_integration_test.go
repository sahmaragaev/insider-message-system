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

func TestIntegration_HealthCheck(t *testing.T) {
	_, router := apphttp.NewServer(config.ServerConfig{}, apphttp.RouteConfig{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := stdhttp.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)

	var response struct {
		Status bool `json:"status"`
		Data   struct {
			Status  string `json:"status"`
			Service string `json:"service"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.True(t, response.Status)
	require.Equal(t, "ok", response.Data.Status)
	require.Equal(t, "insider-message-system", response.Data.Service)
}
