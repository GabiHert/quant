// Package adapter contains integration adapter interfaces that combine multiple usecase interfaces.
package adapter

import (
	"quant/internal/application/usecase"
)

// JobGroupPersistence combines all job-group-related persistence usecase interfaces.
type JobGroupPersistence interface {
	usecase.FindJobGroup
	usecase.SaveJobGroup
	usecase.UpdateJobGroup
	usecase.DeleteJobGroup
}
