package usecase

import (
	"quant/internal/domain/entity"
)

// FindJob defines the interface for job retrieval operations.
type FindJob interface {
	FindJobByID(id string) (*entity.Job, error)
	FindAllJobs() ([]entity.Job, error)
	FindScheduledJobs() ([]entity.Job, error)
}
