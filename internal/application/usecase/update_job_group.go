package usecase

import (
	"quant/internal/domain/entity"
)

// UpdateJobGroup defines the interface for updating an existing job group.
type UpdateJobGroup interface {
	UpdateJobGroup(group entity.JobGroup) error
}
