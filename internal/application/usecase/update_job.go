package usecase

import (
	"quant/internal/domain/entity"
)

// UpdateJob defines the interface for updating an existing job.
type UpdateJob interface {
	UpdateJob(job entity.Job) error
}
