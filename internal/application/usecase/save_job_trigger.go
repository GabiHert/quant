package usecase

import (
	"quant/internal/domain/entity"
)

// SaveJobTrigger defines the interface for persisting and removing job triggers.
type SaveJobTrigger interface {
	SaveJobTrigger(trigger entity.JobTrigger) error
	DeleteJobTrigger(id string) error
	DeleteTriggersBySourceJobID(jobID string) error
}
