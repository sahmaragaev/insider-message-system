package http

import (
	"insider-message-system/internal/infrastructure/webhook"
	"insider-message-system/internal/interfaces/http/handlers"
	"insider-message-system/internal/interfaces/http/middleware"
	"insider-message-system/pkg/response"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouteConfig holds the dependencies for HTTP route handlers.
type RouteConfig struct {
	MessageHandler   *handlers.MessageHandler
	SchedulerHandler *handlers.SchedulerHandler
	WebhookClient    webhook.Client
	AuthKey          string
}

// SetupRoutes configures all HTTP routes and middleware for the application.
func SetupRoutes(router *gin.Engine, config RouteConfig) {
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) {
		resp := response.Success(map[string]any{
			"status":  "ok",
			"service": "insider-message-system",
		})
		response.SendSuccess(c, resp)
	})

	api := router.Group("/api/v1")

	messageRoutes := api.Group("/messages")
	{
		messageRoutes.POST("", config.MessageHandler.CreateMessage)
		messageRoutes.GET("/sent", config.MessageHandler.GetSentMessages)
	}

	schedulerRoutes := api.Group("/scheduler")
	{
		schedulerRoutes.POST("/start", config.SchedulerHandler.StartScheduler)
		schedulerRoutes.POST("/stop", config.SchedulerHandler.StopScheduler)
		schedulerRoutes.GET("/status", config.SchedulerHandler.GetSchedulerStatus)
	}

	api.GET("/circuit-breaker/status", func(c *gin.Context) {
		if config.WebhookClient == nil {
			resp := response.Success(map[string]any{
				"enabled": false,
				"message": "Circuit breaker not configured",
			})
			response.SendSuccess(c, resp)
			return
		}

		metrics := config.WebhookClient.GetCircuitBreakerMetrics()
		state := config.WebhookClient.GetCircuitBreakerState()

		resp := response.Success(map[string]any{
			"state":   state.String(),
			"metrics": metrics,
		})
		response.SendSuccess(c, resp)
	})
}
