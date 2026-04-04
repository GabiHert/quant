package adapter

import (
	"context"

	"quant/internal/integration/entrypoint/dto"
)

// JobGroupController defines the interface for the job group entrypoint controller.
type JobGroupController interface {
	OnStartup(ctx context.Context)
	OnShutdown(ctx context.Context)
	CreateJobGroup(request dto.CreateJobGroupRequest) (*dto.JobGroupResponse, error)
	UpdateJobGroup(request dto.UpdateJobGroupRequest) (*dto.JobGroupResponse, error)
	DeleteJobGroup(id string) error
	ListJobGroupsByWorkspace(workspaceID string) ([]dto.JobGroupResponse, error)
}
