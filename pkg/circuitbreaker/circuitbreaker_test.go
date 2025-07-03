package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_New(t *testing.T) {
	config := Config{
		FailureThreshold:   3,
		SuccessThreshold:   2,
		Timeout:            10 * time.Second,
		MaxConcurrentCalls: 1,
	}

	cb := New(config)

	if cb.GetState() != StateClosed {
		t.Errorf("Expected initial state to be CLOSED, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_Execute_ClosedToOpen(t *testing.T) {
	config := Config{
		FailureThreshold:   3,
		SuccessThreshold:   2,
		Timeout:            100 * time.Millisecond,
		MaxConcurrentCalls: 1,
	}

	cb := New(config)
	testErr := errors.New("test error")

	for i := 0; i < 3; i++ {
		err := cb.Execute(context.Background(), func() error {
			return testErr
		})

		if err != testErr {
			t.Errorf("Expected test error, got %v", err)
		}
	}

	if cb.GetState() != StateOpen {
		t.Errorf("Expected circuit to be OPEN, got %s", cb.GetState())
	}

	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_Execute_OpenToHalfOpen(t *testing.T) {
	config := Config{
		FailureThreshold:   1,
		SuccessThreshold:   1,
		Timeout:            10 * time.Millisecond,
		MaxConcurrentCalls: 1,
	}

	cb := New(config)

	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	time.Sleep(20 * time.Millisecond)

	if cb.GetState() != StateHalfOpen {
		t.Errorf("Expected circuit to be HALF_OPEN, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_Execute_HalfOpenToClosed(t *testing.T) {
	config := Config{
		FailureThreshold:   1,
		SuccessThreshold:   1,
		Timeout:            10 * time.Millisecond,
		MaxConcurrentCalls: 1,
	}

	cb := New(config)

	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	time.Sleep(20 * time.Millisecond)

	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected circuit to be CLOSED, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_GetMetrics(t *testing.T) {
	config := Config{
		FailureThreshold:   2,
		SuccessThreshold:   1,
		Timeout:            10 * time.Millisecond,
		MaxConcurrentCalls: 1,
	}

	cb := New(config)

	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	cb.Execute(context.Background(), func() error {
		return nil
	})

	metrics := cb.GetMetrics()

	if metrics["total_calls"] != int64(2) {
		t.Errorf("Expected total_calls to be 2, got %v", metrics["total_calls"])
	}

	if metrics["total_failures"] != int64(1) {
		t.Errorf("Expected total_failures to be 1, got %v", metrics["total_failures"])
	}

	if metrics["total_successes"] != int64(1) {
		t.Errorf("Expected total_successes to be 1, got %v", metrics["total_successes"])
	}
}

func TestCircuitBreaker_ForceOpen(t *testing.T) {
	cb := New(Config{})

	cb.ForceOpen()

	if cb.GetState() != StateOpen {
		t.Errorf("Expected circuit to be OPEN after ForceOpen, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_ForceClose(t *testing.T) {
	cb := New(Config{})

	cb.ForceOpen()

	cb.ForceClose()

	if cb.GetState() != StateClosed {
		t.Errorf("Expected circuit to be CLOSED after ForceClose, got %s", cb.GetState())
	}
}
