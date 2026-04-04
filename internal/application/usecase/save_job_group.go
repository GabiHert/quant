package usecase

import (
	"quant/internal/domain/entity"
)

// SaveJobGroup defines the interface for persisting a new job group.
type SaveJobGroup interface {
	SaveJobGroup(group entity.JobGroup) error
}
