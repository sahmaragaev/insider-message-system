package usecases

import (
	"context"
	"testing"

	"insider-message-system/internal/application/services"
	pkgerrors "insider-message-system/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockScheduler struct {
	mock.Mock
	running bool
}

func (m *mockScheduler) Start(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.running = true
	}
	return args.Error(0)
}

func (m *mockScheduler) Stop() error {
	args := m.Called()
	if args.Error(0) == nil {
		m.running = false
	}
	return args.Error(0)
}

func (m *mockScheduler) IsRunning() bool {
	return m.running
}

func (m *mockScheduler) SetMessageProcessor(_ services.MessageProcessor) {}

func TestControlSchedulerUseCase_Start(t *testing.T) {
	ctx := context.Background()
	mockSch := new(mockScheduler)
	uc := NewControlSchedulerUseCase(mockSch)

	mockSch.running = false
	mockSch.On("IsRunning").Return(false)
	mockSch.On("Start", ctx).Return(nil)
	resp, err := uc.Start(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "running", resp.Status)
	mockSch.AssertCalled(t, "Start", ctx)

	mockSch.running = true
	mockSch.On("IsRunning").Return(true)
	resp, err = uc.Start(ctx)
	assert.ErrorIs(t, err, pkgerrors.ErrSchedulerAlreadyRunning)
	assert.Nil(t, resp)
}

func TestControlSchedulerUseCase_Stop(t *testing.T) {
	ctx := context.Background()
	mockSch := new(mockScheduler)
	uc := NewControlSchedulerUseCase(mockSch)

	mockSch.running = true
	mockSch.On("IsRunning").Return(true)
	mockSch.On("Stop").Return(nil)
	resp, err := uc.Stop(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "stopped", resp.Status)
	mockSch.AssertCalled(t, "Stop")

	mockSch.running = false
	mockSch.On("IsRunning").Return(false)
	resp, err = uc.Stop(ctx)
	assert.ErrorIs(t, err, pkgerrors.ErrSchedulerNotRunning)
	assert.Nil(t, resp)
}

func TestControlSchedulerUseCase_Start_ErrorFromService(t *testing.T) {
	ctx := context.Background()
	mockSch := new(mockScheduler)
	uc := NewControlSchedulerUseCase(mockSch)

	mockSch.running = false
	mockSch.On("IsRunning").Return(false)
	mockSch.On("Start", ctx).Return(assert.AnError)
	resp, err := uc.Start(ctx)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestControlSchedulerUseCase_Stop_ErrorFromService(t *testing.T) {
	ctx := context.Background()
	mockSch := new(mockScheduler)
	uc := NewControlSchedulerUseCase(mockSch)

	mockSch.running = true
	mockSch.On("IsRunning").Return(true)
	mockSch.On("Stop").Return(assert.AnError)
	resp, err := uc.Stop(ctx)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestControlSchedulerUseCase_GetStatus(t *testing.T) {
	mockSch := new(mockScheduler)
	uc := NewControlSchedulerUseCase(mockSch)

	mockSch.running = true
	resp := uc.GetStatus(context.Background())
	assert.Equal(t, "running", resp.Status)

	mockSch.running = false
	resp = uc.GetStatus(context.Background())
	assert.Equal(t, "stopped", resp.Status)
}
