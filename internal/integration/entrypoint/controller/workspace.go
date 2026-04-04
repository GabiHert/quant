// Package controller contains Wails-bound entrypoint controllers.
package controller

import (
	"context"

	appAdapter "quant/internal/application/adapter"
	"quant/internal/domain/entity"
	intAdapter "quant/internal/integration/adapter"
	"quant/internal/integration/entrypoint/dto"
)

// workspaceController implements the intAdapter.WorkspaceController interface.
type workspaceController struct {
	ctx              context.Context
	workspaceManager appAdapter.WorkspaceManager
}

// NewWorkspaceController creates a new Wails-bound workspace controller.
func NewWorkspaceController(workspaceManager appAdapter.WorkspaceManager) intAdapter.WorkspaceController {
	return &workspaceController{
		workspaceManager: workspaceManager,
	}
}

func (c *workspaceController) OnStartup(ctx context.Context) {
	c.ctx = ctx
}

func (c *workspaceController) OnShutdown(_ context.Context) {}

// CreateWorkspace handles workspace creation requests.
func (c *workspaceController) CreateWorkspace(request dto.CreateWorkspaceRequest) (*dto.WorkspaceResponse, error) {
	workspace := entity.Workspace{
		Name: request.Name,
	}

	created, err := c.workspaceManager.CreateWorkspace(workspace)
	if err != nil {
		return nil, err
	}

	return dto.WorkspaceResponseFromEntityPtr(created), nil
}

// UpdateWorkspace handles workspace update requests.
func (c *workspaceController) UpdateWorkspace(request dto.UpdateWorkspaceRequest) (*dto.WorkspaceResponse, error) {
	workspace := entity.Workspace{
		ID:   request.ID,
		Name: request.Name,
	}

	updated, err := c.workspaceManager.UpdateWorkspace(workspace)
	if err != nil {
		return nil, err
	}

	return dto.WorkspaceResponseFromEntityPtr(updated), nil
}

// DeleteWorkspace handles workspace deletion.
func (c *workspaceController) DeleteWorkspace(id string) error {
	return c.workspaceManager.DeleteWorkspace(id)
}

// GetWorkspace retrieves a workspace by ID.
func (c *workspaceController) GetWorkspace(id string) (*dto.WorkspaceResponse, error) {
	workspace, err := c.workspaceManager.GetWorkspace(id)
	if err != nil {
		return nil, err
	}

	return dto.WorkspaceResponseFromEntityPtr(workspace), nil
}

// ListWorkspaces retrieves all workspaces.
func (c *workspaceController) ListWorkspaces() ([]dto.WorkspaceResponse, error) {
	workspaces, err := c.workspaceManager.ListWorkspaces()
	if err != nil {
		return nil, err
	}

	return dto.WorkspaceResponseListFromEntities(workspaces), nil
}
