package usecase

import (
	"quant/internal/domain/entity"
)

// FindWorkspace defines the interface for workspace retrieval operations.
type FindWorkspace interface {
	FindWorkspaceByID(id string) (*entity.Workspace, error)
	FindAllWorkspaces() ([]entity.Workspace, error)
}
