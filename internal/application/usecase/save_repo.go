package usecase

import (
	"quant/internal/domain/entity"
)

// SaveRepo defines the interface for repo persistence operations.
type SaveRepo interface {
	SaveRepo(repo entity.Repo) error
}
