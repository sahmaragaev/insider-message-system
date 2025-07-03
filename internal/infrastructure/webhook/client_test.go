package webhook

import (
	"context"
	"insider-message-system/internal/domain"
	"insider-message-system/pkg/circuitbreaker"
	"insider-message-system/pkg/config"
	"net/http"
	"testing"

	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRestyClient struct {
	mock.Mock
}

func (m *mockRestyClient) R() restyRequest {
	args := m.Called()
	return args.Get(0).(restyRequest)
}

type mockRestyRequest struct {
	mock.Mock
}

func (m *mockRestyRequest) SetContext(ctx context.Context) restyRequest {
	args := m.Called(ctx)
	return args.Get(0).(restyRequest)
}

func (m *mockRestyRequest) SetBody(body any) restyRequest {
	args := m.Called(body)
	return args.Get(0).(restyRequest)
}

func (m *mockRestyRequest) SetResult(result any) restyRequest {
	args := m.Called(result)
	return args.Get(0).(restyRequest)
}

func (m *mockRestyRequest) Post(url string) (*resty.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*resty.Response), args.Error(1)
}

func (m *mockRestyRequest) Get(url string) (*resty.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*resty.Response), args.Error(1)
}

type mockRestyResponse struct {
	mock.Mock
}

func (m *mockRestyResponse) StatusCode() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockRestyResponse) String() string {
	args := m.Called()
	return args.String(0)
}

func TestSendMessage_Success(t *testing.T) {
	mockClient := new(mockRestyClient)
	mockReq := new(mockRestyRequest)

	mockClient.On("R").Return(mockReq)
	mockReq.On("SetContext", mock.Anything).Return(mockReq)
	mockReq.On("SetBody", mock.Anything).Return(mockReq)
	mockReq.On("SetResult", mock.Anything).Return(mockReq)

	resp := &resty.Response{}
	resp.RawResponse = &http.Response{StatusCode: 200, Status: "200 OK"}
	resp.SetBody([]byte(""))
	mockReq.On("Post", "http://webhook").Return(resp, nil)

	c := &client{
		client: mockClient,
		config: config.WebhookConfig{URL: "http://webhook"},
	}

	response, err := c.sendMessageDirect(context.Background(), domain.WebhookRequest{To: "+123", Content: "hi"})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestSendMessage_HTTPError(t *testing.T) {
	mockClient := new(mockRestyClient)
	mockReq := new(mockRestyRequest)

	mockClient.On("R").Return(mockReq)
	mockReq.On("SetContext", mock.Anything).Return(mockReq)
	mockReq.On("SetBody", mock.Anything).Return(mockReq)
	mockReq.On("SetResult", mock.Anything).Return(mockReq)
	resp := &resty.Response{}
	resp.RawResponse = &http.Response{StatusCode: 200, Status: "200 OK"}
	mockReq.On("Post", "http://webhook").Return(resp, fmt.Errorf("http fail"))

	c := &client{
		client: mockClient,
		config: config.WebhookConfig{URL: "http://webhook"},
	}

	response, err := c.sendMessageDirect(context.Background(), domain.WebhookRequest{To: "+123", Content: "hi"})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestSendMessage_Non2xx(t *testing.T) {
	mockClient := new(mockRestyClient)
	mockReq := new(mockRestyRequest)

	mockClient.On("R").Return(mockReq)
	mockReq.On("SetContext", mock.Anything).Return(mockReq)
	mockReq.On("SetBody", mock.Anything).Return(mockReq)
	mockReq.On("SetResult", mock.Anything).Return(mockReq)
	resp := &resty.Response{}
	resp.RawResponse = &http.Response{StatusCode: 500, Status: "500 Internal Server Error"}
	resp.SetBody([]byte("fail"))
	mockReq.On("Post", "http://webhook").Return(resp, nil)

	c := &client{
		client: mockClient,
		config: config.WebhookConfig{URL: "http://webhook"},
	}

	response, err := c.sendMessageDirect(context.Background(), domain.WebhookRequest{To: "+123", Content: "hi"})
	assert.Error(t, err)
	assert.Nil(t, response)
}

// Circuit breaker test double

type cbTestDouble struct {
	execErr error
	state   circuitbreaker.State
}

func (cb *cbTestDouble) Execute(ctx context.Context, fn func() error) error {
	return cb.execErr
}
func (cb *cbTestDouble) GetMetrics() map[string]any     { return map[string]any{"enabled": true} }
func (cb *cbTestDouble) GetState() circuitbreaker.State { return cb.state }

func TestSendMessage_CBOpen(t *testing.T) {
	c := &client{
		client:         new(mockRestyClient),
		config:         config.WebhookConfig{URL: "http://webhook"},
		circuitBreaker: &cbTestDouble{execErr: circuitbreaker.ErrCircuitOpen, state: circuitbreaker.StateOpen},
	}
	resp, err := c.SendMessage(context.Background(), domain.WebhookRequest{To: "+123", Content: "hi"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "WEBHOOK_CIRCUIT_OPEN")
}

func TestSendMessage_CBHalfOpen(t *testing.T) {
	c := &client{
		client:         new(mockRestyClient),
		config:         config.WebhookConfig{URL: "http://webhook"},
		circuitBreaker: &cbTestDouble{execErr: circuitbreaker.ErrCircuitHalfOpen, state: circuitbreaker.StateHalfOpen},
	}
	resp, err := c.SendMessage(context.Background(), domain.WebhookRequest{To: "+123", Content: "hi"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "WEBHOOK_CIRCUIT_HALF_OPEN")
}

func TestSendMessage_CBGenericError(t *testing.T) {
	c := &client{
		client:         new(mockRestyClient),
		config:         config.WebhookConfig{URL: "http://webhook"},
		circuitBreaker: &cbTestDouble{execErr: fmt.Errorf("cb fail"), state: circuitbreaker.StateOpen},
	}
	resp, err := c.SendMessage(context.Background(), domain.WebhookRequest{To: "+123", Content: "hi"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "cb fail")
}

func TestGetCircuitBreakerMetrics_And_State(t *testing.T) {
	c := &client{circuitBreaker: nil}
	metrics := c.GetCircuitBreakerMetrics()
	assert.False(t, metrics["enabled"].(bool))
	assert.Equal(t, circuitbreaker.StateClosed, c.GetCircuitBreakerState())

	c = &client{circuitBreaker: &cbTestDouble{state: circuitbreaker.StateOpen}}
	metrics = c.GetCircuitBreakerMetrics()
	assert.True(t, metrics["enabled"].(bool))
	assert.Equal(t, circuitbreaker.StateOpen, c.GetCircuitBreakerState())
}

func TestHealthCheck_Success(t *testing.T) {
	mockClient := new(mockRestyClient)
	mockReq := new(mockRestyRequest)
	mockClient.On("R").Return(mockReq)
	mockReq.On("SetContext", mock.Anything).Return(mockReq)
	resp := &resty.Response{}
	resp.RawResponse = &http.Response{StatusCode: 200, Status: "200 OK"}
	mockReq.On("Get", "http://webhook").Return(resp, nil)

	c := &client{
		client: mockClient,
		config: config.WebhookConfig{URL: "http://webhook"},
	}
	assert.NoError(t, c.HealthCheck(context.Background()))
}

func TestHealthCheck_Error(t *testing.T) {
	mockClient := new(mockRestyClient)
	mockReq := new(mockRestyRequest)
	mockClient.On("R").Return(mockReq)
	mockReq.On("SetContext", mock.Anything).Return(mockReq)
	resp := &resty.Response{}
	resp.RawResponse = &http.Response{StatusCode: 200, Status: "200 OK"}
	mockReq.On("Get", "http://webhook").Return(resp, fmt.Errorf("fail"))

	c := &client{
		client: mockClient,
		config: config.WebhookConfig{URL: "http://webhook"},
	}
	assert.Error(t, c.HealthCheck(context.Background()))
}

func TestHealthCheck_Non2xx(t *testing.T) {
	mockClient := new(mockRestyClient)
	mockReq := new(mockRestyRequest)
	mockClient.On("R").Return(mockReq)
	mockReq.On("SetContext", mock.Anything).Return(mockReq)
	resp := &resty.Response{}
	resp.RawResponse = &http.Response{StatusCode: 500, Status: "500 Internal Server Error"}
	mockReq.On("Get", "http://webhook").Return(resp, nil)

	c := &client{
		client: mockClient,
		config: config.WebhookConfig{URL: "http://webhook"},
	}
	assert.Error(t, c.HealthCheck(context.Background()))
}
