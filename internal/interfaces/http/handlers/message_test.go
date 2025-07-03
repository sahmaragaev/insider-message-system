package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"insider-message-system/internal/application/usecases"
	"insider-message-system/internal/domain"
	customerrors "insider-message-system/pkg/errors"
	"insider-message-system/pkg/response"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

type mockMessageService struct {
	mock.Mock
}

func (m *mockMessageService) GetSentMessages(ctx context.Context, page, pageSize int) ([]*domain.Message, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*domain.Message), args.Get(1).(int64), args.Error(2)
}

func (m *mockMessageService) CreateMessage(ctx context.Context, req domain.MessageRequest) (*domain.Message, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *mockMessageService) SendPendingMessages(ctx context.Context, limit int) error {
	return nil
}

type mockSendMessageUseCase struct {
	mock.Mock
}

func (m *mockSendMessageUseCase) Execute(ctx context.Context, req domain.MessageRequest) (*domain.Message, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.Message), args.Error(1)
}

type nilErrorUseCase struct{}

func (n *nilErrorUseCase) Execute(ctx context.Context, req domain.MessageRequest) (*domain.Message, error) {
	return nil, nil
}

type nilErrorGetMessagesUC struct{}

func (n *nilErrorGetMessagesUC) Execute(ctx context.Context, req usecases.GetMessagesRequest) (any, error) {
	return nil, nil
}

type nilMessageService struct{}

func (n *nilMessageService) CreateMessage(ctx context.Context, req domain.MessageRequest) (*domain.Message, error) {
	return nil, nil
}

func (n *nilMessageService) GetSentMessages(ctx context.Context, page, pageSize int) ([]*domain.Message, int64, error) {
	return nil, 0, nil
}

func (n *nilMessageService) SendPendingMessages(ctx context.Context, limit int) error {
	return nil
}

type emptyError struct{}

func (e *emptyError) Error() string {
	return ""
}

type nilSendMessageUseCase struct{}

func (n *nilSendMessageUseCase) Execute(ctx context.Context, req domain.MessageRequest) (*domain.Message, error) {
	return nil, &emptyError{}
}

type nilGetMessagesUseCase struct{}

func (n *nilGetMessagesUseCase) Execute(ctx context.Context, req usecases.GetMessagesRequest) (*usecases.GetMessagesResponse, error) {
	return nil, &emptyError{}
}

// Test-specific interfaces for mocking
type SendMessageUseCaseInterface interface {
	Execute(ctx context.Context, req domain.MessageRequest) (*domain.Message, error)
}

type GetMessagesUseCaseInterface interface {
	Execute(ctx context.Context, req usecases.GetMessagesRequest) (*usecases.GetMessagesResponse, error)
}

// Test-specific message handler
type testMessageHandler struct {
	*MessageHandler
	sendMessageUC SendMessageUseCaseInterface
	getMessagesUC GetMessagesUseCaseInterface
}

func (h *testMessageHandler) CreateMessage(c *gin.Context) {
	var request domain.MessageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		resp := response.ValidationError(err.Error())
		response.SendError(c, resp)
		return
	}

	message, err := h.sendMessageUC.Execute(c.Request.Context(), request)
	if err != nil {
		if customErr, ok := err.(*customerrors.Error); ok {
			resp := response.New(customErr.Status, &response.Body{
				Status: false,
				Msg:    customErr.Message,
				Data: map[string]string{
					"code": customErr.Code,
				},
			}, &response.Log{
				Level: zapcore.InfoLevel,
				Msg:   customErr.Message,
				Type:  response.API,
			})
			response.SendError(c, resp)
		} else if stdErr, ok := err.(error); ok && stdErr.Error() != "" {
			resp := response.InternalServerError(stdErr.Error())
			response.SendError(c, resp)
		} else {
			resp := response.InternalServerError("Unknown error (nil)")
			response.SendError(c, resp)
		}
		return
	}

	resp := response.MessageCreated(message)
	response.SendSuccess(c, resp)
}

func (h *testMessageHandler) GetSentMessages(c *gin.Context) {
	var request usecases.GetMessagesRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		resp := response.ValidationError(err.Error())
		response.SendError(c, resp)
		return
	}

	messagesResponse, err := h.getMessagesUC.Execute(c.Request.Context(), request)
	if err != nil {
		if customErr, ok := err.(*customerrors.Error); ok {
			resp := response.New(customErr.Status, &response.Body{
				Status: false,
				Msg:    customErr.Message,
				Data: map[string]string{
					"code": customErr.Code,
				},
			}, &response.Log{
				Level: zapcore.InfoLevel,
				Msg:   customErr.Message,
				Type:  response.API,
			})
			response.SendError(c, resp)
		} else if stdErr, ok := err.(error); ok && stdErr.Error() != "" {
			resp := response.InternalServerError(stdErr.Error())
			response.SendError(c, resp)
		} else {
			resp := response.InternalServerError("Unknown error (nil)")
			response.SendError(c, resp)
		}
		return
	}

	resp := response.Success(messagesResponse)
	response.SendSuccess(c, resp)
}

func newTestMessageHandler(sendUC SendMessageUseCaseInterface, getUC GetMessagesUseCaseInterface) *testMessageHandler {
	return &testMessageHandler{
		MessageHandler: &MessageHandler{},
		sendMessageUC:  sendUC,
		getMessagesUC:  getUC,
	}
}

func TestGetSentMessages_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockSvc := new(mockMessageService)
	uc := usecases.NewGetMessagesUseCase(mockSvc)
	h := &MessageHandler{getMessagesUC: uc}

	messages := []*domain.Message{
		{ID: uuid.New(), Content: "Hello"},
	}
	total := int64(1)
	mockSvc.On("GetSentMessages", mock.Anything, 1, 10).Return(messages, total, nil)

	r.GET("/v1/messages/sent", h.GetSentMessages)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/messages/sent?page=1&limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.True(t, body["status"].(bool))
	assert.NotNil(t, body["data"])
	mockSvc.AssertExpectations(t)
}

func TestGetSentMessages_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockSvc := new(mockMessageService)
	uc := usecases.NewGetMessagesUseCase(mockSvc)
	h := &MessageHandler{getMessagesUC: uc}

	mockSvc.On("GetSentMessages", mock.Anything, 1, 10).Return([]*domain.Message{}, int64(0), errors.New("fail"))

	r.GET("/v1/messages/sent", h.GetSentMessages)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/messages/sent?page=1&limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("Response is not valid JSON: %v. Body: %s", err, w.Body.String())
	}
	status, statusOk := body["status"].(bool)
	msg, msgOk := body["msg"].(string)
	assert.True(t, statusOk, "status field missing or not a bool")
	assert.False(t, status, "status should be false on error")
	assert.True(t, msgOk, "msg field missing or not a string")
	assert.Equal(t, "Something went wrong", msg)
	mockSvc.AssertExpectations(t)
}

func TestGetSentMessages_InvalidQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockSvc := new(mockMessageService)
	uc := usecases.NewGetMessagesUseCase(mockSvc)
	h := &MessageHandler{getMessagesUC: uc}

	r.GET("/v1/messages/sent", h.GetSentMessages)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/messages/sent?page=abc&limit=xyz", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.False(t, body["status"].(bool))
	assert.NotNil(t, body["msg"])
}

func TestCreateMessage_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockService := new(mockMessageService)
	mockGetUC := new(mockMessageService)
	sendUC := usecases.NewSendMessageUseCase(mockService)
	h := &MessageHandler{sendMessageUC: sendUC, getMessagesUC: usecases.NewGetMessagesUseCase(mockGetUC)}

	msgReq := domain.MessageRequest{To: "+1234567890", Content: "Hello"}
	msg := &domain.Message{ID: uuid.New(), To: msgReq.To, Content: msgReq.Content}
	mockService.On("CreateMessage", mock.Anything, msgReq).Return(msg, nil)

	r.POST("/v1/messages", h.CreateMessage)
	w := httptest.NewRecorder()
	body, _ := json.Marshal(msgReq)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp["status"].(bool))
	assert.NotNil(t, resp["data"])
	mockService.AssertExpectations(t)
}

func TestCreateMessage_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockService := new(mockMessageService)
	mockGetUC := new(mockMessageService)
	sendUC := usecases.NewSendMessageUseCase(mockService)
	h := &MessageHandler{sendMessageUC: sendUC, getMessagesUC: usecases.NewGetMessagesUseCase(mockGetUC)}

	r.POST("/v1/messages", h.CreateMessage)
	w := httptest.NewRecorder()
	// Missing 'to' and 'content' fields
	invalidBody := []byte(`{"foo":"bar"}`)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["status"].(bool))
	assert.NotNil(t, resp["msg"])
}

func TestCreateMessage_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockService := new(mockMessageService)
	mockGetUC := new(mockMessageService)
	sendUC := usecases.NewSendMessageUseCase(mockService)
	h := &MessageHandler{sendMessageUC: sendUC, getMessagesUC: usecases.NewGetMessagesUseCase(mockGetUC)}

	msgReq := domain.MessageRequest{To: "+1234567890", Content: "Hello"}
	mockService.On("CreateMessage", mock.Anything, msgReq).Return((*domain.Message)(nil), errors.New("fail"))

	r.POST("/v1/messages", h.CreateMessage)
	w := httptest.NewRecorder()
	body, _ := json.Marshal(msgReq)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["status"].(bool))
	assert.NotNil(t, resp["msg"])
	mockService.AssertExpectations(t)
}

func TestCreateMessage_UnknownErrorNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := newTestMessageHandler(&nilSendMessageUseCase{}, &nilGetMessagesUseCase{})

	msgReq := domain.MessageRequest{To: "+1234567890", Content: "Hello"}

	r.POST("/v1/messages", h.CreateMessage)
	w := httptest.NewRecorder()
	body, _ := json.Marshal(msgReq)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["status"].(bool))
	assert.Equal(t, "Something went wrong", resp["msg"])
}

func TestGetSentMessages_UnknownErrorNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := newTestMessageHandler(nil, &nilGetMessagesUseCase{})

	r.GET("/v1/messages/sent", h.GetSentMessages)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/messages/sent?page=1&limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["status"].(bool))
	assert.Equal(t, "Something went wrong", resp["msg"])
}

func TestNewMessageHandler_Coverage(t *testing.T) {
	mockService := new(mockMessageService)
	sendUC := usecases.NewSendMessageUseCase(mockService)
	getUC := usecases.NewGetMessagesUseCase(mockService)
	h := NewMessageHandler(sendUC, getUC)
	if h == nil {
		t.Fatal("NewMessageHandler returned nil")
	}
}
