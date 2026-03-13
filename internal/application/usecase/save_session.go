package usecase

import (
	"quant/internal/domain/entity"
)

// SaveSession defines the interface for session persistence operations.
type SaveSession interface {
	Save(session entity.Session) error
}
