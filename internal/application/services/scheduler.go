package services

import (
	"context"
	"insider-message-system/pkg/config"
	"insider-message-system/pkg/errors"
	"insider-message-system/pkg/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Scheduler defines the interface for the message scheduler service.
type Scheduler interface {
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
	SetMessageProcessor(processor MessageProcessor)
}

type scheduler struct {
	config    config.SchedulerConfig
	ticker    *time.Ticker
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   bool
	mu        sync.RWMutex
	processor MessageProcessor
}

// NewScheduler creates a new Scheduler with the given configuration.
func NewScheduler(config config.SchedulerConfig) Scheduler {
	return &scheduler{
		config: config,
	}
}

func (s *scheduler) SetMessageProcessor(processor MessageProcessor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processor = processor
}

func (s *scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		logger.Warn("Scheduler is already running")
		return nil
	}

	if s.processor == nil {
		logger.Error("Message processor not set")
		return errors.NewError("PROCESSOR_NOT_SET", "Message processor not set", 500)
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.ticker = time.NewTicker(s.config.Interval)
	s.running = true

	s.wg.Add(1)
	go s.run()

	logger.Info("Scheduler started successfully", zap.Duration("interval", s.config.Interval))
	return nil
}

func (s *scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		logger.Warn("Scheduler is not running")
		return nil
	}

	s.cancel()
	s.ticker.Stop()
	s.running = false

	logger.Info("Stopping scheduler...")
	s.wg.Wait()
	logger.Info("Scheduler stopped successfully")

	return nil
}

func (s *scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *scheduler) run() {
	defer s.wg.Done()

	logger.Info("Scheduler loop started")

	for {
		select {
		case <-s.ctx.Done():
			logger.Info("Scheduler context cancelled, stopping...")
			return

		case <-s.ticker.C:
			s.processWithRetry()
		}
	}
}

func (s *scheduler) processWithRetry() {
	for attempt := 1; attempt <= s.config.MaxRetries; attempt++ {
		err := s.processor.ProcessMessages(s.ctx)
		if err == nil {
			return
		}

		logger.Error("Message processing failed",
			zap.Error(err),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", s.config.MaxRetries))

		if attempt < s.config.MaxRetries {
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(s.config.RetryDelay):
				continue
			}
		}
	}

	logger.Error("All message processing attempts failed",
		zap.Int("max_retries", s.config.MaxRetries))
}
