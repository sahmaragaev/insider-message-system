package handlers

import (
	"context"
	"insider-message-system/internal/application/usecases"
	_ "insider-message-system/pkg/apidocs"
	"insider-message-system/pkg/errors"
	"insider-message-system/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
)

// SchedulerHandler handles HTTP requests related to the message scheduler.
type SchedulerHandler struct {
	controlSchedulerUC *usecases.ControlSchedulerUseCase
}

// NewSchedulerHandler creates a new SchedulerHandler with the provided use case.
func NewSchedulerHandler(controlSchedulerUC *usecases.ControlSchedulerUseCase) *SchedulerHandler {
	return &SchedulerHandler{
		controlSchedulerUC: controlSchedulerUC,
	}
}

// @Summary Start the message scheduler
// @Description Start automatic message sending process
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} apidocs.SchedulerResponse
// @Failure 400 {object} apidocs.ErrorResponse
// @Failure 500 {object} apidocs.ErrorResponse
// @Router /v1/scheduler/start [post]
// StartScheduler handles POST /v1/scheduler/start to start the message scheduler.
func (h *SchedulerHandler) StartScheduler(c *gin.Context) {
	_, err := h.controlSchedulerUC.Start(context.Background())
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

// @Summary Stop the message scheduler
// @Description Stop automatic message sending process
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} apidocs.SchedulerResponse
// @Failure 400 {object} apidocs.ErrorResponse
// @Failure 500 {object} apidocs.ErrorResponse
// @Router /v1/scheduler/stop [post]
// StopScheduler handles POST /v1/scheduler/stop to stop the message scheduler.
func (h *SchedulerHandler) StopScheduler(c *gin.Context) {
	_, err := h.controlSchedulerUC.Stop(context.Background())
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

// @Summary Get scheduler status
// @Description Get current status of the message scheduler
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} apidocs.SchedulerResponse
// @Router /v1/scheduler/status [get]
// GetSchedulerStatus handles GET /v1/scheduler/status to get the scheduler's current status.
func (h *SchedulerHandler) GetSchedulerStatus(c *gin.Context) {
	statusResponse := h.controlSchedulerUC.GetStatus(context.Background())
	resp := response.Success(statusResponse)
	response.SendSuccess(c, resp)
}
