// Package adapter contains integration adapter interfaces that combine multiple usecase interfaces.
package adapter

import (
	"quant/internal/application/usecase"
)

// WorkspacePersistence combines all workspace-related persistence usecase interfaces.
// Integration persistence implementations must implement this interface.
type WorkspacePersistence interface {
	usecase.FindWorkspace
	usecase.SaveWorkspace
	usecase.UpdateWorkspace
	usecase.DeleteWorkspace
}
