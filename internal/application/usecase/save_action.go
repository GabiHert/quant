package usecase

import (
	"quant/internal/domain/entity"
)

// SaveAction defines the interface for action persistence operations.
type SaveAction interface {
	SaveAction(action entity.Action) error
}
