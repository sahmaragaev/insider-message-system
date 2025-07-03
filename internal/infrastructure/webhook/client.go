package webhook

import (
	"context"
	"fmt"
	"insider-message-system/internal/domain"
	"insider-message-system/pkg/circuitbreaker"
	"insider-message-system/pkg/config"
	"insider-message-system/pkg/errors"
	httppkg "insider-message-system/pkg/http"
	"insider-message-system/pkg/logger"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Client interface {
	SendMessage(ctx context.Context, request domain.WebhookRequest) (*domain.MessageResponse, error)
	GetCircuitBreakerMetrics() map[string]any
	GetCircuitBreakerState() circuitbreaker.State
}

type restyRequester interface {
	R() restyRequest
}

type restyRequest interface {
	SetContext(ctx context.Context) restyRequest
	SetBody(body any) restyRequest
	SetResult(result any) restyRequest
	Post(url string) (*resty.Response, error)
	Get(url string) (*resty.Response, error)
}

type realRestyRequester struct {
	client *resty.Client
}

func (r *realRestyRequester) R() restyRequest {
	return &realRestyRequest{req: r.client.R()}
}

type realRestyRequest struct {
	req *resty.Request
}

func (r *realRestyRequest) SetContext(ctx context.Context) restyRequest {
	r.req = r.req.SetContext(ctx)
	return r
}

func (r *realRestyRequest) SetBody(body any) restyRequest {
	r.req = r.req.SetBody(body)
	return r
}

func (r *realRestyRequest) SetResult(result any) restyRequest {
	r.req = r.req.SetResult(result)
	return r
}

func (r *realRestyRequest) Post(url string) (*resty.Response, error) {
	return r.req.Post(url)
}

func (r *realRestyRequest) Get(url string) (*resty.Response, error) {
	return r.req.Get(url)
}

type CircuitBreaker interface {
	Execute(ctx context.Context, fn func() error) error
	GetMetrics() map[string]any
	GetState() circuitbreaker.State
}

type client struct {
	client         restyRequester
	config         config.WebhookConfig
	circuitBreaker CircuitBreaker
}

func NewClient(cfg config.WebhookConfig, cbConfig config.CircuitBreakerConfig) Client {
	clientConfig := &httppkg.ClientConfig{
		Timeout:          cfg.Timeout,
		RetryCount:       3,
		RetryWaitTime:    1 * time.Second,
		RetryMaxWaitTime: 5 * time.Second,
		DefaultHeaders: map[string]string{
			"Content-Type": "application/json",
		},
		RetryCondition: func(r *resty.Response, err error) bool {
			return err != nil || r.StatusCode() >= 500
		},
	}

	retryClient := httppkg.NewClientWithCustomAuth("x-ins-auth-key", cfg.AuthKey, clientConfig)

	var cb CircuitBreaker
	if cbConfig.Enabled {
		cb = circuitbreaker.New(circuitbreaker.Config{
			FailureThreshold:   5,
			SuccessThreshold:   3,
			Timeout:            cbConfig.HalfOpenAfter,
			MaxConcurrentCalls: 1,
		})
		logger.Info("Circuit breaker enabled for webhook client",
			zap.Duration("timeout", cbConfig.HalfOpenAfter))
	}

	return &client{
		client:         &realRestyRequester{client: retryClient},
		config:         cfg,
		circuitBreaker: cb,
	}
}

func (w *client) SendMessage(ctx context.Context, request domain.WebhookRequest) (*domain.MessageResponse, error) {
	logger.Debug("Sending webhook request",
		zap.String("url", w.config.URL),
		zap.String("to", request.To),
		zap.String("content", request.Content))

	if w.circuitBreaker == nil {
		return w.sendMessageDirect(ctx, request)
	}

	var response *domain.MessageResponse
	var err error

	err = w.circuitBreaker.Execute(ctx, func() error {
		response, err = w.sendMessageDirect(ctx, request)
		return err
	})

	if err != nil {
		if err == circuitbreaker.ErrCircuitOpen {
			logger.Warn("Circuit breaker is open, webhook request rejected",
				zap.String("url", w.config.URL),
				zap.String("to", request.To))
			return nil, errors.NewErrorWithDetails(
				"WEBHOOK_CIRCUIT_OPEN",
				"Webhook service is temporarily unavailable",
				"Circuit breaker is open due to repeated failures",
				http.StatusServiceUnavailable,
			)
		}
		if err == circuitbreaker.ErrCircuitHalfOpen {
			logger.Warn("Circuit breaker is half-open, webhook request rejected",
				zap.String("url", w.config.URL),
				zap.String("to", request.To))
			return nil, errors.NewErrorWithDetails(
				"WEBHOOK_CIRCUIT_HALF_OPEN",
				"Webhook service is testing recovery",
				"Circuit breaker is in half-open state",
				http.StatusServiceUnavailable,
			)
		}
		return nil, err
	}

	return response, nil
}

func (w *client) sendMessageDirect(ctx context.Context, request domain.WebhookRequest) (*domain.MessageResponse, error) {
	var response domain.MessageResponse
	resp, err := w.client.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&response).
		Post(w.config.URL)

	if err != nil {
		logger.Error("Failed to send webhook request", zap.Error(err))
		return nil, errors.WrapError(err, "WEBHOOK_ERROR", "Failed to send request", http.StatusInternalServerError)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		logger.Error("Webhook request failed",
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response", resp.String()))
		return nil, errors.NewErrorWithDetails(
			"WEBHOOK_ERROR",
			fmt.Sprintf("Webhook request failed with status %d", resp.StatusCode()),
			resp.String(),
			http.StatusInternalServerError,
		)
	}

	logger.Info("Webhook request successful",
		zap.String("to", request.To),
		zap.String("message_id", response.MessageID),
		zap.Int("status_code", resp.StatusCode()))

	return &response, nil
}

func (w *client) GetCircuitBreakerMetrics() map[string]any {
	if w.circuitBreaker == nil {
		return map[string]any{
			"enabled": false,
			"message": "Circuit breaker is not enabled",
		}
	}
	return w.circuitBreaker.GetMetrics()
}

func (w *client) GetCircuitBreakerState() circuitbreaker.State {
	if w.circuitBreaker == nil {
		return circuitbreaker.StateClosed
	}
	return w.circuitBreaker.GetState()
}

func (w *client) HealthCheck(ctx context.Context) error {
	resp, err := w.client.R().
		SetContext(ctx).
		Get(w.config.URL)

	if err != nil {
		return err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode())
	}

	return nil
}
