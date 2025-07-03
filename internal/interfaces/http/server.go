package http

import (
	"context"
	"fmt"
	"insider-message-system/pkg/config"
	"insider-message-system/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server wraps the HTTP server and its configuration.
type Server struct {
	httpServer *http.Server
	config     config.ServerConfig
}

// NewServer creates a new HTTP server with the given configuration and routes.
func NewServer(config config.ServerConfig, routeConfig RouteConfig) (*Server, *gin.Engine) {
	if config.Host == "" {
		config.Host = "0.0.0.0"
	}
	if config.Port == "" {
		config.Port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	SetupRoutes(router, routeConfig)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &Server{
		httpServer: httpServer,
		config:     config,
	}, router
}

// Start runs the HTTP server and blocks until it stops or fails.
func (s *Server) Start() error {
	logger.Info("Starting HTTP server",
		zap.String("address", s.httpServer.Addr),
		zap.Duration("read_timeout", s.config.ReadTimeout),
		zap.Duration("write_timeout", s.config.WriteTimeout))

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the HTTP server with a timeout.
func (s *Server) Stop(ctx context.Context) error {
	logger.Info("Stopping HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	logger.Info("HTTP server stopped successfully")
	return nil
}

// Addr returns the address the server is listening on.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
