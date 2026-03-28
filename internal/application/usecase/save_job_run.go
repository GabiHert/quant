package usecase

import (
	"quant/internal/domain/entity"
)

// SaveJobRun defines the interface for persisting and updating job runs.
type SaveJobRun interface {
	SaveJobRun(run entity.JobRun) error
	UpdateJobRun(run entity.JobRun) error
}
