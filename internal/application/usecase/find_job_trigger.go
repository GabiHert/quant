package usecase

import (
	"quant/internal/domain/entity"
)

// FindJobTrigger defines the interface for job trigger retrieval operations.
type FindJobTrigger interface {
	FindTriggersBySourceJobID(jobID string) ([]entity.JobTrigger, error)
	FindTriggersByTargetJobID(jobID string) ([]entity.JobTrigger, error)
}
