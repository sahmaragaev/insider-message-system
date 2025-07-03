//go:build integration
// +build integration

package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"insider-message-system/internal/application/services"
	"insider-message-system/internal/application/usecases"
	"insider-message-system/internal/domain"
	"insider-message-system/internal/infrastructure/database"
	"insider-message-system/internal/infrastructure/database/repos"
	apphttp "insider-message-system/internal/interfaces/http"
	"insider-message-system/internal/interfaces/http/handlers"
	"insider-message-system/pkg/config"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestServer(t *testing.T) (*httptest.Server, *database.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.AutoMigrate(&domain.Message{})

	dbWrapper := &database.DB{DB: db}

	messageRepo := repos.NewMessage(dbWrapper)
	messageService := services.NewMessage(messageRepo, nil, nil)
	sendMessageUC := usecases.NewSendMessageUseCase(messageService)
	getMessagesUC := usecases.NewGetMessagesUseCase(messageService)
	messageHandler := handlers.NewMessageHandler(sendMessageUC, getMessagesUC)

	routeConfig := apphttp.RouteConfig{
		MessageHandler: messageHandler,
	}

	_, router := apphttp.NewServer(config.ServerConfig{}, routeConfig)
	ts := httptest.NewServer(router)

	return ts, dbWrapper
}

func TestIntegration_CreateMessage(t *testing.T) {
	ts, db := setupTestServer(t)
	defer ts.Close()
	defer db.Close()

	payload := map[string]any{
		"to":      "+905551111111",
		"content": "Integration test message",
	}
	body, _ := json.Marshal(payload)

	resp, err := stdhttp.Post(fmt.Sprintf("%s/api/v1/messages", ts.URL), "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusCreated, resp.StatusCode)

	var response struct {
		Status bool   `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			ID      string `json:"id"`
			To      string `json:"to"`
			Content string `json:"content"`
			Status  string `json:"status"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.True(t, response.Status)
	require.Equal(t, payload["to"], response.Data.To)
	require.Equal(t, payload["content"], response.Data.Content)

	var count int64
	db.DB.Model(&domain.Message{}).Where(map[string]any{
		"to":      payload["to"],
		"content": payload["content"],
	}).Count(&count)
	require.Equal(t, int64(1), count, "message should be persisted in DB")
}

func TestIntegration_CreateMessage_InvalidPayload(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	payload := map[string]any{
		"to":      "", // missing required fields
		"content": "",
	}
	body, _ := json.Marshal(payload)

	resp, err := stdhttp.Post(fmt.Sprintf("%s/api/v1/messages", ts.URL), "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusBadRequest, resp.StatusCode)
}

func TestIntegration_CreateMessage_InvalidJSON(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	invalidJSON := `{"to": "+905551111111", "content": "missing quote}` // malformed JSON
	resp, err := stdhttp.Post(fmt.Sprintf("%s/api/v1/messages", ts.URL), "application/json", bytes.NewBufferString(invalidJSON))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusBadRequest, resp.StatusCode)
}

func TestIntegration_GetSentMessages(t *testing.T) {
	ts, db := setupTestServer(t)
	defer ts.Close()
	defer db.Close()

	// Insert a message directly
	db.DB.Create(&domain.Message{
		To:      "+905551111111",
		Content: "Integration test message",
		Status:  "sent",
	})

	resp, err := stdhttp.Get(fmt.Sprintf("%s/api/v1/messages/sent?page=1&limit=10", ts.URL))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)

	var response struct {
		Status bool   `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			Data []struct {
				To      string `json:"to"`
				Content string `json:"content"`
				Status  string `json:"status"`
			} `json:"data"`
			Pagination any `json:"pagination"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.True(t, response.Status)
	require.NotEmpty(t, response.Data.Data)
	require.Equal(t, "+905551111111", response.Data.Data[0].To)
}

func TestIntegration_GetSentMessages_InvalidQueryParams(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	resp, err := stdhttp.Get(fmt.Sprintf("%s/api/v1/messages/sent?page=abc&limit=xyz", ts.URL))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusBadRequest, resp.StatusCode)
}
