package response

import (
	"insider-message-system/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Send(c *gin.Context, resp *Response) {
	switch resp.Log.Level {
	case zap.ErrorLevel:
		logger.Error(resp.Log.Msg)
	case zap.WarnLevel:
		logger.Warn(resp.Log.Msg)
	case zap.InfoLevel:
		logger.Info(resp.Log.Msg)
	case zap.DebugLevel:
		logger.Debug(resp.Log.Msg)
	}

	c.JSON(resp.Code, resp.Body)
}

func SendError(c *gin.Context, resp *Response) {
	logger.Error(resp.Log.Msg, zap.Int("status_code", resp.Code))

	c.JSON(resp.Code, resp.Body)
}

func SendSuccess(c *gin.Context, resp *Response) {
	logger.Info(resp.Log.Msg, zap.Int("status_code", resp.Code))

	c.JSON(resp.Code, resp.Body)
}
