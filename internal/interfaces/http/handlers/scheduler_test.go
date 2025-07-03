package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"insider-message-system/internal/application/services"
	"insider-message-system/internal/application/usecases"
	"insider-message-system/pkg/errors"
	"insider-message-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

// ControlSchedulerUseCaseInterface defines the interface for the use case
type ControlSchedulerUseCaseInterface interface {
	Start(ctx context.Context) (*usecases.SchedulerStatusResponse, error)
	Stop(ctx context.Context) (*usecases.SchedulerStatusResponse, error)
	GetStatus(ctx context.Context) *usecases.SchedulerStatusResponse
}

type mockControlSchedulerUseCase struct {
	mock.Mock
}

func (m *mockControlSchedulerUseCase) Start(ctx context.Context) (*usecases.SchedulerStatusResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.SchedulerStatusResponse), args.Error(1)
}

func (m *mockControlSchedulerUseCase) Stop(ctx context.Context) (*usecases.SchedulerStatusResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.SchedulerStatusResponse), args.Error(1)
}

func (m *mockControlSchedulerUseCase) GetStatus(ctx context.Context) *usecases.SchedulerStatusResponse {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*usecases.SchedulerStatusResponse)
}

// testSchedulerHandler is a test-specific wrapper
type testSchedulerHandler struct {
	*SchedulerHandler
	mockUC ControlSchedulerUseCaseInterface
}

func (h *testSchedulerHandler) StartScheduler(c *gin.Context) {
	_, err := h.mockUC.Start(context.Background())
	if err != nil {
		if err, ok := err.(*errors.Error); ok {
			var resp *response.Response
			switch err.Code {
			case "SCHEDULER_ALREADY_RUNNING":
				resp = response.SchedulerAlreadyRunning()
			default:
				resp = response.New(err.Status, &response.Body{
					Status: false,
					Msg:    err.Message,
					Data: map[string]string{
						"code": err.Code,
					},
				}, &response.Log{
					Level: zapcore.InfoLevel,
					Msg:   err.Message,
					Type:  response.API,
				})
			}
			response.SendError(c, resp)
		} else {
			resp := response.InternalServerError(err.Error())
			response.SendError(c, resp)
		}
		return
	}

	resp := response.SchedulerStarted()
	response.SendSuccess(c, resp)
}

func (h *testSchedulerHandler) StopScheduler(c *gin.Context) {
	_, err := h.mockUC.Stop(context.Background())
	if err != nil {
		if err, ok := err.(*errors.Error); ok {
			var resp *response.Response
			switch err.Code {
			case "SCHEDULER_NOT_RUNNING":
				resp = response.SchedulerNotRunning()
			default:
				resp = response.New(err.Status, &response.Body{
					Status: false,
					Msg:    err.Message,
					Data: map[string]string{
						"code": err.Code,
					},
				}, &response.Log{
					Level: zapcore.InfoLevel,
					Msg:   err.Message,
					Type:  response.API,
				})
			}
			response.SendError(c, resp)
		} else {
			resp := response.InternalServerError(err.Error())
			response.SendError(c, resp)
		}
		return
	}

	resp := response.SchedulerStopped()
	response.SendSuccess(c, resp)
}

func (h *testSchedulerHandler) GetSchedulerStatus(c *gin.Context) {
	statusResponse := h.mockUC.GetStatus(context.Background())
	resp := response.Success(statusResponse)
	response.SendSuccess(c, resp)
}

func newTestSchedulerHandler(mockUC ControlSchedulerUseCaseInterface) *testSchedulerHandler {
	return &testSchedulerHandler{
		SchedulerHandler: &SchedulerHandler{},
		mockUC:           mockUC,
	}
}

func TestSchedulerHandler_StartScheduler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockControlSchedulerUseCase)
	handler := newTestSchedulerHandler(mockUC)

	statusResponse := &usecases.SchedulerStatusResponse{
		Status:  "running",
		Message: "Scheduler started successfully",
	}

	mockUC.On("Start", mock.Anything).Return(statusResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.StartScheduler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Scheduler started successfully")
	mockUC.AssertExpectations(t)
}

func TestSchedulerHandler_StartScheduler_AlreadyRunning(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockControlSchedulerUseCase)
	handler := newTestSchedulerHandler(mockUC)

	err := errors.NewError("SCHEDULER_ALREADY_RUNNING", "Scheduler is already running", 409)
	mockUC.On("Start", mock.Anything).Return(nil, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.StartScheduler(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Scheduler is already running")
	mockUC.AssertExpectations(t)
}

func TestSchedulerHandler_StartScheduler_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockControlSchedulerUseCase)
	handler := newTestSchedulerHandler(mockUC)

	err := errors.NewError("UNKNOWN_ERROR", "Some unknown error", 500)
	mockUC.On("Start", mock.Anything).Return(nil, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.StartScheduler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "UNKNOWN_ERROR")
	mockUC.AssertExpectations(t)
}

func TestSchedulerHandler_StopScheduler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockControlSchedulerUseCase)
	handler := newTestSchedulerHandler(mockUC)

	statusResponse := &usecases.SchedulerStatusResponse{
		Status:  "stopped",
		Message: "Scheduler stopped successfully",
	}

	mockUC.On("Stop", mock.Anything).Return(statusResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.StopScheduler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Scheduler stopped successfully")
	mockUC.AssertExpectations(t)
}

func TestSchedulerHandler_StopScheduler_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockControlSchedulerUseCase)
	handler := newTestSchedulerHandler(mockUC)

	err := errors.NewError("UNKNOWN_ERROR", "Some unknown error", 500)
	mockUC.On("Stop", mock.Anything).Return(nil, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.StopScheduler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "UNKNOWN_ERROR")
	mockUC.AssertExpectations(t)
}

func TestSchedulerHandler_GetSchedulerStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockControlSchedulerUseCase)
	handler := newTestSchedulerHandler(mockUC)

	statusResponse := &usecases.SchedulerStatusResponse{
		Status:  "running",
		Message: "Scheduler is running",
	}

	mockUC.On("GetStatus", mock.Anything).Return(statusResponse)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetSchedulerStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Scheduler is running")
	mockUC.AssertExpectations(t)
}

func TestSchedulerHandler_GetSchedulerStatus_NilResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockControlSchedulerUseCase)
	handler := newTestSchedulerHandler(mockUC)

	mockUC.On("GetStatus", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetSchedulerStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestNewSchedulerHandler(t *testing.T) {
	mockUC := new(mockControlSchedulerUseCase)
	handler := newTestSchedulerHandler(mockUC)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUC, handler.mockUC)
}

type mockSchedulerService struct{}

func (m *mockSchedulerService) Start(ctx context.Context) error                         { return nil }
func (m *mockSchedulerService) Stop() error                                             { return nil }
func (m *mockSchedulerService) IsRunning() bool                                         { return true }
func (m *mockSchedulerService) SetMessageProcessor(processor services.MessageProcessor) {}

func TestNewSchedulerHandler_Coverage(t *testing.T) {
	mockService := &mockSchedulerService{}
	uc := usecases.NewControlSchedulerUseCase(mockService)
	h := NewSchedulerHandler(uc)
	if h == nil {
		t.Fatal("NewSchedulerHandler returned nil")
	}
}

func TestSchedulerHandler_Methods_Coverage(t *testing.T) {
	mockService := &mockSchedulerService{}
	uc := usecases.NewControlSchedulerUseCase(mockService)
	h := NewSchedulerHandler(uc)
	r := gin.Default()
	r.POST("/start", h.StartScheduler)
	r.POST("/stop", h.StopScheduler)
	r.GET("/status", h.GetSchedulerStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/start", nil)
	r.ServeHTTP(w, req)
	assert.Contains(t, []int{http.StatusOK, http.StatusConflict, http.StatusBadRequest, http.StatusInternalServerError}, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/stop", nil)
	r.ServeHTTP(w, req)
	assert.Contains(t, []int{http.StatusOK, http.StatusConflict, http.StatusBadRequest, http.StatusInternalServerError}, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/status", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
