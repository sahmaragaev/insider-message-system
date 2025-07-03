package handlers

import (
	"insider-message-system/internal/application/usecases"
	"insider-message-system/internal/domain"
	_ "insider-message-system/pkg/apidocs"
	"insider-message-system/pkg/errors"
	"insider-message-system/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
)

// MessageHandler handles HTTP requests related to messages.
type MessageHandler struct {
	sendMessageUC *usecases.SendMessageUseCase
	getMessagesUC *usecases.GetMessagesUseCase
}

// NewMessageHandler creates a new MessageHandler with the provided use cases.
func NewMessageHandler(
	sendMessageUC *usecases.SendMessageUseCase,
	getMessagesUC *usecases.GetMessagesUseCase,
) *MessageHandler {
	return &MessageHandler{
		sendMessageUC: sendMessageUC,
		getMessagesUC: getMessagesUC,
	}
}

// CreateMessage handles POST /v1/messages requests to create a new message.
// @Summary Create a new message
// @Description Add a new message to the queue for sending
// @Tags messages
// @Accept json
// @Produce json
// @Param message body domain.MessageRequest true "Message data"
// @Success 201 {object} apidocs.MessageCreatedResponse
// @Failure 400 {object} apidocs.ErrorResponse
// @Failure 500 {object} apidocs.ErrorResponse
// @Router /v1/messages [post]
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var request domain.MessageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		resp := response.ValidationError(err.Error())
		response.SendError(c, resp)
		return
	}

	message, err := h.sendMessageUC.Execute(c.Request.Context(), request)
	if err != nil {
		if customErr, ok := err.(*errors.Error); ok {
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

// GetSentMessages handles GET /v1/messages/sent requests to retrieve sent messages with pagination.
// @Summary Get sent messages
// @Description Retrieve a paginated list of sent messages
// @Tags messages
// @Accept json
// @Produce json
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} apidocs.MessagesListResponse
// @Failure 400 {object} apidocs.ErrorResponse
// @Failure 500 {object} apidocs.ErrorResponse
// @Router /v1/messages/sent [get]
func (h *MessageHandler) GetSentMessages(c *gin.Context) {
	var request usecases.GetMessagesRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		resp := response.ValidationError(err.Error())
		response.SendError(c, resp)
		return
	}

	messagesResponse, err := h.getMessagesUC.Execute(c.Request.Context(), request)
	if err != nil {
		if customErr, ok := err.(*errors.Error); ok {
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
