package usecase

import (
	"quant/internal/domain/entity"
)

// SaveWorkspace defines the interface for persisting a new workspace.
type SaveWorkspace interface {
	SaveWorkspace(workspace entity.Workspace) error
}
