package usecase

import (
	"quant/internal/domain/entity"
)

// UpdateTask defines the interface for task update operations.
type UpdateTask interface {
	UpdateTask(task entity.Task) error
}
