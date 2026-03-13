package usecase

import (
	"quant/internal/domain/entity"
)

// SaveTask defines the interface for task persistence operations.
type SaveTask interface {
	SaveTask(task entity.Task) error
}
