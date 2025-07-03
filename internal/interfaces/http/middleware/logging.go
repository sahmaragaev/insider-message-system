package middleware

import (
	"insider-message-system/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("body_size", bodySize),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		if statusCode >= 500 {
			logger.Error("HTTP request error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request warning", fields...)
		} else {
			logger.Info("HTTP request", fields...)
		}
	}
}
