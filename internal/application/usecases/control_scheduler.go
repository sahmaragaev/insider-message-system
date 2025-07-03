package usecases

import (
	"context"
	"insider-message-system/internal/application/services"
	"insider-message-system/pkg/errors"
	"insider-message-system/pkg/logger"

	"go.uber.org/zap"
)

// SchedulerStatusResponse represents the status of the scheduler.
type SchedulerStatusResponse struct {
	Status  string `json:"status" example:"running"`
	Message string `json:"message" example:"Scheduler is running"`
}

// ControlSchedulerUseCase handles starting, stopping, and querying the scheduler.
type ControlSchedulerUseCase struct {
	schedulerService services.Scheduler
}

// NewControlSchedulerUseCase creates a new ControlSchedulerUseCase with the given scheduler service.
func NewControlSchedulerUseCase(schedulerService services.Scheduler) *ControlSchedulerUseCase {
	return &ControlSchedulerUseCase{
		schedulerService: schedulerService,
	}
}

// Start attempts to start the scheduler and returns its status.
func (uc *ControlSchedulerUseCase) Start(ctx context.Context) (*SchedulerStatusResponse, error) {
	if uc.schedulerService.IsRunning() {
		logger.Warn("Attempted to start scheduler that is already running")
		return nil, errors.ErrSchedulerAlreadyRunning
	}

	if err := uc.schedulerService.Start(ctx); err != nil {
		logger.Error("Failed to start scheduler in use case", zap.Error(err))
		return nil, err
	}

	logger.Info("Scheduler started via use case")
	return &SchedulerStatusResponse{
		Status:  "running",
		Message: "Scheduler started successfully",
	}, nil
}

// Stop attempts to stop the scheduler and returns its status.
func (uc *ControlSchedulerUseCase) Stop(ctx context.Context) (*SchedulerStatusResponse, error) {
	if !uc.schedulerService.IsRunning() {
		logger.Warn("Attempted to stop scheduler that is not running")
		return nil, errors.ErrSchedulerNotRunning
	}

	if err := uc.schedulerService.Stop(); err != nil {
		logger.Error("Failed to stop scheduler in use case", zap.Error(err))
		return nil, err
	}

	logger.Info("Scheduler stopped via use case")
	return &SchedulerStatusResponse{
		Status:  "stopped",
		Message: "Scheduler stopped successfully",
	}, nil
}

// GetStatus returns the current status of the scheduler.
func (uc *ControlSchedulerUseCase) GetStatus(ctx context.Context) *SchedulerStatusResponse {
	isRunning := uc.schedulerService.IsRunning()

	if isRunning {
		return &SchedulerStatusResponse{
			Status:  "running",
			Message: "Scheduler is currently running",
		}
	}

	return &SchedulerStatusResponse{
		Status:  "stopped",
		Message: "Scheduler is currently stopped",
	}
}
