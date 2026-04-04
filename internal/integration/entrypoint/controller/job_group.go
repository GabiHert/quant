// Package controller contains Wails-bound entrypoint controllers.
package controller

import (
	"context"

	appAdapter "quant/internal/application/adapter"
	"quant/internal/domain/entity"
	intAdapter "quant/internal/integration/adapter"
	"quant/internal/integration/entrypoint/dto"
)

// jobGroupController implements the intAdapter.JobGroupController interface.
type jobGroupController struct {
	ctx             context.Context
	jobGroupManager appAdapter.JobGroupManager
}

// NewJobGroupController creates a new Wails-bound job group controller.
func NewJobGroupController(jobGroupManager appAdapter.JobGroupManager) intAdapter.JobGroupController {
	return &jobGroupController{
		jobGroupManager: jobGroupManager,
	}
}

func (c *jobGroupController) OnStartup(ctx context.Context) {
	c.ctx = ctx
}

func (c *jobGroupController) OnShutdown(_ context.Context) {}

// CreateJobGroup handles job group creation requests.
func (c *jobGroupController) CreateJobGroup(request dto.CreateJobGroupRequest) (*dto.JobGroupResponse, error) {
	group := entity.JobGroup{
		Name:        request.Name,
		WorkspaceID: request.WorkspaceID,
		JobIDs:      request.JobIDs,
	}

	created, err := c.jobGroupManager.CreateJobGroup(group)
	if err != nil {
		return nil, err
	}

	return dto.JobGroupResponseFromEntityPtr(created), nil
}

// UpdateJobGroup handles job group update requests.
func (c *jobGroupController) UpdateJobGroup(request dto.UpdateJobGroupRequest) (*dto.JobGroupResponse, error) {
	group := entity.JobGroup{
		ID:          request.ID,
		Name:        request.Name,
		WorkspaceID: request.WorkspaceID,
		JobIDs:      request.JobIDs,
	}

	updated, err := c.jobGroupManager.UpdateJobGroup(group)
	if err != nil {
		return nil, err
	}

	return dto.JobGroupResponseFromEntityPtr(updated), nil
}

// DeleteJobGroup handles job group deletion.
func (c *jobGroupController) DeleteJobGroup(id string) error {
	return c.jobGroupManager.DeleteJobGroup(id)
}

// ListJobGroupsByWorkspace retrieves all job groups for a workspace.
func (c *jobGroupController) ListJobGroupsByWorkspace(workspaceID string) ([]dto.JobGroupResponse, error) {
	groups, err := c.jobGroupManager.ListJobGroupsByWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	return dto.JobGroupResponseListFromEntities(groups), nil
}
