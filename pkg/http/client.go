package http

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type ClientConfig struct {
	Timeout          time.Duration
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
	DefaultHeaders   map[string]string
	RetryCondition   func(*resty.Response, error) bool
}

func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:          30 * time.Second,
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
}

func NewClient(config *ClientConfig) *resty.Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := resty.New().
		SetTimeout(config.Timeout).
		SetRetryCount(config.RetryCount).
		SetRetryWaitTime(config.RetryWaitTime).
		SetRetryMaxWaitTime(config.RetryMaxWaitTime)

	for key, value := range config.DefaultHeaders {
		client.SetHeader(key, value)
	}

	if config.RetryCondition != nil {
		client.AddRetryCondition(config.RetryCondition)
	}

	return client
}

func NewClientWithAuth(authKey string, config *ClientConfig) *resty.Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := NewClient(config)
	client.SetHeader("Authorization", "Bearer "+authKey)

	return client
}

func NewClientWithCustomAuth(headerName, headerValue string, config *ClientConfig) *resty.Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := NewClient(config)
	client.SetHeader(headerName, headerValue)

	return client
}
