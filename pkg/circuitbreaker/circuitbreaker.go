package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"insider-message-system/pkg/logger"

	"go.uber.org/zap"
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrCircuitHalfOpen = errors.New("circuit breaker is half-open")
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

type Config struct {
	FailureThreshold   int
	SuccessThreshold   int
	Timeout            time.Duration
	MaxConcurrentCalls int
}

type CircuitBreaker struct {
	config Config
	state  State
	mu     sync.RWMutex

	failureCount    int
	lastFailureTime time.Time

	successCount int

	concurrentCalls int

	totalCalls     int64
	totalFailures  int64
	totalSuccesses int64
	totalTimeouts  int64
}

func New(config Config) *CircuitBreaker {
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold <= 0 {
		config.SuccessThreshold = 3
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxConcurrentCalls <= 0 {
		config.MaxConcurrentCalls = 1
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.mu.Lock()
	currentState := cb.state
	cb.mu.Unlock()

	switch currentState {
	case StateClosed:
		return cb.executeClosed(ctx, fn)
	case StateHalfOpen:
		return cb.executeHalfOpen(ctx, fn)
	case StateOpen:
		return cb.executeOpen(ctx, fn)
	default:
		return fmt.Errorf("unknown circuit breaker state: %d", currentState)
	}
}

func (cb *CircuitBreaker) executeClosed(ctx context.Context, fn func() error) error {
	cb.mu.Lock()
	cb.totalCalls++
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()
		cb.totalFailures++

		logger.Warn("Circuit breaker failure in closed state",
			zap.Error(err),
			zap.Int("failure_count", cb.failureCount),
			zap.Int("threshold", cb.config.FailureThreshold))

		if cb.failureCount >= cb.config.FailureThreshold {
			cb.transitionToOpen()
		}
	} else {
		cb.failureCount = 0
		cb.totalSuccesses++
		logger.Debug("Circuit breaker success in closed state",
			zap.Int("failure_count", cb.failureCount))
	}

	return err
}

func (cb *CircuitBreaker) executeHalfOpen(ctx context.Context, fn func() error) error {
	cb.mu.Lock()
	if cb.concurrentCalls >= cb.config.MaxConcurrentCalls {
		cb.mu.Unlock()
		logger.Warn("Circuit breaker max concurrent calls reached in half-open state")
		return ErrCircuitHalfOpen
	}

	cb.concurrentCalls++
	cb.totalCalls++
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.concurrentCalls--

	if err != nil {
		cb.totalFailures++
		cb.transitionToOpen()
		logger.Warn("Circuit breaker failure in half-open state, transitioning to open",
			zap.Error(err))
	} else {
		cb.successCount++
		cb.totalSuccesses++
		logger.Info("Circuit breaker success in half-open state",
			zap.Int("success_count", cb.successCount),
			zap.Int("threshold", cb.config.SuccessThreshold))

		if cb.successCount >= cb.config.SuccessThreshold {
			cb.transitionToClosed()
		}
	}

	return err
}

func (cb *CircuitBreaker) executeOpen(ctx context.Context, fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
		cb.transitionToHalfOpen()
		logger.Info("Circuit breaker transitioning to half-open state after timeout")
		return cb.executeHalfOpen(ctx, fn)
	}

	cb.totalTimeouts++
	logger.Warn("Circuit breaker is open, request rejected",
		zap.Duration("time_since_last_failure", time.Since(cb.lastFailureTime)),
		zap.Duration("timeout", cb.config.Timeout))

	return ErrCircuitOpen
}

func (cb *CircuitBreaker) transitionToOpen() {
	if cb.state != StateOpen {
		cb.state = StateOpen
		cb.successCount = 0
		cb.concurrentCalls = 0
		logger.Warn("Circuit breaker transitioning to OPEN state",
			zap.Int("failure_count", cb.failureCount),
			zap.Duration("timeout", cb.config.Timeout))
	}
}

func (cb *CircuitBreaker) transitionToHalfOpen() {
	if cb.state != StateHalfOpen {
		cb.state = StateHalfOpen
		cb.successCount = 0
		cb.concurrentCalls = 0
		logger.Info("Circuit breaker transitioning to HALF_OPEN state")
	}
}

func (cb *CircuitBreaker) transitionToClosed() {
	if cb.state != StateClosed {
		cb.state = StateClosed
		cb.failureCount = 0
		cb.successCount = 0
		cb.concurrentCalls = 0
		logger.Info("Circuit breaker transitioning to CLOSED state")
	}
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) GetMetrics() map[string]any {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]any{
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"concurrent_calls":  cb.concurrentCalls,
		"total_calls":       cb.totalCalls,
		"total_failures":    cb.totalFailures,
		"total_successes":   cb.totalSuccesses,
		"total_timeouts":    cb.totalTimeouts,
		"last_failure_time": cb.lastFailureTime,
		"config": map[string]any{
			"failure_threshold":    cb.config.FailureThreshold,
			"success_threshold":    cb.config.SuccessThreshold,
			"timeout":              cb.config.Timeout,
			"max_concurrent_calls": cb.config.MaxConcurrentCalls,
		},
	}
}

func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionToOpen()
}

func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionToClosed()
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.concurrentCalls = 0
	cb.totalCalls = 0
	cb.totalFailures = 0
	cb.totalSuccesses = 0
	cb.totalTimeouts = 0
	cb.lastFailureTime = time.Time{}

	logger.Info("Circuit breaker reset")
}
