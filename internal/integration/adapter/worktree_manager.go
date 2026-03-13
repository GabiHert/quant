package adapter

import (
	"quant/internal/application/usecase"
)

// WorktreeManager combines worktree-related usecase interfaces.
// Integration worktree implementations must implement this interface.
type WorktreeManager interface {
	usecase.ManageWorktree
}
