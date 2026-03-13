package usecase

import (
	"quant/internal/domain/entity"
)

// FindTask defines the interface for task retrieval operations.
type FindTask interface {
	FindTaskByID(id string) (*entity.Task, error)
	FindTasksByRepoID(repoID string) ([]entity.Task, error)
	FindAllTasks() ([]entity.Task, error)
}
