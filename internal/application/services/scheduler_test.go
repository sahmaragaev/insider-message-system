package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"insider-message-system/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProcessor struct{ mock.Mock }

func (m *mockProcessor) ProcessMessages(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestScheduler_StartStopStatus(t *testing.T) {
	processor := new(mockProcessor)
	cfg := config.SchedulerConfig{
		Interval:   10 * time.Millisecond,
		MaxRetries: 1,
		RetryDelay: 1 * time.Millisecond,
	}
	s := NewScheduler(cfg)
	s.SetMessageProcessor(processor)

	// Initially stopped
	assert.False(t, s.IsRunning())

	// Start
	err := s.Start(context.Background())
	assert.NoError(t, err)
	assert.True(t, s.IsRunning())

	// Double start should not error (returns nil)
	err = s.Start(context.Background())
	assert.NoError(t, err)

	// Stop
	err = s.Stop()
	assert.NoError(t, err)
	assert.False(t, s.IsRunning())

	// Double stop should not error (returns nil)
	err = s.Stop()
	assert.NoError(t, err)
}

func TestScheduler_ProcessError(t *testing.T) {
	processor := new(mockProcessor)
	cfg := config.SchedulerConfig{
		Interval:   1 * time.Millisecond,
		MaxRetries: 1,
		RetryDelay: 1 * time.Millisecond,
	}
	s := NewScheduler(cfg)
	s.SetMessageProcessor(processor)

	processor.On("ProcessMessages", mock.Anything).Return(errors.New("fail")).Maybe()

	err := s.Start(context.Background())
	assert.NoError(t, err)
	time.Sleep(5 * time.Millisecond)
	err = s.Stop()
	assert.NoError(t, err)
	processor.AssertCalled(t, "ProcessMessages", mock.Anything)
}

func TestScheduler_Start_NoProcessorSet(t *testing.T) {
	cfg := config.SchedulerConfig{
		Interval:   10 * time.Millisecond,
		MaxRetries: 1,
		RetryDelay: 1 * time.Millisecond,
	}
	s := NewScheduler(cfg)
	// Do not set processor
	err := s.Start(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "processor not set")
}

func TestScheduler_Start_AlreadyRunning(t *testing.T) {
	processor := new(mockProcessor)
	cfg := config.SchedulerConfig{
		Interval:   10 * time.Millisecond,
		MaxRetries: 1,
		RetryDelay: 1 * time.Millisecond,
	}
	s := NewScheduler(cfg)
	s.SetMessageProcessor(processor)
	assert.NoError(t, s.Start(context.Background()))
	// Should not error, just log warning
	assert.NoError(t, s.Start(context.Background()))
	s.Stop()
}

func TestScheduler_Stop_NotRunning(t *testing.T) {
	cfg := config.SchedulerConfig{
		Interval:   10 * time.Millisecond,
		MaxRetries: 1,
		RetryDelay: 1 * time.Millisecond,
	}
	s := NewScheduler(cfg)
	// Not started
	assert.NoError(t, s.Stop())
}
