package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"insider-message-system/internal/application/services"
	"insider-message-system/internal/application/usecases"
	"insider-message-system/internal/infrastructure/database"
	"insider-message-system/internal/infrastructure/database/repos"
	"insider-message-system/internal/infrastructure/redis"
	"insider-message-system/internal/infrastructure/webhook"
	"insider-message-system/internal/interfaces/http"
	"insider-message-system/internal/interfaces/http/handlers"
	"insider-message-system/pkg/config"
	"insider-message-system/pkg/logger"

	_ "insider-message-system/docs"

	"go.uber.org/zap"
)

// @title Insider Message System API
// @version 1.0
// @description Automatic message sending system
// @host localhost:8080
// @BasePath /api
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	if err := logger.Init(cfg.Logger); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting Insider Message System")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	var cacheService redis.CacheService
	if cfg.Redis.Host != "" {
		cacheService, err = redis.NewCacheService(cfg.Redis)
		if err != nil {
			logger.Warn("Failed to connect to Redis, continuing without cache", zap.Error(err))
		} else {
			logger.Info("Redis cache service initialized")
		}
	}

	messageRepo := repos.NewMessage(db)
	webhookService := webhook.NewClient(cfg.Webhook, cfg.CircuitBreaker)
	messageService := services.NewMessage(messageRepo, webhookService, cacheService)
	schedulerService := services.NewScheduler(cfg.Scheduler)

	processor := services.NewMessageProcessor(messageService, cfg.Scheduler.BatchSize)
	schedulerService.SetMessageProcessor(processor)

	sendMessageUC := usecases.NewSendMessageUseCase(messageService)
	getMessagesUC := usecases.NewGetMessagesUseCase(messageService)
	controlSchedulerUC := usecases.NewControlSchedulerUseCase(schedulerService)

	messageHandler := handlers.NewMessageHandler(sendMessageUC, getMessagesUC)
	schedulerHandler := handlers.NewSchedulerHandler(controlSchedulerUC)

	routeConfig := http.RouteConfig{
		MessageHandler:   messageHandler,
		SchedulerHandler: schedulerHandler,
		WebhookClient:    webhookService,
		AuthKey:          cfg.Webhook.AuthKey,
	}

	server, _ := http.NewServer(cfg.Server, routeConfig)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(); err != nil {
			logger.Error("Server failed to start", zap.Error(err))
			cancel()
		}
	}()

	if cfg.Scheduler.AutoStart {
		logger.Info("Auto-starting scheduler")
		if err := schedulerService.Start(ctx); err != nil {
			logger.Error("Failed to auto-start scheduler", zap.Error(err))
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Received shutdown signal")
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	logger.Info("Shutting down application...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if schedulerService.IsRunning() {
		logger.Info("Stopping scheduler...")
		if err := schedulerService.Stop(); err != nil {
			logger.Error("Failed to stop scheduler", zap.Error(err))
		}
	}

	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Failed to stop server", zap.Error(err))
	}

	cancel()
	wg.Wait()

	logger.Info("Application shutdown complete")
}
