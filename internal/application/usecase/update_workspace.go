package usecase

import (
	"quant/internal/domain/entity"
)

// UpdateWorkspace defines the interface for updating an existing workspace.
type UpdateWorkspace interface {
	UpdateWorkspace(workspace entity.Workspace) error
}
