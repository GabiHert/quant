package usecase

import (
	"quant/internal/domain/entity"
)

// FindAction defines the interface for action retrieval operations.
type FindAction interface {
	FindActionsBySessionID(sessionID string) ([]entity.Action, error)
}
