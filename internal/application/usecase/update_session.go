package usecase

import (
	"quant/internal/domain/entity"
)

// UpdateSession defines the interface for session update operations.
type UpdateSession interface {
	UpdateStatus(id string, status string) error
	Update(session entity.Session) error
}
