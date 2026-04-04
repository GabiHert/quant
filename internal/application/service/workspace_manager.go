// Package service contains application service implementations.
package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"quant/internal/application/adapter"
	"quant/internal/application/usecase"
	"quant/internal/domain/entity"
)

// workspaceManagerService implements the adapter.WorkspaceManager interface.
type workspaceManagerService struct {
	findWorkspace   usecase.FindWorkspace
	saveWorkspace   usecase.SaveWorkspace
	updateWorkspace usecase.UpdateWorkspace
	deleteWorkspace usecase.DeleteWorkspace
}

// NewWorkspaceManagerService creates a new workspace manager service.
func NewWorkspaceManagerService(
	findWorkspace usecase.FindWorkspace,
	saveWorkspace usecase.SaveWorkspace,
	updateWorkspace usecase.UpdateWorkspace,
	deleteWorkspace usecase.DeleteWorkspace,
) adapter.WorkspaceManager {
	return &workspaceManagerService{
		findWorkspace:   findWorkspace,
		saveWorkspace:   saveWorkspace,
		updateWorkspace: updateWorkspace,
		deleteWorkspace: deleteWorkspace,
	}
}

// CreateWorkspace creates a new workspace with a generated ID and timestamps.
func (s *workspaceManagerService) CreateWorkspace(workspace entity.Workspace) (*entity.Workspace, error) {
	now := time.Now()
	workspace.ID = uuid.New().String()
	workspace.CreatedAt = now
	workspace.UpdatedAt = now

	if err := s.saveWorkspace.SaveWorkspace(workspace); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return &workspace, nil
}

// UpdateWorkspace updates an existing workspace.
func (s *workspaceManagerService) UpdateWorkspace(workspace entity.Workspace) (*entity.Workspace, error) {
	existing, err := s.findWorkspace.FindWorkspaceByID(workspace.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("workspace not found: %s", workspace.ID)
	}

	workspace.CreatedAt = existing.CreatedAt
	workspace.UpdatedAt = time.Now()

	if err := s.updateWorkspace.UpdateWorkspace(workspace); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return &workspace, nil
}

// DeleteWorkspace deletes a workspace by ID.
func (s *workspaceManagerService) DeleteWorkspace(id string) error {
	return s.deleteWorkspace.DeleteWorkspace(id)
}

// GetWorkspace retrieves a workspace by ID.
func (s *workspaceManagerService) GetWorkspace(id string) (*entity.Workspace, error) {
	return s.findWorkspace.FindWorkspaceByID(id)
}

// ListWorkspaces retrieves all workspaces.
func (s *workspaceManagerService) ListWorkspaces() ([]entity.Workspace, error) {
	return s.findWorkspace.FindAllWorkspaces()
}
