package usecase

import (
	"quant/internal/domain/entity"
)

// SaveJob defines the interface for persisting a new job.
type SaveJob interface {
	SaveJob(job entity.Job) error
}
